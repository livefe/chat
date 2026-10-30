package service

import (
	"context"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/pkg/config"
	"app/pkg/logger"
	"app/pkg/storage"
)

type MediaUploader struct {
	Kind uint8
	ID   uint
}

func UserUploader(id uint) MediaUploader {
	return MediaUploader{Kind: model.MediaUploaderUser, ID: id}
}

func EmployeeUploader(id uint) MediaUploader {
	return MediaUploader{Kind: model.MediaUploaderEmployee, ID: id}
}

type MediaService struct {
	log        *logger.Logger
	client     storage.Client
	assetRepo  repository.MediaAssetRepository
	usageRepo  repository.MediaUsageRepository
	limitMB    map[uint8]int
	pathPrefix string
	orphanTTL  time.Duration
}

func NewMediaService(log *logger.Logger, client storage.Client, cfg config.MediaConfig, assetRepo repository.MediaAssetRepository, usageRepo repository.MediaUsageRepository) *MediaService {
	return &MediaService{
		log:       log,
		client:    client,
		assetRepo: assetRepo,
		usageRepo: usageRepo,
		limitMB: map[uint8]int{
			model.MediaKindImage: cfg.MaxImageMB,
			model.MediaKindVideo: cfg.MaxVideoMB,
			model.MediaKindAudio: cfg.MaxAudioMB,
			model.MediaKindFile:  cfg.MaxFileMB,
		},
		pathPrefix: strings.Trim(strings.TrimSpace(cfg.PathPrefix), "/"),
		orphanTTL:  time.Duration(cfg.OrphanTTLHours) * time.Hour,
	}
}

func (s *MediaService) MaxUploadBytes(kind uint8) int64 {
	return int64(s.limitMB[kind]) << 20
}

func mediaUsable(asset *model.MediaAsset) bool {
	return asset.Status == model.MediaStatusUploaded || asset.Status == model.MediaStatusBound
}

func (s *MediaService) ResolveURL(ctx context.Context, mediaID uint) string {
	if mediaID == 0 {
		return ""
	}
	asset, err := s.assetRepo.GetByID(ctx, mediaID)
	if err != nil || !mediaUsable(asset) {
		return ""
	}
	return s.client.URL(asset.ObjectKey)
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
		if !mediaUsable(a) {
			continue
		}
		out[a.ID] = s.client.URL(a.ObjectKey)
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
