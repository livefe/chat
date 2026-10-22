package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerRegionRoutes(engine *gin.Engine, h *container.Handlers) {
	rh := h.Region

	auth := engine.Group("/api/region", h.UserAuth)
	{
		auth.GET("/children", rh.Children)
		auth.GET("/search", rh.Search)
		auth.GET("/detail", rh.Detail)
	}

	admin := engine.Group("/api/admin/region", h.EmployeeAuth, h.RequirePermission)
	{
		admin.GET("/children", rh.Children)
		admin.GET("/search", rh.Search)
		admin.GET("/trail", rh.AdminTrail)
		admin.GET("/detail", rh.AdminDetail)
	}
}
