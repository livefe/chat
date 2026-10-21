package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/logger"
	"app/pkg/websocket"

	"github.com/redis/go-redis/v9"
)

const AttrEmployeeID = "employee_id"

type imHub struct {
	log        *logger.Logger
	cache      *cache.Cache
	instanceID string
	broker     *websocket.RedisBroker
	hub        *websocket.Hub[uint]
	maxConns   int
}

func NewIMHub(log *logger.Logger, c *cache.Cache, cfg *websocket.Config, instanceID string) (*imHub, error) {
	broker := websocket.NewRedisBroker(c.Client, instanceID+"-im")
	h := &imHub{
		log:        log,
		cache:      c,
		instanceID: instanceID,
		broker:     broker,
		hub:        websocket.NewHub[uint](),
		maxConns:   cfg.MaxConnections,
	}
	if err := h.broker.Subscribe(types.CacheWSEmployeePatternKey, h.handleBrokerMsg); err != nil {
		_ = broker.Close(context.Background())
		return nil, fmt.Errorf("IM WebSocket broker 订阅失败: %w", err)
	}
	return h, nil
}

type IMRealtimeService struct {
	*imHub
	imSvc    *IMService
	permSvc  *PermissionService
	auditSvc *AuditLogService
}

func NewIMRealtimeService(hub *imHub, imSvc *IMService, permSvc *PermissionService, auditSvc *AuditLogService) *IMRealtimeService {
	return &IMRealtimeService{
		imHub:    hub,
		imSvc:    imSvc,
		permSvc:  permSvc,
		auditSvc: auditSvc,
	}
}

func (s *imHub) Close() error {
	n := s.hub.CloseAll()
	if n > 0 {
		s.log.Info(context.Background(), "IM WebSocket 服务关闭", logger.Int("closed_conns", n))
	}
	return s.broker.Close(context.Background())
}

func (s *imHub) AtCapacity() bool {
	return s.maxConns > 0 && s.hub.ConnCount() >= s.maxConns
}

func (s *imHub) OnlineEmployees(ctx context.Context, ids []uint) map[uint]struct{} {
	out := make(map[uint]struct{}, len(ids))
	if s.cache == nil || len(ids) == 0 {
		return out
	}
	queued := false
	pipe := s.cache.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(ids))
	for i, id := range ids {
		if id == 0 {
			continue
		}
		cmds[i] = pipe.HGetAll(ctx, fmt.Sprintf(types.CacheWSEmployeePresenceKey, id))
		queued = true
	}
	if !queued {
		return out
	}
	if _, err := pipe.Exec(ctx); err != nil {
		s.log.Warn(ctx, "查询员工在线失败", logger.Error(err))
		return out
	}
	for i, id := range ids {
		if cmds[i] == nil {
			continue
		}
		all, err := cmds[i].Result()
		if err != nil || !employeePresenceOnline(all) {
			continue
		}
		out[id] = struct{}{}
	}
	return out
}

func (s *imHub) markEmployeeOnline(ctx context.Context, employeeID uint) {
	if employeeID == 0 || s.cache == nil {
		return
	}
	key := fmt.Sprintf(types.CacheWSEmployeePresenceKey, employeeID)
	now := strconv.FormatInt(time.Now().Unix(), 10)
	pipe := s.cache.Pipeline()
	pipe.HSet(ctx, key, s.instanceID, now)
	pipe.Expire(ctx, key, types.Expire45s*2)
	if _, err := pipe.Exec(ctx); err != nil {
		s.log.Warn(ctx, "员工上线写入失败", logger.Uint("employee_id", employeeID), logger.Error(err))
	}
}

func (s *imHub) clearEmployeeOnline(ctx context.Context, employeeID uint) {
	if employeeID == 0 || s.cache == nil {
		return
	}
	key := fmt.Sprintf(types.CacheWSEmployeePresenceKey, employeeID)
	if err := s.cache.HDel(ctx, key, s.instanceID).Err(); err != nil {
		s.log.Warn(ctx, "清理员工在线失败", logger.Uint("employee_id", employeeID), logger.Error(err))
		return
	}
	all, err := s.cache.HGetAll(ctx, key).Result()
	if err != nil || len(all) == 0 {
		_ = s.cache.Del(ctx, key).Err()
		return
	}
	now := time.Now().Unix()
	ttlSec := int64(types.Expire45s.Seconds())
	stale := make([]string, 0)
	alive := 0
	for field, val := range all {
		ts, parseErr := strconv.ParseInt(val, 10, 64)
		if parseErr != nil || now-ts > ttlSec {
			stale = append(stale, field)
			continue
		}
		alive++
	}
	if len(stale) > 0 {
		_ = s.cache.HDel(ctx, key, stale...).Err()
	}
	if alive == 0 {
		_ = s.cache.Del(ctx, key).Err()
	}
}

func employeePresenceOnline(all map[string]string) bool {
	if len(all) == 0 {
		return false
	}
	now := time.Now().Unix()
	ttlSec := int64(types.Expire45s.Seconds())
	for _, val := range all {
		ts, err := strconv.ParseInt(val, 10, 64)
		if err == nil && now-ts <= ttlSec {
			return true
		}
	}
	return false
}

