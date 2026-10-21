package repository

import (
	"context"
	"strings"

	"app/internal/model"
	"app/pkg/database"
)

type AdCampaignRepository interface {
	Create(ctx context.Context, item *model.AdCampaign) error
	Update(ctx context.Context, item *model.AdCampaign) error
	GetByID(ctx context.Context, id uint) (*model.AdCampaign, error)
	List(ctx context.Context, status *uint8, keyword string, page, pageSize int) ([]*model.AdCampaign, int64, error)
	ListActiveCandidates(ctx context.Context, now int64) ([]*model.AdCampaign, error)
	CountActiveByAdvertiser(ctx context.Context, advertiserName string, excludeID uint) (int64, error)
	UpdateStatus(ctx context.Context, id uint, status uint8, rejectRemark string, landingInvalid bool) error
	BatchUpdateStatus(ctx context.Context, ids []uint, status uint8) error
	MarkEndedExpired(ctx context.Context, now int64) (int64, error)
}

type adCampaignRepository struct{ db *database.DB }

func NewAdCampaignRepository(db *database.DB) AdCampaignRepository {
	return &adCampaignRepository{db: db}
}

func (r *adCampaignRepository) Create(ctx context.Context, item *model.AdCampaign) error {
	return database.Conn(ctx, r.db).Create(item).Error
}

func (r *adCampaignRepository) Update(ctx context.Context, item *model.AdCampaign) error {
	return database.Conn(ctx, r.db).Save(item).Error
}

func (r *adCampaignRepository) GetByID(ctx context.Context, id uint) (*model.AdCampaign, error) {
	var item model.AdCampaign
	err := database.Conn(ctx, r.db).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *adCampaignRepository) List(ctx context.Context, status *uint8, keyword string, page, pageSize int) ([]*model.AdCampaign, int64, error) {
	q := database.Conn(ctx, r.db).Model(&model.AdCampaign{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("advertiser_name LIKE ? OR title LIKE ? OR name LIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.AdCampaign
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *adCampaignRepository) ListActiveCandidates(ctx context.Context, now int64) ([]*model.AdCampaign, error) {
	var list []*model.AdCampaign
	err := database.Conn(ctx, r.db).
		Where("status = ?", model.AdStatusActive).
		Where("start_at <= ? AND end_at >= ?", now, now).
		Where("landing_invalid = ?", false).
		Order("priority DESC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *adCampaignRepository) CountActiveByAdvertiser(ctx context.Context, advertiserName string, excludeID uint) (int64, error) {
	var n int64
	q := database.Conn(ctx, r.db).Model(&model.AdCampaign{}).
		Where("advertiser_name = ? AND status = ?", advertiserName, model.AdStatusActive)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	err := q.Count(&n).Error
	return n, err
}

func (r *adCampaignRepository) UpdateStatus(ctx context.Context, id uint, status uint8, rejectRemark string, landingInvalid bool) error {
	updates := map[string]any{
		"status":          status,
		"reject_remark":   rejectRemark,
		"landing_invalid": landingInvalid,
	}
	return database.Conn(ctx, r.db).Model(&model.AdCampaign{}).Where("id = ?", id).Updates(updates).Error
}

func (r *adCampaignRepository) BatchUpdateStatus(ctx context.Context, ids []uint, status uint8) error {
	return database.Conn(ctx, r.db).Model(&model.AdCampaign{}).Where("id IN ?", ids).Update("status", status).Error
}

func (r *adCampaignRepository) MarkEndedExpired(ctx context.Context, now int64) (int64, error) {
	res := database.Conn(ctx, r.db).Model(&model.AdCampaign{}).
		Where("status = ? AND end_at > 0 AND end_at < ?", model.AdStatusActive, now).
		Update("status", model.AdStatusEnded)
	return res.RowsAffected, res.Error
}
