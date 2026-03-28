package storage

import (
	"context"
	"database/sql"

	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

var _ ports.UnitOfWork = (*Catalog)(nil)

// WithinTx runs fn with a CatalogRepository backed by a single transaction.
// On error or panic, the transaction is rolled back; on success it commits.
func (c *Catalog) WithinTx(ctx context.Context, fn func(ctx context.Context, repo ports.CatalogRepository) error) (err error) {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	repo := &catalogRepo{ex: tx, pg: c.pg}
	err = fn(ctx, repo)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}
