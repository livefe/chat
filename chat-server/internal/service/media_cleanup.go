package service

import (
	"context"
	"time"

	"app/pkg/logger"
)

const mediaCleanupBatch = 100

func (s *MediaService) Cleanup(ctx context.Context) error {
	if err := s.sweepOrphans(ctx); err != nil {
		return err
	}
	return s.purgeDeleted(ctx)
}

func (s *MediaService) sweepOrphans(ctx context.Context) error {
	before := time.Now().Add(-s.orphanTTL)
	var afterID uint
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		ids, err := s.assetRepo.ListOrphanIDs(ctx, before, afterID, mediaCleanupBatch)
		if err != nil {
			return err
		}
		for _, id := range ids {
			if _, err := s.assetRepo.MarkOrphanDeleted(ctx, id, before); err != nil {
				s.log.Error(ctx, "标记孤儿媒体失败", logger.Uint("mediaId", id), logger.Error(err))
			}
		}
		if len(ids) < mediaCleanupBatch {
			return nil
		}
		afterID = ids[len(ids)-1]
	}
}

func (s *MediaService) purgeDeleted(ctx context.Context) error {
	var afterID uint
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		list, err := s.assetRepo.ListPurgeable(ctx, afterID, mediaCleanupBatch)
		if err != nil {
			return err
		}
		for _, asset := range list {
			if err := s.client.Delete(ctx, asset.ObjectKey); err != nil {
				s.log.Warn(ctx, "删除媒体文件失败", logger.String("key", asset.ObjectKey), logger.Error(err))
				continue
			}
			if err := s.assetRepo.MarkPurged(ctx, asset.ID, time.Now()); err != nil {
				s.log.Error(ctx, "标记媒体已清理失败", logger.Uint("mediaId", asset.ID), logger.Error(err))
			}
		}
		if len(list) < mediaCleanupBatch {
			return nil
		}
		afterID = list[len(list)-1].ID
	}
}
