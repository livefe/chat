package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerAdRoutes(engine *gin.Engine, h *container.Handlers) {
	ah := h.Ad

	api := engine.Group("/api/ad", h.UserAuth)
	{
		api.POST("/events", ah.ReportEvents)
	}

	admin := engine.Group("/api/admin/ad", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.GET("/campaign/list", ah.AdminList)
		admin.GET("/campaign/detail", ah.AdminDetail)
		admin.POST("/campaign/save", ah.AdminSave)
		admin.POST("/campaign/update/status", ah.AdminUpdateStatus)
		admin.POST("/campaign/batch/pause", ah.AdminBatchPause)
		admin.POST("/campaign/copy", ah.AdminCopy)
		admin.GET("/campaign/daily-stats", ah.AdminDailyStats)
		admin.GET("/setting", ah.AdminGetSetting)
		admin.POST("/setting", ah.AdminSaveSetting)
	}
}
