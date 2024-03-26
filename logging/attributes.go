package logging

import (
	"github.com/artufi/trader/model"
	"log/slog"
)

func URLAttr(url string) slog.Attr {
	return slog.Any("URL", url)
}

func ErrorAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

func UserIDAttr(userID string) slog.Attr {
	return slog.Any("userID", userID)
}

func ClientID(clientID string) slog.Attr {
	return slog.Any("clientID", clientID)
}

func TraceIDAttr(traceID string) slog.Attr {
	return slog.Any("traceID", traceID)
}

func MsgAttr(msg string) slog.Attr {
	return slog.Any("msg", msg)
}

func SymbolAttr(symbol string) slog.Attr {
	return slog.Any("symbol", symbol)
}

func PredictionDetailsAttr(predictionDetails model.PredictionDetails) slog.Attr {
	return slog.Any("predictionDetails", predictionDetails)
}

func RespAttr(resp interface{}) slog.Attr {
	return slog.Any("resp", resp)
}

func IDAttr(id int) slog.Attr {
	return slog.Any("id", id)
}
