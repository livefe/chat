package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	log     *logger.Logger
	roomSvc *service.RoomService
}

func NewRoomHandler(log *logger.Logger, roomSvc *service.RoomService) *RoomHandler {
	return &RoomHandler{log: log, roomSvc: roomSvc}
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req types.CreateRoomRequest
	if !bindJSON(c, &req) {
		return
	}
	id, err := h.roomSvc.Create(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "创建成功，等待审核", types.IDResponse{ID: id})
}

func (h *RoomHandler) MyRoomApplications(c *gin.Context) {
	var req types.MyRoomApplicationsQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.roomSvc.ListMyApplications(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *RoomHandler) RoomDetail(c *gin.Context) {
	var req types.IDRequest
	if !bindURI(c, &req) {
		return
	}
	detail, err := h.roomSvc.Get(c.Request.Context(), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", detail)
}

func (h *RoomHandler) NearbyBootstrap(c *gin.Context) {
	var req types.NearbyBootstrapQuery
	if !bindQuery(c, &req) {
		return
	}
	userID := middleware.CurrentUserID(c)
	data, err := h.roomSvc.NearbyBootstrap(c.Request.Context(), userID, req.Lat, req.Lng, req.Radius, req.Limit, req.WithMessages)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", data)
}

func (h *RoomHandler) RoomBootstrap(c *gin.Context) {
	var req types.RoomBootstrapQuery
	if !bindQuery(c, &req) {
		return
	}
	userID := middleware.CurrentUserID(c)
	data, err := h.roomSvc.RoomBootstrap(c.Request.Context(), userID, req.RoomID, req.Lat, req.Lng, req.Limit, req.WithMessages)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", data)
}

func (h *RoomHandler) AdminRoomList(c *gin.Context) {
	var req types.RoomListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.roomSvc.AdminList(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *RoomHandler) AdminRoomDetail(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	detail, err := h.roomSvc.AdminGet(c.Request.Context(), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", detail)
}

func (h *RoomHandler) AdminRoomCreate(c *gin.Context) {
	var req types.CreateRoomRequest
	if !bindJSON(c, &req) {
		return
	}
	id, err := h.roomSvc.AdminCreate(c.Request.Context(), middleware.CurrentEmployeeID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "创建成功", types.IDResponse{ID: id})
}

func (h *RoomHandler) AdminRoomUpdate(c *gin.Context) {
	var req types.UpdateRoomRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.roomSvc.AdminUpdate(c.Request.Context(), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *RoomHandler) AdminRoomAudit(c *gin.Context) {
	var req types.AuditRoomRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.roomSvc.AdminAudit(c.Request.Context(), middleware.CurrentEmployeeID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "审核完成", nil)
}

func (h *RoomHandler) AdminRoomBatchAudit(c *gin.Context) {
	var req types.BatchAuditRoomRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.roomSvc.AdminBatchAudit(c.Request.Context(), middleware.CurrentEmployeeID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "批量审核完成", result)
}

func (h *RoomHandler) AdminRoomNeighbors(c *gin.Context) {
	var req types.AdminRoomNeighborsQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.roomSvc.AdminListNeighbors(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *RoomHandler) AdminRoomMap(c *gin.Context) {
	var req types.AdminRoomMapQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.roomSvc.AdminMap(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *RoomHandler) AdminRoomUpdateStatus(c *gin.Context) {
	var req types.UpdateRoomStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.roomSvc.AdminUpdateStatus(c.Request.Context(), req.ID, req.Status); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}
