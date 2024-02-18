package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	ws "github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"path/filepath"
	"trader/config"
	"trader/controller"
	"trader/infrastructure/websocket"
	"trader/logging"
)

var env = "dev"

func main() {
	logFile := logging.Must(logging.GetLogFile("app.log"))
	logger := slog.New(slog.NewJSONHandler(logFile, nil))

	cfg := config.Config{}
	config.Must(config.LoadYAMLConf(filepath.Join("config", fmt.Sprintf("cfg-%s.yaml", env)), &cfg))

	dialer := &ws.Dialer{}
	wsManager := websocket.NewWSManager(dialer, logger)

	r := chi.NewRouter()
	r.Get("/payload", controller.TradeHandler(cfg, wsManager, logger))

	http.ListenAndServe(":3000", r)
}
