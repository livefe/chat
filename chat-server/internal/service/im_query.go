package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/logger"
)

func (s *IMService) ListContacts(ctx context.Context, selfID uint, req *types.IMContactQuery) (*types.IMContactListResponse, error) {
	status := model.EmployeeStatusEnabled
	f := repository.EmployeeListFilter{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Status:    &status,
		ExcludeID: selfID,
	}
	if req.DepartmentID > 0 {
		depID := req.DepartmentID
		f.DepartmentID = &depID
	}
	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		if isAllDigits(keyword) {
			f.Mobile = keyword
		} else {
			f.Name = keyword
		}
	}

	list, total, err := s.employeeRepo.List(ctx, f)
	if err != nil {
		s.log.Error(ctx, "IM通讯录查询失败", logger.Error(err))
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	depIDs := make([]uint, 0, len(list))
	items := make([]types.IMContactItem, 0, len(list))
	for _, e := range list {
		depIDs = append(depIDs, e.DepartmentID)
		items = append(items, types.IMContactItem{
			ID:           e.ID,
			Name:         e.Name,
			AvatarURL:    e.AvatarURL,
			DepartmentID: e.DepartmentID,
		})
	}
	depNames := s.departmentNames(ctx, depIDs)
	ids := make([]uint, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	online := s.OnlineEmployees(ctx, ids)
	for i := range items {
		items[i].DepartmentName = depNames[items[i].DepartmentID]
		if _, ok := online[items[i].ID]; ok {
			items[i].Online = true
		}
	}
	return &types.IMContactListResponse{Total: total, List: items}, nil
}

func (s *IMService) ListMessages(ctx context.Context, employeeID, conversationID, beforeID uint, limit int) (*types.IMMessageListResponse, error) {
	if _, err := s.requireActiveMember(ctx, conversationID, employeeID); err != nil {
		return nil, err
	}
	msgs, err := s.messageRepo.ListByConversation(ctx, conversationID, beforeID, limit)
	if err != nil {
		s.log.Error(ctx, "IM消息列表失败", logger.Error(err))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items, err := s.toMessageItems(ctx, msgs)
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return &types.IMMessageListResponse{List: items}, nil
}

func (s *IMService) UnreadSummary(ctx context.Context, employeeID uint) (*types.IMUnreadSummaryResponse, error) {
	m, err := s.UnreadMap(ctx, employeeID)
	if err != nil {
		return &types.IMUnreadSummaryResponse{}, nil
	}
	var staffTotal, supportTotal int64
	for field, v := range m {
		id64, err := strconv.ParseUint(field, 10, 64)
		if err != nil || id64 == 0 {
			staffTotal += v
			continue
		}
		conv, err := s.convRepo.GetByID(ctx, uint(id64))
		if err != nil {
			staffTotal += v
			continue
		}
		if conv.Type == model.IMConvTypeSupport {
			supportTotal += v
		} else {
			staffTotal += v
		}
	}
	return &types.IMUnreadSummaryResponse{Total: staffTotal, SupportTotal: supportTotal}, nil
}

type imMessageRefs struct {
	senderIDs []uint
	replyIDs  []uint
	mediaIDs  []uint
}

func isIMMediaContent(contentType uint8) bool {
	return contentType == model.MessageContentTypeImage ||
		contentType == model.MessageContentTypeVideo ||
		contentType == model.MessageContentTypeFile
}

func collectIMMessageRefs(msgs []*model.IMMessage) imMessageRefs {
	refs := imMessageRefs{senderIDs: make([]uint, 0, len(msgs))}
	for _, m := range msgs {
		if m.SenderKind == model.IMPrincipalEmployee {
			refs.senderIDs = append(refs.senderIDs, m.SenderID)
		}
		if m.ReplyToID > 0 {
			refs.replyIDs = append(refs.replyIDs, m.ReplyToID)
		}
		if isIMMediaContent(m.ContentType) && m.Recalled == 0 {
			if id, err := parseStoredMediaID(m.ContentType, m.Content); err == nil && id > 0 {
				refs.mediaIDs = append(refs.mediaIDs, id)
			}
		}
	}
	return refs
}

func (s *IMService) toMessageItems(ctx context.Context, msgs []*model.IMMessage) ([]types.IMMessageItem, error) {
	if len(msgs) == 0 {
		return []types.IMMessageItem{}, nil
	}
	refs := collectIMMessageRefs(msgs)
	empMap := s.senderEmployees(ctx, refs.senderIDs)
	replyMap := s.loadReplyTargets(ctx, refs.replyIDs, empMap)
	urlMap := map[uint]string{}
	if s.mediaSvc != nil && len(refs.mediaIDs) > 0 {
		urlMap = s.mediaSvc.ResolveURLs(ctx, refs.mediaIDs)
	}

	items := make([]types.IMMessageItem, 0, len(msgs))
	for _, m := range msgs {
		items = append(items, toIMMessageItem(m, empMap, replyMap, urlMap))
	}
	return items, nil
}

func (s *IMService) senderEmployees(ctx context.Context, ids []uint) map[uint]*model.Employee {
	employees, err := s.employeeRepo.GetByIDs(ctx, ids)
	if err != nil {
		s.log.Warn(ctx, "IM消息发送人查询失败", logger.Error(err))
	}
	out := make(map[uint]*model.Employee, len(employees))
	for _, e := range employees {
		out[e.ID] = e
	}
	return out
}

func (s *IMService) loadReplyTargets(ctx context.Context, ids []uint, empMap map[uint]*model.Employee) map[uint]*model.IMMessage {
	out := make(map[uint]*model.IMMessage, len(ids))
	for _, id := range ids {
		rm, err := s.messageRepo.GetByID(ctx, id)
		if err != nil {
			continue
		}
		out[id] = rm
		if rm.SenderKind != model.IMPrincipalEmployee {
			continue
		}
		if _, ok := empMap[rm.SenderID]; ok {
			continue
		}
		if e, err := s.employeeRepo.GetByID(ctx, rm.SenderID); err == nil {
			empMap[e.ID] = e
		}
	}
	return out
}

func toIMMessageItem(m *model.IMMessage, empMap map[uint]*model.Employee, replyMap map[uint]*model.IMMessage, urlMap map[uint]string) types.IMMessageItem {
	item := types.IMMessageItem{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderKind:     m.SenderKind,
		SenderID:       m.SenderID,
		ContentType:    m.ContentType,
		Content:        m.Content,
		ReplyToID:      m.ReplyToID,
		CreatedAt:      m.CreatedAt.Unix(),
		Recalled:       m.Recalled == 1,
		ClientMsgID:    m.ClientMsgID,
	}
	if e := empMap[m.SenderID]; e != nil {
		item.SenderName = e.Name
		item.SenderAvatarURL = e.AvatarURL
	}
	if item.Recalled {
		item.Content = ""
	} else if isIMMediaContent(m.ContentType) {
		item.Content = withIMMediaURL(m.Content, urlMap)
	}
	if reply := replyMap[m.ReplyToID]; reply != nil {
		item.ReplyToContentType = reply.ContentType
		if reply.Recalled != 1 {
			item.ReplyToContent = reply.Content
		}
		if e := empMap[reply.SenderID]; e != nil {
			item.ReplyToSenderName = e.Name
		}
	}
	return item
}

func withIMMediaURL(content string, urlMap map[uint]string) string {
	var body messageMediaContent
	if err := json.Unmarshal([]byte(content), &body); err != nil || body.MediaID == 0 {
		return content
	}
	if u := urlMap[body.MediaID]; u != "" {
		body.URL = u
	}
	out, err := json.Marshal(body)
	if err != nil {
		return content
	}
	return string(out)
}
