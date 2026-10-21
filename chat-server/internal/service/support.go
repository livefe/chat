package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

type employeeMessenger interface {
	DeliverMessage(ctx context.Context, employeeIDs []uint, item *types.IMMessageItem)
	BroadcastRaw(ctx context.Context, payload []byte)
}

type userMessenger interface {
	DeliverToUser(ctx context.Context, userID uint, payload []byte)
}

type SupportService struct {
	log               *logger.Logger
	cache             *cache.Cache
	db                *database.DB
	supportStateRepo  repository.IMSupportStateRepository
	supportConfigRepo repository.IMSupportConfigRepository
	userRepo          *UserStore
	employeeRepo      repository.EmployeeRepository
	imSvc             *IMService
	employees         employeeMessenger
	users             userMessenger
	permSvc           *PermissionService
}

func NewSupportService(
	log *logger.Logger,
	cache *cache.Cache,
	db *database.DB,
	supportStateRepo repository.IMSupportStateRepository,
	supportConfigRepo repository.IMSupportConfigRepository,
	userRepo *UserStore,
	employeeRepo repository.EmployeeRepository,
	imSvc *IMService,
	employees employeeMessenger,
	users userMessenger,
	permSvc *PermissionService,
) *SupportService {
	return &SupportService{
		log:               log,
		cache:             cache,
		db:                db,
		supportStateRepo:  supportStateRepo,
		supportConfigRepo: supportConfigRepo,
		userRepo:          userRepo,
		employeeRepo:      employeeRepo,
		imSvc:             imSvc,
		employees:         employees,
		users:             users,
		permSvc:           permSvc,
	}
}

func (s *SupportService) loadConfig(ctx context.Context) (*model.IMSupportConfig, error) {
	cfg, err := s.supportConfigRepo.Get(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.Unavailable("客服配置未初始化", nil)
		}
		return nil, fmt.Errorf("读取配置失败: %w", err)
	}
	return cfg, nil
}

func (s *SupportService) sendAs(ctx context.Context, senderKind uint8, senderID uint, req *types.SupportSendRequest) (*types.IMMessageItem, error) {
	item, fresh, err := s.imSvc.ReserveOutgoing(ctx, senderKind, senderID, req.ClientMsgID)
	if err != nil || !fresh {
		return item, err
	}
	item, err = s.imSvc.CommitOutgoing(ctx, IMDraft{
		ConversationID: req.ConversationID,
		SenderKind:     senderKind,
		SenderID:       senderID,
		ContentType:    req.ContentType,
		Raw:            req.Content,
		ReplyToID:      req.ReplyToID,
		ClientMsgID:    req.ClientMsgID,
		Prepare:        true,
		BindMedia:      senderKind != model.IMPrincipalSystem,
	})
	if err != nil {
		return nil, err
	}
	if item.SenderKind == model.IMPrincipalUser {
		items := []types.IMMessageItem{*item}
		s.enrichUserSenderNames(ctx, items)
		item = &items[0]
	}
	s.fanoutMessage(ctx, req.ConversationID, senderKind, senderID, item)
	return item, nil
}

func (s *SupportService) writeSystemMessage(ctx context.Context, conversationID uint, text string) (*types.IMMessageItem, error) {
	item, err := s.imSvc.CommitOutgoing(ctx, IMDraft{
		ConversationID: conversationID,
		SenderKind:     model.IMPrincipalSystem,
		ContentType:    model.MessageContentTypeText,
		Raw:            text,
	})
	if err != nil {
		return nil, err
	}
	item.SenderName = "系统"
	return item, nil
}

func (s *SupportService) listMessages(ctx context.Context, conversationID, beforeID uint, limit int) (*types.IMMessageListResponse, error) {
	items, err := s.imSvc.ListMessageItems(ctx, conversationID, beforeID, limit)
	if err != nil {
		return nil, err
	}
	s.enrichUserSenderNames(ctx, items)
	for i := range items {
		if items[i].SenderKind == model.IMPrincipalSystem {
			items[i].SenderName = "系统"
		}
	}
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return &types.IMMessageListResponse{List: items}, nil
}

