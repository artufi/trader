package logging

import "log/slog"

func URLAttr(url string) slog.Attr {
	return slog.Any("URL", url)
}

func ErrorAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

func ClientIPAttr(ip string) slog.Attr {
	return slog.Any("client_ip", ip)
}

func UserIDAttr(id string) slog.Attr {
	return slog.Any("user_id", id)
}
