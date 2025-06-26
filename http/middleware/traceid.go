package middleware

import (
	"context"
	"github.com/google/uuid"
	"net/http"
)

type traceIDKeyType string

const traceIDKey traceIDKeyType = "traceID"

const XTraceIDHeader = "X-Trace-ID"

func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			uuidTraceID, err := uuid.NewRandom()
			if err != nil {
				http.Error(w, "Failed to generate traceID for the request", http.StatusInternalServerError)
				return
			}
			traceID = uuidTraceID.String()
		}
		ctx := context.WithValue(r.Context(), traceIDKey, traceID)
		w.Header().Set("X-Trace-ID", traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}
