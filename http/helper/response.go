package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/artufi/trader/logging"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type Response struct {
	Logger *slog.Logger
	Writer http.ResponseWriter
}

func (rh Response) WriteJSON(ctx context.Context, statusCode int, jsonData any) {
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

func (rh Response) WriteErrorJSON(ctx context.Context, statusCode int, errMsg string) {
	resp := ErrorResponse{Error: errMsg}
	rh.WriteJSON(ctx, statusCode, resp)
}

func (rh Response) WriteAndLogError(ctx context.Context, statusCode int, msg string, err error) {
	rh.Logger.ErrorContext(ctx, msg, logging.ErrorAttr(err))
	rh.WriteErrorJSON(ctx, statusCode, msg)
}

func (rh Response) writeRawJSON(ctx context.Context, statusCode int, data []byte) {
	rh.setHeaders()
	rh.Writer.WriteHeader(statusCode)
	if _, err := rh.Writer.Write(data); err != nil {
		rh.Logger.ErrorContext(ctx, "Failed to write response", logging.ErrorAttr(err))
	}
}

func (rh Response) setHeaders() {
	rh.Writer.Header().Set("Content-Type", "application/json")
}
