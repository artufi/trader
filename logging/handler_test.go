package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestLogHandler(t *testing.T) {
	newLogger := func(buffer *bytes.Buffer) *slog.Logger {
		return slog.New(LogHandler{slog.NewJSONHandler(buffer, nil)})
	}

	t.Run("adds context attributes to the Record", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newLogger(&buf)

		ctx := AppendAttrsCtx(context.Background(), ConnNo(1))

		logger.InfoContext(ctx, "test message")

		if !strings.Contains(buf.String(), `"connNo":1`) {
			t.Errorf("missing: connNo in output, got: %s", buf.String())
		}
	})

	t.Run("preserves LogHandler after WithAttrs", func(t *testing.T) {
		var buf bytes.Buffer
		expectedAttr := "staticAttr"
		// goes through WithAttrs internally
		logger := newLogger(&buf).With(expectedAttr, "t")

		ctx := AppendAttrsCtx(context.Background(), TraceIDAttr("abc"))
		logger.InfoContext(ctx, "test message")

		out := buf.String()
		if !strings.Contains(out, `"staticAttr":"t"`) {
			t.Errorf("missing: %s, got: %s", expectedAttr, out)
		}
		if !strings.Contains(out, `"traceID":"abc"`) {
			t.Errorf("LogHandler wrapper lost after WithAttrs, got: %s", out)
		}
	})

	t.Run("preserves LogHandler after WithGroup", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newLogger(&buf).WithGroup("group")

		ctx := AppendAttrsCtx(context.Background(), TraceIDAttr("abc"))
		logger.InfoContext(ctx, "test message")

		if !strings.Contains(buf.String(), `"group":{"traceID":"abc"}`) {
			t.Errorf("LogHandler wrapper lost after WithGroup, got: %s", buf.String())
		}
	})
}

func TestAppendAttrsCtx(t *testing.T) {
	t.Run("returns the same context when no attributes provided", func(t *testing.T) {
		parent := context.Background()
		child := AppendAttrsCtx(parent)

		if parent != child {
			t.Errorf("expected parent: %v to be equal to child: %v", parent, child)
		}
	})

	t.Run("adds attributes", func(t *testing.T) {
		traceIDAttr := TraceIDAttr("traceIDTest")
		userIDAttr := UserIDAttr("userIDTest")
		ctx := AppendAttrsCtx(context.Background(), traceIDAttr, userIDAttr)

		attrs := attrsOf(ctx)

		expected := []slog.Attr{traceIDAttr, userIDAttr}
		if len(attrs) != len(expected) {
			t.Fatalf("expected length: %d, got: %d", len(expected), len(attrs))
		}
		for _, e := range expected {
			if !containsAttr(attrs, e) {
				t.Errorf("missing expected attr: %v in: %v", e, attrs)
			}
		}
	})

	t.Run("keeps the last value on duplicate keys", func(t *testing.T) {
		oldAttr := TraceIDAttr("oldValue")
		newAttr := TraceIDAttr("newValue")
		ctx := AppendAttrsCtx(context.Background(), oldAttr, newAttr)

		attrs := attrsOf(ctx)

		if len(attrs) != 1 {
			t.Fatalf("expected length: 1, got: %d", len(attrs))
		}
		if containsAttr(attrs, oldAttr) {
			t.Errorf("unexpected old attr still present: %v, got: %v", oldAttr, attrs)
		}
		if !containsAttr(attrs, newAttr) {
			t.Errorf("missing expected attr: %v in: %v", newAttr, attrs)
		}
	})

	t.Run("keeps order on override", func(t *testing.T) {
		ctx := AppendAttrsCtx(context.Background(), TraceIDAttr("trace"), UserIDAttr("user1"), ConnNo(1))
		updatedUser := UserIDAttr("user2")
		ctx = AppendAttrsCtx(ctx, updatedUser)

		attrs := attrsOf(ctx)

		expectedOrder := []string{"traceID", "userID", "connNo"}
		for i, key := range expectedOrder {
			if attrs[i].Key != key {
				t.Errorf("index: %d, expected: %s, got: %s", i, key, attrs[i].Key)
			}
		}
		if !containsAttr(attrs, updatedUser) {
			t.Errorf("missing expected updated attr: %v in: %v", updatedUser, attrs)
		}
	})

	t.Run("does not mutate parent slice", func(t *testing.T) {
		base := make([]slog.Attr, 2, 4)
		base[0] = TraceIDAttr("trace")
		base[1] = UserIDAttr("user")
		parent := context.WithValue(context.Background(), logAttrs, base)

		if cap(base) <= len(base) {
			t.Fatalf("test setup invalid: need cap > len, got len=%d cap=%d", len(base), cap(base))
		}

		_ = AppendAttrsCtx(parent, ConnNo(1))

		if got := len(attrsOf(parent)); got != 2 {
			t.Errorf("parent slice modified, expected length: 2, got: %d", got)
		}
	})

	t.Run("isolates sibling contexts", func(t *testing.T) {
		base := make([]slog.Attr, 2, 4)
		base[0] = TraceIDAttr("trace")
		base[1] = UserIDAttr("user")
		parent := context.WithValue(context.Background(), logAttrs, base)

		if cap(base) <= len(base) {
			t.Fatalf("test setup invalid: need cap > len, got len=%d cap=%d", len(base), cap(base))
		}

		child1 := AppendAttrsCtx(parent, ConnNo(1))
		_ = AppendAttrsCtx(parent, ConnNo(2))

		child1Attrs := attrsOf(child1)
		if containsAttr(child1Attrs, ConnNo(2)) {
			t.Errorf("child1 contains connNo=2 set by sibling child2 (slice aliasing), got: %v", child1Attrs)
		}
	})
}

func attrsOf(ctx context.Context) []slog.Attr {
	return ctx.Value(logAttrs).([]slog.Attr)
}

func containsAttr(attrs []slog.Attr, attr slog.Attr) bool {
	for _, a := range attrs {
		if a.Equal(attr) {
			return true
		}
	}
	return false
}
