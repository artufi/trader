package logging

import (
	"context"
	"log/slog"
	"testing"
)

func TestAppendAttrsCtx(t *testing.T) {
	ctx := context.Background()
	traceIDAttr := TraceIDAttr("traceIDTest")
	userIDAttr := UserIDAttr("userIDTest")
	appendCtx := AppendAttrsCtx(ctx, traceIDAttr, userIDAttr)

	expectedAttrs := []slog.Attr{traceIDAttr, userIDAttr}

	appendedAttrs := appendCtx.Value(logAttrs).([]slog.Attr)

	if len(appendedAttrs) != len(expectedAttrs) {
		t.Errorf("expected %d slog.Attrs, got: %d", len(expectedAttrs), len(appendedAttrs))
	}

	for _, expectedAttr := range expectedAttrs {
		if !containsAttr(appendedAttrs, expectedAttr) {
			t.Errorf("not found expected attr: %v", expectedAttr)
		}
	}
}

func containsAttr(attrs []slog.Attr, attr slog.Attr) bool {
	for _, a := range attrs {
		if a.Equal(attr) {
			return true
		}
	}
	return false
}
