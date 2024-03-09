package main

import (
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/controller"
	"github.com/artufi/trader/controller/middleware"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/logging"
	"github.com/go-chi/chi/v5"
	ws "github.com/gorilla/websocket"
	"log/slog"
	"net/http"
)

func main() {
	loader := config.DotEnvCfg{Filename: "config/.env"}
	cfg := config.MustLoad(loader)

	logFile := logging.Must(logging.GetLogFile("app.log"))
	logger := slog.New(logging.LogHandler{Handler: slog.NewJSONHandler(logFile, nil)})

	dialer := &ws.Dialer{}
	wsManager := websocket.NewWSManager(dialer, logger)

	r := chi.NewRouter()
	r.Use(middleware.ConnDetailsMiddleware(cfg, logger))
	r.Get("/purchase", controller.PurchaseHandler(cfg, wsManager, logger))
	r.Get("/purchase/status", controller.TransactionStatusHandler(cfg, wsManager, logger))

	logger.Info("Starting application port: 3000...")
	http.ListenAndServe(":3000", r)
}
