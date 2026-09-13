package response

import (
	"errors"
	"testing"
)

func statusWith(requestStatus RequestStatus, message *string) TradeTransactionStatus {
	var tts TradeTransactionStatus
	tts.ReturnData.RequestStatus = requestStatus
	tts.ReturnData.Message = message
	return tts
}

func TestTradeTransactionStatus_CheckRequestStatus(t *testing.T) {
	t.Run("reports no error for a transaction", func(t *testing.T) {
		tests := map[string]RequestStatus{
			"accepted": ACCEPTED,
			"pending":  PENDING,
		}

		for name, requestStatus := range tests {
			t.Run(name, func(t *testing.T) {
				got, err := statusWith(requestStatus, nil).CheckRequestStatus()
				if err != nil {
					t.Fatalf("expected nil error, got: %v", err)
				}
				if got != requestStatus {
					t.Errorf("expected: %v, got: %v", requestStatus, got)
				}
			})
		}
	})

	t.Run("builds an error for a refused transaction", func(t *testing.T) {
		marketClosed := "market closed"
		invalidPrice := "invalid price"
		unknownReason := "unknown reason"

		tests := map[string]struct {
			requestStatus  RequestStatus
			message        *string
			wantStatusName string
			wantMessage    string
		}{
			"rejected without a message":    {REJECTED, nil, "REJECTED", ""},
			"rejected with a message":       {REJECTED, &marketClosed, "REJECTED", "market closed"},
			"error without a message":       {ERROR, nil, "ERROR", ""},
			"error with a message":          {ERROR, &invalidPrice, "ERROR", "invalid price"},
			"status XTB does not document":  {RequestStatus(2), nil, "UNKNOWN", ""},
			"status beyond the known range": {RequestStatus(99), &unknownReason, "UNKNOWN", "unknown reason"},
		}

		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				got, err := statusWith(tc.requestStatus, tc.message).CheckRequestStatus()

				var rsErr *RequestStatusError
				if !errors.As(err, &rsErr) {
					t.Fatalf("expected *RequestStatusError, got: %T (%v)", err, err)
				}
				if got != tc.requestStatus {
					t.Errorf("expected requestStatus: %v, got: %v", tc.requestStatus, got)
				}
				if rsErr.RequestStatus != tc.wantStatusName {
					t.Errorf("expected name: %q, got: %q", tc.wantStatusName, rsErr.RequestStatus)
				}
				if rsErr.Message != tc.wantMessage {
					t.Errorf("expected message: %q, got: %q", tc.wantMessage, rsErr.Message)
				}
			})
		}
	})
}
