package service

import (
	"context"
	"fmt"
	"time"

	"app/internal/repository"
	"app/pkg/logger"
)

type AdExpiryService struct {
	log          *logger.Logger
	campaignRepo repository.AdCampaignRepository
}

func NewAdExpiryService(log *logger.Logger, campaignRepo repository.AdCampaignRepository) *AdExpiryService {
	return &AdExpiryService{log: log, campaignRepo: campaignRepo}
}

func (s *AdExpiryService) MarkExpired(ctx context.Context) error {
	n, err := s.campaignRepo.MarkEndedExpired(ctx, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("标记过期广告失败: %w", err)
	}
	if n > 0 {
		s.log.Info(ctx, "已结束过期广告", logger.Int64("count", n))
	}
	return nil
}
