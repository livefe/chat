package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"
)

type ConversationService struct {
	log         *logger.Logger
	cache       *cache.Cache
	convRepo    repository.ConversationRepository
	convDelRepo repository.ConversationDeletionRepository
	convPinRepo repository.ConversationPinRepository
	msgRepo     repository.MessageRepository
	userSvc     *UserService
}

func NewConversationService(
	log *logger.Logger,
	c *cache.Cache,
	convRepo repository.ConversationRepository,
	convDelRepo repository.ConversationDeletionRepository,
	convPinRepo repository.ConversationPinRepository,
	msgRepo repository.MessageRepository,
	userSvc *UserService,
) *ConversationService {
	return &ConversationService{
		log:         log,
		cache:       c,
		convRepo:    convRepo,
		convDelRepo: convDelRepo,
		convPinRepo: convPinRepo,
		msgRepo:     msgRepo,
		userSvc:     userSvc,
	}
}

func orderUserIDs(a, b uint) (uint, uint) {
	if a < b {
		return a, b
	}
	return b, a
}

func (s *ConversationService) List(ctx context.Context, userID uint, page, pageSize int) (*types.ConversationListResponse, error) {
	list, total, err := s.convRepo.ListByUser(ctx, userID, page, pageSize)
	if err != nil {
		s.log.Error(ctx, "查询会话列表失败", logger.Error(err), logger.Uint("userId", userID))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	unread, err := s.UnreadMap(ctx, userID)
	if err != nil {
		s.log.Error(ctx, "查询会话未读失败", logger.Error(err), logger.Uint("userId", userID))
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	peerIDs := make([]uint, 0, len(list))
	for _, conv := range list {
		peerID := conv.UserMaxID
		if peerID == userID {
			peerID = conv.UserMinID
		}
		peerIDs = append(peerIDs, peerID)
	}
	profiles, err := s.userSvc.GetProfiles(ctx, peerIDs)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	items := make([]types.ConversationItem, 0, len(list))
	for _, conv := range list {
		peerID := conv.UserMaxID
		if peerID == userID {
			peerID = conv.UserMinID
		}
		peer, ok := profiles[peerID]
		peerNickname := deactivatedDisplayName()
		peerAvatar := ""
		if ok {
			peerNickname = utils.DisplayName(peer.Nickname)
			peerAvatar = peer.AvatarURL
		}
		items = append(items, types.ConversationItem{
			PeerID:       peerID,
			PeerNickname: peerNickname,
			PeerAvatar:   peerAvatar,
			LastContent:  conv.LastContent,
			LastActiveAt: conv.LastActiveAt,
			UnreadCount:  unread[peerID],
			Pinned:       conv.Pinned,
		})
	}

	return &types.ConversationListResponse{Total: total, List: items}, nil
}

func (s *ConversationService) UnreadSummary(ctx context.Context, userID uint) (*types.ConversationUnreadSummary, error) {
	unread, err := s.UnreadMap(ctx, userID)
	if err != nil {
		s.log.Error(ctx, "查询会话未读失败", logger.Error(err), logger.Uint("userId", userID))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items := make([]types.ConversationUnreadItem, 0, len(unread))
	var total int64
	for peerID, count := range unread {
		if count <= 0 {
			continue
		}
		total += count
		items = append(items, types.ConversationUnreadItem{
			PeerID: peerID,
			Count:  count,
		})
	}
	return &types.ConversationUnreadSummary{Total: total, Items: items}, nil
}

func (s *ConversationService) Delete(ctx context.Context, userID, peerID uint) (uint, error) {
	if userID == 0 || peerID == 0 {
		return 0, errs.Invalid("参数错误")
	}
	clearedBeforeID, err := s.msgRepo.MaxPrivateIDBetween(ctx, userID, peerID)
	if err != nil {
		s.log.Error(ctx, "查询私聊消息上限失败", logger.Error(err))
		return 0, fmt.Errorf("删除会话失败: %w", err)
	}
	if err := s.markDeleted(ctx, userID, peerID, clearedBeforeID); err != nil {
		return 0, err
	}
	return clearedBeforeID, nil
}

func (s *ConversationService) markDeleted(ctx context.Context, userID, peerID, clearedBeforeID uint) error {
	row := &model.ConversationDeletion{
		UserID:          userID,
		PeerID:          peerID,
		ClearedBeforeID: clearedBeforeID,
	}
	if err := s.convDelRepo.Upsert(ctx, row); err != nil {
		s.log.Error(ctx, "标记删除会话失败", logger.Error(err))
		return fmt.Errorf("删除会话失败: %w", err)
	}
	if err := s.clearUnread(ctx, userID, peerID); err != nil {
		s.log.Warn(ctx, "清除会话未读失败", logger.Error(err),
			logger.Uint("userId", userID), logger.Uint("peerId", peerID))
	}
	if err := s.convPinRepo.Delete(ctx, userID, peerID); err != nil {
		s.log.Warn(ctx, "清除会话置顶失败", logger.Error(err),
			logger.Uint("userId", userID), logger.Uint("peerId", peerID))
	}
	return nil
}

func (s *ConversationService) Pin(ctx context.Context, userID, peerID uint, pinned bool) error {
	if userID == 0 || peerID == 0 {
		return errs.Invalid("参数错误")
	}
	if pinned {
		if err := s.convPinRepo.Upsert(ctx, &model.ConversationPin{UserID: userID, PeerID: peerID}); err != nil {
			s.log.Error(ctx, "置顶会话失败", logger.Error(err),
				logger.Uint("userId", userID), logger.Uint("peerId", peerID))
			return fmt.Errorf("置顶失败: %w", err)
		}
		return nil
	}
	if err := s.convPinRepo.Delete(ctx, userID, peerID); err != nil {
		s.log.Error(ctx, "取消置顶失败", logger.Error(err),
			logger.Uint("userId", userID), logger.Uint("peerId", peerID))
		return fmt.Errorf("取消置顶失败: %w", err)
	}
	return nil
}

func (s *ConversationService) PrivateClearedBeforeID(ctx context.Context, userID, peerID uint) uint {
	row, err := s.convDelRepo.Get(ctx, userID, peerID)
	if err != nil || row == nil {
		return 0
	}
	return row.ClearedBeforeID
}

func (s *ConversationService) upsert(ctx context.Context, fromUserID, toUserID uint, msg *model.Message) error {
	minID, maxID := orderUserIDs(fromUserID, toUserID)
	preview := conversationPreview(msg.ContentType, msg.Content)
	conv := &model.Conversation{
		UserMinID:     minID,
		UserMaxID:     maxID,
		LastMessageID: msg.ID,
		LastContent:   preview,
		LastActiveAt:  time.Now().Unix(),
	}
	return s.convRepo.Upsert(ctx, conv)
}

func (s *ConversationService) updateLastPreview(ctx context.Context, userA, userB, lastID uint, lastContent string) error {
	minID, maxID := orderUserIDs(userA, userB)
	if err := s.convRepo.UpdateLastPreview(ctx, minID, maxID, lastID, lastContent); err != nil {
		s.log.Warn(ctx, "更新会话预览失败", logger.Error(err))
		return err
	}
	return nil
}

func (s *ConversationService) IncrUnread(ctx context.Context, userID, fromUserID uint) error {
	if userID == 0 || fromUserID == 0 || userID == fromUserID {
		return nil
	}
	key := fmt.Sprintf(types.CacheUnread, userID)
	field := strconv.FormatUint(uint64(fromUserID), 10)
	pipe := s.cache.Pipeline()
	pipe.HIncrBy(ctx, key, field, 1)
	pipe.Expire(ctx, key, types.Expire1d*31)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *ConversationService) ClearUnread(ctx context.Context, userID, peerID uint) error {
	if userID == 0 || peerID == 0 {
		return errs.Invalid("参数错误")
	}
	if err := s.clearUnread(ctx, userID, peerID); err != nil {
		s.log.Error(ctx, "清除会话未读失败", logger.Error(err),
			logger.Uint("userId", userID), logger.Uint("peerId", peerID))
		return fmt.Errorf("清除未读失败: %w", err)
	}
	return nil
}

func (s *ConversationService) clearUnread(ctx context.Context, userID, peerID uint) error {
	key := fmt.Sprintf(types.CacheUnread, userID)
	return s.cache.HDel(ctx, key, strconv.FormatUint(uint64(peerID), 10)).Err()
}

func (s *ConversationService) AdminList(ctx context.Context, req *types.AdminConversationListRequest) (*types.AdminConversationListResponse, error) {
	list, total, err := s.convRepo.List(ctx, repository.ConversationListFilter{
		UserID: req.UserID,
		Page:   req.Page, PageSize: req.PageSize,
	})
	if err != nil {
		s.log.Error(ctx, "查询会话列表失败", logger.Error(err))
		return nil, fmt.Errorf("查询会话列表失败: %w", err)
	}

	userIDs := make([]uint, 0, len(list)*2)
	for _, conv := range list {
		userIDs = append(userIDs, conv.UserMinID, conv.UserMaxID)
	}
	profiles, err := s.userSvc.GetProfiles(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("查询会话列表失败: %w", err)
	}

	items := make([]types.AdminConversationItem, 0, len(list))
	for _, conv := range list {
		items = append(items, toAdminConversationItem(conv, profiles))
	}
	return &types.AdminConversationListResponse{Total: total, List: items}, nil
}

func toAdminConversationItem(conv *model.Conversation, profiles map[uint]types.UserProfile) types.AdminConversationItem {
	return types.AdminConversationItem{
		ID:            conv.ID,
		UserMinID:     conv.UserMinID,
		UserMaxID:     conv.UserMaxID,
		UserMinName:   profileDisplayName(profiles, conv.UserMinID),
		UserMaxName:   profileDisplayName(profiles, conv.UserMaxID),
		UserMinMobile: profileMobile(profiles, conv.UserMinID),
		UserMaxMobile: profileMobile(profiles, conv.UserMaxID),
		LastContent:   conv.LastContent,
		LastActiveAt:  conv.LastActiveAt,
		CreatedAt:     utils.FormatDateTime(conv.CreatedAt),
	}
}

func profileDisplayName(profiles map[uint]types.UserProfile, id uint) string {
	if p, ok := profiles[id]; ok {
		return utils.DisplayName(p.Nickname)
	}
	return deactivatedDisplayName()
}

func profileMobile(profiles map[uint]types.UserProfile, id uint) string {
	if p, ok := profiles[id]; ok {
		return utils.DesensitizeMobile(p.Mobile)
	}
	return ""
}

func (s *ConversationService) UnreadMap(ctx context.Context, userID uint) (map[uint]int64, error) {
	if userID == 0 {
		return nil, nil
	}
	key := fmt.Sprintf(types.CacheUnread, userID)
	raw, err := s.cache.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[uint]int64, len(raw))
	for k, v := range raw {
		id, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			continue
		}
		count, _ := strconv.ParseInt(v, 10, 64)
		out[uint(id)] = count
	}
	return out, nil
}
