package ginapi

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func adminAuthMiddleware(apiKey string, lg *slog.Logger) gin.HandlerFunc {
	key := strings.TrimSpace(apiKey)
	return func(c *gin.Context) {
		if key == "" {
			if lg != nil {
				lg.Warn("admin API key unset — /api/v1 admin routes are open")
			}
			c.Next()
			return
		}
		auth := c.GetHeader("Authorization")
		const p = "Bearer "
		if strings.HasPrefix(auth, p) && strings.TrimSpace(auth[len(p):]) == key {
			c.Next()
			return
		}
		if c.GetHeader("X-Admin-API-Key") == key {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}
