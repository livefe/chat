package handler

import (
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type RegionHandler struct {
	log       *logger.Logger
	regionSvc *service.RegionService
}

func NewRegionHandler(log *logger.Logger, regionSvc *service.RegionService) *RegionHandler {
	return &RegionHandler{log: log, regionSvc: regionSvc}
}

func (h *RegionHandler) Children(c *gin.Context) {
	var req types.RegionChildrenRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.regionSvc.ListChildren(c.Request.Context(), req.ParentCode)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *RegionHandler) Search(c *gin.Context) {
	var req types.RegionSearchRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.regionSvc.Search(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *RegionHandler) AdminTrail(c *gin.Context) {
	var req types.RegionDetailRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.regionSvc.Trail(c.Request.Context(), req.Code)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *RegionHandler) Detail(c *gin.Context) {
	var req types.RegionDetailRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.regionSvc.Get(c.Request.Context(), req.Code)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *RegionHandler) AdminDetail(c *gin.Context) {
	var req types.RegionDetailRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.regionSvc.AdminGet(c.Request.Context(), req.Code)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}
