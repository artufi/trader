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

func ClientIPAttr(clientIP string) slog.Attr {
	return slog.Any("clientIP", clientIP)
}

func UserIDAttr(userID string) slog.Attr {
	return slog.Any("userID", userID)
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

func PurchaseInstrAttr(purchaseInstr model.PurchaseInstruction) slog.Attr {
	return slog.Any("purchaseInstr", purchaseInstr)
}

func RespAttr(resp interface{}) slog.Attr {
	return slog.Any("resp", resp)
}
