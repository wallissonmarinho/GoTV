package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	ginapi "github.com/wallissonmarinho/GoTV/internal/adapters/http/ginapi"
	"github.com/wallissonmarinho/GoTV/internal/adapters/observability"
	"github.com/wallissonmarinho/GoTV/internal/adapters/scheduler"
	"github.com/wallissonmarinho/GoTV/internal/adapters/state"
	"github.com/wallissonmarinho/GoTV/internal/app"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "migrate" {
		os.Exit(runMigrateCLI(os.Args[2:]))
	}

	otelShutdown, lg, err := observability.Setup(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	slog.SetDefault(lg)

	dataDir := getenv("GOTV_DATA_DIR", "./data")
	if mkErr := os.MkdirAll(dataDir, 0o755); mkErr != nil {
		slog.Error("data dir", slog.Any("err", mkErr))
		os.Exit(1)
	}
	dbPath := filepath.Join(dataDir, "gotv.db")
	sqliteDSN := getenv("GOTV_SQLITE_DSN", "file:"+dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = sqliteDSN
	}

	cat, err := app.OpenCatalog(dsn)
	if err != nil {
		slog.Error("open catalog", slog.Any("err", err))
		os.Exit(1)
	}
	defer cat.Close()

	mem := &state.MemoryStore{}
	app.HydrateStore(context.Background(), cat, mem)

	merge := app.NewPlaylistMergeService(cat, mem, app.MergeRuntimeOptions{
		HTTPTimeout:  durationEnv("GOTV_HTTP_TIMEOUT", 45*time.Second),
		MaxBodyBytes: int64Env("GOTV_MAX_BODY_BYTES", 50<<20),
		UserAgent:    getenv("GOTV_USER_AGENT", "GoTV/1.0"),
		ProbeTimeout: durationEnv("GOTV_PROBE_TIMEOUT", 8*time.Second),
		ProbeStreams: boolEnv("GOTV_PROBE_STREAMS", false),
		MaxProbes:    intEnv("GOTV_MAX_PROBES", 500),
		ProbeWorkers: intEnv("GOTV_PROBE_WORKERS", 32),
	})
	catalogAdmin := app.NewCatalogAdmin(cat)

	if getenv("GIN_MODE", "") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	serviceName := getenv("OTEL_SERVICE_NAME", "gotv")
	observability.RegisterGin(engine, serviceName)
	ginapi.Register(engine, ginapi.Config{AdminAPIKey: app.AdminAPIKey()}, ginapi.Deps{
		Merge:   merge,
		Catalog: catalogAdmin,
		Store:   mem,
		Log:     lg,
	})

	addr := listenAddr()

	srv := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       0,
		WriteTimeout:      0,
	}

	mergeInterval := durationEnv("GOTV_MERGE_INTERVAL", 30*time.Minute)
	schedCtx, schedCancel := context.WithCancel(context.Background())
	defer schedCancel()
	loop := &scheduler.MergeLoop{Merge: merge, Interval: mergeInterval, Log: lg}
	go loop.Run(schedCtx)

	go func() {
		slog.Info("listening", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	// Optional first run shortly after boot (does not block listen).
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("initial merge panic", slog.Any("panic", r))
			}
		}()
		time.Sleep(2 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		ctx, span := observability.StartMergeSpan(ctx, "merge.initial")
		defer span.End()
		res := merge.Run(ctx)
		if len(res.Errors) > 0 {
			slog.Log(ctx, slog.LevelWarn, "initial merge warnings",
				slog.Any("errors", res.Errors))
		}
		slog.Log(ctx, slog.LevelInfo, "initial merge",
			slog.String("message", res.Message))
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	slog.Info("shutting down")
	schedCancel()
	shCtx, shCancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer shCancel()
	_ = srv.Shutdown(shCtx)
	_ = otelShutdown(shCtx)
}

func listenAddr() string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return getenv("GOTV_ADDR", ":8080")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func boolEnv(k string, def bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func intEnv(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func int64Env(k string, def int64) int64 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func durationEnv(k string, def time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
