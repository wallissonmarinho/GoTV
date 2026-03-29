package ports

import (
	"context"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
)

// CatalogRepository persists M3U/EPG source URLs and the last merged snapshot.
type CatalogRepository interface {
	// FindM3USourceByURL returns the source when url matches exactly; nil, nil if none.
	FindM3USourceByURL(ctx context.Context, url string) (*domain.M3USource, error)
	CreateM3USource(ctx context.Context, url, label string) (*domain.M3USource, error)
	ListM3USources(ctx context.Context) ([]domain.M3USource, error)
	DeleteM3USource(ctx context.Context, id string) error

	CreateEPGSource(ctx context.Context, url, label string) (*domain.EPGSource, error)
	ListEPGSources(ctx context.Context) ([]domain.EPGSource, error)
	DeleteEPGSource(ctx context.Context, id string) error

	SaveSnapshot(ctx context.Context, snap domain.MergeSnapshot) error
	LoadSnapshot(ctx context.Context) (domain.MergeSnapshot, error)
}
