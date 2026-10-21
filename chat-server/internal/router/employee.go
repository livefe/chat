package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerEmployeeRoutes(engine *gin.Engine, h *container.Handlers) {
	eh := h.Employee

	me := engine.Group("/api/admin/employee", h.EmployeeAuth)
	{
		me.GET("/me", eh.Detail)
		me.POST("/me/update", eh.UpdateProfile)
		me.POST("/me/avatar", h.Media.UploadImageAdmin)
	}

	auth := engine.Group("/api/admin/employee", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		auth.GET("/list", eh.List)
		auth.POST("/create", eh.Create)
		auth.POST("/update", eh.Update)
		auth.POST("/update/status", eh.UpdateStatus)
		auth.GET("/roles", eh.GetRoles)
		auth.POST("/assign/roles", eh.AssignRoles)
	}
}
