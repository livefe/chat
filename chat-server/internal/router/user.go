package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(engine *gin.Engine, h *container.Handlers) {
	uh := h.User

	auth := engine.Group("/api/user", h.UserAuth)
	{
		auth.GET("/me", uh.Detail)
		auth.GET("/profile", uh.PublicProfile)
		auth.POST("/profile/update", uh.UpdateProfile)
		auth.POST("/mobile/send/old-code", uh.SendChangeMobileOldCode)
		auth.POST("/mobile/verify/old-code", uh.VerifyChangeMobileOldCode)
		auth.POST("/mobile/send/new-code", uh.SendChangeMobileNewCode)
		auth.POST("/mobile/change", uh.ChangeMobile)
		auth.POST("/deactivation/send/code", uh.SendDeactivationCode)
		auth.POST("/deactivate", uh.Deactivate)
		auth.POST("/block", uh.Block)
		auth.POST("/unblock", uh.Unblock)
		auth.GET("/blocks", uh.BlockList)
	}

	admin := engine.Group("/api/admin/user", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.GET("/list", uh.List)
		admin.GET("/detail", uh.AdminDetail)
		admin.GET("/punishments", uh.AdminPunishments)
		admin.GET("/login-records", uh.AdminLoginRecords)
		admin.POST("/mobile/reveal", uh.AdminRevealMobile)
		admin.POST("/create", uh.Create)
		admin.POST("/update", uh.Update)
		admin.POST("/update/status", uh.UpdateStatus)
		admin.POST("/update/mute", uh.AdminMute)
		admin.POST("/batch/status", uh.AdminBatchStatus)
		admin.POST("/batch/mute", uh.AdminBatchMute)
		admin.POST("/deactivate", uh.AdminDeactivate)
	}
}
