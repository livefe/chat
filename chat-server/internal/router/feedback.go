package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerFeedbackRoutes(engine *gin.Engine, h *container.Handlers) {
	fh := h.Feedback

	api := engine.Group("/api/feedback", h.UserAuth)
	{
		api.POST("/submit", fh.Submit)
		api.GET("/my/list", fh.MyList)
		api.GET("/my/detail", fh.MyDetail)
	}

	admin := engine.Group("/api/admin/feedback", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.GET("/list", fh.AdminList)
		admin.GET("/detail", fh.AdminDetail)
		admin.POST("/update/status", fh.AdminUpdateStatus)
	}
}
