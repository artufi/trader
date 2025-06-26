package controller

import (
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/http/helper"
	"github.com/artufi/trader/http/middleware"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/xtb/processor"
	"log/slog"
	"net/http"
)

func TransactionStatusHandler(cfg config.AppConfig, wsManager *websocket.WSManager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := middleware.GetTraceID(r.Context())
		userID := cfg.XTB.Demo.UserID
		ctx := logging.AppendAttrsCtx(r.Context(), logging.TraceIDAttr(traceID), logging.UserIDAttr(userID))

		logger.InfoContext(ctx, "Start processing transaction status request")

		rh := helper.Response{Logger: logger, Writer: w, TraceID: traceID}

		wsClient, err := wsManager.DialForNewClient(ctx, cfg.XTB.Demo.WebSocketURL, nil, userID)
		if err != nil {
			rh.WriteAndLogError(ctx, http.StatusInternalServerError, "Failed to establish connection", err)
			return
		}
		defer func() {
			wsManager.RemoveClient(ctx, wsClient)
		}()

		apiH := APIHandler{
			Proc:    processor.NewProc(wsClient),
			TraceID: traceID,
		}

		loginResponse, err := apiH.Login(ctx, cfg.XTB.Demo.UserID, cfg.XTB.Demo.Password)
		if err != nil {
			rh.WriteAndLogError(ctx, http.StatusBadRequest, "Login failed", err)
			return
		}
		logger.InfoContext(ctx, "Successfully processed Login", logging.RespAttr(loginResponse))

		order := struct {
			Number int `json:"number"`
		}{}
		err = json.NewDecoder(r.Body).Decode(&order)
		if err != nil {
			rh.WriteAndLogError(ctx, http.StatusBadRequest, "Failed to decode request body with order number", err)
			return
		}

		tradeResponseStatus, err := apiH.TradeTransactionStatus(ctx, "", order.Number)
		if err != nil {
			rh.WriteAndLogError(ctx, http.StatusBadRequest, fmt.Sprintf("Unable to obtain transaction status for order: %d", order.Number), err)
			return
		}
		logger.InfoContext(ctx, "Successfully processed TradeTransactionStatus", logging.RespAttr(tradeResponseStatus))

		// Logout user after processing.
		logoutResponse, err := apiH.Logout(ctx)
		if err != nil {
			logger.WarnContext(ctx, "Failed to process Logout - killing client", logging.ErrorAttr(err))
			return
		}
		logger.InfoContext(ctx, "Successfully processed Logout", logging.RespAttr(logoutResponse))

		logger.Info("Transaction Status request fully processed")
	}
}
