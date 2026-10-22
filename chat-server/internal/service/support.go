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
	"app/pkg/utils"

	"gorm.io/gorm"
)

const (
	permCodeSupportOverview = "im:support_overview"
	permCodeSupportAccept   = "im:support:accept"
)

var supportCloseReasonText = map[string]string{
	types.SupportCloseReasonResolved:   "已解决",
	types.SupportCloseReasonNoResponse: "无应答",
	types.SupportCloseReasonDuplicate:  "重复",
	types.SupportCloseReasonOther:      "其他",
	types.SupportCloseReasonUserCancel: "用户取消",
}

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

func (s *SupportService) hasViewAll(ctx context.Context, employeeID uint) bool {
	if s.permSvc == nil {
		return false
	}
	ok, err := s.permSvc.HasPermissionCode(ctx, employeeID, permCodeSupportOverview)
	return err == nil && ok
}

func (s *SupportService) requireViewAll(ctx context.Context, employeeID uint) error {
	if !s.hasViewAll(ctx, employeeID) {
		return errs.Forbidden("无权查看全部客服会话")
	}
	return nil
}

// assertAgentAccess 校验员工能否查看会话消息：督导可看全部；坐席可看自己参与过的会话，
// 以及自己服务过的用户的已结案历史会话；排队中会话只有督导可看。
func (s *SupportService) assertAgentAccess(ctx context.Context, employeeID, conversationID uint) (*model.IMSupportState, *model.IMConversation, error) {
	st, err := s.supportStateRepo.GetByConversationID(ctx, conversationID)
	if err != nil {
		return nil, nil, errs.NotFound("会话不存在")
	}
	conv, err := s.imSvc.GetConversation(ctx, conversationID)
	if err != nil || conv.Type != model.IMConvTypeSupport {
		return nil, nil, errs.NotFound("会话不存在")
	}
	if s.hasViewAll(ctx, employeeID) {
		return st, conv, nil
	}
	if conv.Status == model.IMConvStatusQueued {
		return nil, nil, errs.Forbidden("请先接入会话后再查看消息")
	}
	if st.AssigneeEmployeeID == employeeID {
		return st, conv, nil
	}
	if err := s.imSvc.RequireMember(ctx, conversationID, model.IMPrincipalEmployee, employeeID); err == nil {
		return st, conv, nil
	}
	if conv.Status == model.IMConvStatusClosed {
		if ok, err := s.supportStateRepo.HasServedUser(ctx, st.UserID, employeeID); err == nil && ok {
			return st, conv, nil
		}
	}
	return nil, nil, errs.Forbidden("无权访问该会话")
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

func (s *SupportService) employeeNames(ctx context.Context, rows []repository.SupportListRow) map[uint]string {
	ids := make([]uint, 0, len(rows))
	for _, r := range rows {
		if r.AssigneeEmployeeID > 0 {
			ids = append(ids, r.AssigneeEmployeeID)
		}
	}
	out := make(map[uint]string, len(ids))
	emps, _ := s.employeeRepo.GetByIDs(ctx, ids)
	for _, e := range emps {
		out[e.ID] = e.Name
	}
	return out
}

// toAdminList 组装后台会话列表；employeeID > 0 时附带该员工的未读数。
func (s *SupportService) toAdminList(ctx context.Context, employeeID uint, rows []repository.SupportListRow, total int64) *types.SupportAdminSessionListResponse {
	userIDs := make([]uint, 0, len(rows))
	for _, r := range rows {
		userIDs = append(userIDs, r.UserID)
	}
	users, _ := s.userRepo.GetByIDs(ctx, userIDs)
	userMap := map[uint]*model.User{}
	for _, u := range users {
		userMap[u.ID] = u
	}
	empNames := s.employeeNames(ctx, rows)
	var unread map[string]int64
	if employeeID > 0 {
		unread, _ = s.imSvc.UnreadMap(ctx, employeeID)
	}
	list := make([]types.SupportAdminSessionItem, 0, len(rows))
	for _, r := range rows {
		item := types.SupportAdminSessionItem{
			ID:                 r.ID,
			Status:             r.Status,
			StatusText:         supportStatusText(r.Status),
			UserID:             r.UserID,
			AssigneeEmployeeID: r.AssigneeEmployeeID,
			AssigneeName:       empNames[r.AssigneeEmployeeID],
			LastMessageID:      r.LastMessageID,
			LastContent:        r.LastContent,
			LastActiveAt:       r.LastActiveAt,
			QueueAt:            r.QueueAt,
			AcceptAt:           r.AcceptAt,
			CloseAt:            r.CloseAt,
			CloseReason:        r.CloseReason,
			CloseReasonText:    supportCloseReasonText[r.CloseReason],
			CloseNote:          r.CloseNote,
			UnreadCount:        unread[strconv.FormatUint(uint64(r.ID), 10)],
		}
		if u := userMap[r.UserID]; u != nil {
			item.UserNickname = u.Nickname
			item.UserMobile = utils.DesensitizeMobile(u.Mobile)
		}
		list = append(list, item)
	}
	return &types.SupportAdminSessionListResponse{Total: total, List: list}
}

func (s *SupportService) toUserSessionList(ctx context.Context, rows []repository.SupportListRow, total int64) *types.SupportUserSessionListResponse {
	empNames := s.employeeNames(ctx, rows)
	list := make([]types.SupportUserSessionItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.SupportUserSessionItem{
			ID:              r.ID,
			Status:          r.Status,
			StatusText:      supportStatusText(r.Status),
			AssigneeName:    empNames[r.AssigneeEmployeeID],
			CreatedAt:       utils.FormatDateTime(r.CreatedAt),
			AcceptedAt:      formatUnixTime(r.AcceptAt),
			ClosedAt:        formatUnixTime(r.CloseAt),
			CloseReason:     r.CloseReason,
			CloseReasonText: supportCloseReasonText[r.CloseReason],
			CloseNote:       r.CloseNote,
			LastMessage:     r.LastContent,
			LastMessageAt:   formatUnixTime(r.LastActiveAt),
		})
	}
	return &types.SupportUserSessionListResponse{Total: total, List: list}
}

