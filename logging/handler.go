package logging

import (
	"context"
	"log/slog"
)

type ctxKey string

const (
	logAttrs ctxKey = "logAttrs"
)

type LogHandler struct {
	slog.Handler
}

func (h LogHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(logAttrs).([]slog.Attr); ok {
		for _, v := range attrs {
			r.AddAttrs(v)
		}
	}
	return h.Handler.Handle(ctx, r)
}

func (h LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlerWithAttrs := h.Handler.(*slog.JSONHandler).WithAttrs(attrs)
	return LogHandler{Handler: handlerWithAttrs}
}

// AppendAttrsCtx TODO what if attr already written?
// create newCtx := or make changes here to detect adding the same key
// and override value when detected
func AppendAttrsCtx(parent context.Context, attrs ...slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if a, ok := parent.Value(logAttrs).([]slog.Attr); ok {
		for _, attr := range attrs {
			a = append(a, attr)
		}
		return context.WithValue(parent, logAttrs, a)
	}

	var a []slog.Attr
	for _, attr := range attrs {
		a = append(a, attr)
	}
	return context.WithValue(parent, logAttrs, a)
}
