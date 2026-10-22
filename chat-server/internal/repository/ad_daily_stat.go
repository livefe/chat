package repository

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdDailyStatRepository interface {
	Incr(ctx context.Context, campaignID uint, date string, impressions, clicks int64) error
	Get(ctx context.Context, campaignID uint, date string) (*model.AdDailyStat, error)
	ListByCampaign(ctx context.Context, campaignID uint, fromDate, toDate string) ([]*model.AdDailyStat, error)
	MapTodayByCampaignIDs(ctx context.Context, campaignIDs []uint, date string) (map[uint]*model.AdDailyStat, error)
}

type adDailyStatRepository struct{ db *database.DB }

func NewAdDailyStatRepository(db *database.DB) AdDailyStatRepository {
	return &adDailyStatRepository{db: db}
}

func (r *adDailyStatRepository) Incr(ctx context.Context, campaignID uint, date string, impressions, clicks int64) error {
	item := &model.AdDailyStat{
		CampaignID:  campaignID,
		StatDate:    date,
		Impressions: impressions,
		Clicks:      clicks,
	}
	return database.Conn(ctx, r.db).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "campaign_id"}, {Name: "stat_date"}},
		DoUpdates: clause.Assignments(map[string]any{
			"impressions": gorm.Expr("impressions + ?", impressions),
			"clicks":      gorm.Expr("clicks + ?", clicks),
		}),
	}).Create(item).Error
}

func (r *adDailyStatRepository) Get(ctx context.Context, campaignID uint, date string) (*model.AdDailyStat, error) {
	var item model.AdDailyStat
	err := database.Conn(ctx, r.db).Where("campaign_id = ? AND stat_date = ?", campaignID, date).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *adDailyStatRepository) ListByCampaign(ctx context.Context, campaignID uint, fromDate, toDate string) ([]*model.AdDailyStat, error) {
	var list []*model.AdDailyStat
	err := database.Conn(ctx, r.db).
		Where("campaign_id = ? AND stat_date >= ? AND stat_date <= ?", campaignID, fromDate, toDate).
		Order("stat_date DESC").
		Find(&list).Error
	return list, err
}

func (r *adDailyStatRepository) MapTodayByCampaignIDs(ctx context.Context, campaignIDs []uint, date string) (map[uint]*model.AdDailyStat, error) {
	out := make(map[uint]*model.AdDailyStat, len(campaignIDs))
	if len(campaignIDs) == 0 {
		return out, nil
	}
	var list []*model.AdDailyStat
	if err := database.Conn(ctx, r.db).Where("campaign_id IN ? AND stat_date = ?", campaignIDs, date).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, it := range list {
		out[it.CampaignID] = it
	}
	return out, nil
}

func TodayDate() string {
	return time.Now().Format("2006-01-02")
}

func AdStatCTR(impressions, clicks int64) float64 {
	if impressions <= 0 {
		return 0
	}
	return float64(clicks) / float64(impressions)
}
