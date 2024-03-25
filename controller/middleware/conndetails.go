package middleware

import (
	"context"
	"github.com/artufi/trader/config"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type connDetailsKey string

const ConnDetailsKey connDetailsKey = "connDetails"

type ConnDetails struct {
	TraceID uuid.UUID
	UserID  string
}

func ConnDetailsMiddleware(cfg config.AppConfig, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := uuid.New()

			connDetails := ConnDetails{
				TraceID: traceID,
				UserID:  cfg.XTB.Demo.UserID,
			}

			ctx := context.WithValue(r.Context(), ConnDetailsKey, connDetails)
			logger.InfoContext(ctx, "Got request to process")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
