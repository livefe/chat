package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerOpsRoutes(engine *gin.Engine, h *container.Handlers) {

	admin := engine.Group("/api/admin/ops", h.EmployeeAuth)
	{
		admin.GET("/pending-counts", h.Ops.PendingCounts)
	}
}
