package observability

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	logglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// TracerName is the default OpenTelemetry tracer scope name.
const TracerName = "github.com/wallissonmarinho/GoTV"

func otlpConfigured() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("OTEL_SDK_DISABLED")), "true") {
		return false
	}
	for _, k := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT",
	} {
		if strings.TrimSpace(os.Getenv(k)) != "" {
			return true
		}
	}
	return false
}

// Setup configures slog and, when OTLP endpoint environment variables are set,
// registers TracerProvider and LoggerProvider with OTLP HTTP exporters.
// Always returns a slog.Logger that writes JSON to stderr; when OTLP is on,
// the same records are also bridged to OpenTelemetry logs (with trace correlation).
func Setup(ctx context.Context) (shutdown func(context.Context) error, logger *slog.Logger, err error) {
	stderr := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})

	nopShutdown := func(context.Context) error { return nil }

	if !otlpConfigured() {
		return nopShutdown, slog.New(stderr), nil
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithProcess(),
	)
	if err != nil {
		return nil, nil, err
	}

	texp, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(texp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	lexp, err := otlploghttp.New(ctx)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return nil, nil, err
	}
	batch := sdklog.NewBatchProcessor(lexp)
	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(batch),
	)
	logglobal.SetLoggerProvider(lp)

	otelHandler := otelslog.NewHandler("gotv", otelslog.WithLoggerProvider(lp))
	logger = slog.New(newTeeHandler(stderr, otelHandler))

	shutdown = func(shutdownCtx context.Context) error {
		return errors.Join(tp.Shutdown(shutdownCtx), lp.Shutdown(shutdownCtx))
	}
	return shutdown, logger, nil
}

// StartMergeSpan starts a span for merge jobs (scheduler, HTTP rebuild, boot).
func StartMergeSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return otel.Tracer(TracerName).Start(ctx, spanName)
}
