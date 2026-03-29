package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// CatalogAdminService implements ports.CatalogAdmin using a repository and optional unit of work.
type CatalogAdminService struct {
	Repo ports.CatalogRepository
	UoW  ports.UnitOfWork
	Log  ports.AppLog
}

// NewCatalogAdminService returns a service with transactional writes when uow is non-nil.
func NewCatalogAdminService(repo ports.CatalogRepository, uow ports.UnitOfWork, log ports.AppLog) *CatalogAdminService {
	return &CatalogAdminService{Repo: repo, UoW: uow, Log: log}
}

func (s *CatalogAdminService) CreateM3USource(ctx context.Context, url, label string) (*domain.M3USource, error) {
	if err := domain.ValidateM3USourceURL(url); err != nil {
		return nil, err
	}
	var created *domain.M3USource
	err := s.withWritableRepo(ctx, func(ctx context.Context, repo ports.CatalogRepository) error {
		existing, e := repo.FindM3USourceByURL(ctx, url)
		if e != nil {
			return e
		}
		if existing != nil {
			return domain.ErrDuplicateM3USourceURL
		}
		var e2 error
		created, e2 = repo.CreateM3USource(ctx, url, label)
		return e2
	})
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateM3USourceURL) {
			s.Log.Warning(ctx, "m3u source duplicate",
				slog.Any("err", err),
				slog.String("m3u_source.url", url),
				slog.String("m3u_source.label", label),
			)

			return nil, err
		}
		s.Log.Error(ctx, "create m3u source failed", slog.Any("err", err))
		return nil, err
	}
	s.Log.Info(ctx, "m3u source created",
		slog.String("m3u_source.id", created.ID),
		slog.String("m3u_source.url", created.URL),
		slog.String("m3u_source.label", created.Label),
	)
	return created, nil
}

func (s *CatalogAdminService) ListM3USources(ctx context.Context) ([]domain.M3USource, error) {
	return s.Repo.ListM3USources(ctx)
}

func (s *CatalogAdminService) DeleteM3USource(ctx context.Context, id string) error {
	return s.withWritableRepo(ctx, func(ctx context.Context, repo ports.CatalogRepository) error {
		return repo.DeleteM3USource(ctx, id)
	})
}

func (s *CatalogAdminService) CreateEPGSource(ctx context.Context, url, label string) (*domain.EPGSource, error) {
	if err := domain.ValidateEPGSourceURL(url); err != nil {
		return nil, err
	}
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

func (s *CatalogAdminService) DeleteEPGSource(ctx context.Context, id string) error {
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
