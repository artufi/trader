package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"trader/config"
	"trader/infrastructure/websocket"
	"trader/logging"
	"trader/model"
	"trader/xtb/processor"
)

func PurchaseHandler(cfg config.AppConfig, wsManager *websocket.WSManager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.InfoContext(ctx, "Start processing purchase request")

		wsClient, err := wsManager.DialForNewClient(ctx, cfg.XTB.Demo.WebSocketURL, nil)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create a new client", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, "Failed to establish connection", http.StatusInternalServerError)
			return
		}

		// read client messages
		go wsClient.ReadMessages(ctx)

		// log user into XTB
		loginResponse, err := processor.Login(ctx, cfg, wsClient)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to process Login", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, "Failed to login", http.StatusInternalServerError)
			return
		}
		err = loginResponse.CheckStatus()
		if err != nil {
			logger.ErrorContext(ctx, "Failed to process Login", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, "Failed to login", http.StatusInternalServerError)
			return
		}
		logger.InfoContext(ctx, "Successfully processed Login", logging.RespAttr(loginResponse))

		// get purchase instructions
		purchaseInstructions := make([]model.PurchaseInstruction, 0)
		err = json.NewDecoder(r.Body).Decode(&purchaseInstructions)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to read request body with purchase instruction", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		for _, purchaseInstruction := range purchaseInstructions {
			logger.InfoContext(ctx, "Will process purchase instruction", logging.PurchaseInstrAttr(purchaseInstruction))

			// get symbol details
			symbolResponse, err := processor.GetSymbolExtended(ctx, purchaseInstruction.Symbol, wsClient)
			if err != nil {
				logger.WarnContext(ctx, "Failed to process GetSymbol",
					logging.ErrorAttr(err),
					logging.SymbolAttr(purchaseInstruction.Symbol))
				continue
			}
			err = symbolResponse.CheckStatus()
			if err != nil {
				logger.WarnContext(ctx, "Failed to process GetSymbol",
					logging.ErrorAttr(err),
					logging.SymbolAttr(purchaseInstruction.Symbol))
				continue
			}
			logger.InfoContext(ctx, "Successfully processed GetSymbol", logging.RespAttr(symbolResponse))

			// prepare TradeTransactionInfo to pass it to TradeTransaction as argument
			tradeTransInfo, err := purchaseInstruction.PrepareTradeTransInfo(symbolResponse, 0.2, 0.2)
			if err != nil {
				logger.WarnContext(ctx, "Failed to prepare TradeTransInfo", logging.ErrorAttr(err),
					logging.SymbolAttr(purchaseInstruction.Symbol))
				continue
			}

			// process TradeTransaction
			tradeResponse, err := processor.TradeTransaction(ctx, tradeTransInfo, wsClient)
			if err != nil {
				logger.WarnContext(ctx, "Failed to process TradeTransaction",
					logging.ErrorAttr(err),
					logging.SymbolAttr(purchaseInstruction.Symbol))
				continue
			}
			err = tradeResponse.CheckStatus()
			if err != nil {
				logger.WarnContext(ctx, "Failed to process TradeTransaction",
					logging.ErrorAttr(err),
					logging.SymbolAttr(purchaseInstruction.Symbol))
				continue
			}
			logger.InfoContext(ctx, "Successfully processed TradeTransaction", logging.RespAttr(tradeResponse))

			// tradeTransactionStatus
			tradeResponseStatus, err := processor.TradeTransactionStatus(ctx, tradeResponse.ReturnData.Order, wsClient)
			if err != nil {
				logger.WarnContext(ctx, "Failed to process TradeTransactionStatus",
					logging.ErrorAttr(err),
					logging.SymbolAttr(purchaseInstruction.Symbol))
				continue
			}
			err = tradeResponseStatus.CheckStatus()
			if err != nil {
				logger.WarnContext(ctx, "Failed to process TradeTransactionStatus",
					logging.ErrorAttr(err),
					logging.SymbolAttr(purchaseInstruction.Symbol))
				continue
			}
			err = tradeResponseStatus.CheckRequestStatus()
			if err != nil {
				logger.WarnContext(ctx, "Failed to process TradeTransactionStatus",
					logging.ErrorAttr(err),
					logging.SymbolAttr(purchaseInstruction.Symbol))
				continue
			}
			logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeResponseStatus))
		}

		logoutResponse, err := processor.Logout(ctx, wsClient)
		if err != nil {
			logger.WarnContext(ctx, "Failed to process Logout - killing client", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			return
		}
		logger.InfoContext(ctx, "Successfully processed Logout", logging.RespAttr(logoutResponse))

		logger.Info("Request fully processed")
	}
}
