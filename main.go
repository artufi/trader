package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	ws "github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"trader/config"
	"trader/controller"
	"trader/controller/middleware"
	"trader/infrastructure/websocket"
	"trader/logging"
)

var env = "dev"

func main() {
	logFile := logging.Must(logging.GetLogFile("app.log"))
	logger := slog.New(logging.LogHandler{Handler: slog.NewJSONHandler(logFile, nil)})

	cfg := config.AppConfig{}
	config.Must(config.LoadFSYAMLConf(config.FSAppConfig, fmt.Sprintf("cfg-%s.yaml", env), &cfg))

	dialer := &ws.Dialer{}
	wsManager := websocket.NewWSManager(dialer, logger)

	r := chi.NewRouter()
	r.Use(middleware.ConnDetailsMiddleware(cfg, logger))
	r.Get("/payload", controller.TradeHandler(cfg, wsManager, logger))

	logger.Info("Starting application port: 3000...")
	http.ListenAndServe(":3000", r)
}
