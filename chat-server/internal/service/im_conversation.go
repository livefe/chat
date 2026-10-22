package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

func (s *IMService) OpenDM(ctx context.Context, selfID, peerID uint) (*types.IMConversationItem, error) {
	if selfID == 0 || peerID == 0 || selfID == peerID {
		return nil, errs.Invalid("无效的会话对象")
	}
	peer, err := s.employeeRepo.GetByID(ctx, peerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("员工不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	if peer.Status != model.EmployeeStatusEnabled {
		return nil, errs.Invalid("对方账号已禁用")
	}

	minID, maxID := orderPair(selfID, peerID)
	pair, err := s.staffDMPairRepo.GetByEmployees(ctx, minID, maxID)
	if err == nil && pair != nil {
		return s.conversationItemFor(ctx, selfID, pair.ConversationID, peer)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.log.Error(ctx, "查询员工私聊对失败", logger.Error(err))
		return nil, fmt.Errorf("打开会话失败: %w", err)
	}

	now := time.Now().Unix()
	conv := &model.IMConversation{
		Type:         model.IMConvTypeStaffDM,
		Status:       model.IMConvStatusActive,
		LastActiveAt: now,
	}
	if err := s.convRepo.Create(ctx, conv); err != nil {
		s.log.Error(ctx, "创建IM会话失败", logger.Error(err))
		return nil, fmt.Errorf("打开会话失败: %w", err)
	}
	members := []*model.IMMember{
		{ConversationID: conv.ID, PrincipalKind: model.IMPrincipalEmployee, PrincipalID: selfID, Role: model.IMMemberRoleMember, JoinedAt: now},
		{ConversationID: conv.ID, PrincipalKind: model.IMPrincipalEmployee, PrincipalID: peerID, Role: model.IMMemberRoleMember, JoinedAt: now},
	}
	if err := s.memberRepo.CreateBatch(ctx, members); err != nil {
		s.log.Error(ctx, "创建IM成员失败", logger.Error(err))
		return nil, fmt.Errorf("打开会话失败: %w", err)
	}
	pair = &model.IMStaffDMPair{EmpMinID: minID, EmpMaxID: maxID, ConversationID: conv.ID}
	if err := s.staffDMPairRepo.Create(ctx, pair); err != nil {
		s.log.Error(ctx, "创建员工私聊对失败", logger.Error(err))
		return nil, fmt.Errorf("打开会话失败: %w", err)
	}

	if existing, err := s.staffDMPairRepo.GetByEmployees(ctx, minID, maxID); err == nil && existing != nil {
		return s.conversationItemFor(ctx, selfID, existing.ConversationID, peer)
	}
	return s.conversationItemFor(ctx, selfID, conv.ID, peer)
}

func (s *IMService) ListConversations(ctx context.Context, employeeID uint, req *types.IMConversationQuery) (*types.IMConversationListResponse, error) {
	rows, total, err := s.memberRepo.ListConversationsByEmployee(ctx, employeeID, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		s.log.Error(ctx, "IM会话列表失败", logger.Error(err))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	peerIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		if row.PeerID > 0 {
			peerIDs = append(peerIDs, row.PeerID)
		}
	}
	peers, _ := s.employeeRepo.GetByIDs(ctx, peerIDs)
	peerMap := make(map[uint]*model.Employee, len(peers))
	depIDs := make([]uint, 0, len(peers))
	for _, p := range peers {
		peerMap[p.ID] = p
		depIDs = append(depIDs, p.DepartmentID)
	}
	depNames := s.departmentNames(ctx, depIDs)
	online := s.OnlineEmployees(ctx, peerIDs)
	unreadMap, _ := s.UnreadMap(ctx, employeeID)

	list := make([]types.IMConversationItem, 0, len(rows))
	for _, row := range rows {
		item := types.IMConversationItem{
			ID:            row.ID,
			Type:          row.Type,
			Title:         row.Title,
			PeerID:        row.PeerID,
			LastMessageID: row.LastMessageID,
			LastContent:   row.LastContent,
			LastActiveAt:  row.LastActiveAt,
			UnreadCount:   unreadMap[strconv.FormatUint(uint64(row.ID), 10)],
			Pinned:        row.Pinned,
		}
		if p := peerMap[row.PeerID]; p != nil {
			item.PeerName = p.Name
			item.PeerAvatarURL = p.AvatarURL
			item.PeerDepartmentName = depNames[p.DepartmentID]
			_, item.PeerOnline = online[p.ID]
			if item.Title == "" {
				item.Title = p.Name
			}
		}
		list = append(list, item)
	}
	return &types.IMConversationListResponse{Total: total, List: list}, nil
}

func (s *IMService) Read(ctx context.Context, employeeID uint, req *types.IMReadRequest) error {
	if req == nil || req.ConversationID == 0 {
		return errs.Invalid("参数错误")
	}
	if _, err := s.requireActiveMember(ctx, req.ConversationID, employeeID); err != nil {
		return err
	}
	lastID := req.LastReadMsgID
	if lastID == 0 {
		conv, err := s.convRepo.GetByID(ctx, req.ConversationID)
		if err != nil {
			return errs.NotFound("会话不存在")
		}
		lastID = conv.LastMessageID
	}
	if err := s.memberRepo.UpdateRead(ctx, req.ConversationID, model.IMPrincipalEmployee, employeeID, lastID); err != nil {
		return fmt.Errorf("标记已读失败: %w", err)
	}
	_ = s.ClearUnread(ctx, employeeID, req.ConversationID)
	return nil
}

func (s *IMService) conversationItemFor(ctx context.Context, selfID, conversationID uint, peer *model.Employee) (*types.IMConversationItem, error) {
	conv, err := s.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("打开会话失败: %w", err)
	}
	unreadMap, _ := s.UnreadMap(ctx, selfID)
	_, online := s.OnlineEmployees(ctx, []uint{peer.ID})[peer.ID]
	item := &types.IMConversationItem{
		ID:                 conv.ID,
		Type:               conv.Type,
		Title:              peer.Name,
		PeerID:             peer.ID,
		PeerName:           peer.Name,
		PeerAvatarURL:      peer.AvatarURL,
		PeerDepartmentName: s.departmentNames(ctx, []uint{peer.DepartmentID})[peer.DepartmentID],
		PeerOnline:         online,
		LastMessageID:      conv.LastMessageID,
		LastContent:        conv.LastContent,
		LastActiveAt:       conv.LastActiveAt,
		UnreadCount:        unreadMap[strconv.FormatUint(uint64(conv.ID), 10)],
	}
	return item, nil
}
