package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(engine *gin.Engine, h *container.Handlers) {
	ah := h.Auth

	api := engine.Group("/api/auth")
	{
		api.POST("/login", ah.LoginUser)
		api.POST("/refresh", ah.RefreshUser)
		api.POST("/sms/code", ah.SendUserSMSCode)
	}

	admin := engine.Group("/api/admin/auth")
	{
		admin.POST("/login", ah.LoginEmployee)
		admin.POST("/refresh", ah.RefreshEmployee)
		admin.POST("/sms/code", ah.SendEmployeeSMSCode)
	}
}
