package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

func (s *AppReleaseService) CheckUpdate(ctx context.Context, req *types.AppUpdateCheckRequest) (*types.AppUpdateCheckResponse, error) {
	platform := strings.TrimSpace(req.Platform)
	clientVer := strings.TrimSpace(req.Version)
	if !validPlatform(platform) {
		return nil, errs.Invalid("平台不正确")
	}
	if !utils.ValidSemver(clientVer) {
		return nil, errs.Invalid("版本号格式不正确，需为 x.y.z")
	}

	policy, err := s.policyRepo.GetByPlatform(ctx, platform)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("未配置该平台更新策略")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	latestVer := ""
	downloadURL := ""
	var fileSize int64
	changelog := ""
	current, err := s.appReleaseRepo.GetCurrent(ctx, platform)
	if err == nil && current != nil {
		latestVer = current.Version
		downloadURL = current.DownloadURL
		fileSize = current.FileSize
		changelog = current.Changelog
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	minVer := strings.TrimSpace(policy.MinSupportedVersion)
	if !utils.ValidSemver(minVer) {
		return nil, errs.Unavailable("最低支持版本未配置", nil)
	}

	cmpMin, err := utils.CompareSemver(clientVer, minVer)
	if err != nil {
		return nil, errs.Invalid("版本号格式不正确")
	}
	force := cmpMin < 0
	hasUpdate := force
	if latestVer != "" {
		cmpLatest, err := utils.CompareSemver(clientVer, latestVer)
		if err != nil {
			return nil, errs.Invalid("版本号格式不正确")
		}
		hasUpdate = cmpLatest < 0 || force
	}

	if force && strings.TrimSpace(downloadURL) == "" && platform != repository.AppPlatformH5 {
		force = false
		hasUpdate = false
	}
	if hasUpdate && strings.TrimSpace(downloadURL) == "" && platform != repository.AppPlatformH5 {
		hasUpdate = false
	}

	title := strings.TrimSpace(policy.Title)
	message := strings.TrimSpace(policy.Message)
	if message == "" {
		message = strings.TrimSpace(changelog)
	}

	return &types.AppUpdateCheckResponse{
		LatestVersion:       latestVer,
		MinSupportedVersion: minVer,
		DownloadURL:         downloadURL,
		FileSize:            fileSize,
		HasUpdate:           hasUpdate,
		ForceUpdate:         force,
		Title:               title,
		Message:             message,
	}, nil
}

func (s *AppReleaseService) AdminListPolicies(ctx context.Context) (*types.AppUpdatePolicyListResponse, error) {
	policies, err := s.policyRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	out := make([]*types.AppUpdatePolicyItem, 0, len(policies))
	for _, p := range policies {
		item := &types.AppUpdatePolicyItem{
			Platform:            p.Platform,
			PlatformText:        platformLabel(p.Platform),
			MinSupportedVersion: p.MinSupportedVersion,
			Title:               p.Title,
			Message:             p.Message,
			UpdatedAt:           utils.FormatDateTime(p.UpdatedAt),
		}
		if cur, err := s.appReleaseRepo.GetCurrent(ctx, p.Platform); err == nil && cur != nil {
			item.LatestVersion = cur.Version
			item.DownloadURL = cur.DownloadURL
		}
		out = append(out, item)
	}
	return &types.AppUpdatePolicyListResponse{List: out}, nil
}

func (s *AppReleaseService) AdminSavePolicy(ctx context.Context, req *types.AppUpdatePolicySaveRequest) error {
	platform := strings.TrimSpace(req.Platform)
	minVer := strings.TrimSpace(req.MinSupportedVersion)
	if !validPlatform(platform) {
		return errs.Invalid("平台不正确")
	}
	if !utils.ValidSemver(minVer) {
		return errs.Invalid("最低版本格式不正确，需为 x.y.z")
	}

	cur, curErr := s.appReleaseRepo.GetCurrent(ctx, platform)
	if errors.Is(curErr, gorm.ErrRecordNotFound) || cur == nil {
		return errs.Invalid("请先设置当前对外版本后再调整最低支持版本")
	}
	if curErr != nil {
		return fmt.Errorf("查询失败: %w", curErr)
	}
	cmp, err := utils.CompareSemver(minVer, cur.Version)
	if err != nil {
		return errs.Invalid("版本号格式不正确")
	}
	if cmp > 0 {
		return errs.Invalid("最低支持版本不能高于当前对外版本 " + cur.Version)
	}
	if strings.TrimSpace(cur.DownloadURL) == "" && platform != repository.AppPlatformH5 {
		return errs.Invalid("当前对外版本缺少下载地址，无法启用强制更新门槛")
	}

	title := strings.TrimSpace(req.Title)
	message := strings.TrimSpace(req.Message)

	row := &model.AppUpdatePolicy{
		Platform:            platform,
		MinSupportedVersion: minVer,
		Title:               title,
		Message:             message,
	}
	if err := s.policyRepo.Upsert(ctx, row); err != nil {
		s.log.Error(ctx, "保存更新策略失败", logger.Error(err))
		return fmt.Errorf("保存失败: %w", err)
	}
	return nil
}
