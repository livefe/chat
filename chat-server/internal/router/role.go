package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerRoleRoutes(engine *gin.Engine, h *container.Handlers) {
	rh := h.Role

	role := engine.Group("/api/admin/role", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		role.GET("/list", rh.ListRoles)
		role.GET("/detail", rh.RoleDetail)
		role.POST("/create", rh.CreateRole)
		role.POST("/update", rh.UpdateRole)
		role.POST("/delete", rh.DeleteRole)
		role.GET("/permissions", rh.RolePermissions)
		role.POST("/assign/permissions", rh.AssignRolePermissions)
	}
}
