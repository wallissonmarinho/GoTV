package observability

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// SlogTrace implements ports.AppLog: slog + OpenTelemetry span event on the current span.
type SlogTrace struct{}

var _ ports.AppLog = SlogTrace{}

func (SlogTrace) Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
	addSpanEvent(ctx, msg, attrs, attribute.String("log.severity", "info"))
}

func (SlogTrace) Warning(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
	addSpanEvent(ctx, msg, attrs, attribute.String("log.severity", "warn"))
}

func (SlogTrace) Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelError, msg, attrs...)
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	addSpanEvent(ctx, msg, attrs, attribute.String("log.severity", "error"))
	span.SetStatus(codes.Error, msg)
	if e, ok := errFromAttrs(attrs); ok {
		span.RecordError(e)
	}
}

func errFromAttrs(attrs []slog.Attr) (error, bool) {
	for _, a := range attrs {
		if a.Key != "err" {
			continue
		}
		if a.Value.Kind() != slog.KindAny {
			return nil, false
		}
		if e, ok := a.Value.Any().(error); ok && e != nil {
			return e, true
		}
		return nil, false
	}
	return nil, false
}

func addSpanEvent(ctx context.Context, msg string, attrs []slog.Attr, extra ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}
	otel := make([]attribute.KeyValue, 0, len(extra)+len(attrs))
	otel = append(otel, extra...)
	for _, a := range attrs {
		otel = append(otel, slogAttrToAttribute(a))
	}
	span.AddEvent(msg, trace.WithAttributes(otel...))
}

func slogAttrToAttribute(a slog.Attr) attribute.KeyValue {
	switch a.Value.Kind() {
	case slog.KindGroup:
		return attribute.String(a.Key, a.Value.String())
	case slog.KindString:
		return attribute.String(a.Key, a.Value.String())
	case slog.KindInt64:
		return attribute.Int64(a.Key, a.Value.Int64())
	case slog.KindUint64:
		return attribute.String(a.Key, strconv.FormatUint(a.Value.Uint64(), 10))
	case slog.KindFloat64:
		return attribute.Float64(a.Key, a.Value.Float64())
	case slog.KindBool:
		return attribute.Bool(a.Key, a.Value.Bool())
	case slog.KindDuration:
		return attribute.String(a.Key, a.Value.Duration().String())
	case slog.KindTime:
		return attribute.String(a.Key, a.Value.Time().String())
	case slog.KindAny:
		return attribute.String(a.Key, fmt.Sprint(a.Value.Any()))
	default:
		return attribute.String(a.Key, a.Value.String())
	}
}
