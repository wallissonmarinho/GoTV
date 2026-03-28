package ginapi

import "github.com/gin-gonic/gin"

func (h *handlers) registerAdminV1(engine *gin.Engine) {
	v1 := engine.Group("/api/v1")
	admin := v1.Group("")
	admin.Use(adminAuthMiddleware(h.cfg.AdminAPIKey, h.deps.Log))

	h.registerM3USourceRoutes(admin)
	h.registerEPGSourceRoutes(admin)
	h.registerMergeRoutes(admin)
}
