package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/pkg/logger"
	"app/pkg/storage"
)

type MediaService struct {
	log            *logger.Logger
	client         storage.Client
	assetRepo      repository.MediaAssetRepository
	usageRepo      repository.MediaUsageRepository
	maxSizeMB      int
	maxVideoSizeMB int
	maxAudioSizeMB int
	pathPrefix     string
	orphanTTL      time.Duration
}

func NewMediaService(log *logger.Logger, client storage.Client, cfg *storage.Config, assetRepo repository.MediaAssetRepository, usageRepo repository.MediaUsageRepository) (*MediaService, error) {
	if client == nil {
		return nil, fmt.Errorf("对象存储未初始化")
	}
	svc := &MediaService{
		log:            log,
		client:         client,
		assetRepo:      assetRepo,
		usageRepo:      usageRepo,
		maxVideoSizeMB: 50,
		maxAudioSizeMB: 5,
		orphanTTL:      24 * time.Hour,
	}
	if cfg != nil {
		svc.maxSizeMB = cfg.MaxSizeMB
		if cfg.MaxVideoSizeMB > 0 {
			svc.maxVideoSizeMB = cfg.MaxVideoSizeMB
		}
		if cfg.MaxAudioSizeMB > 0 {
			svc.maxAudioSizeMB = cfg.MaxAudioSizeMB
		}
		svc.pathPrefix = strings.Trim(strings.TrimSpace(cfg.PathPrefix), "/")
		if cfg.OrphanTTLHours > 0 {
			svc.orphanTTL = time.Duration(cfg.OrphanTTLHours) * time.Hour
		}
	}
	return svc, nil
}

func (s *MediaService) ResolveURL(ctx context.Context, mediaID uint) string {
	if mediaID == 0 {
		return ""
	}
	asset, err := s.assetRepo.GetByID(ctx, mediaID)
	if err != nil || asset.Status == model.MediaStatusDeleted {
		return ""
	}
	return asset.URL
}

func (s *MediaService) ResolveURLs(ctx context.Context, ids []uint) map[uint]string {
	out := make(map[uint]string, len(ids))
	if len(ids) == 0 {
		return out
	}
	uniq := make([]uint, 0, len(ids))
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	list, err := s.assetRepo.GetByIDs(ctx, uniq)
	if err != nil {
		return out
	}
	for _, a := range list {
		if a.Status == model.MediaStatusDeleted {
			continue
		}
		out[a.ID] = a.URL
	}
	return out
}

func (s *MediaService) ResolveURLList(ctx context.Context, ids []uint) []string {
	urls := s.ResolveURLs(ctx, ids)
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if u := urls[id]; u != "" {
			out = append(out, u)
		}
	}
	return out
}
