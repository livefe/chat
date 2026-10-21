package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerConversationRoutes(engine *gin.Engine, h *container.Handlers) {
	conv := h.Conversation

	auth := engine.Group("/api/conversation", h.UserAuth)
	{
		auth.GET("/list", conv.Conversations)
		auth.GET("/unread", conv.ConversationUnread)
		auth.POST("/read", conv.ReadConversation)
		auth.POST("/delete", conv.DeleteConversation)
		auth.POST("/pin", conv.PinConversation)
	}

	admin := engine.Group("/api/admin/conversation", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.GET("/list", conv.AdminConversationList)
	}
}
