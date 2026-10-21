package service

import (
	"context"
	"encoding/json"
	"fmt"

	"app/internal/model"
	"app/internal/types"
)

func (s *SupportService) fanoutMessage(ctx context.Context, conversationID uint, senderKind uint8, senderID uint, item *types.IMMessageItem) {
	members, _ := s.imSvc.ListActiveMembers(ctx, conversationID)
	empSeen := map[uint]struct{}{}
	empIDs := make([]uint, 0, 2)
	var userID uint
	for _, m := range members {
		switch m.PrincipalKind {
		case model.IMPrincipalEmployee:
			if _, ok := empSeen[m.PrincipalID]; ok {
				continue
			}
			empSeen[m.PrincipalID] = struct{}{}
			empIDs = append(empIDs, m.PrincipalID)
			if !(senderKind == model.IMPrincipalEmployee && m.PrincipalID == senderID) {
				_ = s.imSvc.IncrUnread(ctx, m.PrincipalID, conversationID)
			}
		case model.IMPrincipalUser:
			userID = m.PrincipalID
			if !(senderKind == model.IMPrincipalUser && m.PrincipalID == senderID) {
				_ = s.incrUserUnread(ctx, m.PrincipalID)
			}
		}
	}

	if st, err := s.supportStateRepo.GetByConversationID(ctx, conversationID); err == nil && st != nil {
		if userID == 0 && st.UserID > 0 {
			userID = st.UserID
			if !(senderKind == model.IMPrincipalUser && st.UserID == senderID) {
				_ = s.incrUserUnread(ctx, st.UserID)
			}
		}
		if st.AssigneeEmployeeID > 0 {
			if _, ok := empSeen[st.AssigneeEmployeeID]; !ok {
				empSeen[st.AssigneeEmployeeID] = struct{}{}
				empIDs = append(empIDs, st.AssigneeEmployeeID)
				if !(senderKind == model.IMPrincipalEmployee && st.AssigneeEmployeeID == senderID) {
					_ = s.imSvc.IncrUnread(ctx, st.AssigneeEmployeeID, conversationID)
				}
			}
		}
	}
	if s.employees != nil && len(empIDs) > 0 {
		s.employees.DeliverMessage(ctx, empIDs, item)
	}
	if s.users != nil && userID > 0 {

		wrapped, _ := json.Marshal(map[string]any{
			"type":  types.MsgTypeMessage,
			"scope": types.ScopeSupport,
			"data":  item,
		})
		s.users.DeliverToUser(ctx, userID, wrapped)
	}
}

func (s *SupportService) broadcastQueueUpdate(ctx context.Context) {
	if s.employees == nil {
		return
	}
	payload, err := json.Marshal(types.IMServerMsg{Type: types.IMMsgTypeSupportQueue})
	if err != nil {
		return
	}
	s.employees.BroadcastRaw(ctx, payload)
}

func (s *SupportService) notifyUserSession(ctx context.Context, userID, conversationID uint) {
	if s.users == nil || userID == 0 {
		return
	}
	item, err := s.sessionItemForUser(ctx, userID, conversationID)
	if err != nil {
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"type": "support.session",
		"data": item,
	})
	s.users.DeliverToUser(ctx, userID, payload)
}

func (s *SupportService) incrUserUnread(ctx context.Context, userID uint) error {
	key := fmt.Sprintf(types.CacheSupportUnreadKey, userID)
	pipe := s.cache.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, types.Expire1d*31)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *SupportService) clearUserUnread(ctx context.Context, userID uint) error {
	return s.cache.Del(ctx, fmt.Sprintf(types.CacheSupportUnreadKey, userID)).Err()
}

func (s *SupportService) getUserUnread(ctx context.Context, userID uint) int64 {
	n, err := s.cache.Get(ctx, fmt.Sprintf(types.CacheSupportUnreadKey, userID)).Int64()
	if err != nil {
		return 0
	}
	return n
}
