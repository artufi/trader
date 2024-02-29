package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"trader/config"
	"trader/infrastructure/websocket"
	"trader/logging"
	"trader/xtb/processor"
)

func TransactionStatusHandler(cfg config.AppConfig, wsManager *websocket.WSManager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.InfoContext(ctx, "Start processing transaction status request")

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
		logger.InfoContext(ctx, "Successfully processed Login", logging.RespAttr(loginResponse))

		order := struct {
			Number int `json:"number"`
		}{}
		err = json.NewDecoder(r.Body).Decode(&order)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to read request body with purchase instruction", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		tradeResponseStatus, err := processor.TradeTransactionStatus(ctx, order.Number, wsClient)
		if err != nil {
			logger.WarnContext(ctx, "Failed to process TradeTransactionStatus", logging.ErrorAttr(err))
		}
		logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeResponseStatus))
	}
}
