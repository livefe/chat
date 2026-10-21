package container

import (
	"context"
	"fmt"

	"app/internal/repository"
	"app/internal/service"
	"app/pkg/scheduler"
	"app/pkg/storage"
)

type WorkerContainer struct {
	Infra     *Infra
	Scheduler *scheduler.Scheduler
}

func NewWorker(ctx context.Context) (*WorkerContainer, error) {
	in, err := NewInfra()
	if err != nil {
		return nil, err
	}

	regionSvc := service.NewRegionService(in.Log, repository.NewRegionRepository(in.DB))

	store, err := storage.New(&in.Cfg.Storage)
	if err != nil {
		in.Close()
		return nil, fmt.Errorf("对象存储初始化失败: %w", err)
	}
	mediaSvc, err := service.NewMediaService(
		in.Log,
		store,
		&in.Cfg.Storage,
		repository.NewMediaAssetRepository(in.DB),
		repository.NewMediaUsageRepository(in.DB),
	)
	if err != nil {
		in.Close()
		return nil, fmt.Errorf("媒体服务初始化失败: %w", err)
	}

	adExpirySvc := service.NewAdExpiryService(in.Log, repository.NewAdCampaignRepository(in.DB))

	sched, err := scheduler.New(in.Log, in.Cache, ctx)
	if err != nil {
		in.Close()
		return nil, err
	}
	if err := sched.Register(map[string]scheduler.JobFunc{
		"region.sync":           regionSvc.Sync,
		"media.cleanup_orphans": mediaSvc.CleanupOrphans,
		"ad.mark_expired":       adExpirySvc.MarkExpired,
	}); err != nil {
		in.Close()
		return nil, err
	}

	return &WorkerContainer{
		Infra:     in,
		Scheduler: sched,
	}, nil
}

func (c *WorkerContainer) Close() {
	c.Infra.Close()
}
