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

func (s *MediaService) AssertOwn(ctx context.Context, uploader MediaUploader, mediaID uint, expectKind uint8) (*model.MediaAsset, error) {
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
	if !mediaUsable(asset) {
		return nil, errs.Invalid("媒体已失效")
	}
	if asset.UploaderKind != uploader.Kind || asset.UploaderID != uploader.ID {
		return nil, errs.Forbidden("无权使用该媒体")
	}
	if expectKind != 0 && asset.Kind != expectKind {
		return nil, errs.Invalid("媒体类型不匹配")
	}
	return asset, nil
}

func (s *MediaService) Bind(ctx context.Context, uploader MediaUploader, mediaID uint, expectKind uint8, bizType string, bizID uint) error {
	if bizType == "" || bizID == 0 {
		return errs.Invalid("业务引用无效")
	}
	if _, err := s.AssertOwn(ctx, uploader, mediaID, expectKind); err != nil {
		return err
	}
	return s.attach(ctx, mediaID, bizType, bizID)
}

func (s *MediaService) attach(ctx context.Context, mediaID uint, bizType string, bizID uint) error {
	if err := s.usageRepo.Create(ctx, &model.MediaUsage{MediaID: mediaID, BizType: bizType, BizID: bizID}); err != nil {
		s.log.Error(ctx, "绑定媒体失败", logger.Uint("mediaId", mediaID), logger.Error(err))
		return fmt.Errorf("绑定媒体失败: %w", err)
	}
	if err := s.assetRepo.PromoteBound(ctx, mediaID, time.Now()); err != nil {
		s.log.Error(ctx, "更新媒体状态失败", logger.Uint("mediaId", mediaID), logger.Error(err))
		return fmt.Errorf("绑定媒体失败: %w", err)
	}
	bound, err := s.assetRepo.LockBound(ctx, mediaID)
	if err != nil {
		s.log.Error(ctx, "确认媒体状态失败", logger.Uint("mediaId", mediaID), logger.Error(err))
		return fmt.Errorf("绑定媒体失败: %w", err)
	}
	if !bound {
		if err := s.usageRepo.Delete(ctx, mediaID, bizType, bizID); err != nil {
			s.log.Error(ctx, "撤回媒体引用失败", logger.Uint("mediaId", mediaID), logger.Error(err))
		}
		return errs.Invalid("媒体已失效")
	}
	return nil
}

func (s *MediaService) BindShared(ctx context.Context, mediaID uint, bizType string, bizID uint) error {
	if mediaID == 0 || bizType == "" || bizID == 0 {
		return errs.Invalid("业务引用无效")
	}
	return s.attach(ctx, mediaID, bizType, bizID)
}

func (s *MediaService) BindMany(ctx context.Context, uploader MediaUploader, mediaIDs []uint, expectKind uint8, bizType string, bizID uint) error {
	for _, id := range uniqueUints(mediaIDs) {
		if err := s.Bind(ctx, uploader, id, expectKind, bizType, bizID); err != nil {
			return err
		}
	}
	return nil
}

func (s *MediaService) ReplaceBizMedia(ctx context.Context, uploader MediaUploader, mediaIDs []uint, expectKind uint8, bizType string, bizID uint) error {
	if bizType == "" || bizID == 0 {
		return errs.Invalid("业务引用无效")
	}
	oldIDs, err := s.usageRepo.ListMediaIDsByBiz(ctx, bizType, bizID)
	if err != nil {
		s.log.Error(ctx, "查询业务媒体引用失败", logger.Error(err))
		return fmt.Errorf("更新媒体失败: %w", err)
	}
	wanted := uniqueUints(mediaIDs)
	for _, id := range wanted {
		if containsUint(oldIDs, id) {
			continue
		}
		if err := s.Bind(ctx, uploader, id, expectKind, bizType, bizID); err != nil {
			return err
		}
	}
	for _, id := range oldIDs {
		if containsUint(wanted, id) {
			continue
		}
		if err := s.UnbindOne(ctx, id, bizType, bizID); err != nil {
			return err
		}
	}
	return nil
}

func (s *MediaService) ReplaceSingle(ctx context.Context, uploader MediaUploader, oldID, newID uint, expectKind uint8, bizType string, bizID uint, persist func() error) error {
	if newID != 0 && newID != oldID {
		if err := s.Bind(ctx, uploader, newID, expectKind, bizType, bizID); err != nil {
			return err
		}
	}
	if err := persist(); err != nil {
		return err
	}
	if oldID != 0 && oldID != newID {
		if err := s.UnbindOne(ctx, oldID, bizType, bizID); err != nil {
			s.log.Warn(ctx, "解绑旧媒体失败", logger.Uint("mediaId", oldID), logger.Error(err))
		}
	}
	return nil
}

func (s *MediaService) UnbindOne(ctx context.Context, mediaID uint, bizType string, bizID uint) error {
	if mediaID == 0 {
		return nil
	}
	if err := s.usageRepo.Delete(ctx, mediaID, bizType, bizID); err != nil {
		s.log.Error(ctx, "解除媒体引用失败", logger.Error(err), logger.Uint("mediaId", mediaID))
		return fmt.Errorf("解除媒体引用失败: %w", err)
	}
	if err := s.assetRepo.MarkDeletedIfUnused(ctx, mediaID); err != nil {
		s.log.Error(ctx, "释放媒体失败", logger.Error(err), logger.Uint("mediaId", mediaID))
		return fmt.Errorf("解除媒体引用失败: %w", err)
	}
	return nil
}
