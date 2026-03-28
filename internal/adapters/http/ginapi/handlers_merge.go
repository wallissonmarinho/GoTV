package ginapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wallissonmarinho/GoTV/internal/adapters/observability"
)

func (h *handlers) registerMergeRoutes(admin *gin.RouterGroup) {
	admin.POST("/rebuild", h.postRebuild)
	admin.GET("/merge-status", h.getMergeStatus)
}

func (h *handlers) postRebuild(c *gin.Context) {
	baseCtx := context.WithoutCancel(c.Request.Context())
	go func() {
		defer func() {
			if r := recover(); r != nil {
				h.deps.Log.Error("merge panic", slog.Any("panic", r))
			}
		}()
		ctx, cancel := context.WithTimeout(baseCtx, 30*time.Minute)
		defer cancel()
		ctx, span := observability.StartMergeSpan(ctx, "merge.manual")
		defer span.End()
		res := h.deps.Merge.Run(ctx)
		if len(res.Errors) > 0 {
			h.deps.Log.Log(ctx, slog.LevelWarn, "merge warnings",
				slog.String("message", res.Message),
				slog.Any("errors", res.Errors))
		}
		h.deps.Log.Log(ctx, slog.LevelInfo, "merge done",
			slog.String("message", res.Message))
		if c := h.deps.ManualMergeDone; c != nil {
			select {
			case c <- struct{}{}:
			default:
			}
		}
	}()
	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

func (h *handlers) getMergeStatus(c *gin.Context) {
	if h.deps.Catalog == nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "message": "no repository"})
		return
	}
	snap, err := h.deps.Catalog.LoadMergeStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":              snap.OK,
		"message":         snap.Message,
		"last_error":      snap.LastError,
		"channel_count":   snap.ChannelCount,
		"programme_count": snap.ProgrammeCount,
		"started_at":      snap.StartedAt,
		"finished_at":     snap.FinishedAt,
	})
}
