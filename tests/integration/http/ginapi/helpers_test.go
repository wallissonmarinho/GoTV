package ginapi_test

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wallissonmarinho/GoTV/internal/adapters/http/ginapi"
	"github.com/wallissonmarinho/GoTV/internal/adapters/state"
	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// Shared stubs and engine setup for this package. Run: go test ./tests/integration/http/ginapi -v

type stubMerge struct{}

func (stubMerge) Run(ctx context.Context) domain.MergeResult {
	now := time.Now()
	return domain.MergeResult{OK: true, Message: "stub", StartedAt: now, FinishedAt: now}
}

type stubCatalog struct{}

func (stubCatalog) CreateM3USource(ctx context.Context, url, label string) (*domain.M3USource, error) {
	return &domain.M3USource{ID: 1, URL: url, Label: label}, nil
}

func (stubCatalog) ListM3USources(ctx context.Context) ([]domain.M3USource, error) {
	return []domain.M3USource{{ID: 10, URL: "http://list/m3u", Label: "L"}}, nil
}

func (stubCatalog) DeleteM3USource(ctx context.Context, id int64) error { return nil }

func (stubCatalog) CreateEPGSource(ctx context.Context, url, label string) (*domain.EPGSource, error) {
	return &domain.EPGSource{ID: 2, URL: url, Label: label}, nil
}

func (stubCatalog) ListEPGSources(ctx context.Context) ([]domain.EPGSource, error) {
	return []domain.EPGSource{{ID: 20, URL: "http://list/epg", Label: "E"}}, nil
}

func (stubCatalog) DeleteEPGSource(ctx context.Context, id int64) error { return nil }

func (stubCatalog) LoadMergeStatus(ctx context.Context) (domain.MergeSnapshot, error) {
	return domain.MergeSnapshot{OK: true, Message: "merged"}, nil
}

func testDeps(store *state.MemoryStore) ginapi.Deps {
	return ginapi.Deps{
		Merge:   stubMerge{},
		Catalog: stubCatalog{},
		Store:   store,
		Log:     nil,
	}
}

func depsWithCatalog(catalog ports.CatalogAdmin, store *state.MemoryStore) ginapi.Deps {
	if store == nil {
		store = &state.MemoryStore{}
	}
	return ginapi.Deps{
		Merge:   stubMerge{},
		Catalog: catalog,
		Store:   store,
		Log:     nil,
	}
}

func newTestEngine(cfg ginapi.Config, d ginapi.Deps) *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	ginapi.Register(e, cfg, d)
	return e
}
