package ports

import "context"

// UnitOfWork runs a block of repository work inside a single database transaction.
type UnitOfWork interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, repo CatalogRepository) error) error
}
