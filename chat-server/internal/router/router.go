package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func Register(engine *gin.Engine, h *container.Handlers) {
	registerStaticRoutes(engine)

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	engine.Use(h.Recovery, h.ClientContext, h.Logger)

	registerAuthRoutes(engine, h)
	registerUserRoutes(engine, h)
	registerEmployeeRoutes(engine, h)
	registerDeviceRoutes(engine, h)

	registerDepartmentRoutes(engine, h)
	registerRoleRoutes(engine, h)
	registerPermissionRoutes(engine, h)

	registerMediaRoutes(engine, h)

	registerRoomRoutes(engine, h)
	registerMessageRoutes(engine, h)
	registerConversationRoutes(engine, h)
	registerRealtimeRoutes(engine, h)

	registerPostRoutes(engine, h)
	registerAdRoutes(engine, h)

	registerFeedbackRoutes(engine, h)
	registerReportRoutes(engine, h)
	registerSensitiveWordRoutes(engine, h)

	registerIMRoutes(engine, h)
	registerSupportRoutes(engine, h)

	registerRegionRoutes(engine, h)
	registerAuditLogRoutes(engine, h)
	registerAppReleaseRoutes(engine, h)

	registerOpsRoutes(engine, h)
}
