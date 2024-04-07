package main

import (
	"context"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/controller"
	"github.com/artufi/trader/infrastructure/database"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/model"
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

	db := database.MustOpen(database.OpenPool(context.Background(), cfg.Database.PostgresConfig))
	defer db.Close()
	err := database.Connect(db, cfg.Database.Connection.MaxAttempts, time.Duration(cfg.Database.Connection.NextTrySec))
	if err != nil {
		logger.Info("Failed to connect with database", logging.ErrorAttr(err))
		panic(err)
	}
	logger.Info("Connected to database", logging.URLAttr(cfg.Database.Host+":"+cfg.Database.Port))

	dialer := &ws.Dialer{}
	wsManager := websocket.NewWSManager(dialer, logger)

	connService := websocket.ConnService{
		Cfg:       cfg,
		WSManager: wsManager,
		Logger:    logger,
	}
	connService.OpenConnPool(cfg.XTB.Demo.UserID, cfg.XTB.Demo.Password, cfg.Client.Connection.Pool.Size)

	purchaseC := controller.Purchase{
		Cfg:               cfg,
		WSManager:         wsManager,
		Logger:            logger,
		PredictionService: model.PredictionService{DB: db},
		OrderService:      model.OrderService{DB: db},
	}

	r := chi.NewRouter()
	r.Get("/purchases", purchaseC.PurchasesHandler())
	r.Get("/purchase", purchaseC.PurchaseHandler())

	r.Get("/purchase/status", controller.TransactionStatusHandler(cfg, wsManager, logger))

	logger.Info("Starting application port: 4000...")
	err = http.ListenAndServe(":4000", r)
	if err != nil {
		panic(err)
	}
}
