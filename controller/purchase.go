package controller

import (
	"encoding/json"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/controller/middleware"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/model"
	"github.com/artufi/trader/xtb/processor"
	"log/slog"
	"net/http"
)

func PurchaseHandler(cfg config.AppConfig, wsManager *websocket.WSManager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.InfoContext(ctx, "Start processing purchase request")

		connDetails, ok := ctx.Value(middleware.ConnDetailsKey).(middleware.ConnDetails)
		if !ok {
			logger.ErrorContext(ctx, "Failed to get ConnDetails")
			http.Error(w, "Could not retrieve client details", http.StatusInternalServerError)
			return
		}
		traceID := connDetails.TraceID.String()

		wsClient, err := wsManager.DialForNewClient(ctx, cfg.XTB.Demo.WebSocketURL, nil)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create a new client", logging.ErrorAttr(err))
			http.Error(w, "Failed to establish client connection", http.StatusInternalServerError)
			return
		}
		defer func() {
			wsManager.RemoveClient(ctx, wsClient)
		}()

		h := Handler{
			Proc:    processor.NewProc(wsClient),
			TraceID: traceID,
		}

		// log user into XTB
		loginResponse, err := h.Login(ctx, cfg.XTB.Demo.UserID, cfg.XTB.Demo.Password)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to login")
			http.Error(w, "Failed to login", http.StatusBadRequest)
			return
		}
		logger.InfoContext(ctx, "Successfully processed Login", logging.RespAttr(loginResponse))

		// get purchase instructions
		purchaseInstructions := make([]model.PurchaseInstruction, 0)
		err = json.NewDecoder(r.Body).Decode(&purchaseInstructions)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to read request body with purchase instruction", logging.ErrorAttr(err))
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		for _, purchaseInstruction := range purchaseInstructions {
			logger.InfoContext(ctx, "Will process purchase instruction", logging.PurchaseInstrAttr(purchaseInstruction))

			symbol := purchaseInstruction.Symbol

			// get symbol details
			symbolResponse, err := h.GetSymbolExtended(ctx, symbol)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to process GetSymbol",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			logger.InfoContext(ctx, "Successfully processed GetSymbol", logging.RespAttr(symbolResponse))

			// prepare TradeTransactionInfo to pass it to TradeTransaction as argument
			tradeTransInfo, err := purchaseInstruction.PrepareTradeTransInfo(symbolResponse, 0.2, 0.2)
			if err != nil {
				logger.WarnContext(ctx, "Failed to prepare TradeTransInfo", logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}

			// process TradeTransaction
			tradeResponse, err := h.TradeTransaction(ctx, symbol, tradeTransInfo)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to process TradeTransaction",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			logger.InfoContext(ctx, "Successfully processed TradeTransaction", logging.RespAttr(tradeResponse))

			// process TradeTransactionStatus
			tradeResponseStatus, err := h.TradeTransactionStatus(ctx, symbol, tradeResponse.ReturnData.Order)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}

			//reqStatus, err := tradeResponseStatus.CheckRequestStatus()
			//if reqStatus == response.ACCEPTED || reqStatus == response.PENDING || reqStatus == response.REJECTED {
			//	go sendPing(ctx, wsClient, logger, 5)
			//}
			logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeResponseStatus))
		}

		// logout user after processing
		logoutResponse, err := h.Logout(ctx)
		if err != nil {
			logger.WarnContext(ctx, "Failed to process Logout - killing client", logging.ErrorAttr(err))
			return
		}
		logger.InfoContext(ctx, "Successfully processed Logout", logging.RespAttr(logoutResponse))

		logger.Info("Purchase request fully processed")
	}
}
