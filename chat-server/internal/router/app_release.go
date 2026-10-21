package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerAppReleaseRoutes(engine *gin.Engine, h *container.Handlers) {
	ah := h.AppRelease

	engine.GET("/api/app/release/current", ah.PublicCurrent)
	engine.GET("/api/app/update/check", ah.PublicCheck)

	admin := engine.Group("/api/admin/app/release", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.GET("/list", ah.AdminList)
		admin.POST("/upload", ah.AdminUpload)
		admin.POST("/publish-meta", ah.AdminPublishMeta)
		admin.POST("/set-current", ah.AdminSetCurrent)
		admin.POST("/unset-current", ah.AdminUnsetCurrent)
		admin.GET("/policy/list", ah.AdminPolicyList)
		admin.POST("/policy/save", ah.AdminPolicySave)
	}
}