func (s *SupportService) adminSessionItem(ctx context.Context, employeeID, conversationID uint) (*types.SupportAdminSessionItem, error) {
	row, err := s.supportStateRepo.GetRow(ctx, conversationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("会话不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	resp := s.toAdminList(ctx, employeeID, []repository.SupportListRow{*row}, 1)
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

func formatUnixTime(sec int64) string {
	if sec <= 0 {
		return ""
	}
	return utils.FormatDateTime(time.Unix(sec, 0))
}

func toQuickReplyItems(list []model.SupportQuickReply) []types.SupportQuickReplyItem {
	out := make([]types.SupportQuickReplyItem, 0, len(list))
	for _, q := range list {
		out = append(out, types.SupportQuickReplyItem{Title: q.Title, Content: q.Content})
	}
	return out
}

func (s *SupportService) AdminGetConfig(ctx context.Context) (*types.SupportAdminConfigResponse, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &types.SupportAdminConfigResponse{
		SupportProfileResponse: types.SupportProfileResponse{
			BrandName:      cfg.BrandName,
			BrandAvatarURL: cfg.BrandAvatarURL,
			WelcomeText:    cfg.WelcomeText,
			OfflineText:    cfg.OfflineText,
			Enabled:        cfg.Enabled == 1,
		},
		QuickReplies: toQuickReplyItems(cfg.QuickReplies),
	}, nil
}

func (s *SupportService) AdminQuickReplies(ctx context.Context) (*types.SupportQuickReplyListResponse, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &types.SupportQuickReplyListResponse{List: toQuickReplyItems(cfg.QuickReplies)}, nil
}

func (s *SupportService) AdminUpdateConfig(ctx context.Context, req *types.SupportConfigUpdateRequest) (*types.SupportAdminConfigResponse, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	brand := strings.TrimSpace(req.BrandName)
	if brand == "" {
		return nil, errs.Invalid("请填写品牌名称")
	}
	replies := make([]model.SupportQuickReply, 0, len(req.QuickReplies))
	for _, q := range req.QuickReplies {
		title := strings.TrimSpace(q.Title)
		content := strings.TrimSpace(q.Content)
		if title == "" || content == "" {
			return nil, errs.Invalid("快捷回复的标题和内容不能为空")
		}
		replies = append(replies, model.SupportQuickReply{Title: title, Content: content})
	}
	cfg.BrandName = brand
	cfg.BrandAvatarURL = strings.TrimSpace(req.BrandAvatarURL)
	cfg.WelcomeText = strings.TrimSpace(req.WelcomeText)
	cfg.OfflineText = strings.TrimSpace(req.OfflineText)
	cfg.QuickReplies = replies
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}
	if err := s.supportConfigRepo.Save(ctx, cfg); err != nil {
		return nil, fmt.Errorf("保存失败: %w", err)
	}
	return s.AdminGetConfig(ctx)
}
