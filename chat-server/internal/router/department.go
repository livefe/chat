package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerDepartmentRoutes(engine *gin.Engine, h *container.Handlers) {
	dh := h.Department

	engine.GET("/api/admin/department/options", h.EmployeeAuth, dh.DepartmentOptions)

	auth := engine.Group("/api/admin/department", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		auth.GET("/list", dh.ListDepartments)
		auth.POST("/create", dh.CreateDepartment)
		auth.POST("/update", dh.UpdateDepartment)
		auth.POST("/update/status", dh.UpdateDepartmentStatus)
	}
}
