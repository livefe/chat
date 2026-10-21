package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerMessageRoutes(engine *gin.Engine, h *container.Handlers) {
	msg := h.Message

	auth := engine.Group("/api/message", h.UserAuth)
	{
		auth.GET("/room", msg.RoomMessages)
		auth.GET("/post", msg.PostMessages)
		auth.GET("/nearby", msg.NearbyMessages)
		auth.GET("/private", msg.PrivateMessages)
		auth.POST("/recall", msg.RecallMessage)
	}

	admin := engine.Group("/api/admin/message", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.GET("/list", msg.AdminMessageList)
		admin.GET("/context", msg.AdminMessageContext)
		admin.POST("/delete", msg.AdminMessageDelete)
	}
}
