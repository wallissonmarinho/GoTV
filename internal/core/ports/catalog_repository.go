package ports

import (
	"context"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
)

// CatalogRepository persists M3U/EPG source URLs and the last merged snapshot.
type CatalogRepository interface {
	CreateM3USource(ctx context.Context, url, label string) (*domain.M3USource, error)
	ListM3USources(ctx context.Context) ([]domain.M3USource, error)
	DeleteM3USource(ctx context.Context, id int64) error

	CreateEPGSource(ctx context.Context, url, label string) (*domain.EPGSource, error)
	ListEPGSources(ctx context.Context) ([]domain.EPGSource, error)
	DeleteEPGSource(ctx context.Context, id int64) error

	SaveSnapshot(ctx context.Context, snap domain.MergeSnapshot) error
	LoadSnapshot(ctx context.Context) (domain.MergeSnapshot, error)
}
