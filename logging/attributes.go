package logging

import (
	"github.com/artufi/trader/model"
	"log/slog"
)

func AttemptAttr(attempt int) slog.Attr {
	return slog.Int("attempt", attempt)
}

func URLAttr(url string) slog.Attr {
	return slog.String("URL", url)
}

func ErrorAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

func UserIDAttr(userID string) slog.Attr {
	return slog.String("userID", userID)
}

func ConnNo(connNumber int) slog.Attr {
	return slog.Int("connNo", connNumber)
}

func TraceIDAttr(traceID string) slog.Attr {
	return slog.String("traceID", traceID)
}

func MsgAttr(msg string) slog.Attr {
	return slog.String("msg", msg)
}

func SymbolAttr(symbol string) slog.Attr {
	return slog.String("symbol", symbol)
}

func ServiceName(service string) slog.Attr {
	return slog.String("service", service)
}

func PredictionDetailsAttr(predictionDetails model.PredictionDetails) slog.Attr {
	return slog.Any("predictionDetails", predictionDetails)
}

func RespAttr(resp interface{}) slog.Attr {
	return slog.Any("resp", resp)
}

func IDAttr(id int) slog.Attr {
	return slog.Int("id", id)
}

func StreamID(ssid string) slog.Attr {
	return slog.String("streamID", ssid)
}
