package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerSupportRoutes(engine *gin.Engine, h *container.Handlers) {
	user := engine.Group("/api/support", h.UserAuth)
	{
		user.GET("/profile", h.Support.Profile)
		user.GET("/session/current", h.Support.Current)
		user.POST("/session/open", h.Support.Open)
		user.POST("/session/cancel", h.Support.Cancel)
		user.GET("/messages", h.Support.Messages)
		user.POST("/messages", h.Support.Send)
		user.POST("/read", h.Support.Read)
	}

	admin := engine.Group("/api/admin/support", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.GET("/queue", h.Support.AdminQueue)
		admin.GET("/mine", h.Support.AdminMine)
		admin.GET("/closed", h.Support.AdminClosed)
		admin.GET("/active-all", h.Support.AdminActiveAll)
		admin.GET("/closed-all", h.Support.AdminClosedAll)
		admin.GET("/overview/summary", h.Support.AdminOverviewSummary)
		admin.GET("/session", h.Support.AdminSession)
		admin.GET("/by-user", h.Support.AdminByUser)
		admin.GET("/history", h.Support.AdminHistory)
		admin.GET("/agents", h.Support.AdminAgents)
		admin.GET("/quick-replies", h.Support.AdminQuickReplies)
		admin.POST("/accept", h.Support.AdminAccept)
		admin.POST("/transfer", h.Support.AdminTransfer)
		admin.POST("/close", h.Support.AdminClose)
		admin.GET("/messages", h.Support.AdminMessages)
		admin.POST("/messages", h.Support.AdminSend)
		admin.POST("/read", h.Support.AdminRead)
		admin.GET("/config", h.Support.AdminGetConfig)
		admin.POST("/config", h.Support.AdminUpdateConfig)
	}
}
