package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/http/middleware"
	"github.com/artufi/trader/logging"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ResponseHelper struct {
	Logger  *slog.Logger
	Writer  http.ResponseWriter
	TraceID string
}

func (rh ResponseHelper) WriteJSON(ctx context.Context, statusCode int, jsonData any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(jsonData); err != nil {
		rh.Logger.ErrorContext(ctx, "Failed to encode response", logging.ErrorAttr(err))

		// Fallback error response to prevent infinite recursion if encoding fails.
		fallback := ErrorResponse{Error: http.StatusText(http.StatusInternalServerError)}
		raw, err := json.Marshal(fallback)
		if err != nil {
			rh.Logger.ErrorContext(ctx, "Failed to marshal fallback error", logging.ErrorAttr(err))
			rh.setHeaders()
			rh.Writer.WriteHeader(http.StatusInternalServerError)
			rh.Writer.Write([]byte(fmt.Sprintf(`{"error":%q}`, err)))
			return
		}

		rh.writeRawJSON(ctx, http.StatusInternalServerError, raw)
		return
	}

	rh.writeRawJSON(ctx, statusCode, buf.Bytes())
}

func (rh ResponseHelper) WriteErrorJSON(ctx context.Context, statusCode int, errMsg string) {
	resp := ErrorResponse{Error: errMsg}
	rh.WriteJSON(ctx, statusCode, resp)
}

func (rh ResponseHelper) WriteAndLogError(ctx context.Context, statusCode int, msg string, err error) {
	rh.Logger.ErrorContext(ctx, msg, logging.ErrorAttr(err))
	rh.WriteErrorJSON(ctx, statusCode, msg)
}

func (rh ResponseHelper) writeRawJSON(ctx context.Context, statusCode int, data []byte) {
	rh.setHeaders()
	rh.Writer.WriteHeader(statusCode)
	if _, err := rh.Writer.Write(data); err != nil {
		rh.Logger.ErrorContext(ctx, "Failed to write response", logging.ErrorAttr(err))
	}
}

func (rh ResponseHelper) setHeaders() {
	rh.Writer.Header().Set("Content-Type", "application/json")
	if rh.TraceID != "" {
		rh.Writer.Header().Set(middleware.XTraceIDHeader, rh.TraceID)
	}
}
