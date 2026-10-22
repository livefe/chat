package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerPostRoutes(engine *gin.Engine, h *container.Handlers) {
	ph := h.Post

	auth := engine.Group("/api/post", h.UserAuth)
	{
		auth.POST("/create", ph.Create)
		auth.POST("/update", ph.Update)
		auth.POST("/withdraw", ph.Withdraw)
		auth.POST("/delete", ph.DeleteMine)
		auth.GET("/list", ph.List)
		auth.GET("/my/list", ph.MyList)
		auth.GET("/my/joined/list", ph.MyJoinedList)
		auth.POST("/my/joined/remove", ph.HideJoined)
		auth.GET("/user/list", ph.UserList)
		auth.GET("/detail", ph.Detail)
		auth.POST("/signup", ph.Signup)
		auth.POST("/signup/cancel", ph.CancelSignup)
		auth.GET("/signup/check-in-code", ph.GetCheckInCode)
		auth.POST("/signup/check-in", ph.CheckInSignup)
		auth.GET("/signup/list", ph.PublisherSignupList)
		auth.POST("/signup/remark", ph.UpdateSignupRemark)
	}

	admin := engine.Group("/api/admin/post", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.GET("/list", ph.AdminList)
		admin.GET("/map", ph.AdminMap)
		admin.GET("/detail", ph.AdminDetail)
		admin.GET("/signup/list", ph.AdminSignupList)
		admin.GET("/signup/by-user", ph.AdminUserSignupList)
		admin.POST("/update/status", ph.AdminUpdateStatus)
		admin.POST("/batch/status", ph.AdminBatchUpdateStatus)
	}
}
