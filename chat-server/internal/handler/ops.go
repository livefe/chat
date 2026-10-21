package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type OpsHandler struct {
	log    *logger.Logger
	opsSvc *service.OpsService
}

func NewOpsHandler(log *logger.Logger, opsSvc *service.OpsService) *OpsHandler {
	return &OpsHandler{log: log, opsSvc: opsSvc}
}

func (h *OpsHandler) PendingCounts(c *gin.Context) {
	result, err := h.opsSvc.PendingCounts(c.Request.Context(), middleware.CurrentEmployeeID(c))
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}
