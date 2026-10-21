package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type SupportHandler struct {
	log        *logger.Logger
	supportSvc *service.SupportService
}

func NewSupportHandler(log *logger.Logger, supportSvc *service.SupportService) *SupportHandler {
	return &SupportHandler{log: log, supportSvc: supportSvc}
}

func (h *SupportHandler) Profile(c *gin.Context) {
	result, err := h.supportSvc.GetProfile(c.Request.Context())
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) Current(c *gin.Context) {
	result, err := h.supportSvc.GetCurrent(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) Open(c *gin.Context) {
	var req types.SupportOpenRequest
	if c.Request.ContentLength != 0 && !bindJSON(c, &req) {
		return
	}
	result, err := h.supportSvc.Open(c.Request.Context(), middleware.CurrentUserID(c), req.Source)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) Cancel(c *gin.Context) {
	var req types.IDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.supportSvc.Cancel(c.Request.Context(), middleware.CurrentUserID(c), req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "ok", nil)
}

func (h *SupportHandler) Messages(c *gin.Context) {
	var req types.SupportMessageQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.supportSvc.UserListMessages(c.Request.Context(), middleware.CurrentUserID(c), req.ConversationID, req.BeforeID, req.Limit)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) Send(c *gin.Context) {
	var req types.SupportSendRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.supportSvc.UserSend(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", item)
}

func (h *SupportHandler) Read(c *gin.Context) {
	var req types.SupportReadRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.supportSvc.UserRead(c.Request.Context(), middleware.CurrentUserID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "ok", nil)
}

func (h *SupportHandler) AdminGetConfig(c *gin.Context) {
	result, err := h.supportSvc.GetProfile(c.Request.Context())
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) AdminUpdateConfig(c *gin.Context) {
	var req types.SupportConfigUpdateRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.supportSvc.AdminUpdateConfig(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) AdminQueue(c *gin.Context) {
	var req types.PageQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.supportSvc.AdminListQueue(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) AdminMine(c *gin.Context) {
	var req types.PageQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.supportSvc.AdminListMine(c.Request.Context(), middleware.CurrentEmployeeID(c), req.Page, req.PageSize)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) AdminClosed(c *gin.Context) {
	var req types.PageQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.supportSvc.AdminListMyClosed(c.Request.Context(), middleware.CurrentEmployeeID(c), req.Page, req.PageSize)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) AdminActiveAll(c *gin.Context) {
	var req types.SupportOverviewQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.supportSvc.AdminListActive(c.Request.Context(), middleware.CurrentEmployeeID(c), req.Page, req.PageSize, req.AssigneeEmployeeID, req.Keyword)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) AdminClosedAll(c *gin.Context) {
	var req types.SupportOverviewQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.supportSvc.AdminListClosed(c.Request.Context(), middleware.CurrentEmployeeID(c), req.Page, req.PageSize, req.AssigneeEmployeeID, req.Keyword)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) AdminOverviewSummary(c *gin.Context) {
	result, err := h.supportSvc.AdminOverviewSummary(c.Request.Context(), middleware.CurrentEmployeeID(c))
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) AdminAccept(c *gin.Context) {
	var req types.SupportAcceptRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.supportSvc.AdminAccept(c.Request.Context(), middleware.CurrentEmployeeID(c), req.ConversationID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", item)
}

func (h *SupportHandler) AdminTransfer(c *gin.Context) {
	var req types.SupportTransferRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.supportSvc.AdminTransfer(c.Request.Context(), middleware.CurrentEmployeeID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "ok", nil)
}

func (h *SupportHandler) AdminClose(c *gin.Context) {
	var req types.SupportCloseRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.supportSvc.AdminClose(c.Request.Context(), middleware.CurrentEmployeeID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "ok", nil)
}

func (h *SupportHandler) AdminMessages(c *gin.Context) {
	var req types.SupportMessageQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.supportSvc.AdminListMessages(c.Request.Context(), middleware.CurrentEmployeeID(c), req.ConversationID, req.BeforeID, req.Limit)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *SupportHandler) AdminSend(c *gin.Context) {
	var req types.SupportSendRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.supportSvc.AdminSend(c.Request.Context(), middleware.CurrentEmployeeID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", item)
}

func (h *SupportHandler) AdminRead(c *gin.Context) {
	var req types.SupportReadRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.supportSvc.AdminRead(c.Request.Context(), middleware.CurrentEmployeeID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "ok", nil)
}
