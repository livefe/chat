package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"app/internal/middleware"
	"app/internal/model"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

const multipartOverhead = 1 << 20

type MediaHandler struct {
	log      *logger.Logger
	mediaSvc *service.MediaService
}

type uploadFunc func(ctx context.Context, uploader service.MediaUploader, reader io.Reader, size int64, filename string) (*types.MediaResponse, error)

func NewMediaHandler(log *logger.Logger, mediaSvc *service.MediaService) *MediaHandler {
	return &MediaHandler{log: log, mediaSvc: mediaSvc}
}

func userUploader(c *gin.Context) service.MediaUploader {
	return service.UserUploader(middleware.CurrentUserID(c))
}

func employeeUploader(c *gin.Context) service.MediaUploader {
	return service.EmployeeUploader(middleware.CurrentEmployeeID(c))
}

func (h *MediaHandler) UploadImage(c *gin.Context) {
	h.upload(c, userUploader(c), model.MediaKindImage, "请选择要上传的图片", unnamedUpload(h.mediaSvc.UploadImage))
}

func (h *MediaHandler) UploadImageAdmin(c *gin.Context) {
	h.upload(c, employeeUploader(c), model.MediaKindImage, "请选择要上传的图片", unnamedUpload(h.mediaSvc.UploadImage))
}

func (h *MediaHandler) UploadVideo(c *gin.Context) {
	h.upload(c, userUploader(c), model.MediaKindVideo, "请选择要上传的视频", unnamedUpload(h.mediaSvc.UploadVideo))
}

func (h *MediaHandler) UploadVideoAdmin(c *gin.Context) {
	h.upload(c, employeeUploader(c), model.MediaKindVideo, "请选择要上传的视频", unnamedUpload(h.mediaSvc.UploadVideo))
}

func (h *MediaHandler) UploadAudio(c *gin.Context) {
	h.upload(c, userUploader(c), model.MediaKindAudio, "请选择要上传的语音", unnamedUpload(h.mediaSvc.UploadAudio))
}

func (h *MediaHandler) UploadFile(c *gin.Context) {
	h.upload(c, userUploader(c), model.MediaKindFile, "请选择要上传的文件", h.mediaSvc.UploadFile)
}

func (h *MediaHandler) UploadFileAdmin(c *gin.Context) {
	h.upload(c, employeeUploader(c), model.MediaKindFile, "请选择要上传的文件", h.mediaSvc.UploadFile)
}

func unnamedUpload(fn func(ctx context.Context, uploader service.MediaUploader, reader io.Reader, size int64) (*types.MediaResponse, error)) uploadFunc {
	return func(ctx context.Context, uploader service.MediaUploader, reader io.Reader, size int64, _ string) (*types.MediaResponse, error) {
		return fn(ctx, uploader, reader, size)
	}
}

func (h *MediaHandler) upload(c *gin.Context, uploader service.MediaUploader, kind uint8, missing string, fn uploadFunc) {
	maxBytes := h.mediaSvc.MaxUploadBytes(kind)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes+multipartOverhead)
	file, err := c.FormFile("file")
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			response.WriteError(c, errs.Invalid(fmt.Sprintf("文件大小不能超过 %dMB", maxBytes>>20)))
			return
		}
		response.WriteError(c, errs.Invalid(missing))
		return
	}
	f, err := file.Open()
	if err != nil {
		response.WriteError(c, errs.Invalid("读取文件失败"))
		return
	}
	defer f.Close()
	result, err := fn(c.Request.Context(), uploader, f, file.Size, file.Filename)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "上传成功", result)
}
