package app

import (
	"context"
	"os"
	"time"

	"github.com/wallissonmarinho/GoTV/internal/adapters/clock"
	"github.com/wallissonmarinho/GoTV/internal/adapters/httpclient"
	"github.com/wallissonmarinho/GoTV/internal/adapters/m3u"
	"github.com/wallissonmarinho/GoTV/internal/adapters/state"
	"github.com/wallissonmarinho/GoTV/internal/adapters/storage"
	xmltvad "github.com/wallissonmarinho/GoTV/internal/adapters/xmltv"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
	"github.com/wallissonmarinho/GoTV/internal/core/services"
)

// OpenCatalog opens the catalog database (SQLite file or PostgreSQL via DSN).
func OpenCatalog(dsn string) (*storage.Catalog, error) {
	return storage.Open(dsn)
}

// HydrateStore loads the last persisted snapshot into the in-memory output store when present.
func HydrateStore(ctx context.Context, repo ports.CatalogRepository, mem *state.MemoryStore) {
	snap, err := repo.LoadSnapshot(ctx)
	if err != nil || (len(snap.M3U) == 0 && len(snap.EPG) == 0) {
		return
	}
	mem.Set(snap.M3U, snap.EPG, ports.OutputMeta{
		OK:             snap.OK,
		Message:        snap.Message,
		ChannelCount:   snap.ChannelCount,
		ProgrammeCount: snap.ProgrammeCount,
	})
}

// NewCatalogAdmin wires catalog admin use case with the same DB handle as repository and unit of work.
func NewCatalogAdmin(repo *storage.Catalog) *services.CatalogAdminService {
	return services.NewCatalogAdminService(repo, repo)
}

// MergeRuntimeOptions configures outbound HTTP, probing, and encoding for the merge service.
type MergeRuntimeOptions struct {
	HTTPTimeout  time.Duration
	MaxBodyBytes int64
	UserAgent    string
	ProbeTimeout time.Duration
	ProbeStreams bool
	MaxProbes    int
	ProbeWorkers int
}

// NewPlaylistMergeService builds the merge runner with concrete adapters (HTTP client, parsers, writers, probe).
func NewPlaylistMergeService(repo *storage.Catalog, mem *state.MemoryStore, o MergeRuntimeOptions) *services.PlaylistMergeService {
	getter := httpclient.NewGetter(o.HTTPTimeout, o.UserAgent, o.MaxBodyBytes)
	probe := httpclient.NewStreamProbe(o.ProbeTimeout, o.UserAgent)
	return services.NewPlaylistMergeService(services.PlaylistMergeDeps{
		Repo:         repo,
		UoW:          repo,
		HTTP:         getter,
		M3U:          m3u.Parser{},
		XMLTV:        xmltvad.Parser{},
		Probe:        probe,
		Clock:        &clock.System{},
		Store:        mem,
		M3UOut:       &m3u.Writer{},
		EPGOut:       &xmltvad.Writer{},
		ProbeStreams: o.ProbeStreams,
		MaxProbes:    o.MaxProbes,
		ProbeWorkers: o.ProbeWorkers,
	})
}

// AdminAPIKey returns GOTV_ADMIN_API_KEY or ADMIN_API_KEY for Gin admin routes.
func AdminAPIKey() string {
	if v := os.Getenv("GOTV_ADMIN_API_KEY"); v != "" {
		return v
	}
	return os.Getenv("ADMIN_API_KEY")
}
