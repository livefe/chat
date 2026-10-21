package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	log     *logger.Logger
	userSvc *service.UserService
}

func NewUserHandler(log *logger.Logger, userSvc *service.UserService) *UserHandler {
	return &UserHandler{log: log, userSvc: userSvc}
}

func (h *UserHandler) Detail(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	info, err := h.userSvc.Get(c.Request.Context(), userID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", info)
}

func (h *UserHandler) PublicProfile(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	viewerID := middleware.CurrentUserID(c)
	profile, err := h.userSvc.PublicProfile(c.Request.Context(), viewerID, req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", profile)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req types.UpdateProfileRequest
	if !bindJSON(c, &req) {
		return
	}
	userID := middleware.CurrentUserID(c)
	if err := h.userSvc.UpdateProfile(c.Request.Context(), userID, &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *UserHandler) List(c *gin.Context) {
	var req types.UserListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.userSvc.AdminList(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req types.CreateUserRequest
	if !bindJSON(c, &req) {
		return
	}
	id, err := h.userSvc.AdminCreate(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "创建成功", types.IDResponse{ID: id})
}

func (h *UserHandler) Update(c *gin.Context) {
	var req types.UpdateUserRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.userSvc.AdminUpdate(c.Request.Context(), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
	var req types.UpdateUserStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.userSvc.AdminUpdateStatus(c.Request.Context(), req.ID, req.Status); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *UserHandler) SendDeactivationCode(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	if err := h.userSvc.SendDeactivationCode(c.Request.Context(), userID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "验证码已发送", nil)
}

func (h *UserHandler) SendChangeMobileOldCode(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	if err := h.userSvc.SendChangeMobileOldCode(c.Request.Context(), userID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "验证码已发送", nil)
}

func (h *UserHandler) VerifyChangeMobileOldCode(c *gin.Context) {
	var req types.VerifyOldMobileCodeRequest
	if !bindJSON(c, &req) {
		return
	}
	userID := middleware.CurrentUserID(c)
	data, err := h.userSvc.VerifyChangeMobileOldCode(c.Request.Context(), userID, req.Code)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "验证成功", data)
}

func (h *UserHandler) SendChangeMobileNewCode(c *gin.Context) {
	var req types.SendNewMobileCodeRequest
	if !bindJSON(c, &req) {
		return
	}
	userID := middleware.CurrentUserID(c)
	if err := h.userSvc.SendChangeMobileNewCode(c.Request.Context(), userID, req.Mobile, req.Ticket); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "验证码已发送", nil)
}

func (h *UserHandler) ChangeMobile(c *gin.Context) {
	var req types.ChangeMobileRequest
	if !bindJSON(c, &req) {
		return
	}
	userID := middleware.CurrentUserID(c)
	if err := h.userSvc.ChangeMobile(c.Request.Context(), userID, &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更换成功", nil)
}

func (h *UserHandler) Deactivate(c *gin.Context) {
	var req types.DeactivationRequest
	if !bindJSON(c, &req) {
		return
	}
	userID := middleware.CurrentUserID(c)
	if err := h.userSvc.Deactivate(c.Request.Context(), userID, req.Code, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "注销成功", nil)
}

func (h *UserHandler) AdminDetail(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	detail, err := h.userSvc.AdminGet(c.Request.Context(), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", detail)
}

func (h *UserHandler) AdminMute(c *gin.Context) {
	var req types.AdminUserMuteRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.userSvc.AdminMute(c.Request.Context(), req.ID, req.Duration, req.Reason); err != nil {
		response.WriteError(c, err)
		return
	}
	msg := "解除禁言成功"
	if req.Duration != 0 {
		msg = "禁言设置成功"
	}
	response.Success(c, msg, nil)
}

func (h *UserHandler) AdminDeactivate(c *gin.Context) {
	var req types.AdminDeactivateUserRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.userSvc.AdminDeactivate(c.Request.Context(), middleware.CurrentEmployeeID(c), req.ID, req.Reason, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "注销成功", nil)
}

func (h *UserHandler) Block(c *gin.Context) {
	var req types.BlockUserRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.userSvc.Block(c.Request.Context(), middleware.CurrentUserID(c), req.UserID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已拉黑", nil)
}

func (h *UserHandler) Unblock(c *gin.Context) {
	var req types.BlockUserRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.userSvc.Unblock(c.Request.Context(), middleware.CurrentUserID(c), req.UserID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已取消拉黑", nil)
}

func (h *UserHandler) BlockList(c *gin.Context) {
	var req types.BlockListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.userSvc.ListBlocked(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}
