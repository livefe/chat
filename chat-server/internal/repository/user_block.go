package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type UserBlockRepository interface {
	Exists(ctx context.Context, blockerID, blockedID uint) (bool, error)
	ExistsEither(ctx context.Context, userA, userB uint) (bool, error)
	CountByBlocker(ctx context.Context, blockerID uint) (int64, error)
	ListEitherPeerIDs(ctx context.Context, userID uint) ([]uint, error)
	ListByBlocker(ctx context.Context, blockerID uint, page, pageSize int) ([]*model.UserBlock, int64, error)
	DeleteByUserID(ctx context.Context, userID uint) error
	ListActivePairsByUserID(ctx context.Context, userID uint) ([]model.UserBlock, error)
	ApplyBlock(ctx context.Context, blockerID, blockedID uint) error
	ApplyUnblock(ctx context.Context, blockerID, blockedID uint) error
}

type userBlockRepository struct {
	db *database.DB
}

func NewUserBlockRepository(db *database.DB) UserBlockRepository {
	return &userBlockRepository{db: db}
}

func (r *userBlockRepository) Exists(ctx context.Context, blockerID, blockedID uint) (bool, error) {
	var n int64
	err := database.Conn(ctx, r.db).Model(&model.UserBlock{}).
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Count(&n).Error
	return n > 0, err
}

func (r *userBlockRepository) ExistsEither(ctx context.Context, userA, userB uint) (bool, error) {
	if userA == 0 || userB == 0 || userA == userB {
		return false, nil
	}
	var n int64
	err := database.Conn(ctx, r.db).Model(&model.UserBlock{}).
		Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
			userA, userB, userB, userA).
		Limit(1).
		Count(&n).Error
	return n > 0, err
}

func (r *userBlockRepository) CountByBlocker(ctx context.Context, blockerID uint) (int64, error) {
	var n int64
	err := database.Conn(ctx, r.db).Model(&model.UserBlock{}).
		Where("blocker_id = ?", blockerID).
		Count(&n).Error
	return n, err
}

func (r *userBlockRepository) ListEitherPeerIDs(ctx context.Context, userID uint) ([]uint, error) {
	if userID == 0 {
		return nil, nil
	}
	var rows []model.UserBlock
	err := database.Conn(ctx, r.db).
		Where("blocker_id = ? OR blocked_id = ?", userID, userID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	seen := make(map[uint]struct{}, len(rows))
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		peer := row.BlockedID
		if row.BlockerID != userID {
			peer = row.BlockerID
		}
		if peer == 0 || peer == userID {
			continue
		}
		if _, ok := seen[peer]; ok {
			continue
		}
		seen[peer] = struct{}{}
		ids = append(ids, peer)
	}
	return ids, nil
}

func (r *userBlockRepository) ListByBlocker(ctx context.Context, blockerID uint, page, pageSize int) ([]*model.UserBlock, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.UserBlock{}).Where("blocker_id = ?", blockerID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.UserBlock
	err := db.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *userBlockRepository) ListActivePairsByUserID(ctx context.Context, userID uint) ([]model.UserBlock, error) {
	var rows []model.UserBlock
	err := database.Conn(ctx, r.db).
		Where("blocker_id = ? OR blocked_id = ?", userID, userID).
		Find(&rows).Error
	return rows, err
}

func (r *userBlockRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return database.Conn(ctx, r.db).
		Where("blocker_id = ? OR blocked_id = ?", userID, userID).
		Delete(&model.UserBlock{}).Error
}

func (r *userBlockRepository) ApplyBlock(ctx context.Context, blockerID, blockedID uint) error {
	return database.Conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.UserBlock{BlockerID: blockerID, BlockedID: blockedID}).Error; err != nil {
			return err
		}
		return tx.Create(&model.UserBlockEvent{
			BlockerID: blockerID,
			BlockedID: blockedID,
			Action:    model.UserBlockEventActionBlock,
		}).Error
	})
}

func (r *userBlockRepository) ApplyUnblock(ctx context.Context, blockerID, blockedID uint) error {
	return database.Conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
			Delete(&model.UserBlock{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		return tx.Create(&model.UserBlockEvent{
			BlockerID: blockerID,
			BlockedID: blockedID,
			Action:    model.UserBlockEventActionUnblock,
		}).Error
	})
}
