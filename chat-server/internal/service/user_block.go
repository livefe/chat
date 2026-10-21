package service

import (
	"context"
	"errors"
	"fmt"

	"app/internal/types"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"
)

const maxBlockedUsers = 50

var errBlockLimitReached = errs.Invalid("最多拉黑50人")

func (s *UserService) Block(ctx context.Context, blockerID, blockedID uint) error {
	if blockerID == 0 || blockedID == 0 {
		return errs.Invalid("参数错误")
	}
	if blockerID == blockedID {
		return errs.Invalid("不能拉黑自己")
	}
	if err := s.EnsureActive(ctx, blockedID); err != nil {
		return err
	}
	err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		exists, err := s.blockRepo.Exists(ctx, blockerID, blockedID)
		if err != nil || exists {
			return err
		}
		n, err := s.blockRepo.CountByBlocker(ctx, blockerID)
		if err != nil {
			return err
		}
		if n >= maxBlockedUsers {
			return errBlockLimitReached
		}
		return s.blockRepo.ApplyBlock(ctx, blockerID, blockedID)
	})
	if err != nil {
		if errors.Is(err, errBlockLimitReached) {
			return err
		}
		s.log.Error(ctx, "拉黑用户失败", logger.Error(err))
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

func (s *UserService) Unblock(ctx context.Context, blockerID, blockedID uint) error {
	if blockerID == 0 || blockedID == 0 {
		return errs.Invalid("参数错误")
	}
	if err := s.blockRepo.ApplyUnblock(ctx, blockerID, blockedID); err != nil {
		s.log.Error(ctx, "取消拉黑失败", logger.Error(err))
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

func (s *UserService) ListBlocked(ctx context.Context, blockerID uint, req *types.BlockListRequest) (*types.BlockListResponse, error) {
	list, total, err := s.blockRepo.ListByBlocker(ctx, blockerID, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	ids := make([]uint, 0, len(list))
	for _, row := range list {
		ids = append(ids, row.BlockedID)
	}
	profiles, err := s.GetProfiles(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items := make([]types.BlockListItem, 0, len(list))
	for _, row := range list {
		nickname := deactivatedDisplayName()
		avatar := ""
		if p, ok := profiles[row.BlockedID]; ok {
			nickname = utils.DisplayName(p.Nickname)
			avatar = p.AvatarURL
		}
		items = append(items, types.BlockListItem{
			UserID:    row.BlockedID,
			Nickname:  nickname,
			AvatarURL: avatar,
			CreatedAt: utils.FormatDateTime(row.CreatedAt),
		})
	}
	return &types.BlockListResponse{Total: total, List: items}, nil
}

func (s *UserService) IsBlockedBy(ctx context.Context, blockerID, blockedID uint) (bool, error) {
	ok, err := s.blockRepo.Exists(ctx, blockerID, blockedID)
	if err != nil {
		return false, fmt.Errorf("调用 Exists 失败: %w", err)
	}
	return ok, nil
}

func (s *UserService) IsBlockedEither(ctx context.Context, userA, userB uint) (bool, error) {
	ok, err := s.blockRepo.ExistsEither(ctx, userA, userB)
	if err != nil {
		return false, fmt.Errorf("调用 ExistsEither 失败: %w", err)
	}
	return ok, nil
}

func (s *UserService) ListBlockedEitherPeerIDs(ctx context.Context, userID uint) ([]uint, error) {
	ids, err := s.blockRepo.ListEitherPeerIDs(ctx, userID)
	if err != nil {
		s.log.Warn(ctx, "查询拉黑关系失败", logger.Error(err))
		return nil, err
	}
	return ids, nil
}

func (s *UserService) AssertCanPrivateChat(ctx context.Context, fromUserID, toUserID uint) error {
	ok, err := s.IsBlockedEither(ctx, fromUserID, toUserID)
	if err != nil {
		return fmt.Errorf("发送失败: %w", err)
	}
	if !ok {
		return nil
	}
	blockedByMe, err := s.IsBlockedBy(ctx, fromUserID, toUserID)
	if err != nil {
		return fmt.Errorf("发送失败: %w", err)
	}
	if blockedByMe {
		return errs.Forbidden("你已拉黑对方，请先取消拉黑")
	}
	return errs.Forbidden("无法发送私信")
}

func (s *UserService) clearUserBlocksOnDeactivation(ctx context.Context, userID uint) error {
	pairs, err := s.blockRepo.ListActivePairsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("调用 ListActivePairsByUserID 失败: %w", err)
	}
	for _, pair := range pairs {
		if err := s.blockRepo.ApplyUnblock(ctx, pair.BlockerID, pair.BlockedID); err != nil {
			return err
		}
	}
	return nil
}
