package observability

import (
	"context"
	"log/slog"
)

// teeHandler sends each record to two handlers (e.g. stderr JSON + OTLP).
type teeHandler struct {
	a, b slog.Handler
}

func newTeeHandler(a, b slog.Handler) slog.Handler {
	return &teeHandler{a: a, b: b}
}

func (t *teeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return t.a.Enabled(ctx, level) || t.b.Enabled(ctx, level)
}

func (t *teeHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	if t.a.Enabled(ctx, r.Level) {
		if err := t.a.Handle(ctx, r); err != nil {
			firstErr = err
		}
	}
	if t.b.Enabled(ctx, r.Level) {
		if err := t.b.Handle(ctx, r); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (t *teeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return newTeeHandler(t.a.WithAttrs(attrs), t.b.WithAttrs(attrs))
}

func (t *teeHandler) WithGroup(name string) slog.Handler {
	return newTeeHandler(t.a.WithGroup(name), t.b.WithGroup(name))
}
