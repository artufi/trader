package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/controller/middleware"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/model"
	"github.com/artufi/trader/xtb/processor"
	"github.com/artufi/trader/xtb/response"
	"log/slog"
	"net/http"
)

type Purchase struct {
	Cfg       config.AppConfig
	WSManager *websocket.WSManager
	Logger    *slog.Logger

	PositionService model.PositionService
	OrderService    model.OrderService
}

func (p *Purchase) PurchasesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		p.Logger.InfoContext(ctx, "Start processing purchases request")

		connDetails, ok := ctx.Value(middleware.ConnDetailsKey).(middleware.ConnDetails)
		if !ok {
			p.Logger.ErrorContext(ctx, "Failed to get ConnDetails")
			http.Error(w, "Could not retrieve client details", http.StatusInternalServerError)
			return
		}
		traceID := connDetails.TraceID.String()

		wsClient, err := p.WSManager.DialForNewClient(ctx, p.Cfg.XTB.Demo.WebSocketURL, nil)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to create a new client", logging.ErrorAttr(err))
			http.Error(w, "Failed to establish client connection", http.StatusInternalServerError)
			return
		}
		defer func() {
			p.WSManager.RemoveClient(ctx, wsClient)
		}()

		// init api handler
		apiH := APIHandler{
			Proc:    processor.NewProc(wsClient),
			TraceID: traceID,
		}

		dbH := DBHandler{
			TraceID:         traceID,
			Logger:          p.Logger,
			PositionService: p.PositionService,
			OrderService:    p.OrderService,
		}

		// log user into XTB
		loginResponse, err := apiH.Login(ctx, p.Cfg.XTB.Demo.UserID, p.Cfg.XTB.Demo.Password)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to login", logging.ErrorAttr(err))
			http.Error(w, "Failed to login", http.StatusBadRequest)
			return
		}
		p.Logger.InfoContext(ctx, "Successfully processed Login", logging.RespAttr(loginResponse))

		// get purchase instructions
		purchaseInstructions := make([]model.PurchaseInstruction, 0)
		err = json.NewDecoder(r.Body).Decode(&purchaseInstructions)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to read request body with purchase instruction", logging.ErrorAttr(err))
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		for _, purchaseInstruction := range purchaseInstructions {
			p.Logger.InfoContext(ctx, "Will process purchase instruction", logging.PurchaseInstrAttr(purchaseInstruction))

			symbol := purchaseInstruction.Symbol

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
			tradeTransInfo, err := purchaseInstruction.PrepareTradeTransInfo(symbolResponse, 0.01)
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

			reqStatus, err := tradeResponseStatus.CheckRequestStatus()
			order := model.Order{
				Number:         tradeResponse.ReturnData.Order,
				UserID:         1,
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
						RequestStatus:  rsErr.RequestStatus,
						Message:        rsErr.Message,
						TradeTransInfo: tradeTransInfo,
					}
				}
			}

			ordID, err := dbH.InsertOrder(ctx, order)
			if err != nil {
				p.Logger.ErrorContext(ctx, "Failed to Insert Order",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			p.Logger.InfoContext(ctx, "Inserted order", logging.IDAttr(ordID))

			position := model.Position{
				UserID:              1,
				OrderID:             ordID,
				PurchaseInstruction: purchaseInstruction,
			}
			posID, err := dbH.InsertPosition(ctx, position)
			if err != nil {
				p.Logger.ErrorContext(ctx, "Failed to Insert Position",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			p.Logger.InfoContext(ctx, "Inserted position", logging.IDAttr(posID))
		}

		// logout user after processing
		logoutResponse, err := apiH.Logout(ctx)
		if err != nil {
			p.Logger.WarnContext(ctx, "Failed to process Logout - killing client", logging.ErrorAttr(err))
			return
		}
		p.Logger.InfoContext(ctx, "Successfully processed Logout", logging.RespAttr(logoutResponse))

		p.Logger.Info("Purchases request fully processed")
	}
}

func (p *Purchase) PurchaseHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		p.Logger.InfoContext(ctx, "Start processing purchase request")

		connDetails, ok := ctx.Value(middleware.ConnDetailsKey).(middleware.ConnDetails)
		if !ok {
			p.Logger.ErrorContext(ctx, "Failed to get ConnDetails")
			http.Error(w, "Could not retrieve client details", http.StatusInternalServerError)
			return
		}
		traceID := connDetails.TraceID.String()

		wsClient, err := p.WSManager.DialForNewClient(ctx, p.Cfg.XTB.Demo.WebSocketURL, nil)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to create a new client", logging.ErrorAttr(err))
			http.Error(w, "Failed to establish client connection", http.StatusInternalServerError)
			return
		}
		defer func() {
			p.WSManager.RemoveClient(ctx, wsClient)
		}()

		h := APIHandler{
			Proc:    processor.NewProc(wsClient),
			TraceID: traceID,
		}

		// log user into XTB
		loginResponse, err := h.Login(ctx, p.Cfg.XTB.Demo.UserID, p.Cfg.XTB.Demo.Password)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to login", logging.ErrorAttr(err))
			http.Error(w, "Failed to login", http.StatusBadRequest)
			return
		}
		p.Logger.InfoContext(ctx, "Successfully processed Login", logging.RespAttr(loginResponse))

		// get purchase instructions
		var purchaseInstruction model.PurchaseInstruction
		err = json.NewDecoder(r.Body).Decode(&purchaseInstruction)
		if err != nil {
			p.Logger.ErrorContext(ctx, "Failed to read request body with purchase instruction", logging.ErrorAttr(err))
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		p.Logger.InfoContext(ctx, "Will process purchase instruction", logging.PurchaseInstrAttr(purchaseInstruction))

		symbol := purchaseInstruction.Symbol

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
		tradeTransInfo, err := purchaseInstruction.PrepareTradeTransInfo(symbolResponse, 0.01)
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
