package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	log    *logger.Logger
	msgSvc *service.MessageService
}

func NewMessageHandler(log *logger.Logger, msgSvc *service.MessageService) *MessageHandler {
	return &MessageHandler{log: log, msgSvc: msgSvc}
}

func (h *MessageHandler) RoomMessages(c *gin.Context) {
	var req types.RoomMessageQuery
	if !bindQuery(c, &req) {
		return
	}
	list, err := h.msgSvc.ListRoom(c.Request.Context(), middleware.CurrentUserID(c), req.RoomID, req.Lat, req.Lng, req.BeforeID, req.Limit)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", types.MessageListResponse{List: list})
}

func (h *MessageHandler) PostMessages(c *gin.Context) {
	var req types.PostMessageQuery
	if !bindQuery(c, &req) {
		return
	}
	list, err := h.msgSvc.ListPost(c.Request.Context(), middleware.CurrentUserID(c), req.PostID, req.BeforeID, req.Limit)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", types.MessageListResponse{List: list})
}

func (h *MessageHandler) NearbyMessages(c *gin.Context) {
	var req types.NearbyMessageQuery
	if !bindQuery(c, &req) {
		return
	}
	list, err := h.msgSvc.ListNearby(c.Request.Context(), req.Lat, req.Lng, req.Radius, req.BeforeID, req.Limit)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", types.MessageListResponse{List: list})
}

func (h *MessageHandler) PrivateMessages(c *gin.Context) {
	var req types.PrivateMessageQuery
	if !bindQuery(c, &req) {
		return
	}
	userID := middleware.CurrentUserID(c)
	list, clearedBefore, err := h.msgSvc.ListPrivate(c.Request.Context(), userID, req.PeerID, req.BeforeID, req.Limit)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", types.PrivateMessagesResponse{
		List:            list,
		ClearedBeforeID: clearedBefore,
	})
}

func (h *MessageHandler) RecallMessage(c *gin.Context) {
	var req types.IDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.msgSvc.Recall(c.Request.Context(), middleware.CurrentUserID(c), req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已撤回", nil)
}

func (h *MessageHandler) AdminMessageList(c *gin.Context) {
	var req types.AdminMessageListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.msgSvc.AdminList(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *MessageHandler) AdminMessageContext(c *gin.Context) {
	var req types.AdminMessageContextRequest
	if !bindQuery(c, &req) {
		return
	}
	list, err := h.msgSvc.AdminGetContext(c.Request.Context(), req.ID, req.Limit)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", types.AdminMessageContextResponse{List: list})
}

func (h *MessageHandler) AdminMessageDelete(c *gin.Context) {
	var req types.IDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.msgSvc.AdminDelete(c.Request.Context(), req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "删除成功", nil)
}
