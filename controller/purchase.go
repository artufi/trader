package controller

import (
	"encoding/json"
	"errors"
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
		ctx = logging.AppendAttrsCtx(ctx, logging.ConnNo(wsClient.ConnID), logging.StreamID(wsClient.StreamSessionID))

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
			tradeTransInfo, err := predictionDetails.PrepareTradeTransInfo(symbolResponse, p.Cfg.BuySellParams.Volume)
			if err != nil {
				if errors.Is(err, model.ErrNoAction) {
					p.Logger.InfoContext(ctx, "ModelType: NoAction - stop processing",
						logging.SymbolAttr(symbol))
				} else {
					p.Logger.WarnContext(ctx, "Failed to prepare TradeTransInfo",
						logging.ErrorAttr(err),
						logging.SymbolAttr(symbol))
				}
				continue
			}

			// process TradeTransaction
			tradeTransResp, err := apiH.TradeTransaction(ctx, symbol, tradeTransInfo)
			if err != nil {
				p.Logger.ErrorContext(ctx, "Failed to process TradeTransaction",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			p.Logger.InfoContext(ctx, "Successfully processed TradeTransaction", logging.RespAttr(tradeTransResp))

			// process TradeTransactionStatus
			tradeTransStatusResp, err := apiH.TradeTransactionStatus(ctx, symbol, tradeTransResp.ReturnData.Order)
			if err != nil {
				p.Logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			p.Logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeTransStatusResp))

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

			var order model.Order
			reqStatus, errReq := tradeTransStatusResp.CheckRequestStatus()
			if reqStatus != response.REJECTED {
				// getTrades to acquire position required to forcibly close purchase on demand in history service
				getTradesResp, err := apiH.GetTrades(ctx, symbol, true)
				if err != nil {
					p.Logger.WarnContext(ctx, "Failed to process GetTrades, position number will not be added",
						logging.ErrorAttr(err),
						logging.SymbolAttr(symbol))
				} else {
					p.Logger.InfoContext(ctx, "Successfully processed GetTrades", logging.RespAttr(getTradesResp))
				}

				var position int
				for _, tradeRecord := range getTradesResp.ReturnData {
					// Each transaction in the XTB system is assigned a unique order number. When a TradeTransaction
					// is executed by the client, the system returns an order number for that transaction.
					// Subsequently, request made by the client is processed by the system, which marks the start
					// of a new transaction in their system. As a result, the previous order number is transformed
					// into 'order2', indicating that the previous system transaction has simply ended,
					// and a new transaction with a different order number is being processed. When the previous
					// order number matches the 'order2' number retrieved from XTB, it indicates that transactions
					// in their system are connected, and it is possible to obtain a position number that represents
					// the entire sequence of transactions.
					//
					// This process resembles a chain of transactions: a TradeTransaction from the client assigns
					// an order number to that transaction, then the system processes the TradeTransaction,
					// assigns it a new order number, and saves the previous order number of transaction in the
					// 'order2' field. Finally, it returns the position number that represents the complete
					// chain of transactions in the system.
					if tradeTransResp.ReturnData.Order == tradeRecord.Order2 {
						position = tradeRecord.Position
						p.Logger.InfoContext(ctx, "Acquired order position", logging.PositionAttr(position))
						break
					}
				}

				order = model.Order{
					Number:         tradeTransResp.ReturnData.Order,
					UserID:         1,
					Position:       &position,
					PredictionID:   predID,
					RequestStatus:  response.RequestStatusName[reqStatus],
					Message:        tradeTransStatusResp.ReturnData.Message,
					TradeTransInfo: tradeTransInfo,
				}
			} else {
				rsErr := &response.RequestStatusError{}
				if errors.As(errReq, &rsErr) {
					order = model.Order{
						Number:         tradeTransResp.ReturnData.Order,
						UserID:         1,
						PredictionID:   predID,
						RequestStatus:  rsErr.RequestStatus,
						Message:        &rsErr.Message,
						TradeTransInfo: tradeTransInfo,
						OrderClosedDetails: model.OrderClosedDetails{
							Closed: true,
						},
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
		ctx = logging.AppendAttrsCtx(ctx, logging.ConnNo(wsClient.ConnID), logging.StreamID(wsClient.StreamSessionID))

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
		symbolResponse, err := apiH.GetSymbolExtended(ctx, symbol)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to process GetSymbol",
				logging.ErrorAttr(err),
				logging.SymbolAttr(symbol))
			http.Error(w, "Failed to execute GetSymbol", http.StatusInternalServerError)
			return
		}
		p.Logger.InfoContext(ctx, "Successfully processed GetSymbol", logging.RespAttr(symbolResponse))

		// prepare TradeTransactionInfo to pass it to TradeTransaction as argument
		tradeTransInfo, err := predictionDetails.PrepareTradeTransInfo(symbolResponse, p.Cfg.BuySellParams.Volume)
		if err != nil {
			if errors.Is(err, model.ErrNoAction) {
				p.Logger.InfoContext(ctx, "ModelType: NoAction - stop processing",
					logging.SymbolAttr(symbol))
			} else {
				p.Logger.WarnContext(ctx, "Failed to prepare TradeTransInfo",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
			}
			return
		}

		// process TradeTransaction
		tradeTransResp, err := apiH.TradeTransaction(ctx, symbol, tradeTransInfo)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to process TradeTransaction",
				logging.ErrorAttr(err),
				logging.SymbolAttr(symbol))
			http.Error(w, "Failed to execute TradeTransaction", http.StatusInternalServerError)
			return
		}
		p.Logger.InfoContext(ctx, "Successfully processed TradeTransaction", logging.RespAttr(tradeTransResp))

		// process TradeTransactionStatus
		tradeTransStatusResp, err := apiH.TradeTransactionStatus(ctx, symbol, tradeTransResp.ReturnData.Order)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus",
				logging.ErrorAttr(err),
				logging.SymbolAttr(symbol))
			http.Error(w, "Failed to execute TradeTransactionStatus", http.StatusInternalServerError)
			return
		}
		p.Logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeTransStatusResp))

		prediction := model.Prediction{
			UserID:            1,
			PredictionDetails: predictionDetails,
		}
		predID, err := dbH.InsertPrediction(ctx, prediction)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to insert prediction",
				logging.ErrorAttr(err),
				logging.SymbolAttr(symbol))
			http.Error(w, "Failed to insert prediction", http.StatusInternalServerError)
			return
		}
		p.Logger.InfoContext(ctx, "Inserted prediction", logging.IDAttr(predID))

		var order model.Order
		reqStatus, errReq := tradeTransStatusResp.CheckRequestStatus()
		if reqStatus != response.REJECTED {
			// getTrades to acquire position required to forcibly close purchase on demand in history service
			getTradesResp, err := apiH.GetTrades(ctx, symbol, true)
			if err != nil {
				p.Logger.WarnContext(ctx, "Failed to process GetTrades, position number will not be added",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
			} else {
				p.Logger.InfoContext(ctx, "Successfully processed GetTrades", logging.RespAttr(getTradesResp))
			}

			var position int
			for _, tradeRecord := range getTradesResp.ReturnData {
				// Each transaction in the XTB system is assigned a unique order number. When a TradeTransaction
				// is executed by the client, the system returns an order number for that transaction.
				// Subsequently, request made by the client is processed by the system, which marks the start
				// of a new transaction in their system. As a result, the previous order number is transformed
				// into 'order2', indicating that the previous system transaction has simply ended,
				// and a new transaction with a different order number is being processed. When the previous
				// order number matches the 'order2' number retrieved from XTB, it indicates that transactions
				// in their system are connected, and it is possible to obtain a position number that represents
				// the entire sequence of transactions.
				//
				// This process resembles a chain of transactions: a TradeTransaction from the client assigns
				// an order number to that transaction, then the system processes the TradeTransaction,
				// assigns it a new order number, and saves the previous order number of transaction in the
				// 'order2' field. Finally, it returns the position number that represents the complete
				// chain of transactions in the system.
				if tradeTransResp.ReturnData.Order == tradeRecord.Order2 {
					position = tradeRecord.Position
					p.Logger.InfoContext(ctx, "Acquired order position", logging.PositionAttr(position))
					break
				}
			}

			order = model.Order{
				Number:         tradeTransResp.ReturnData.Order,
				UserID:         1,
				Position:       &position,
				PredictionID:   predID,
				RequestStatus:  response.RequestStatusName[reqStatus],
				Message:        tradeTransStatusResp.ReturnData.Message,
				TradeTransInfo: tradeTransInfo,
			}
		} else {
			rsErr := &response.RequestStatusError{}
			if errors.As(errReq, &rsErr) {
				order = model.Order{
					Number:         tradeTransResp.ReturnData.Order,
					UserID:         1,
					PredictionID:   predID,
					RequestStatus:  rsErr.RequestStatus,
					Message:        &rsErr.Message,
					TradeTransInfo: tradeTransInfo,
					OrderClosedDetails: model.OrderClosedDetails{
						Closed: true,
					},
				}
			}
		}
		ordID, err := dbH.InsertOrder(ctx, order)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to insert order",
				logging.ErrorAttr(err),
				logging.SymbolAttr(symbol))
			http.Error(w, "Failed to insert order", http.StatusInternalServerError)
			return
		}
		p.Logger.InfoContext(ctx, "Inserted order", logging.IDAttr(ordID))

		p.Logger.InfoContext(ctx, "Purchases request fully processed")
	}
}
