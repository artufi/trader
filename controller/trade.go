package controller

import (
	"log/slog"
	"net/http"
	"trader/config"
	"trader/controller/middleware"
	"trader/infrastructure/websocket"
	"trader/xtb"
	"trader/xtb/command"
)

func TradeHandler(cfg config.Config, wsManager *websocket.WSManager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger.InfoContext(ctx, "Starting processing TRADE request")

		connDetails, ok := ctx.Value(middleware.ConnDetailsKey).(middleware.ConnDetails)
		if !ok {
			http.Error(w, "Connection details are not available", http.StatusBadRequest)
			return
		}

		wsClient, err := wsManager.Dial(ctx, cfg.XTB.Demo.WebSocketURL, nil)
		if err != nil {
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//defer wsManager.RemoveClient(wsClient)

		loginJSON, err := xtb.Login(command.LoginArgs{
			UserID:   connDetails.UserID,
			Password: cfg.Testing.Password,
		})
		if err != nil {
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		go wsClient.ReadMessages(ctx)

		// login
		err = wsClient.WriteText(loginJSON)
		if err != nil {
			wsManager.RemoveClient(ctx, wsClient)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
