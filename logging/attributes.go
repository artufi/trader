package logging

import (
	"github.com/google/uuid"
	"log/slog"
)

func URLAttr(url string) slog.Attr {
	return slog.Any("URL", url)
}

func ErrorAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

func ClientIPAttr(clientIP string) slog.Attr {
	return slog.Any("clientIP", clientIP)
}

func UserIDAttr(userID string) slog.Attr {
	return slog.Any("userID", userID)
}

func TransactionIDAttr(transactionID uuid.UUID) slog.Attr {
	return slog.Any("transactionID", transactionID)
}
