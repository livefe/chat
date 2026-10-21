package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	log     *logger.Logger
	convSvc *service.ConversationService
}

func NewConversationHandler(log *logger.Logger, convSvc *service.ConversationService) *ConversationHandler {
	return &ConversationHandler{log: log, convSvc: convSvc}
}

func (h *ConversationHandler) Conversations(c *gin.Context) {
	var req types.ConversationListQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.convSvc.List(c.Request.Context(), middleware.CurrentUserID(c), req.Page, req.PageSize)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *ConversationHandler) ConversationUnread(c *gin.Context) {
	summary, err := h.convSvc.UnreadSummary(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", summary)
}

func (h *ConversationHandler) DeleteConversation(c *gin.Context) {
	var req types.PeerIDRequest
	if !bindJSON(c, &req) {
		return
	}
	clearedBefore, err := h.convSvc.Delete(c.Request.Context(), middleware.CurrentUserID(c), req.PeerID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "删除成功", types.DeleteConversationResponse{ClearedBeforeID: clearedBefore})
}

func (h *ConversationHandler) PinConversation(c *gin.Context) {
	var req types.PinConversationRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.convSvc.Pin(c.Request.Context(), middleware.CurrentUserID(c), req.PeerID, req.Pinned); err != nil {
		response.WriteError(c, err)
		return
	}
	if req.Pinned {
		response.Success(c, "已置顶", nil)
		return
	}
	response.Success(c, "已取消置顶", nil)
}

func (h *ConversationHandler) ReadConversation(c *gin.Context) {
	var req types.PeerIDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.convSvc.ClearUnread(c.Request.Context(), middleware.CurrentUserID(c), req.PeerID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已读", nil)
}

func (h *ConversationHandler) AdminConversationList(c *gin.Context) {
	var req types.AdminConversationListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.convSvc.AdminList(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}
