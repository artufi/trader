package logging

import (
	"context"
	"log/slog"
	"trader/controller/middleware"
)

type LogHandler struct {
	slog.Handler
}

func (h LogHandler) Handle(ctx context.Context, r slog.Record) error {
	if connDetails, ok := ctx.Value(middleware.ConnDetailsKey).(middleware.ConnDetails); ok {
		r.Add(TraceIDAttr(connDetails.TraceID.String()))
		r.Add(ClientIPAttr(connDetails.IPAddr))
		r.Add(UserIDAttr(connDetails.UserID))
	}
	return h.Handler.Handle(ctx, r)
}
