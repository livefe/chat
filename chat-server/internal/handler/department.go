package handler

import (
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type DepartmentHandler struct {
	log     *logger.Logger
	deptSvc *service.DepartmentService
}

func NewDepartmentHandler(log *logger.Logger, deptSvc *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{log: log, deptSvc: deptSvc}
}

func (h *DepartmentHandler) ListDepartments(c *gin.Context) {
	var req types.DepartmentListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.deptSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *DepartmentHandler) DepartmentOptions(c *gin.Context) {
	list, err := h.deptSvc.ListOptions(c.Request.Context())
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", list)
}

func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req types.CreateDepartmentRequest
	if !bindJSON(c, &req) {
		return
	}
	id, err := h.deptSvc.Create(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "创建成功", types.IDResponse{ID: id})
}

func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	var req types.UpdateDepartmentRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.deptSvc.Update(c.Request.Context(), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *DepartmentHandler) UpdateDepartmentStatus(c *gin.Context) {
	var req types.UpdateDepartmentStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.deptSvc.UpdateStatus(c.Request.Context(), req.ID, req.Status); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}
