package repository

import (
	"context"
	"strings"

	"app/internal/model"
	"app/pkg/database"
)

// AdCampaignListFilter 的 From/To 为 Unix 秒，筛选排期与 [From, To) 有交集的计划。
type AdCampaignListFilter struct {
	Status         *uint8
	Keyword        string
	AdvertiserName string
	From           int64
	To             int64
	Page           int
	PageSize       int
}

type AdCampaignRepository interface {
	Create(ctx context.Context, item *model.AdCampaign) error
	Update(ctx context.Context, item *model.AdCampaign) error
	GetByID(ctx context.Context, id uint) (*model.AdCampaign, error)
	List(ctx context.Context, f AdCampaignListFilter) ([]*model.AdCampaign, int64, error)
	ListAdvertisers(ctx context.Context) ([]string, error)
	ListActiveCandidates(ctx context.Context, now int64) ([]*model.AdCampaign, error)
	CountActiveByAdvertiser(ctx context.Context, advertiserName string, excludeID uint) (int64, error)
	UpdateStatus(ctx context.Context, id uint, status uint8, rejectRemark string, landingInvalid bool) error
	PauseActive(ctx context.Context, ids []uint) (int64, error)
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

func (r *adCampaignRepository) List(ctx context.Context, f AdCampaignListFilter) ([]*model.AdCampaign, int64, error) {
	q := database.Conn(ctx, r.db).Model(&model.AdCampaign{})
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if kw := strings.TrimSpace(f.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("(advertiser_name LIKE ? OR title LIKE ? OR name LIKE ?)", like, like, like)
	}
	if name := strings.TrimSpace(f.AdvertiserName); name != "" {
		q = q.Where("advertiser_name = ?", name)
	}
	if f.From > 0 {
		q = q.Where("end_at >= ?", f.From)
	}
	if f.To > 0 {
		q = q.Where("start_at < ?", f.To)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.AdCampaign
	err := q.Order("id DESC").Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

func (r *adCampaignRepository) ListAdvertisers(ctx context.Context) ([]string, error) {
	var names []string
	err := database.Conn(ctx, r.db).Model(&model.AdCampaign{}).
		Where("advertiser_name <> ''").
		Distinct("advertiser_name").
		Order("advertiser_name ASC").
		Pluck("advertiser_name", &names).Error
	return names, err
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

// PauseActive 只暂停投放中的计划，返回实际暂停的条数。
func (r *adCampaignRepository) PauseActive(ctx context.Context, ids []uint) (int64, error) {
	res := database.Conn(ctx, r.db).Model(&model.AdCampaign{}).
		Where("id IN ? AND status = ?", ids, model.AdStatusActive).
		Update("status", model.AdStatusPaused)
	return res.RowsAffected, res.Error
}

func (r *adCampaignRepository) MarkEndedExpired(ctx context.Context, now int64) (int64, error) {
	res := database.Conn(ctx, r.db).Model(&model.AdCampaign{}).
		Where("status = ? AND end_at > 0 AND end_at < ?", model.AdStatusActive, now).
		Update("status", model.AdStatusEnded)
	return res.RowsAffected, res.Error
}
