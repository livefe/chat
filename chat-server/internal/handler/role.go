package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	log     *logger.Logger
	roleSvc *service.RoleService
}

func NewRoleHandler(log *logger.Logger, roleSvc *service.RoleService) *RoleHandler {
	return &RoleHandler{log: log, roleSvc: roleSvc}
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req types.CreateRoleRequest
	if !bindJSON(c, &req) {
		return
	}
	id, err := h.roleSvc.Create(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "创建成功", types.IDResponse{ID: id})
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	var req types.UpdateRoleRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.roleSvc.Update(c.Request.Context(), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {
	var req types.IDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.roleSvc.Delete(c.Request.Context(), req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "删除成功", nil)
}

func (h *RoleHandler) RoleDetail(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	role, err := h.roleSvc.Get(c.Request.Context(), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", role)
}

func (h *RoleHandler) ListRoles(c *gin.Context) {
	var req types.RoleListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.roleSvc.List(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *RoleHandler) RolePermissions(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	perms, err := h.roleSvc.ListPermissions(c.Request.Context(), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", perms)
}

func (h *RoleHandler) AssignRolePermissions(c *gin.Context) {
	var req types.AssignPermissionsRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.roleSvc.AssignPermissions(c.Request.Context(), middleware.CurrentEmployeeID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "分配成功", nil)
}
