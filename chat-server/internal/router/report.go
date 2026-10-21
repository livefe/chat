package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerReportRoutes(engine *gin.Engine, h *container.Handlers) {
	rh := h.Report

	api := engine.Group("/api/report", h.UserAuth)
	{
		api.POST("/submit", rh.Submit)
	}

	admin := engine.Group("/api/admin/report", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.GET("/list", rh.AdminList)
		admin.GET("/detail", rh.AdminDetail)
		admin.POST("/handle", rh.AdminHandle)
	}
}
