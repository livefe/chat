package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	log     *logger.Logger
	permSvc *service.PermissionService
}

func NewPermissionHandler(log *logger.Logger, permSvc *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{log: log, permSvc: permSvc}
}

func (h *PermissionHandler) Create(c *gin.Context) {
	var req types.CreatePermissionRequest
	if !bindJSON(c, &req) {
		return
	}
	id, err := h.permSvc.AdminCreate(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "创建成功", types.IDResponse{ID: id})
}

func (h *PermissionHandler) Update(c *gin.Context) {
	var req types.UpdatePermissionRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.permSvc.AdminUpdate(c.Request.Context(), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *PermissionHandler) Delete(c *gin.Context) {
	var req types.IDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.permSvc.AdminDelete(c.Request.Context(), req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "删除成功", nil)
}

func (h *PermissionHandler) Tree(c *gin.Context) {
	var req types.PermissionTreeRequest
	if !bindQuery(c, &req) {
		return
	}
	tree, err := h.permSvc.AdminTree(c.Request.Context(), req.Assignable)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", tree)
}

func (h *PermissionHandler) GetEmployeePermissions(c *gin.Context) {
	employeeID := middleware.CurrentEmployeeID(c)
	perms, err := h.permSvc.GetEmployeePermissions(c.Request.Context(), employeeID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", perms)
}
