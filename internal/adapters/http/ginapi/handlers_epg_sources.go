package ginapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *handlers) registerEPGSourceRoutes(admin *gin.RouterGroup) {
	admin.POST("/epg-sources", h.postEPGSource)
	admin.GET("/epg-sources", h.listEPGSources)
	admin.DELETE("/epg-sources/:id", h.deleteEPGSource)
}

func (h *handlers) postEPGSource(c *gin.Context) {
	var body sourceCreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	u := strings.TrimSpace(body.URL)
	if u == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url required"})
		return
	}
	created, err := h.deps.Catalog.CreateEPGSource(c.Request.Context(), u, strings.TrimSpace(body.Label))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *handlers) listEPGSources(c *gin.Context) {
	list, err := h.deps.Catalog.ListEPGSources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *handlers) deleteEPGSource(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if delErr := h.deps.Catalog.DeleteEPGSource(c.Request.Context(), id); delErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": delErr.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
