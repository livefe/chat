package handler

import (
	"context"
	"net/http"
	"time"

	"app/internal/middleware"
	"app/internal/service"
	"app/pkg/auth"
	"app/pkg/logger"
	"app/pkg/websocket"

	"github.com/gin-gonic/gin"
	gorillaws "github.com/gorilla/websocket"
)

const wsCloseAuthFailed = 4401

type RealtimeHandler struct {
	log         *logger.Logger
	tokenAuth   *auth.Auth
	authSvc     *service.AuthService
	upgrader    *gorillaws.Upgrader
	realtimeSvc *service.RealtimeService
	wsCfg       websocket.Config
}

func NewRealtimeHandler(
	log *logger.Logger,
	tokenAuth *auth.Auth,
	authSvc *service.AuthService,
	realtimeSvc *service.RealtimeService,
	wsCfg websocket.Config,
) *RealtimeHandler {
	upgrader := websocket.NewUpgrader(&wsCfg, func(origin string) {
		log.Warn(context.Background(), "WebSocket upgrade Origin 被拒",
			logger.String("origin", origin),
			logger.Any("allow_origins", wsCfg.AllowOrigins))
	})
	return &RealtimeHandler{
		log:         log,
		tokenAuth:   tokenAuth,
		authSvc:     authSvc,
		upgrader:    upgrader,
		realtimeSvc: realtimeSvc,
		wsCfg:       wsCfg,
	}
}

func (h *RealtimeHandler) Upgrade(c *gin.Context) {
	if h.realtimeSvc.AtCapacity() {
		h.log.Warn(c.Request.Context(), "WebSocket upgrade 拒绝：实例连接已满",
			logger.String("remote", c.ClientIP()))
		c.String(http.StatusServiceUnavailable, "服务繁忙")
		return
	}
	token, selectedProto := extractWSAccessToken(c)

	claims, fail := middleware.ValidateAccessToken(h.tokenAuth, token)
	if fail != nil {
		origin := c.Request.Header.Get("Origin")
		switch fail.Kind {
		case middleware.TokenErrExpired:
			h.log.Warn(c.Request.Context(), "WebSocket upgrade token 已过期", logger.String("origin", origin))

			closeAfterUpgrade(h.log, h.upgrader, c, selectedProto, wsCloseAuthFailed, "token expired")
		case middleware.TokenErrMissing:
			h.log.Warn(c.Request.Context(), "WebSocket upgrade 拒绝：缺少令牌",
				logger.String("origin", origin),
				logger.String("remote", c.ClientIP()))
			c.String(http.StatusUnauthorized, fail.Message)
		case middleware.TokenErrTooLong:
			h.log.Warn(c.Request.Context(), "WebSocket upgrade 拒绝：令牌过长", logger.Int("len", len(token)))
			c.String(http.StatusUnauthorized, fail.Message)
		default:
			h.log.Warn(c.Request.Context(), "WebSocket upgrade 拒绝：令牌无效",
				logger.String("origin", origin),
				logger.String("reason", fail.Message))
			c.String(http.StatusUnauthorized, fail.Message)
		}
		return
	}
	if claims == nil || claims.SubjectType != auth.SubjectTypeUser || claims.SubjectID == 0 {
		h.log.Warn(c.Request.Context(), "WebSocket upgrade 拒绝：用户无效",
			logger.String("origin", c.Request.Header.Get("Origin")))
		c.String(http.StatusUnauthorized, "无效的令牌")
		return
	}
	if fail := middleware.AssertAccessTokenNotRevoked(c.Request.Context(), h.authSvc, claims); fail != nil {
		h.log.Warn(c.Request.Context(), "WebSocket upgrade 拒绝：令牌已失效",
			logger.Uint("user_id", claims.SubjectID))
		c.String(http.StatusUnauthorized, fail.Message)
		return
	}

	raw, err := h.upgrader.Upgrade(c.Writer, c.Request, wsProtocolHeader(selectedProto))
	if err != nil {
		h.log.Warn(c.Request.Context(), "WebSocket upgrade 失败",
			logger.String("origin", c.Request.Header.Get("Origin")),
			logger.Error(err))
		return
	}

	conn := websocket.NewConn(raw, h.wsCfg.ConnConfig, h.realtimeSvc)
	conn.Set(service.AttrUserID, claims.SubjectID)
	if rid, ok := c.Get(logger.TraceIDKey); ok {
		if s, ok := rid.(string); ok && s != "" {
			conn.Set(service.AttrRequestID, s)
		}
	}
	conn.Start()
}

func extractWSAccessToken(c *gin.Context) (token string, selectedProto string) {
	token = auth.ExtractTokenFromHeader(c.GetHeader(auth.AuthorizationHeader))
	if token != "" {
		return token, ""
	}
	protos := gorillaws.Subprotocols(c.Request)
	for i, p := range protos {
		if p == "bearer" && i+1 < len(protos) {
			return protos[i+1], "bearer"
		}
	}
	return "", ""
}

func wsProtocolHeader(selectedProto string) http.Header {
	if selectedProto == "" {
		return nil
	}
	h := make(http.Header)
	h.Set("Sec-WebSocket-Protocol", selectedProto)
	return h
}

func closeAfterUpgrade(log *logger.Logger, upgrader *gorillaws.Upgrader, c *gin.Context, selectedProto string, code int, reason string) {
	raw, err := upgrader.Upgrade(c.Writer, c.Request, wsProtocolHeader(selectedProto))
	if err != nil {
		log.Warn(c.Request.Context(), "WebSocket upgrade 失败（鉴权失败关闭）",
			logger.String("origin", c.Request.Header.Get("Origin")),
			logger.Error(err))
		return
	}
	_ = raw.WriteControl(
		gorillaws.CloseMessage,
		gorillaws.FormatCloseMessage(code, reason),
		time.Now().Add(2*time.Second),
	)
	_ = raw.Close()
}