func (s *SupportService) enrichUserSenderNames(ctx context.Context, items []types.IMMessageItem) {
	ids := make([]uint, 0)
	for _, it := range items {
		if it.SenderKind == model.IMPrincipalUser && it.SenderName == "" {
			ids = append(ids, it.SenderID)
		}
	}
	if len(ids) == 0 {
		return
	}
	users, _ := s.userRepo.GetByIDs(ctx, ids)
	m := map[uint]string{}
	for _, u := range users {
		m[u.ID] = u.Nickname
	}
	for i := range items {
		if items[i].SenderKind == model.IMPrincipalUser && items[i].SenderName == "" {
			items[i].SenderName = m[items[i].SenderID]
		}
	}
}

func (s *SupportService) loadSupportOwnedByUser(ctx context.Context, userID, conversationID uint) (*model.IMSupportState, *model.IMConversation, error) {
	st, err := s.supportStateRepo.GetByConversationID(ctx, conversationID)
	if err != nil {
		return nil, nil, errs.NotFound("会话不存在")
	}
	if st.UserID != userID {
		return nil, nil, errs.Forbidden("无权访问该会话")
	}
	conv, err := s.imSvc.GetConversation(ctx, conversationID)
	if err != nil || conv.Type != model.IMConvTypeSupport {
		return nil, nil, errs.NotFound("会话不存在")
	}
	return st, conv, nil
}

func (s *SupportService) hasViewAll(ctx context.Context, employeeID uint) (bool, error) {
	if s.permSvc == nil {
		return false, nil
	}
	return s.permSvc.HasPermissionCode(ctx, employeeID, "im:support_overview")
}

func (s *SupportService) assertAgentAccess(ctx context.Context, employeeID, conversationID uint) error {
	st, err := s.supportStateRepo.GetByConversationID(ctx, conversationID)
	if err != nil {
		return errs.NotFound("会话不存在")
	}
	conv, err := s.imSvc.GetConversation(ctx, conversationID)
	if err != nil || conv.Type != model.IMConvTypeSupport {
		return errs.NotFound("会话不存在")
	}
	if conv.Status == model.IMConvStatusQueued {
		return errs.Forbidden("请先接入会话后再查看消息")
	}
	if ok, err := s.hasViewAll(ctx, employeeID); err == nil && ok {
		return nil
	}
	if conv.Status == model.IMConvStatusClosed {
		if st.AssigneeEmployeeID == employeeID {
			return nil
		}

		if err := s.imSvc.RequireMember(ctx, conversationID, model.IMPrincipalEmployee, employeeID); err != nil {
			return err
		}
		return nil
	}
	if st.AssigneeEmployeeID != employeeID {
		if err := s.imSvc.RequireMember(ctx, conversationID, model.IMPrincipalEmployee, employeeID); err != nil {
			return err
		}
	}
	return nil
}

func (s *SupportService) sessionItemForUser(ctx context.Context, userID, conversationID uint) (*types.SupportSessionItem, error) {
	conv, err := s.imSvc.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	st, _ := s.supportStateRepo.GetByConversationID(ctx, conversationID)
	item := &types.SupportSessionItem{
		ID:            conv.ID,
		Status:        conv.Status,
		StatusText:    supportStatusText(conv.Status),
		LastMessageID: conv.LastMessageID,
		LastContent:   conv.LastContent,
		LastActiveAt:  conv.LastActiveAt,
		UnreadCount:   s.getUserUnread(ctx, userID),
	}
	if cfg, err := s.loadConfig(ctx); err == nil && cfg != nil {
		item.BrandName = cfg.BrandName
		item.BrandAvatarURL = cfg.BrandAvatarURL
	}
	if st != nil {
		item.AssigneeEmployeeID = st.AssigneeEmployeeID
		item.QueueAt = st.QueueAt
		if st.AssigneeEmployeeID > 0 {
			if e, err := s.employeeRepo.GetByID(ctx, st.AssigneeEmployeeID); err == nil {
				item.AssigneeName = e.Name
			}
		}
	}
	return item, nil
}

