package ports

import (
	"context"
	"log/slog"
)

// AppLog is a driven port: structured logs plus span events when a trace is active (Jaeger, etc.).
type AppLog interface {
	Info(ctx context.Context, msg string, attrs ...slog.Attr)
	Warning(ctx context.Context, msg string, attrs ...slog.Attr)
	Error(ctx context.Context, msg string, attrs ...slog.Attr)
}

// NoopAppLog satisfies AppLog and discards nothing (tests / optional wiring).
type NoopAppLog struct{}

func (NoopAppLog) Info(context.Context, string, ...slog.Attr)    {}
func (NoopAppLog) Warning(context.Context, string, ...slog.Attr) {}
func (NoopAppLog) Error(context.Context, string, ...slog.Attr)    {}
