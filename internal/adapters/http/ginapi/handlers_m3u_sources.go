package ginapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *handlers) registerM3USourceRoutes(admin *gin.RouterGroup) {
	admin.POST("/m3u-sources", h.postM3USource)
	admin.GET("/m3u-sources", h.listM3USources)
	admin.DELETE("/m3u-sources/:id", h.deleteM3USource)
}

func (h *handlers) postM3USource(c *gin.Context) {
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
	created, err := h.deps.Catalog.CreateM3USource(c.Request.Context(), u, strings.TrimSpace(body.Label))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *handlers) listM3USources(c *gin.Context) {
	list, err := h.deps.Catalog.ListM3USources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *handlers) deleteM3USource(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if delErr := h.deps.Catalog.DeleteM3USource(c.Request.Context(), id); delErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": delErr.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
