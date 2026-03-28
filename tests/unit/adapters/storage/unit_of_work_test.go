package storage_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wallissonmarinho/GoTV/internal/adapters/storage"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

func TestCatalog_WithinTx_commit(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "gotv.db")
	c, err := storage.Open("file:" + dbPath + "?_pragma=busy_timeout(5000)")
	require.NoError(t, err)
	defer c.Close()

	ctx := context.Background()
	err = c.WithinTx(ctx, func(ctx context.Context, repo ports.CatalogRepository) error {
		_, e := repo.CreateM3USource(ctx, "http://example.com/a.m3u", "a")
		return e
	})
	require.NoError(t, err)

	list, err := c.ListM3USources(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestCatalog_WithinTx_rollback(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "gotv.db")
	c, err := storage.Open("file:" + dbPath + "?_pragma=busy_timeout(5000)")
	require.NoError(t, err)
	defer c.Close()

	ctx := context.Background()
	boom := errors.New("fail")
	err = c.WithinTx(ctx, func(ctx context.Context, repo ports.CatalogRepository) error {
		_, e := repo.CreateM3USource(ctx, "http://example.com/b.m3u", "b")
		if e != nil {
			return e
		}
		return boom
	})
	require.ErrorIs(t, err, boom)

	list, err := c.ListM3USources(ctx)
	require.NoError(t, err)
	require.Empty(t, list)
}
