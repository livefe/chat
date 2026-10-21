package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"app/internal/model"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

func (s *MediaService) AssertOwn(ctx context.Context, userID, mediaID uint, expectKind uint8) (*model.MediaAsset, error) {
	if mediaID == 0 {
		return nil, errs.Invalid("媒体无效")
	}
	asset, err := s.assetRepo.GetByID(ctx, mediaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.Invalid("媒体不存在")
		}
		return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if asset.Status == model.MediaStatusDeleted {
		return nil, errs.Invalid("媒体已失效")
	}
	if asset.UserID != userID {
		return nil, errs.Forbidden("无权使用该媒体")
	}
	if expectKind != 0 && asset.Kind != expectKind {
		return nil, errs.Invalid("媒体类型不匹配")
	}
	return asset, nil
}

func (s *MediaService) Bind(ctx context.Context, userID, mediaID uint, expectKind uint8, bizType string, bizID uint) (*model.MediaAsset, error) {
	asset, err := s.AssertOwn(ctx, userID, mediaID, expectKind)
	if err != nil {
		return nil, err
	}
	if bizType == "" || bizID == 0 {
		return nil, errs.Invalid("业务引用无效")
	}
	if err := s.usageRepo.Create(ctx, &model.MediaUsage{
		MediaID: mediaID,
		BizType: bizType,
		BizID:   bizID,
		UserID:  userID,
	}); err != nil {
		s.log.Error(ctx, "绑定媒体失败", logger.Error(err))
		return nil, fmt.Errorf("绑定媒体失败: %w", err)
	}
	if asset.Status == model.MediaStatusUploaded {
		now := time.Now()
		asset.Status = model.MediaStatusBound
		asset.BoundAt = &now
		if err := s.assetRepo.Update(ctx, asset); err != nil {
			s.log.Error(ctx, "更新媒体状态失败", logger.Error(err))
			return nil, fmt.Errorf("绑定媒体失败: %w", err)
		}
	}
	return asset, nil
}

func (s *MediaService) BindMany(ctx context.Context, userID uint, mediaIDs []uint, expectKind uint8, bizType string, bizID uint) ([]*model.MediaAsset, error) {
	out := make([]*model.MediaAsset, 0, len(mediaIDs))
	seen := make(map[uint]struct{}, len(mediaIDs))
	for _, id := range mediaIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		asset, err := s.Bind(ctx, userID, id, expectKind, bizType, bizID)
		if err != nil {
			return nil, err
		}
		out = append(out, asset)
	}
	return out, nil
}

func (s *MediaService) ReplaceBizMedia(ctx context.Context, userID uint, mediaIDs []uint, expectKind uint8, bizType string, bizID uint) ([]*model.MediaAsset, error) {
	oldIDs, err := s.usageRepo.ListMediaIDsByBiz(ctx, bizType, bizID)
	if err != nil {
		s.log.Error(ctx, "查询业务媒体引用失败", logger.Error(err))
		return nil, fmt.Errorf("更新媒体失败: %w", err)
	}
	if err := s.usageRepo.DeleteByBiz(ctx, bizType, bizID); err != nil {
		s.log.Error(ctx, "清理业务媒体引用失败", logger.Error(err))
		return nil, fmt.Errorf("更新媒体失败: %w", err)
	}
	for _, id := range oldIDs {
		_ = s.releaseIfUnused(ctx, id)
	}
	return s.BindMany(ctx, userID, mediaIDs, expectKind, bizType, bizID)
}

func (s *MediaService) UnbindOne(ctx context.Context, mediaID uint, bizType string, bizID uint) error {
	if mediaID == 0 {
		return nil
	}
	if err := s.usageRepo.Delete(ctx, mediaID, bizType, bizID); err != nil {
		s.log.Error(ctx, "解除媒体引用失败", logger.Error(err), logger.Uint("mediaId", mediaID))
		return fmt.Errorf("解除媒体引用失败: %w", err)
	}
	if err := s.releaseIfUnused(ctx, mediaID); err != nil {
		s.log.Error(ctx, "释放媒体失败", logger.Error(err), logger.Uint("mediaId", mediaID))
		return fmt.Errorf("解除媒体引用失败: %w", err)
	}
	return nil
}

func (s *MediaService) releaseIfUnused(ctx context.Context, mediaID uint) error {
	n, err := s.usageRepo.CountByMediaID(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("调用 CountUsages 失败: %w", err)
	}
	if n > 0 {
		return nil
	}
	asset, err := s.assetRepo.GetByID(ctx, mediaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if asset.Status == model.MediaStatusDeleted {
		return nil
	}
	asset.Status = model.MediaStatusDeleted
	if err := s.assetRepo.Update(ctx, asset); err != nil {
		return err
	}
	if err := s.client.Delete(ctx, asset.ObjectKey); err != nil {
		s.log.Warn(ctx, "删除媒体文件失败", logger.String("key", asset.ObjectKey), logger.Error(err))
	}
	return nil
}
