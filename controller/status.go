package controller

import (
	"encoding/json"
	"fmt"
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
			http.Error(w, "Could not retrieve client details", http.StatusInternalServerError)
			return
		}
		traceID := connDetails.TraceID.String()

		wsClient, err := wsManager.DialForNewClient(ctx, cfg.XTB.Demo.WebSocketURL, nil)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create a new client", logging.ErrorAttr(err))
			http.Error(w, "Failed to establish connection", http.StatusInternalServerError)
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
			logger.ErrorContext(ctx, "Failed to login", logging.ErrorAttr(err))
			http.Error(w, "Failed to login", http.StatusBadRequest)
			return
		}
		logger.InfoContext(ctx, "Successfully processed Login", logging.RespAttr(loginResponse))

		order := struct {
			Number int `json:"number"`
		}{}
		err = json.NewDecoder(r.Body).Decode(&order)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to read request body with order number", logging.ErrorAttr(err))
			http.Error(w, "Failed to read body, expected order number", http.StatusBadRequest)
			return
		}

		// process TradeTransactionStatus
		tradeResponseStatus, err := h.TradeTransactionStatus(ctx, "", order.Number)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to process TradeTransactionStatus", logging.ErrorAttr(err))
			http.Error(w, fmt.Sprintf("Unable to obtain transaction status for order: %d", order.Number), http.StatusBadRequest)
			return
		}
		logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeResponseStatus))

		// logout user after processing
		logoutResponse, err := h.Logout(ctx)
		if err != nil {
			logger.WarnContext(ctx, "Failed to process Logout - killing client", logging.ErrorAttr(err))
			return
		}
		logger.InfoContext(ctx, "Successfully processed Logout", logging.RespAttr(logoutResponse))

		logger.Info("Transaction Status request fully processed")
	}
}
