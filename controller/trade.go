package controller

import (
	"log/slog"
	"net/http"
	"trader/config"
	"trader/infrastructure/websocket"
	"trader/logging"
	"trader/xtb"
	"trader/xtb/command"
)

func TradeHandler(cfg config.Config, wsManager *websocket.WSManager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := cfg.Testing.UserID
		logger.Info("Starting processing trade request.", logging.UserIDAttr(userID), logging.ClientIPAttr(r.RemoteAddr))

		wsClient, err := wsManager.Dial(cfg.XTB.Demo.WebSocketURL, nil, r.RemoteAddr, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer wsManager.RemoveClient(wsClient)

		loginJSON, err := xtb.Login(command.LoginArgs{
			UserID:   userID,
			Password: cfg.Testing.Password,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = wsClient.WriteText(loginJSON)
	}
}
