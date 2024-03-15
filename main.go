package main

import (
	"context"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/controller"
	"github.com/artufi/trader/controller/middleware"
	"github.com/artufi/trader/infrastructure/database"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/logging"
	"github.com/go-chi/chi/v5"
	ws "github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"time"
)

func main() {
	loader := config.DotEnvCfg{Filename: "config/.env"}
	cfg := config.MustLoad(loader)

	logFile := logging.Must(logging.GetLogFile("app.log"))
	logger := slog.New(logging.LogHandler{Handler: slog.NewJSONHandler(logFile, nil)})

	db := database.MustOpen(database.OpenPool(context.Background(), cfg.Database))
	defer db.Close()
	err := database.Connect(db, 10, time.Second*2)
	if err != nil {
		panic(err)
	}
	logger.Info("Connected to database", logging.URLAttr(cfg.Database.Host+cfg.Database.Port))

	dialer := &ws.Dialer{}
	wsManager := websocket.NewWSManager(dialer, logger)

	r := chi.NewRouter()
	r.Use(middleware.ConnDetailsMiddleware(cfg, logger))
	r.Get("/purchases", controller.PurchasesHandler(cfg, wsManager, logger))
	r.Get("/purchase", controller.PurchaseHandler(cfg, wsManager, logger))
	r.Get("/purchase/status", controller.TransactionStatusHandler(cfg, wsManager, logger))

	logger.Info("Starting application port: 4000...")
	err = http.ListenAndServe(":4000", r)
	if err != nil {
		panic(err)
	}
}
