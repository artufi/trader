package main

import (
	"context"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/controller"
	"github.com/artufi/trader/history"
	"github.com/artufi/trader/infrastructure/database"
	"github.com/artufi/trader/infrastructure/database/migration"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/model"
	"github.com/go-chi/chi/v5"
	ws "github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

func main() {
	loader := config.DotEnvCfg{Filename: "config/.env"}
	cfg := config.MustLoad(loader)

	handlerOptions := &slog.HandlerOptions{
		AddSource: false,
	}
	logger := slog.New(logging.LogHandler{
		Handler: slog.NewJSONHandler(logging.Must(logging.GetLogFile("app.log")), handlerOptions),
	})

	db := database.MustOpen(database.OpenPool(context.Background(), cfg.Database.PostgresConfig))
	defer db.Close()
	err := database.Connect(db, cfg.Database.Connection.MaxAttempts, time.Duration(cfg.Database.Connection.NextTrySec))
	if err != nil {
		logger.Info("Failed to connect with database", logging.ErrorAttr(err))
		panic(err)
	}
	logger.Info("Connected to database", logging.URLAttr(cfg.Database.Host+":"+cfg.Database.Port))

	filenames := migration.Must(migration.Up(db))
	logger.Info("Loaded migrations", "migrations", filenames)

	userService := model.UserService{
		Logger: logger,
		DB:     db,
	}
	userID, err := strconv.Atoi(cfg.XTB.Demo.UserID)
	if err != nil {
		panic(err)
	}
	err = userService.Insert(userID)
	if err != nil {
		panic(err)
	}

	dialer := &ws.Dialer{}
	wsManager := websocket.NewWSManager(cfg, dialer, logger)

	connManager := websocket.ConnManager{
		Cfg:       cfg,
		WSManager: wsManager,
		Logger:    logger,
	}
	connManager.OpenConnPool(cfg.XTB.Demo.UserID, cfg.XTB.Demo.Password, cfg.Client.Connection.Pool.Size)

	historyService := &history.HService{
		Cfg:              cfg,
		Logger:           logger,
		WSManager:        wsManager,
		Dialer:           dialer,
		OrderService:     model.OrderService{DB: db},
		OrdersByPosition: make(map[int]int),
	}
	go historyService.GetTradesStream(cfg.XTB.Demo.UserID)

	go func() {
		ticker := time.NewTicker(time.Second * 30)
		for {
			select {
			case <-ticker.C:
				historyService.CloseEligibleOrders(cfg.XTB.Demo.UserID)
			}
		}
	}()

	purchaseC := controller.Purchase{
		Cfg:               cfg,
		WSManager:         wsManager,
		Logger:            logger,
		PredictionService: model.PredictionService{DB: db},
		OrderService:      model.OrderService{DB: db},
		UserService:       userService,
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
