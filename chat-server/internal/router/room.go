package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerRoomRoutes(engine *gin.Engine, h *container.Handlers) {
	room := h.Room

	auth := engine.Group("/api/room", h.UserAuth)
	{
		auth.POST("/create", room.CreateRoom)
		auth.GET("/applications", room.MyRoomApplications)
		auth.GET("/bootstrap", room.RoomBootstrap)
		auth.GET("/nearby/bootstrap", room.NearbyBootstrap)
		auth.GET("/:id", room.RoomDetail)
	}

	admin := engine.Group("/api/admin/room", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.GET("/list", room.AdminRoomList)
		admin.GET("/detail", room.AdminRoomDetail)
		admin.GET("/neighbors", room.AdminRoomNeighbors)
		admin.GET("/map", room.AdminRoomMap)
		admin.POST("/create", room.AdminRoomCreate)
		admin.POST("/update", room.AdminRoomUpdate)
		admin.POST("/audit", room.AdminRoomAudit)
		admin.POST("/update/status", room.AdminRoomUpdateStatus)
	}
}
