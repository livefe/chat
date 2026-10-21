package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerAuditLogRoutes(engine *gin.Engine, h *container.Handlers) {
	ah := h.AuditLog

	admin := engine.Group("/api/admin/audit", h.EmployeeAuth, h.RequirePermission)
	{
		admin.GET("/list", ah.List)
		admin.GET("/detail", ah.Detail)
	}
}
