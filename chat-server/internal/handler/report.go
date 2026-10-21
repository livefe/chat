package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	log       *logger.Logger
	reportSvc *service.ReportService
}

func NewReportHandler(log *logger.Logger, reportSvc *service.ReportService) *ReportHandler {
	return &ReportHandler{log: log, reportSvc: reportSvc}
}

func (h *ReportHandler) Submit(c *gin.Context) {
	var req types.SubmitReportRequest
	if !bindJSON(c, &req) {
		return
	}
	userID := middleware.CurrentUserID(c)
	if err := h.reportSvc.Submit(c.Request.Context(), userID, &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "提交成功", nil)
}

func (h *ReportHandler) AdminList(c *gin.Context) {
	var req types.AdminReportListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.reportSvc.AdminList(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *ReportHandler) AdminDetail(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	item, err := h.reportSvc.AdminGet(c.Request.Context(), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", item)
}

func (h *ReportHandler) AdminHandle(c *gin.Context) {
	var req types.AdminReportHandleRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.reportSvc.AdminHandle(c.Request.Context(), middleware.CurrentEmployeeID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "处理成功", nil)
}
