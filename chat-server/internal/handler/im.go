package handler

import (
	"context"
	"net/http"

	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/auth"
	"app/pkg/logger"
	"app/pkg/response"
	"app/pkg/websocket"

	"github.com/gin-gonic/gin"
	gorillaws "github.com/gorilla/websocket"
)

type IMHandler struct {
	log         *logger.Logger
	tokenAuth   *auth.Auth
	authSvc     *service.AuthService
	employeeSvc *service.EmployeeService
	imSvc       *service.IMService
	imRealtime  *service.IMRealtimeService
	upgrader    *gorillaws.Upgrader
	wsCfg       websocket.Config
}

func NewIMHandler(
	log *logger.Logger,
	tokenAuth *auth.Auth,
	authSvc *service.AuthService,
	employeeSvc *service.EmployeeService,
	imSvc *service.IMService,
	imRealtime *service.IMRealtimeService,
	wsCfg websocket.Config,
) *IMHandler {
	upgrader := websocket.NewUpgrader(&wsCfg, func(origin string) {
		log.Warn(context.Background(), "IM WebSocket upgrade Origin 被拒",
			logger.String("origin", origin),
			logger.Any("allow_origins", wsCfg.AllowOrigins))
	})
	return &IMHandler{
		log:         log,
		tokenAuth:   tokenAuth,
		authSvc:     authSvc,
		employeeSvc: employeeSvc,
		imSvc:       imSvc,
		imRealtime:  imRealtime,
		upgrader:    upgrader,
		wsCfg:       wsCfg,
	}
}

func (h *IMHandler) Upgrade(c *gin.Context) {
	if h.imRealtime.AtCapacity() {
		c.String(http.StatusServiceUnavailable, "服务繁忙")
		return
	}
	token, selectedProto := extractWSAccessToken(c)
	employeeID, reason := h.authorize(c, token)
	if reason != "" {
		closeAfterUpgrade(h.log, h.upgrader, c, selectedProto, wsCloseAuthFailed, reason)
		return
	}

	raw, err := h.upgrader.Upgrade(c.Writer, c.Request, wsProtocolHeader(selectedProto))
	if err != nil {
		h.log.Warn(c.Request.Context(), "IM WebSocket upgrade 失败", logger.Error(err))
		return
	}

	conn := websocket.NewConn(raw, h.wsCfg.ConnConfig, h.imRealtime)
	conn.Set(service.AttrEmployeeID, employeeID)
	if rid, ok := c.Get(logger.TraceIDKey); ok {
		if s, ok := rid.(string); ok && s != "" {
			conn.Set(service.AttrRequestID, s)
		}
	}
	conn.Start()
}

func (h *IMHandler) authorize(c *gin.Context, token string) (uint, string) {
	claims, fail := middleware.ValidateAccessToken(h.tokenAuth, token)
	if fail != nil {
		return 0, fail.Message
	}
	if claims == nil || claims.SubjectType != auth.SubjectTypeEmployee || claims.SubjectID == 0 {
		return 0, "无效的令牌"
	}
	if fail := middleware.AssertAccessTokenNotRevoked(c.Request.Context(), h.authSvc, claims); fail != nil {
		return 0, fail.Message
	}
	if err := h.employeeSvc.EnsureActive(c.Request.Context(), claims.SubjectID); err != nil {
		return 0, "员工不可用"
	}
	return claims.SubjectID, ""
}

func (h *IMHandler) Contacts(c *gin.Context) {
	var req types.IMContactQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.imSvc.ListContacts(c.Request.Context(), middleware.CurrentEmployeeID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *IMHandler) OpenDM(c *gin.Context) {
	var req types.IMOpenDMRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.imSvc.OpenDM(c.Request.Context(), middleware.CurrentEmployeeID(c), req.PeerEmployeeID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", item)
}

func (h *IMHandler) Conversations(c *gin.Context) {
	var req types.PageQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.imSvc.ListConversations(c.Request.Context(), middleware.CurrentEmployeeID(c), req.Page, req.PageSize)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *IMHandler) Messages(c *gin.Context) {
	var req types.IMMessageQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.imSvc.ListMessages(c.Request.Context(), middleware.CurrentEmployeeID(c), req.ConversationID, req.BeforeID, req.Limit)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *IMHandler) Send(c *gin.Context) {
	var req types.IMSendMessageRequest
	if !bindJSON(c, &req) {
		return
	}
	item, _, err := h.imSvc.Send(c.Request.Context(), middleware.CurrentEmployeeID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", item)
}

func (h *IMHandler) Read(c *gin.Context) {
	var req types.IMReadRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.imSvc.Read(c.Request.Context(), middleware.CurrentEmployeeID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "ok", nil)
}

func (h *IMHandler) Recall(c *gin.Context) {
	var req types.IMRecallRequest
	if !bindJSON(c, &req) {
		return
	}
	if _, _, err := h.imSvc.Recall(c.Request.Context(), middleware.CurrentEmployeeID(c), req.MessageID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "ok", nil)
}
