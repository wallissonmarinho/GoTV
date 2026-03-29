package ginapi_test

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wallissonmarinho/GoTV/internal/adapters/http/ginapi"
	"github.com/wallissonmarinho/GoTV/internal/adapters/state"
	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
	"github.com/wallissonmarinho/GoTV/internal/core/services"
)

// Shared stubs and engine setup for this package. Run: go test ./tests/integration/http/ginapi -v

type stubMerge struct{}

func (stubMerge) Run(ctx context.Context) domain.MergeResult {
	now := time.Now()
	return domain.MergeResult{OK: true, Message: "stub", StartedAt: now, FinishedAt: now}
}

type stubCatalog struct{}

func (*stubCatalog) FindM3USourceByURL(ctx context.Context, url string) (*domain.M3USource, error) {
	return nil, nil
}

func (*stubCatalog) CreateM3USource(ctx context.Context, url, label string) (*domain.M3USource, error) {
	return &domain.M3USource{ID: "00000000-0000-4000-8000-000000000001", URL: url, Label: label}, nil
}

func (*stubCatalog) ListM3USources(ctx context.Context) ([]domain.M3USource, error) {
	return []domain.M3USource{{ID: "00000000-0000-4000-8000-00000000000a", URL: "http://list/feed.m3u", Label: "L"}}, nil
}

func (*stubCatalog) DeleteM3USource(ctx context.Context, id string) error { return nil }

func (*stubCatalog) CreateEPGSource(ctx context.Context, url, label string) (*domain.EPGSource, error) {
	return &domain.EPGSource{ID: "00000000-0000-4000-8000-000000000002", URL: url, Label: label}, nil
}

func (*stubCatalog) ListEPGSources(ctx context.Context) ([]domain.EPGSource, error) {
	return []domain.EPGSource{{ID: "00000000-0000-4000-8000-000000000014", URL: "http://list/guide.xml", Label: "E"}}, nil
}

func (*stubCatalog) DeleteEPGSource(ctx context.Context, id string) error { return nil }

func (*stubCatalog) SaveSnapshot(ctx context.Context, snap domain.MergeSnapshot) error { return nil }

func (*stubCatalog) LoadSnapshot(ctx context.Context) (domain.MergeSnapshot, error) {
	return domain.MergeSnapshot{OK: true, Message: "merged"}, nil
}

func testDeps(store *state.MemoryStore) ginapi.Deps {
	stub := &stubCatalog{}
	admin := services.NewCatalogAdminService(stub, nil, ports.NoopAppLog{})
	return ginapi.Deps{
		Merge:   stubMerge{},
		Catalog: admin,
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
