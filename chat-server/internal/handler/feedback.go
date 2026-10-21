package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type FeedbackHandler struct {
	log         *logger.Logger
	feedbackSvc *service.FeedbackService
}

func NewFeedbackHandler(log *logger.Logger, feedbackSvc *service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{log: log, feedbackSvc: feedbackSvc}
}

func (h *FeedbackHandler) Submit(c *gin.Context) {
	var req types.SubmitFeedbackRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.feedbackSvc.Submit(c.Request.Context(), middleware.CurrentUserID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "提交成功", nil)
}

func (h *FeedbackHandler) MyList(c *gin.Context) {
	var req types.MyFeedbackListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.feedbackSvc.ListMine(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *FeedbackHandler) MyDetail(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	item, err := h.feedbackSvc.GetMine(c.Request.Context(), middleware.CurrentUserID(c), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", item)
}

func (h *FeedbackHandler) AdminList(c *gin.Context) {
	var req types.AdminFeedbackListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.feedbackSvc.AdminList(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *FeedbackHandler) AdminDetail(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	item, err := h.feedbackSvc.AdminGet(c.Request.Context(), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", item)
}

func (h *FeedbackHandler) AdminUpdateStatus(c *gin.Context) {
	var req types.AdminFeedbackUpdateStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.feedbackSvc.AdminUpdateStatus(c.Request.Context(), middleware.CurrentEmployeeID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}
