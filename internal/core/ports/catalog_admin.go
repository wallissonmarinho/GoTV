package ports

import (
	"context"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
)

// CatalogAdmin manages M3U/EPG source registration and exposes persisted merge status.
// Implementations encapsulate whether work runs inside UnitOfWork transactions.
type CatalogAdmin interface {
	CreateM3USource(ctx context.Context, url, label string) (*domain.M3USource, error)
	ListM3USources(ctx context.Context) ([]domain.M3USource, error)
	DeleteM3USource(ctx context.Context, id int64) error

	CreateEPGSource(ctx context.Context, url, label string) (*domain.EPGSource, error)
	ListEPGSources(ctx context.Context) ([]domain.EPGSource, error)
	DeleteEPGSource(ctx context.Context, id int64) error

	LoadMergeStatus(ctx context.Context) (domain.MergeSnapshot, error)
}
