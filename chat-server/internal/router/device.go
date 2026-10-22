package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerDeviceRoutes(engine *gin.Engine, h *container.Handlers) {
	auth := engine.Group("/api/device", h.UserAuth)
	{
		auth.POST("/register", h.Device.Register)
		auth.POST("/unbind", h.Device.Unbind)
	}

	admin := engine.Group("/api/admin/user", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.GET("/devices", h.Device.AdminList)
	}
}
