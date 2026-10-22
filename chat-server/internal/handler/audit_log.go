package handler

import (
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuditLogHandler struct {
	log         *logger.Logger
	auditLogSvc *service.AuditLogService
}

func NewAuditLogHandler(log *logger.Logger, auditLogSvc *service.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{log: log, auditLogSvc: auditLogSvc}
}

func (h *AuditLogHandler) List(c *gin.Context) {
	var req types.AuditLogListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.auditLogSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *AuditLogHandler) Options(c *gin.Context) {
	result, err := h.auditLogSvc.Options(c.Request.Context())
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *AuditLogHandler) EmployeeOptions(c *gin.Context) {
	var req types.AuditEmployeeOptionsRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.auditLogSvc.EmployeeOptions(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *AuditLogHandler) Detail(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.auditLogSvc.Get(c.Request.Context(), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}
