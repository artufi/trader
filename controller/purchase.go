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
			http.Error(w, "Unspecified parameters", http.StatusBadRequest)
			return
		}
		traceID := connDetails.TraceID.String()

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
		loginCustomTag := traceID + "login"
		loginResponse, err := processor.Login(ctx, cfg, wsClient, loginCustomTag)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to process Login", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, "Failed to login", http.StatusInternalServerError)
			return
		}
		err = loginResponse.CheckCustomTag(loginCustomTag)
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

			symbol := purchaseInstruction.Symbol

			// get symbol details
			gsCustomTag := traceID + "gse" + symbol
			symbolResponse, err := processor.GetSymbolExtended(ctx, symbol, wsClient, gsCustomTag)
			if err != nil {
				logger.WarnContext(ctx, "Failed to process GetSymbol",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			err = symbolResponse.CheckCustomTag(gsCustomTag)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to process GetSymbol",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			err = symbolResponse.CheckStatus()
			if err != nil {
				logger.WarnContext(ctx, "Failed to process GetSymbol",
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
			ttCustomTag := traceID + "tt" + symbol
			tradeResponse, err := processor.TradeTransaction(ctx, tradeTransInfo, wsClient, ttCustomTag)
			if err != nil {
				logger.WarnContext(ctx, "Failed to process TradeTransaction",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			err = tradeResponse.CheckCustomTag(ttCustomTag)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to process TradeTransaction",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			err = tradeResponse.CheckStatus()
			if err != nil {
				logger.WarnContext(ctx, "Failed to process TradeTransaction",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			logger.InfoContext(ctx, "Successfully processed TradeTransaction", logging.RespAttr(tradeResponse))

			// process tradeTransactionStatus
			ttsCustomTag := traceID + "tts" + symbol
			tradeResponseStatus, err := processor.TradeTransactionStatus(ctx, tradeResponse.ReturnData.Order, wsClient,
				ttsCustomTag)
			if err != nil {
				logger.WarnContext(ctx, "Failed to process TradeTransactionStatus",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			err = tradeResponseStatus.CheckCustomTag(ttsCustomTag)
			if err != nil {
				logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			err = tradeResponseStatus.CheckStatus()
			if err != nil {
				logger.WarnContext(ctx, "Failed to process TradeTransactionStatus",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			_, err = tradeResponseStatus.CheckRequestStatus()
			if err != nil {
				logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus",
					logging.ErrorAttr(err),
					logging.SymbolAttr(symbol))
				continue
			}
			logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeResponseStatus))

			//if reqStatus == response.ACCEPTED || reqStatus == response.PENDING || reqStatus == response.REJECTED {
			//	go sendPing(ctx, wsClient, logger, 5)
			//}
		}

		// logout user after processing
		logoutCustomTag := traceID + "logout"
		logoutResponse, err := processor.Logout(ctx, wsClient, logoutCustomTag)
		if err != nil {
			logger.WarnContext(ctx, "Failed to process Logout - killing client", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			return
		}
		err = logoutResponse.CheckStatus()
		if err != nil {
			logger.WarnContext(ctx, "Failed to process Logout - killing client", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			return
		}
		logger.InfoContext(ctx, "Successfully processed Logout", logging.RespAttr(logoutResponse))

		logger.Info("Purchase request fully processed")
	}
}