func (s *IMRealtimeService) OnOpen(c *websocket.Conn) {
	ctx := context.Background()
	eid := connEmployeeID(c)
	if !s.hub.RegisterIfUnder(eid, c, s.maxConns) {
		s.log.Warn(ctx, "IM WebSocket 连接数已满",
			logger.Uint("employee_id", eid),
			logger.Int("max_connections", s.maxConns))
		c.Close()
		return
	}
	s.markEmployeeOnline(ctx, eid)
	s.log.Info(ctx, "IM WebSocket 已连接", logger.Uint("employee_id", eid), logger.String("remote", c.RemoteAddr()))
}

func (s *IMRealtimeService) OnClose(c *websocket.Conn) {
	eid := connEmployeeID(c)
	s.hub.Unregister(c)
	remaining := s.hub.CountByID(eid)
	if remaining == 0 {
		s.clearEmployeeOnline(context.Background(), eid)
	}
	s.log.Info(context.Background(), "IM WebSocket 已断开",
		logger.Uint("employee_id", eid),
		logger.String("remote", c.RemoteAddr()),
		logger.Int("remaining_conns", remaining))
}

func (s *IMRealtimeService) OnMessage(c *websocket.Conn, raw []byte) {
	var msg types.IMClientMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		s.replyError(c, "消息格式错误", "")
		return
	}
	ctx := context.Background()
	eid := connEmployeeID(c)
	switch msg.Type {
	case types.IMMsgTypePing:
		s.markEmployeeOnline(ctx, eid)
		return
	case types.IMMsgTypeSend:
		if s.permSvc != nil {
			ok, err := s.permSvc.CheckPermission(ctx, eid, "/api/admin/im/messages", "POST")
			if err != nil || !ok {
				s.replyError(c, "无发送权限", msg.ClientMsgID)
				return
			}
		}
		item, _, err := s.imSvc.Send(ctx, eid, &types.IMSendMessageRequest{
			ConversationID: msg.ConversationID,
			ContentType:    msg.ContentType,
			Content:        msg.Content,
			ReplyToID:      msg.ReplyToID,
			ClientMsgID:    msg.ClientMsgID,
		})
		if s.auditSvc != nil {
			body, _ := json.Marshal(map[string]any{
				"conversationId": msg.ConversationID,
				"contentType":    msg.ContentType,
				"clientMsgId":    msg.ClientMsgID,
			})
			s.auditSvc.Record(ctx, AuditLogInput{
				EmployeeID:    eid,
				RequestPath:   "/api/admin/im/messages",
				RequestMethod: "POST",
				RequestBody:   body,
				Success:       err == nil,
			})
		}
		if err != nil {
			s.replyError(c, clientErrorText(ctx, s.log, err, "发送失败"), msg.ClientMsgID)
			return
		}
		ack, _ := json.Marshal(types.IMServerMsg{
			Type:        types.IMMsgTypeAck,
			ClientMsgID: msg.ClientMsgID,
			ID:          item.ID,
		})
		_ = c.Send(ack)
	default:
		s.replyError(c, "未知的消息类型", msg.ClientMsgID)
	}
}

func (s *imHub) BroadcastRaw(ctx context.Context, payload []byte) {
	if len(payload) == 0 {
		return
	}
	s.deliverLocal(payload)
	_ = s.broker.Publish(ctx, types.CacheWSEmployeeBroadcastChannel, payload)
}

func (s *imHub) deliverLocal(payload []byte) {
	for id := range s.hub.SnapshotByUser() {
		s.hub.SendToID(id, payload)
	}
}

func (s *imHub) DeliverMessage(_ context.Context, employeeIDs []uint, item *types.IMMessageItem) {
	if item == nil || len(employeeIDs) == 0 {
		return
	}
	payload, err := json.Marshal(types.IMServerMsg{
		Type: types.IMMsgTypeMessage,
		Data: item,
	})
	if err != nil {
		return
	}
	for _, id := range employeeIDs {
		s.deliverEmployee(context.Background(), id, payload)
	}
}

func (s *imHub) DeliverRecall(_ context.Context, employeeIDs []uint, conversationID, messageID uint) {
	payload, err := json.Marshal(types.IMServerMsg{
		Type:           types.IMMsgTypeRecall,
		ID:             messageID,
		ConversationID: conversationID,
	})
	if err != nil {
		return
	}
	for _, id := range employeeIDs {
		s.deliverEmployee(context.Background(), id, payload)
	}
}

func (s *imHub) deliverEmployee(ctx context.Context, employeeID uint, payload []byte) {
	if employeeID == 0 {
		return
	}
	s.hub.SendToID(employeeID, payload)
	ch := fmt.Sprintf(types.CacheWSEmployeeChannel, employeeID)
	_ = s.broker.Publish(ctx, ch, payload)
}

func (s *imHub) handleBrokerMsg(channel string, payload []byte) {
	if channel == types.CacheWSEmployeeBroadcastChannel {
		s.deliverLocal(payload)
		return
	}
	prefix := fmt.Sprintf(types.CacheWSEmployeeChannel, 0)
	prefix = prefix[:len(prefix)-1]
	if len(channel) <= len(prefix) || channel[:len(prefix)] != prefix {
		return
	}
	id, err := strconv.ParseUint(channel[len(prefix):], 10, 64)
	if err != nil {
		return
	}
	s.hub.SendToID(uint(id), payload)
}

func (s *IMRealtimeService) replyError(c *websocket.Conn, content, clientMsgID string) {
	payload, _ := json.Marshal(types.IMServerMsg{
		Type:        types.IMMsgTypeError,
		Content:     content,
		ClientMsgID: clientMsgID,
	})
	_ = c.Send(payload)
}

func connEmployeeID(c *websocket.Conn) uint {
	if v, ok := c.Get(AttrEmployeeID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}
