package handler

import (
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type SensitiveWordHandler struct {
	log *logger.Logger
	svc *service.SensitiveWordService
}

func NewSensitiveWordHandler(log *logger.Logger, svc *service.SensitiveWordService) *SensitiveWordHandler {
	return &SensitiveWordHandler{log: log, svc: svc}
}

func (h *SensitiveWordHandler) List(c *gin.Context) {
	var req types.SensitiveWordListRequest
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

func (h *SensitiveWordHandler) Create(c *gin.Context) {
	var req types.CreateSensitiveWordRequest
	if !bindJSON(c, &req) {
		return
	}
	id, err := h.svc.AdminCreate(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "创建成功", types.IDResponse{ID: id})
}

func (h *SensitiveWordHandler) Update(c *gin.Context) {
	var req types.UpdateSensitiveWordRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.AdminUpdate(c.Request.Context(), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *SensitiveWordHandler) UpdateStatus(c *gin.Context) {
	var req types.UpdateSensitiveWordStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.AdminUpdateStatus(c.Request.Context(), req.ID, req.Status); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *SensitiveWordHandler) Delete(c *gin.Context) {
	var req types.IDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.AdminDelete(c.Request.Context(), req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "删除成功", nil)
}

func (h *SensitiveWordHandler) BatchDelete(c *gin.Context) {
	var req types.BatchSensitiveWordRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.AdminBatchDelete(c.Request.Context(), req.IDs); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "删除成功", nil)
}

func (h *SensitiveWordHandler) BatchUpdateStatus(c *gin.Context) {
	var req types.BatchSensitiveWordStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.AdminBatchUpdateStatus(c.Request.Context(), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *SensitiveWordHandler) Import(c *gin.Context) {
	var req types.ImportSensitiveWordsRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.svc.AdminImport(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "导入完成", result)
}
