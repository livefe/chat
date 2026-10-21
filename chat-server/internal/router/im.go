package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerIMRoutes(engine *gin.Engine, h *container.Handlers) {
	ih := h.IM
	engine.GET("/api/admin/ws", ih.Upgrade)

	im := engine.Group("/api/admin/im", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		im.GET("/contacts", ih.Contacts)
		im.POST("/conversations/dm", ih.OpenDM)
		im.GET("/conversations", ih.Conversations)
		im.GET("/messages", ih.Messages)
		im.POST("/messages", ih.Send)
		im.POST("/read", ih.Read)
		im.POST("/recall", ih.Recall)
	}
}
