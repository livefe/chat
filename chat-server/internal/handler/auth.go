package handler

import (
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	log     *logger.Logger
	authSvc *service.AuthService
	smsSvc  *service.SMSService
}

func NewAuthHandler(log *logger.Logger, authSvc *service.AuthService, smsSvc *service.SMSService) *AuthHandler {
	return &AuthHandler{log: log, authSvc: authSvc, smsSvc: smsSvc}
}

func (h *AuthHandler) LoginUser(c *gin.Context) {
	var req types.LoginRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.authSvc.LoginUser(c.Request.Context(), &req, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "登录成功", result)
}

func (h *AuthHandler) RefreshUser(c *gin.Context) {
	var req types.RefreshTokenRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.authSvc.RefreshUser(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.log.Warn(c.Request.Context(), "令牌刷新失败", logger.Error(err))
		response.WriteError(c, err)
		return
	}
	response.Success(c, "刷新成功", result)
}

func (h *AuthHandler) LoginEmployee(c *gin.Context) {
	var req types.EmployeeLoginRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.authSvc.LoginEmployee(c.Request.Context(), &req, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "登录成功", result)
}

func (h *AuthHandler) RefreshEmployee(c *gin.Context) {
	var req types.RefreshTokenRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.authSvc.RefreshEmployee(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.log.Warn(c.Request.Context(), "员工令牌刷新失败", logger.Error(err))
		response.WriteError(c, err)
		return
	}
	response.Success(c, "刷新成功", result)
}

func (h *AuthHandler) SendUserSMSCode(c *gin.Context) {
	h.sendSMSCode(c, types.SourceAppLogin)
}

func (h *AuthHandler) SendEmployeeSMSCode(c *gin.Context) {
	h.sendSMSCode(c, types.SourceAdminLogin)
}

func (h *AuthHandler) sendSMSCode(c *gin.Context, source string) {
	var req struct {
		Mobile string `json:"mobile" binding:"required,mobile"`
	}
	if !bindJSON(c, &req) {
		return
	}
	smsReq := &types.SendCodeRequest{Mobile: req.Mobile, Source: source}
	if err := h.smsSvc.SendLoginCode(c.Request.Context(), smsReq, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "验证码发送成功", nil)
}
