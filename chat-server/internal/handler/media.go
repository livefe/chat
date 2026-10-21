package handler

import (
	"context"
	"io"

	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	log      *logger.Logger
	mediaSvc *service.MediaService
}

type uploadFunc func(ctx context.Context, uploaderID uint, reader io.Reader, size int64, filename string) (*types.MediaResponse, error)

func NewMediaHandler(log *logger.Logger, mediaSvc *service.MediaService) *MediaHandler {
	return &MediaHandler{log: log, mediaSvc: mediaSvc}
}

func (h *MediaHandler) UploadImage(c *gin.Context) {
	h.upload(c, middleware.CurrentUserID(c), "请选择要上传的图片", unnamedUpload(h.mediaSvc.UploadImage))
}

func (h *MediaHandler) UploadImageAdmin(c *gin.Context) {
	h.upload(c, middleware.CurrentEmployeeID(c), "请选择要上传的图片", unnamedUpload(h.mediaSvc.UploadImage))
}

func (h *MediaHandler) UploadVideo(c *gin.Context) {
	h.upload(c, middleware.CurrentUserID(c), "请选择要上传的视频", unnamedUpload(h.mediaSvc.UploadVideo))
}

func (h *MediaHandler) UploadVideoAdmin(c *gin.Context) {
	h.upload(c, middleware.CurrentEmployeeID(c), "请选择要上传的视频", unnamedUpload(h.mediaSvc.UploadVideo))
}

func (h *MediaHandler) UploadAudio(c *gin.Context) {
	h.upload(c, middleware.CurrentUserID(c), "请选择要上传的语音", unnamedUpload(h.mediaSvc.UploadAudio))
}

func (h *MediaHandler) UploadFile(c *gin.Context) {
	h.upload(c, middleware.CurrentUserID(c), "请选择要上传的文件", h.mediaSvc.UploadFile)
}

func (h *MediaHandler) UploadFileAdmin(c *gin.Context) {
	h.upload(c, middleware.CurrentEmployeeID(c), "请选择要上传的文件", h.mediaSvc.UploadFile)
}

func unnamedUpload(fn func(ctx context.Context, uploaderID uint, reader io.Reader, size int64) (*types.MediaResponse, error)) uploadFunc {
	return func(ctx context.Context, uploaderID uint, reader io.Reader, size int64, _ string) (*types.MediaResponse, error) {
		return fn(ctx, uploaderID, reader, size)
	}
}

func (h *MediaHandler) upload(c *gin.Context, uploaderID uint, missing string, fn uploadFunc) {
	file, err := c.FormFile("file")
	if err != nil {
		response.WriteError(c, errs.Invalid(missing))
		return
	}
	f, err := file.Open()
	if err != nil {
		response.WriteError(c, errs.Invalid("读取文件失败"))
		return
	}
	defer f.Close()
	result, err := fn(c.Request.Context(), uploaderID, f, file.Size, file.Filename)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "上传成功", result)
}
