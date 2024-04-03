package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/model"
	"github.com/artufi/trader/xtb/processor"
	"github.com/artufi/trader/xtb/response"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type Purchase struct {
	Cfg       config.AppConfig
	WSManager *websocket.WSManager
	Logger    *slog.Logger

	PredictionService model.PredictionService
	OrderService      model.OrderService
}

func (p *Purchase) PurchasesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New().String()
		userID := p.Cfg.XTB.Demo.UserID
		ctx := logging.AppendAttrsCtx(r.Context(), logging.TraceIDAttr(traceID), logging.UserIDAttr(userID))

		p.Logger.InfoContext(ctx, "Start processing purchases request")

		wsClient, err := p.WSManager.GetUserRandomClient(userID)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to get client to process purchase request", logging.ErrorAttr(err))
			http.Error(w, "No clients assigned to specified user", http.StatusInternalServerError)
			return
		}

		// init api handler
		apiH := APIHandler{
			Proc:    processor.NewProc(wsClient),
			TraceID: traceID,
		}

		dbH := DBHandler{
			TraceID:           traceID,
			Logger:            p.Logger,
			PredictionService: p.PredictionService,
			OrderService:      p.OrderService,
		}

		// get prediction detail
		predictions := make([]model.PredictionDetails, 0)
		err = json.NewDecoder(r.Body).Decode(&predictions)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to read request body with prediction details list", logging.ErrorAttr(err))
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		for _, predictionDetails := range predictions {
			p.Logger.InfoContext(ctx, "Will process prediction details", logging.PredictionDetailsAttr(predictionDetails))

			symbol := predictionDetails.Symbol

			// get symbol details
			symbolResponse, err := apiH.GetSymbolExtended(ctx, symbol)
			if err != nil {
				p.Logger.ErrorContext(ctx, "Failed to process GetSymbol",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			p.Logger.InfoContext(ctx, "Successfully processed GetSymbol", logging.RespAttr(symbolResponse))

			// prepare TradeTransactionInfo to pass it to TradeTransaction as argument
			tradeTransInfo, err := predictionDetails.PrepareTradeTransInfo(symbolResponse, 0.01)
			if err != nil {
				p.Logger.WarnContext(ctx, "Failed to prepare TradeTransInfo", logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}

			// process TradeTransaction
			tradeResponse, err := apiH.TradeTransaction(ctx, symbol, tradeTransInfo)
			if err != nil {
				p.Logger.ErrorContext(ctx, "Failed to process TradeTransaction",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			p.Logger.InfoContext(ctx, "Successfully processed TradeTransaction", logging.RespAttr(tradeResponse))

			// process TradeTransactionStatus
			tradeResponseStatus, err := apiH.TradeTransactionStatus(ctx, symbol, tradeResponse.ReturnData.Order)
			if err != nil {
				p.Logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			p.Logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeResponseStatus))

			prediction := model.Prediction{
				UserID:            1,
				PredictionDetails: predictionDetails,
			}
			predID, err := dbH.InsertPrediction(ctx, prediction)
			if err != nil {
				p.Logger.ErrorContext(ctx, "Failed to insert prediction",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			p.Logger.InfoContext(ctx, "Inserted prediction", logging.IDAttr(predID))

			reqStatus, err := tradeResponseStatus.CheckRequestStatus()
			order := model.Order{
				Number:         tradeResponse.ReturnData.Order,
				UserID:         1,
				PredictionID:   predID,
				RequestStatus:  response.RequestStatusName[reqStatus],
				Message:        tradeResponseStatus.ReturnData.Message,
				TradeTransInfo: tradeTransInfo,
			}
			if reqStatus == response.REJECTED {
				rsErr := &response.RequestStatusError{}
				if errors.As(err, &rsErr) {
					order = model.Order{
						Number:         tradeResponse.ReturnData.Order,
						UserID:         1,
						PredictionID:   predID,
						RequestStatus:  rsErr.RequestStatus,
						Message:        rsErr.Message,
						TradeTransInfo: tradeTransInfo,
					}
				}
			}
			ordID, err := dbH.InsertOrder(ctx, order)
			if err != nil {
				p.Logger.ErrorContext(ctx, "Failed to insert order",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			p.Logger.InfoContext(ctx, "Inserted order", logging.IDAttr(ordID))
		}

		// logout user after processing
		logoutResponse, err := apiH.Logout(ctx)
		if err != nil {
			p.Logger.WarnContext(ctx, "Failed to process Logout - killing client", logging.ErrorAttr(err))
			return
		}
		p.Logger.InfoContext(ctx, "Successfully processed Logout", logging.RespAttr(logoutResponse))

		p.Logger.InfoContext(ctx, "Purchases request fully processed")
	}
}

func (p *Purchase) PurchaseHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New().String()
		userID := p.Cfg.XTB.Demo.UserID
		ctx := logging.AppendAttrsCtx(r.Context(), logging.TraceIDAttr(traceID), logging.UserIDAttr(userID))

		p.Logger.InfoContext(ctx, "Start processing purchase request")

		wsClient, err := p.WSManager.GetUserRandomClient(userID)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to get client to process purchase request", logging.ErrorAttr(err))
			http.Error(w, "No clients assigned to specified user", http.StatusInternalServerError)
			return
		}

		h := APIHandler{
			Proc:    processor.NewProc(wsClient),
			TraceID: traceID,
		}

		// get predictions
		var predictionDetails model.PredictionDetails
		err = json.NewDecoder(r.Body).Decode(&predictionDetails)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to read request body with prediction details", logging.ErrorAttr(err))
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		p.Logger.InfoContext(ctx, "Will process prediction details", logging.PredictionDetailsAttr(predictionDetails))

		symbol := predictionDetails.Symbol

		// get symbol details
		symbolResponse, err := h.GetSymbolExtended(ctx, symbol)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to process GetSymbol",
				logging.ErrorAttr(err),
				logging.SymbolAttr(symbol))
			http.Error(w, "Failed to execute GetSymbol", http.StatusInternalServerError)
			return
		}
		p.Logger.InfoContext(ctx, "Successfully processed GetSymbol", logging.RespAttr(symbolResponse))

		// prepare TradeTransactionInfo to pass it to TradeTransaction as argument
		tradeTransInfo, err := predictionDetails.PrepareTradeTransInfo(symbolResponse, 0.01)
		if err != nil {
			p.Logger.WarnContext(ctx, "Failed to prepare TradeTransInfo",
				logging.ErrorAttr(err),
				logging.SymbolAttr(symbol))
			http.Error(w, fmt.Sprintf("Failed to prepare TradeTransInfo: %v", err), http.StatusBadRequest)
			return
		}

		// process TradeTransaction
		tradeResponse, err := h.TradeTransaction(ctx, symbol, tradeTransInfo)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to process TradeTransaction",
				logging.ErrorAttr(err),
				logging.SymbolAttr(symbol))
			http.Error(w, "Failed to execute TradeTransaction", http.StatusInternalServerError)
			return
		}
		p.Logger.InfoContext(ctx, "Successfully processed TradeTransaction", logging.RespAttr(tradeResponse))

		// process TradeTransactionStatus
		tradeResponseStatus, err := h.TradeTransactionStatus(ctx, symbol, tradeResponse.ReturnData.Order)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus",
				logging.ErrorAttr(err),
				logging.SymbolAttr(symbol))
			http.Error(w, "Failed to execute TradeTransactionStatus", http.StatusInternalServerError)
			return
		}

		//reqStatus, err := tradeResponseStatus.CheckRequestStatus()
		//if reqStatus == response.ACCEPTED || reqStatus == response.PENDING || reqStatus == response.REJECTED {
		//	go sendPing(ctx, wsClient, p.Logger, 5)
		//}
		p.Logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeResponseStatus))

		// logout user after processing
		logoutResponse, err := h.Logout(ctx)
		if err != nil {
			p.Logger.WarnContext(ctx, "Failed to process Logout - killing client", logging.ErrorAttr(err))
			http.Error(w, "Failed to execute TradeTransactionStatus", http.StatusInternalServerError)
			return
		}
		p.Logger.InfoContext(ctx, "Successfully processed Logout", logging.RespAttr(logoutResponse))

		p.Logger.Info("Purchase request fully processed")
	}
}
