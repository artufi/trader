package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/artufi/trader/config"
	"github.com/artufi/trader/history"
	"github.com/artufi/trader/http/controller"
	"github.com/artufi/trader/http/middleware"
	"github.com/artufi/trader/infrastructure/database"
	"github.com/artufi/trader/infrastructure/database/migration"
	"github.com/artufi/trader/infrastructure/websocket"
	"github.com/artufi/trader/logging"
	"github.com/artufi/trader/model"
	"github.com/go-chi/chi/v5"
	ws "github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := loadConfig()
	logger := initLogger()

	listenForShutdownSignal(logger, cancel)

	db, err := initDatabase(ctx, cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize database", logging.ErrorAttr(err))
		os.Exit(1)
	}
	defer func() {
		if dbErr := db.Close(); dbErr != nil {
			logger.Error("Failed to close database", logging.ErrorAttr(dbErr))
		}
	}()

	err = runMigrations(db, logger)
	if err != nil {
		logger.Error("Failed to run database migrations", logging.ErrorAttr(err))
		os.Exit(1)
	}

	userService := model.UserService{
		Logger: logger,
		DB:     db,
	}
	if err = initTestUser(cfg, logger, userService); err != nil {
		logger.Error("Failed to initialize user", logging.ErrorAttr(err))
		os.Exit(1)
	}

	// Initialize WebSocket manager responsible for clients lifecycle.
	dialer := &ws.Dialer{}
	wsManager := websocket.NewWSManager(cfg, dialer, logger)

	connManager := websocket.ConnManager{
		Cfg:       cfg,
		WSManager: wsManager,
		Logger:    logger,
	}
	// Open connection pool to keep user's clients connected.
	connManager.OpenConnPool(ctx, cfg.XTB.Demo.UserID, cfg.XTB.Demo.Password, cfg.Client.Connection.Pool.Size)

	historyService := &history.HService{
		Cfg:              cfg,
		Logger:           logger,
		WSManager:        wsManager,
		Dialer:           dialer,
		OrderService:     model.OrderService{DB: db},
		OrdersByPosition: make(map[int]int),
	}
	// Start streaming user's trades.
	go historyService.GetTradesStream(ctx, cfg.XTB.Demo.UserID)
	// Close user's eligible orders every minute.
	go historyService.StartOrderCloseLoop(ctx, cfg.XTB.Demo.UserID, time.Duration(cfg.Client.Order.Close.IntervalSec))

	purchaseC := controller.Purchase{
		Cfg:               cfg,
		WSManager:         wsManager,
		Logger:            logger,
		PredictionService: model.PredictionService{DB: db},
		OrderService:      model.OrderService{DB: db},
		UserService:       userService,
	}
	router := setupHTTPRouter(cfg, logger, purchaseC, wsManager)
	srv := runHTTPServer(router, logger, cancel)

	<-ctx.Done()
	err = gracefulShutdown(logger, srv)
	if err != nil {
		logger.Error("Failed to shutdown server", logging.ErrorAttr(err))
		os.Exit(1)
	}
}

func loadConfig() config.AppConfig {
	loader := config.DotEnvCfg{Filename: "config/.env"}
	cfg := config.MustLoad(loader)
	return cfg
}

func initLogger() *slog.Logger {
	handlerOptions := &slog.HandlerOptions{
		AddSource: true,
	}
	logger := slog.New(logging.LogHandler{
		Handler: slog.NewJSONHandler(logging.Must(logging.GetLogFile("app.log")), handlerOptions),
	})
	return logger
}

func listenForShutdownSignal(logger *slog.Logger, cancel context.CancelFunc) {
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit
		logger.Error("Received shutdown signal", slog.String("signal", sig.String()))
		cancel()
	}()
}

func initDatabase(ctx context.Context, cfg config.AppConfig, logger *slog.Logger) (*sql.DB, error) {
	db := database.MustOpen(database.OpenPool(ctx, cfg.Database.PostgresConfig))
	err := database.Connect(db, cfg.Database.Connection.MaxAttempts, time.Duration(cfg.Database.Connection.NextTrySec))
	if err != nil {
		return nil, fmt.Errorf("database connection: %w", err)
	}

	logger.Info("Connected to database", logging.URLAttr(fmt.Sprintf("%s:%s", cfg.Database.Host, cfg.Database.Port)))
	return db, nil
}

func runMigrations(db *sql.DB, logger *slog.Logger) error {
	filenames, err := migration.Up(db)
	if err != nil {
		return err
	}
	logger.Info("Loaded migrations", slog.Any("migrations", filenames))
	return nil
}

func initTestUser(cfg config.AppConfig, logger *slog.Logger, userService model.UserService) error {
	userID, err := strconv.Atoi(cfg.XTB.Demo.UserID)
	if err != nil {
		return fmt.Errorf("invalid XTB user ID: %w", err)
	}
	if err = userService.Insert(userID); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	logger.Info("Initialized user", logging.UserIDAttr(cfg.XTB.Demo.UserID))
	return nil
}

func setupHTTPRouter(cfg config.AppConfig, logger *slog.Logger, purchaseC controller.Purchase, wsManager *websocket.WSManager) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.TraceIDMiddleware(logger))
	r.Get("/purchases", purchaseC.PurchasesHandler)
	r.Route("/purchase", func(r chi.Router) {
		r.Get("/", purchaseC.PurchaseHandler)
		r.Get("/status", controller.TransactionStatusHandler(cfg, wsManager, logger))
	})
	return r
}

func runHTTPServer(r http.Handler, logger *slog.Logger, cancel context.CancelFunc) *http.Server {
	port := ":4000"
	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	logger.Info("Starting HTTP server", slog.String("port", port))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", logging.ErrorAttr(err))
			cancel()
		}
	}()
	return srv
}

func gracefulShutdown(logger *slog.Logger, srv *http.Server) error {
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	return nil
}
