package ginapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *handlers) registerPublic(engine *gin.Engine) {
	engine.GET("/health", h.getHealth)
	engine.GET("/playlist.m3u", h.getPlaylistM3U)
	engine.GET("/epg.xml", h.getEPGXML)
}

func (h *handlers) getHealth(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (h *handlers) getPlaylistM3U(c *gin.Context) {
	m3u, _, ok := h.deps.Store.Get()
	if !ok || len(m3u) == 0 {
		c.String(http.StatusNotFound, "no playlist yet — wait for merge job\n")
		return
	}
	c.Data(http.StatusOK, "audio/x-mpegurl; charset=utf-8", m3u)
}

func (h *handlers) getEPGXML(c *gin.Context) {
	_, epg, ok := h.deps.Store.Get()
	if !ok || len(epg) == 0 {
		c.String(http.StatusNotFound, "no epg yet — wait for merge job\n")
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", epg)
}
