package container

import (
	"context"

	"app/internal/repository"
	"app/internal/service"
	"app/pkg/scheduler"
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
	mediaSvc := service.NewMediaService(
		in.Log,
		in.Storage,
		in.Cfg.Media,
		repository.NewMediaAssetRepository(in.DB),
		repository.NewMediaUsageRepository(in.DB),
	)
	adExpirySvc := service.NewAdExpiryService(in.Log, repository.NewAdCampaignRepository(in.DB))

	sched, err := scheduler.New(in.Log, in.Cache, ctx)
	if err != nil {
		in.Close()
		return nil, err
	}
	if err := sched.Register(map[string]scheduler.JobFunc{
		"region.sync":     regionSvc.Sync,
		"media.cleanup":   mediaSvc.Cleanup,
		"ad.mark_expired": adExpirySvc.MarkExpired,
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
