package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerPermissionRoutes(engine *gin.Engine, h *container.Handlers) {
	ph := h.Permission

	engine.GET("/api/admin/permission/me", h.EmployeeAuth, ph.GetEmployeePermissions)

	admin := engine.Group("/api/admin/permission", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.GET("/list", ph.List)
		admin.GET("/detail", ph.Detail)
		admin.GET("/tree", ph.Tree)
		admin.POST("/create", ph.Create)
		admin.POST("/update", ph.Update)
		admin.POST("/delete", ph.Delete)
	}
}
