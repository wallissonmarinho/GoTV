package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/wallissonmarinho/GoTV/internal/adapters/observability"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// MergeLoop runs merge on a fixed interval (serialized with other callers via MergeRunner.Run).
type MergeLoop struct {
	Merge    ports.MergeRunner
	Interval time.Duration
	Log      *slog.Logger
}

// Run blocks until ctx is cancelled, invoking merge on each tick.
func (l *MergeLoop) Run(ctx context.Context) {
	if l.Log == nil {
		l.Log = slog.Default()
	}
	if l.Interval <= 0 {
		l.Interval = 30 * time.Minute
	}
	t := time.NewTicker(l.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			l.fire()
		}
	}
}

func (l *MergeLoop) fire() {
	defer func() {
		if r := recover(); r != nil {
			l.Log.Error("merge job panic", slog.Any("panic", r))
		}
	}()
	runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	runCtx, span := observability.StartMergeSpan(runCtx, "merge.scheduled")
	defer span.End()
	res := l.Merge.Run(runCtx)
	if len(res.Errors) > 0 {
		l.Log.Log(runCtx, slog.LevelWarn, "merge job warnings",
			slog.String("message", res.Message),
			slog.Any("errors", res.Errors))
	} else {
		l.Log.Log(runCtx, slog.LevelInfo, "merge job ok",
			slog.String("message", res.Message))
	}
}
