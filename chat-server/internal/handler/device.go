package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	log       *logger.Logger
	deviceSvc *service.DeviceService
}

func NewDeviceHandler(log *logger.Logger, deviceSvc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{log: log, deviceSvc: deviceSvc}
}

func (h *DeviceHandler) Register(c *gin.Context) {
	var req types.DeviceRegisterRequest
	if !bindJSON(c, &req) {
		return
	}
	platform := middleware.ClientPlatform(c)
	if !middleware.ValidDevicePlatform(platform) {
		response.WriteError(c, errs.Invalid("当前客户端不支持设备推送绑定"))
		return
	}
	cmd := &types.DeviceRegisterCommand{
		DeviceID:   middleware.ClientDeviceID(c),
		Platform:   platform,
		Vendor:     req.Vendor,
		PushToken:  req.PushToken,
		AppVersion: middleware.ClientVersion(c),
		OSVersion:  req.OSVersion,
	}
	if err := h.deviceSvc.Register(c.Request.Context(), middleware.CurrentUserID(c), cmd); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "绑定成功", nil)
}

func (h *DeviceHandler) AdminList(c *gin.Context) {
	var req types.AdminUserSubListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.deviceSvc.AdminListByUser(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *DeviceHandler) Unbind(c *gin.Context) {
	if err := h.deviceSvc.Unbind(c.Request.Context(), middleware.CurrentUserID(c), middleware.ClientDeviceID(c)); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "解绑成功", nil)
}
