package services

import (
	"context"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// CatalogAdminService implements ports.CatalogAdmin using a repository and optional unit of work.
type CatalogAdminService struct {
	Repo ports.CatalogRepository
	UoW  ports.UnitOfWork
}

// NewCatalogAdminService returns a service with transactional writes when uow is non-nil.
func NewCatalogAdminService(repo ports.CatalogRepository, uow ports.UnitOfWork) *CatalogAdminService {
	return &CatalogAdminService{Repo: repo, UoW: uow}
}

func (s *CatalogAdminService) CreateM3USource(ctx context.Context, url, label string) (*domain.M3USource, error) {
	var created *domain.M3USource
	err := s.withWritableRepo(ctx, func(ctx context.Context, repo ports.CatalogRepository) error {
		var e error
		created, e = repo.CreateM3USource(ctx, url, label)
		return e
	})
	return created, err
}

func (s *CatalogAdminService) ListM3USources(ctx context.Context) ([]domain.M3USource, error) {
	return s.Repo.ListM3USources(ctx)
}

func (s *CatalogAdminService) DeleteM3USource(ctx context.Context, id int64) error {
	return s.withWritableRepo(ctx, func(ctx context.Context, repo ports.CatalogRepository) error {
		return repo.DeleteM3USource(ctx, id)
	})
}

func (s *CatalogAdminService) CreateEPGSource(ctx context.Context, url, label string) (*domain.EPGSource, error) {
	var created *domain.EPGSource
	err := s.withWritableRepo(ctx, func(ctx context.Context, repo ports.CatalogRepository) error {
		var e error
		created, e = repo.CreateEPGSource(ctx, url, label)
		return e
	})
	return created, err
}

func (s *CatalogAdminService) ListEPGSources(ctx context.Context) ([]domain.EPGSource, error) {
	return s.Repo.ListEPGSources(ctx)
}

func (s *CatalogAdminService) DeleteEPGSource(ctx context.Context, id int64) error {
	return s.withWritableRepo(ctx, func(ctx context.Context, repo ports.CatalogRepository) error {
		return repo.DeleteEPGSource(ctx, id)
	})
}

func (s *CatalogAdminService) LoadMergeStatus(ctx context.Context) (domain.MergeSnapshot, error) {
	return s.Repo.LoadSnapshot(ctx)
}

func (s *CatalogAdminService) withWritableRepo(ctx context.Context, fn func(ctx context.Context, repo ports.CatalogRepository) error) error {
	if s.UoW != nil {
		return s.UoW.WithinTx(ctx, fn)
	}
	return fn(ctx, s.Repo)
}

var _ ports.CatalogAdmin = (*CatalogAdminService)(nil)
