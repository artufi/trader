package helper

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func TestResponse_WriteJSON_Success(t *testing.T) {
	recorder := httptest.NewRecorder()

	respHelper := Response{
		Logger: newDiscardLogger(),
		Writer: recorder,
	}

	respHelper.WriteJSON(context.Background(), http.StatusCreated, nil)

	if recorder.Code != http.StatusCreated {
		t.Errorf("expected: %d, got: %d", http.StatusCreated, recorder.Code)
	}
	if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected: application/json, got: %s", ct)
	}
}

func TestResponse_WriteJSON_SuccessWithValidData(t *testing.T) {
	recorder := httptest.NewRecorder()

	respHelper := Response{
		Logger: newDiscardLogger(),
		Writer: recorder,
	}

	type dummyResponse struct {
		Message string `json:"message"`
	}
	expectedData := dummyResponse{Message: "test-message"}

	respHelper.WriteJSON(context.Background(), http.StatusCreated, expectedData)

	var decoded dummyResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &decoded)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if expectedData != decoded {
		t.Errorf("expected: %s, got: %s", expectedData, decoded)
	}
}

func TestResponse_WriteJSON_FailedToEncodeJSONData(t *testing.T) {
	recorder := httptest.NewRecorder()

	respHelper := Response{
		Logger: newDiscardLogger(),
		Writer: recorder,
	}

	wrongData := struct {
		Wrong chan struct{}
	}{
		Wrong: make(chan struct{}),
	}

	respHelper.WriteJSON(context.Background(), http.StatusCreated, wrongData)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("expected: %d, got: %d", http.StatusInternalServerError, recorder.Code)
	}

	var decoded ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &decoded)
	if err != nil {
		t.Fatalf("failed to decode fallback response: %v", err)
	}
	if decoded.Error != http.StatusText(http.StatusInternalServerError) {
		t.Errorf("expected fallback error: %s, got: %s", http.StatusText(http.StatusInternalServerError), decoded.Error)
	}
}

func TestResponse_WriteErrorJSON_SuccessWithValidData(t *testing.T) {
	recorder := httptest.NewRecorder()

	respHelper := Response{
		Logger: newDiscardLogger(),
		Writer: recorder,
	}

	expectedData := ErrorResponse{Error: "test-error-message"}

	respHelper.WriteErrorJSON(context.Background(), http.StatusCreated, "test-error-message")

	var decoded ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &decoded)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if expectedData != decoded {
		t.Errorf("expected: %s, got: %s", expectedData, decoded)
	}
}
