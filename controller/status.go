package controller

import (
	"encoding/json"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/controller/middleware"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/xtb/processor"
	"log/slog"
	"net/http"
)

func TransactionStatusHandler(cfg config.AppConfig, wsManager *websocket.WSManager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.InfoContext(ctx, "Start processing transaction status request")

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
		lCustomTag := traceID + "login"
		loginResponse, err := processor.Login(ctx, cfg, wsClient, lCustomTag)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to process Login", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, "Failed to login", http.StatusInternalServerError)
			return
		}
		err = loginResponse.CheckCustomTag(lCustomTag)
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

		// process tradeTransactionStatus
		ttsCustomTag := traceID + "tts"
		tradeResponseStatus, err := processor.TradeTransactionStatus(ctx, order.Number, wsClient, ttsCustomTag)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, "Failed to process TradeTransactionStatus", http.StatusBadRequest)
			return
		}
		err = tradeResponseStatus.CheckCustomTag(ttsCustomTag)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, "Failed to process TradeTransactionStatus", http.StatusBadRequest)
			return
		}
		err = tradeResponseStatus.CheckStatus()
		if err != nil {
			logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, "Failed to process TradeTransactionStatus", http.StatusBadRequest)
			return
		}
		_, err = tradeResponseStatus.CheckRequestStatus()
		if err != nil {
			logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus", logging.ErrorAttr(err))
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, "Failed to process TradeTransactionStatus", http.StatusBadRequest)
			return
		}
		logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeResponseStatus))

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

		logger.Info("Transaction Status request fully processed")
	}
}
