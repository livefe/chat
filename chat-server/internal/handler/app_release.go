package handler

import (
	"errors"
	"fmt"
	"net/http"

	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type AppReleaseHandler struct {
	log           *logger.Logger
	appReleaseSvc *service.AppReleaseService
}

func NewAppReleaseHandler(log *logger.Logger, appReleaseSvc *service.AppReleaseService) *AppReleaseHandler {
	return &AppReleaseHandler{log: log, appReleaseSvc: appReleaseSvc}
}

func (h *AppReleaseHandler) PublicCurrent(c *gin.Context) {
	result, err := h.appReleaseSvc.ListCurrent(c.Request.Context())
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *AppReleaseHandler) PublicCheck(c *gin.Context) {
	platform := middleware.ClientPlatform(c)
	if !middleware.ValidAppPlatform(platform) {
		response.WriteError(c, errs.Invalid("当前客户端不支持版本检查"))
		return
	}
	req := &types.AppUpdateCheckRequest{
		Platform: platform,
		Version:  middleware.ClientVersion(c),
	}
	result, err := h.appReleaseSvc.CheckUpdate(c.Request.Context(), req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *AppReleaseHandler) AdminList(c *gin.Context) {
	var req types.AppReleaseAdminListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.appReleaseSvc.AdminList(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *AppReleaseHandler) AdminPolicyList(c *gin.Context) {
	result, err := h.appReleaseSvc.AdminListPolicies(c.Request.Context())
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *AppReleaseHandler) AdminPolicySave(c *gin.Context) {
	var req types.AppUpdatePolicySaveRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.appReleaseSvc.AdminSavePolicy(c.Request.Context(), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "策略已保存", nil)
}

func (h *AppReleaseHandler) AdminUpload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.AppPackageMaxBytes+multipartOverhead)
	file, err := c.FormFile("file")
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			response.WriteError(c, errs.Invalid(fmt.Sprintf("安装包大小不能超过 %dMB", service.AppPackageMaxMB)))
			return
		}
		response.WriteError(c, errs.Invalid("请选择安装包文件"))
		return
	}
	platform := c.PostForm("platform")
	version := c.PostForm("version")
	changelog := c.PostForm("changelog")
	f, err := file.Open()
	if err != nil {
		response.WriteError(c, errs.Invalid("读取文件失败"))
		return
	}
	defer f.Close()

	item, err := h.appReleaseSvc.AdminUpload(c.Request.Context(), platform, version, changelog, file.Filename, f, file.Size)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "上传成功", item)
}

func (h *AppReleaseHandler) AdminPublishMeta(c *gin.Context) {
	var req types.AppReleasePublishMetaRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.appReleaseSvc.AdminPublishMeta(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "登记成功", item)
}

func (h *AppReleaseHandler) AdminSetCurrent(c *gin.Context) {
	var req types.IDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.appReleaseSvc.AdminSetCurrent(c.Request.Context(), req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已设为当前版本", nil)
}

func (h *AppReleaseHandler) AdminUnsetCurrent(c *gin.Context) {
	var req types.IDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.appReleaseSvc.AdminUnsetCurrent(c.Request.Context(), req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已下架", nil)
}

func (h *AppReleaseHandler) AdminDelete(c *gin.Context) {
	var req types.IDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.appReleaseSvc.AdminDelete(c.Request.Context(), req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已删除", nil)
}
