package middleware

import (
	"context"
	"github.com/artufi/trader/http/helper"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type traceIDCtxKey string

const (
	traceIDKey     traceIDCtxKey = "traceID"
	XTraceIDHeader               = "X-Trace-ID"
)

func TraceIDMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Header.Get(XTraceIDHeader)
			if traceID == "" {
				var err error
				traceID, err = generateTraceID()
				if err != nil {
					rh := helper.Response{Logger: logger, Writer: w}
					rh.WriteErrorJSON(r.Context(), http.StatusInternalServerError, "Failed to generate traceID for the request")
					return
				}
			}
			ctx := context.WithValue(r.Context(), traceIDKey, traceID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func generateTraceID() (string, error) {
	uuidTraceID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return uuidTraceID.String(), nil
}

func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}
