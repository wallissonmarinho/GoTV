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

func otlpDisabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("OTEL_SDK_DISABLED")), "true")
}

// otlpTracesConfigured reports whether OTLP trace export should run (HTTP exporter reads OTEL_* env).
func otlpTracesConfigured() bool {
	if otlpDisabled() {
		return false
	}
	return strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")) != "" ||
		strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")) != ""
}

// otlpLogsConfigured reports whether OTLP log export should run.
// Logs are opt-in via OTEL_EXPORTER_OTLP_LOGS_ENDPOINT only: backends like Jaeger accept traces
// at :4318 but return 404 on /v1/logs if we derive logs from the generic OTLP endpoint.
func otlpLogsConfigured() bool {
	if otlpDisabled() {
		return false
	}
	return strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT")) != ""
}

func otlpConfigured() bool {
	return otlpTracesConfigured() || otlpLogsConfigured()
}

// Setup configures slog and, when OTLP environment variables are set, registers exporters.
// Traces: OTEL_EXPORTER_OTLP_ENDPOINT and/or OTEL_EXPORTER_OTLP_TRACES_ENDPOINT.
// Logs: only OTEL_EXPORTER_OTLP_LOGS_ENDPOINT (optional; use a collector that supports OTLP logs).
// Always emits JSON logs to stderr; with OTLP logs on, the same records are bridged to OTel logs.
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

	var shutdownFns []func(context.Context) error

	if otlpTracesConfigured() {
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
		shutdownFns = append(shutdownFns, tp.Shutdown)
	}

	logger = slog.New(stderr)
	if otlpLogsConfigured() {
		lexp, err := otlploghttp.New(ctx)
		if err != nil {
			for _, fn := range shutdownFns {
				_ = fn(ctx)
			}
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
		shutdownFns = append(shutdownFns, lp.Shutdown)
	}

	shutdown = func(shutdownCtx context.Context) error {
		var errs []error
		for _, fn := range shutdownFns {
			errs = append(errs, fn(shutdownCtx))
		}
		return errors.Join(errs...)
	}
	return shutdown, logger, nil
}

// StartMergeSpan starts a span for merge jobs (scheduler, HTTP rebuild, boot).
func StartMergeSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return otel.Tracer(TracerName).Start(ctx, spanName)
}
