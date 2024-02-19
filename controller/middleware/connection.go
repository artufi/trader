package middleware

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"trader/config"
)

type connDetailsKey string

const ConnDetailsKey connDetailsKey = "connDetails"

type ConnDetails struct {
	TraceID uuid.UUID
	IPAddr  string
	UserID  string
}

func ConnDetailsMiddleware(cfg config.Config, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := uuid.New()
			ipAddr := r.RemoteAddr

			connDetails := ConnDetails{
				TraceID: traceID,
				IPAddr:  ipAddr,
				UserID:  cfg.Testing.UserID,
			}

			ctx := context.WithValue(r.Context(), ConnDetailsKey, connDetails)
			logger.InfoContext(ctx, "Got request to process")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
