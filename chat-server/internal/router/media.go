package router

import (
	"app/internal/container"

	"github.com/gin-gonic/gin"
)

func registerMediaRoutes(engine *gin.Engine, h *container.Handlers) {
	mh := h.Media

	auth := engine.Group("/api/upload", h.UserAuth)
	{
		auth.POST("/image", mh.UploadImage)
		auth.POST("/video", mh.UploadVideo)
		auth.POST("/audio", mh.UploadAudio)
		auth.POST("/file", mh.UploadFile)
	}

	admin := engine.Group("/api/admin/upload", h.EmployeeAuth, h.RequirePermission, h.AuditMiddleware)
	{
		admin.POST("/image", mh.UploadImageAdmin)
		admin.POST("/video", mh.UploadVideoAdmin)
		admin.POST("/file", mh.UploadFileAdmin)
	}
}
