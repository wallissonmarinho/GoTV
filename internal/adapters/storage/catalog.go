package storage

import (
	"context"
	"database/sql"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
	persistmigrate "github.com/wallissonmarinho/GoTV/internal/persistence/migrate"
)

// Catalog implements ports.CatalogRepository and ports.UnitOfWork for SQLite or PostgreSQL.
type Catalog struct {
	db *sql.DB
	pg bool
}

// OpenDB opens and pings a database handle without running migrations.
func OpenDB(dsn string) (*sql.DB, bool, error) {
	dsn = strings.TrimSpace(dsn)
	var (
		db  *sql.DB
		err error
		pg  bool
	)
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		db, err = sql.Open("pgx", dsn)
		pg = true
	} else {
		db, err = sql.Open("sqlite", dsn)
	}
	if err != nil {
		return nil, false, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, false, err
	}
	return db, pg, nil
}

// Open connects using a DSN. Use postgres:// or postgresql:// for Postgres;
// otherwise opens as SQLite (e.g. file:path?_pragma=busy_timeout(5000)).
func Open(dsn string) (*Catalog, error) {
	db, pg, err := OpenDB(dsn)
	if err != nil {
		return nil, err
	}
	if err := persistmigrate.RunMigrations(context.Background(), db, pg); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Catalog{db: db, pg: pg}, nil
}

// Close releases the DB handle.
func (c *Catalog) Close() error {
	return c.db.Close()
}

func (c *Catalog) base() *catalogRepo {
	return &catalogRepo{ex: c.db, pg: c.pg}
}

func (c *Catalog) FindM3USourceByURL(ctx context.Context, url string) (*domain.M3USource, error) {
	return c.base().FindM3USourceByURL(ctx, url)
}

func (c *Catalog) CreateM3USource(ctx context.Context, url, label string) (*domain.M3USource, error) {
	return c.base().CreateM3USource(ctx, url, label)
}

func (c *Catalog) ListM3USources(ctx context.Context) ([]domain.M3USource, error) {
	return c.base().ListM3USources(ctx)
}

func (c *Catalog) DeleteM3USource(ctx context.Context, id string) error {
	return c.base().DeleteM3USource(ctx, id)
}

func (c *Catalog) CreateEPGSource(ctx context.Context, url, label string) (*domain.EPGSource, error) {
	return c.base().CreateEPGSource(ctx, url, label)
}

func (c *Catalog) ListEPGSources(ctx context.Context) ([]domain.EPGSource, error) {
	return c.base().ListEPGSources(ctx)
}

func (c *Catalog) DeleteEPGSource(ctx context.Context, id string) error {
	return c.base().DeleteEPGSource(ctx, id)
}

func (c *Catalog) SaveSnapshot(ctx context.Context, snap domain.MergeSnapshot) error {
	return c.base().SaveSnapshot(ctx, snap)
}

func (c *Catalog) LoadSnapshot(ctx context.Context) (domain.MergeSnapshot, error) {
	return c.base().LoadSnapshot(ctx)
}

var _ ports.CatalogRepository = (*Catalog)(nil)
