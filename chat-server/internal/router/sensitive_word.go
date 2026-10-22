package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerSensitiveWordRoutes(engine *gin.Engine, h *container.Handlers) {
	sh := h.SensitiveWord
	auth := engine.Group("/api/admin/sensitive-word", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		auth.GET("/list", sh.List)
		auth.POST("/create", sh.Create)
		auth.POST("/update", sh.Update)
		auth.POST("/update/status", sh.UpdateStatus)
		auth.POST("/delete", sh.Delete)
		auth.POST("/batch/delete", sh.BatchDelete)
		auth.POST("/batch/status", sh.BatchUpdateStatus)
		auth.POST("/import", sh.Import)
	}
}