func (s *SupportService) toAdminList(ctx context.Context, employeeID uint, rows []repository.SupportListRow, total int64) (*types.SupportAdminSessionListResponse, error) {
	userIDs := make([]uint, 0, len(rows))
	empIDs := make([]uint, 0, len(rows))
	for _, r := range rows {
		userIDs = append(userIDs, r.UserID)
		if r.AssigneeEmployeeID > 0 {
			empIDs = append(empIDs, r.AssigneeEmployeeID)
		}
	}
	users, _ := s.userRepo.GetByIDs(ctx, userIDs)
	userMap := map[uint]*model.User{}
	for _, u := range users {
		userMap[u.ID] = u
	}
	emps, _ := s.employeeRepo.GetByIDs(ctx, empIDs)
	empMap := map[uint]*model.Employee{}
	for _, e := range emps {
		empMap[e.ID] = e
	}
	now := time.Now().Unix()
	list := make([]types.SupportAdminSessionItem, 0, len(rows))
	for _, r := range rows {
		item := types.SupportAdminSessionItem{
			ID:                 r.ID,
			Status:             r.Status,
			StatusText:         supportStatusText(r.Status),
			UserID:             r.UserID,
			AssigneeEmployeeID: r.AssigneeEmployeeID,
			LastMessageID:      r.LastMessageID,
			LastContent:        r.LastContent,
			LastActiveAt:       r.LastActiveAt,
			QueueAt:            r.QueueAt,
			AcceptAt:           r.AcceptAt,
			CloseAt:            r.CloseAt,
		}
		if r.Status == model.IMConvStatusQueued && r.QueueAt > 0 {
			item.QueueWaitSec = now - r.QueueAt
		}
		if u := userMap[r.UserID]; u != nil {
			item.UserNickname = u.Nickname
			item.UserMobile = u.Mobile
		}
		if e := empMap[r.AssigneeEmployeeID]; e != nil {
			item.AssigneeName = e.Name
		}
		if employeeID > 0 {
			unread, _ := s.imSvc.UnreadMap(ctx, employeeID)
			item.UnreadCount = unread[strconv.FormatUint(uint64(r.ID), 10)]
		}
		list = append(list, item)
	}
	return &types.SupportAdminSessionListResponse{Total: total, List: list}, nil
}

func (s *SupportService) adminSessionItem(ctx context.Context, employeeID, conversationID uint) (*types.SupportAdminSessionItem, error) {
	st, err := s.supportStateRepo.GetByConversationID(ctx, conversationID)
	if err != nil {
		return nil, errs.NotFound("会话不存在")
	}
	conv, err := s.imSvc.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	row := repository.SupportListRow{
		IMConversation:     *conv,
		UserID:             st.UserID,
		AssigneeEmployeeID: st.AssigneeEmployeeID,
		QueueAt:            st.QueueAt,
		AcceptAt:           st.AcceptAt,
		CloseAt:            st.CloseAt,
	}
	resp, err := s.toAdminList(ctx, employeeID, []repository.SupportListRow{row}, 1)
	if err != nil || len(resp.List) == 0 {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return &resp.List[0], nil
}

func supportStatusText(status uint8) string {
	switch status {
	case model.IMConvStatusQueued:
		return "排队中"
	case model.IMConvStatusActive:
		return "进行中"
	case model.IMConvStatusClosed:
		return "已结束"
	default:
		return "未知"
	}
}

func supportTruncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func (s *SupportService) AdminUpdateConfig(ctx context.Context, req *types.SupportConfigUpdateRequest) (*types.SupportProfileResponse, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	cfg.BrandName = supportTruncate(strings.TrimSpace(req.BrandName), 32)
	cfg.BrandAvatarURL = strings.TrimSpace(req.BrandAvatarURL)
	cfg.WelcomeText = supportTruncate(strings.TrimSpace(req.WelcomeText), 500)
	cfg.OfflineText = supportTruncate(strings.TrimSpace(req.OfflineText), 500)
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}
	if err := s.supportConfigRepo.Save(ctx, cfg); err != nil {
		return nil, fmt.Errorf("保存失败: %w", err)
	}
	return s.GetProfile(ctx)
}
