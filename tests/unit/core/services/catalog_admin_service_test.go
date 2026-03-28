package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
	svc "github.com/wallissonmarinho/GoTV/internal/core/services"
)

type fakeCatalogRepo struct {
	m3u     []domain.M3USource
	epg     []domain.EPGSource
	snap    domain.MergeSnapshot
	txCalls int
}

func (f *fakeCatalogRepo) CreateM3USource(ctx context.Context, url, label string) (*domain.M3USource, error) {
	ch := domain.M3USource{ID: 1, URL: url, Label: label, CreatedAt: time.Unix(1, 0)}
	f.m3u = append(f.m3u, ch)
	return &ch, nil
}

func (f *fakeCatalogRepo) ListM3USources(ctx context.Context) ([]domain.M3USource, error) {
	return f.m3u, nil
}

func (f *fakeCatalogRepo) DeleteM3USource(ctx context.Context, id int64) error {
	return nil
}

func (f *fakeCatalogRepo) CreateEPGSource(ctx context.Context, url, label string) (*domain.EPGSource, error) {
	ch := domain.EPGSource{ID: 1, URL: url, Label: label, CreatedAt: time.Unix(1, 0)}
	f.epg = append(f.epg, ch)
	return &ch, nil
}

func (f *fakeCatalogRepo) ListEPGSources(ctx context.Context) ([]domain.EPGSource, error) {
	return f.epg, nil
}

func (f *fakeCatalogRepo) DeleteEPGSource(ctx context.Context, id int64) error {
	return nil
}

func (f *fakeCatalogRepo) SaveSnapshot(ctx context.Context, snap domain.MergeSnapshot) error {
	return nil
}

func (f *fakeCatalogRepo) LoadSnapshot(ctx context.Context) (domain.MergeSnapshot, error) {
	return f.snap, nil
}

type txUoW struct {
	repo ports.CatalogRepository
}

func (t *txUoW) WithinTx(ctx context.Context, fn func(ctx context.Context, repo ports.CatalogRepository) error) error {
	if r, ok := t.repo.(*fakeCatalogRepo); ok {
		r.txCalls++
	}
	return fn(ctx, t.repo)
}

func TestCatalogAdminService_writesWithoutTxWhenUoWNil(t *testing.T) {
	repo := &fakeCatalogRepo{}
	s := svc.NewCatalogAdminService(repo, nil)
	ctx := context.Background()

	_, err := s.CreateM3USource(ctx, "http://a", "A")
	require.NoError(t, err)
	require.Equal(t, 0, repo.txCalls)
}

func TestCatalogAdminService_writesUseUoWWhenSet(t *testing.T) {
	repo := &fakeCatalogRepo{}
	uow := &txUoW{repo: repo}
	s := svc.NewCatalogAdminService(repo, uow)
	ctx := context.Background()

	_, err := s.CreateM3USource(ctx, "http://a", "A")
	require.NoError(t, err)
	require.Equal(t, 1, repo.txCalls)
}

func TestCatalogAdminService_listUsesRepoDirectly(t *testing.T) {
	repo := &fakeCatalogRepo{}
	uow := &txUoW{repo: repo}
	s := svc.NewCatalogAdminService(repo, uow)
	ctx := context.Background()

	_, err := s.ListM3USources(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, repo.txCalls, "read path should not open a transaction")
}

func TestCatalogAdminService_LoadMergeStatus(t *testing.T) {
	repo := &fakeCatalogRepo{snap: domain.MergeSnapshot{OK: true, Message: "ok"}}
	s := svc.NewCatalogAdminService(repo, nil)
	snap, err := s.LoadMergeStatus(context.Background())
	require.NoError(t, err)
	require.True(t, snap.OK)
	require.Equal(t, "ok", snap.Message)
}
