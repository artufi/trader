package logging

import (
	"context"
	"github.com/artufi/trader/controller/middleware"
	"log/slog"
)

type LogHandler struct {
	slog.Handler
}

func (h LogHandler) Handle(ctx context.Context, r slog.Record) error {
	if connDetails, ok := ctx.Value(middleware.ConnDetailsKey).(middleware.ConnDetails); ok {
		r.Add(TraceIDAttr(connDetails.TraceID.String()))
		r.Add(UserIDAttr(connDetails.UserID))
	}
	return h.Handler.Handle(ctx, r)
}
