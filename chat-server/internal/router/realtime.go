package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerRealtimeRoutes(engine *gin.Engine, h *container.Handlers) {
	engine.GET("/api/ws", h.Realtime.Upgrade)
}
