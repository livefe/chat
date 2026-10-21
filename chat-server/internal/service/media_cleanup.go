package service

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/logger"
)

func (s *MediaService) CleanupOrphans(ctx context.Context) error {
	before := time.Now().Add(-s.orphanTTL)
	const batch = 100
	for {
		list, err := s.assetRepo.ListOrphans(ctx, before, batch)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			break
		}
		for _, asset := range list {
			asset.Status = model.MediaStatusDeleted
			if err := s.assetRepo.Update(ctx, asset); err != nil {
				s.log.Error(ctx, "标记孤儿媒体失败", logger.Uint("id", asset.ID), logger.Error(err))
				continue
			}
			if err := s.client.Delete(ctx, asset.ObjectKey); err != nil {
				s.log.Warn(ctx, "删除孤儿文件失败", logger.String("key", asset.ObjectKey), logger.Error(err))
			}
		}
		if len(list) < batch {
			break
		}
	}

	for {
		list, err := s.assetRepo.ListDeletedPendingPurge(ctx, batch)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			break
		}
		for _, asset := range list {
			if err := s.client.Delete(ctx, asset.ObjectKey); err != nil {
				s.log.Warn(ctx, "重试删除媒体文件失败", logger.String("key", asset.ObjectKey), logger.Error(err))
				continue
			}
			asset.ObjectKey = ""
			if err := s.assetRepo.Update(ctx, asset); err != nil {
				s.log.Error(ctx, "清空已删媒体 key 失败", logger.Uint("id", asset.ID), logger.Error(err))
			}
		}
		if len(list) < batch {
			break
		}
	}
	return nil
}
