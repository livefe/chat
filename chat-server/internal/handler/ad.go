package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type AdHandler struct {
	log *logger.Logger
	svc *service.AdService
}

func NewAdHandler(log *logger.Logger, svc *service.AdService) *AdHandler {
	return &AdHandler{log: log, svc: svc}
}

func (h *AdHandler) ReportEvents(c *gin.Context) {
	var req types.AdEventsRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.ReportEvents(c.Request.Context(), middleware.CurrentUserID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", nil)
}

func (h *AdHandler) AdminList(c *gin.Context) {
	var req types.AdminAdCampaignListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.svc.AdminList(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *AdHandler) AdminDetail(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.svc.AdminGet(c.Request.Context(), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *AdHandler) AdminSave(c *gin.Context) {
	var req types.AdminAdCampaignSaveRequest
	if !bindJSON(c, &req) {
		return
	}
	id, err := h.svc.AdminSave(c.Request.Context(), middleware.CurrentEmployeeID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "保存成功", types.IDResponse{ID: id})
}

func (h *AdHandler) AdminUpdateStatus(c *gin.Context) {
	var req types.AdminAdUpdateStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.AdminUpdateStatus(c.Request.Context(), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *AdHandler) AdminAudit(c *gin.Context) {
	var req types.AdminAdAuditRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.AdminAudit(c.Request.Context(), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "审核完成", nil)
}

func (h *AdHandler) AdminBatchPause(c *gin.Context) {
	var req types.AdminAdBatchPauseRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.svc.AdminBatchPause(c.Request.Context(), req.IDs)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *AdHandler) AdminAdvertisers(c *gin.Context) {
	result, err := h.svc.AdminListAdvertisers(c.Request.Context())
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *AdHandler) AdminCopy(c *gin.Context) {
	var req types.AdminAdCopyRequest
	if !bindJSON(c, &req) {
		return
	}
	id, err := h.svc.AdminCopy(c.Request.Context(), middleware.CurrentEmployeeID(c), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "复制成功", types.IDResponse{ID: id})
}

func (h *AdHandler) AdminGetSetting(c *gin.Context) {
	response.Success(c, "", h.svc.AdminGetSetting(c.Request.Context()))
}

func (h *AdHandler) AdminSaveSetting(c *gin.Context) {
	var req types.AdminAdSetting
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.AdminUpdateSetting(c.Request.Context(), req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "保存成功", nil)
}

func (h *AdHandler) AdminDailyStats(c *gin.Context) {
	var req types.AdminAdDailyStatRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.svc.AdminListDailyStats(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}
