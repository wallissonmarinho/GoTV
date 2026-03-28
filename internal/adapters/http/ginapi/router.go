package ginapi

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/wallissonmarinho/GoTV/internal/adapters/state"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// Config holds HTTP settings.
type Config struct {
	AdminAPIKey string
}

// Deps wires handlers. Catalog must be set for /api/v1 admin routes (CRUD + merge-status).
type Deps struct {
	Merge   ports.MergeRunner
	Catalog ports.CatalogAdmin
	Store   *state.MemoryStore
	Log     *slog.Logger
	// ManualMergeDone optional: after POST /api/v1/rebuild, one send is made when the
	// background merge goroutine finishes (including logs). Buffered (cap ≥1) so the handler never blocks.
	ManualMergeDone chan struct{}
}

// handlers binds Gin routes to ports (HTTP adapter). See handlers_public.go, handlers_admin.go, handlers_*_sources.go, handlers_merge.go, middleware.go.
type handlers struct {
	cfg  Config
	deps Deps
}

func newHandlers(cfg Config, d Deps) *handlers {
	if d.Log == nil {
		d.Log = slog.Default()
	}
	return &handlers{cfg: cfg, deps: d}
}

// Register attaches routes to the engine.
func Register(engine *gin.Engine, cfg Config, d Deps) {
	h := newHandlers(cfg, d)
	h.registerPublic(engine)
	h.registerAdminV1(engine)
}
