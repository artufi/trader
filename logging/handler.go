package logging

import (
	"context"
	"log/slog"
)

type ctxKey string

const logAttrs ctxKey = "logAttrs"

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
	return LogHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h LogHandler) WithGroup(name string) slog.Handler {
	return LogHandler{Handler: h.Handler.WithGroup(name)}
}

func AppendAttrsCtx(ctx context.Context, attrs ...slog.Attr) context.Context {
	if len(attrs) == 0 {
		return ctx
	}
	// Copy into a fresh slice before appending. Appending to the parent's
	// slice could mutate its shared backing array when cap > len,
	// corrupting sibling contexts derived from the same parent (slice aliasing).
	existingAttrs, _ := ctx.Value(logAttrs).([]slog.Attr)
	merged := make([]slog.Attr, 0, len(existingAttrs)+len(attrs))
	merged = append(merged, existingAttrs...)
	// Keep order and do deduplication.
	for _, incomingAttr := range attrs {
		replaced := false
		for i := range merged {
			if merged[i].Key == incomingAttr.Key {
				merged[i] = incomingAttr
				replaced = true
				break
			}
		}
		if !replaced {
			merged = append(merged, incomingAttr)
		}
	}
	return context.WithValue(ctx, logAttrs, merged)
}
