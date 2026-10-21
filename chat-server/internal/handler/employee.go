package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type EmployeeHandler struct {
	log         *logger.Logger
	employeeSvc *service.EmployeeService
}

func NewEmployeeHandler(log *logger.Logger, employeeSvc *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{log: log, employeeSvc: employeeSvc}
}

func (h *EmployeeHandler) Detail(c *gin.Context) {
	employeeID := middleware.CurrentEmployeeID(c)
	info, err := h.employeeSvc.Get(c.Request.Context(), employeeID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", info)
}

func (h *EmployeeHandler) UpdateProfile(c *gin.Context) {
	var req types.UpdateMyProfileRequest
	if !bindJSON(c, &req) {
		return
	}
	info, err := h.employeeSvc.UpdateProfile(c.Request.Context(), middleware.CurrentEmployeeID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "保存成功", info)
}

func (h *EmployeeHandler) List(c *gin.Context) {
	var req types.EmployeeListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.employeeSvc.AdminList(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *EmployeeHandler) Create(c *gin.Context) {
	var req types.CreateEmployeeRequest
	if !bindJSON(c, &req) {
		return
	}
	id, err := h.employeeSvc.AdminCreate(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "创建成功", types.IDResponse{ID: id})
}

func (h *EmployeeHandler) Update(c *gin.Context) {
	var req types.UpdateEmployeeRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.employeeSvc.AdminUpdate(c.Request.Context(), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *EmployeeHandler) UpdateStatus(c *gin.Context) {
	var req types.UpdateEmployeeStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.employeeSvc.AdminUpdateStatus(c.Request.Context(), middleware.CurrentEmployeeID(c), req.ID, req.Status); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *EmployeeHandler) GetRoles(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	roles, err := h.employeeSvc.AdminListRoles(c.Request.Context(), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", roles)
}

func (h *EmployeeHandler) AssignRoles(c *gin.Context) {
	var req types.AssignEmployeeRolesRequest
	if !bindJSON(c, &req) {
		return
	}
	operatorID := middleware.CurrentEmployeeID(c)
	if err := h.employeeSvc.AdminAssignRoles(c.Request.Context(), operatorID, &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "分配成功", nil)
}
