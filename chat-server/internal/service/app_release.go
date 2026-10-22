package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/storage"
	"app/pkg/utils"

	"gorm.io/gorm"
)

const appObjectDir = "app"

var platformText = map[string]string{
	repository.AppPlatformAndroid: "Android",
	repository.AppPlatformIOS:     "iOS",
	repository.AppPlatformH5:      "H5",
}

type AppReleaseService struct {
	log            *logger.Logger
	db             *database.DB
	appReleaseRepo repository.AppReleaseRepository
	policyRepo     repository.AppUpdatePolicyRepository
	store          storage.Client
}

func NewAppReleaseService(
	log *logger.Logger,
	db *database.DB,
	appReleaseRepo repository.AppReleaseRepository,
	policyRepo repository.AppUpdatePolicyRepository,
	store storage.Client,
) *AppReleaseService {
	return &AppReleaseService{
		log:            log,
		db:             db,
		appReleaseRepo: appReleaseRepo,
		policyRepo:     policyRepo,
		store:          store,
	}
}

func (s *AppReleaseService) ListCurrent(ctx context.Context) (*types.AppReleasePublicResponse, error) {
	list, err := s.appReleaseRepo.ListCurrent(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items := make([]*types.AppReleasePublicItem, 0, len(list))
	for _, row := range list {
		items = append(items, &types.AppReleasePublicItem{
			Platform:     row.Platform,
			PlatformText: platformLabel(row.Platform),
			Version:      row.Version,
			DownloadURL:  row.DownloadURL,
			FileSize:     row.FileSize,
			Changelog:    row.Changelog,
		})
	}
	return &types.AppReleasePublicResponse{List: items}, nil
}

func (s *AppReleaseService) AdminList(ctx context.Context, req *types.AppReleaseAdminListRequest) (*types.AppReleaseListResponse, error) {
	list, total, err := s.appReleaseRepo.List(ctx, req.Platform, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items := make([]*types.AppReleaseItem, 0, len(list))
	for _, row := range list {
		items = append(items, toAppReleaseItem(row))
	}
	return &types.AppReleaseListResponse{Total: total, List: items}, nil
}

func (s *AppReleaseService) AdminUpload(ctx context.Context, platform, version, changelog, filename string, reader io.Reader, size int64) (*types.AppReleaseItem, error) {
	platform = strings.TrimSpace(platform)
	version = strings.TrimSpace(version)
	changelog = strings.TrimSpace(changelog)
	if !validPlatform(platform) {
		return nil, errs.Invalid("请选择平台")
	}
	if platform == repository.AppPlatformH5 {
		return nil, errs.Invalid("H5 请使用发布入口登记版本")
	}
	if !utils.ValidSemver(version) {
		return nil, errs.Invalid("版本号格式不正确，需为 x.y.z")
	}
	ext := packageExtFor(platform)
	if strings.ToLower(path.Ext(filename)) != ext {
		return nil, errs.Invalid(fmt.Sprintf("%s 安装包须为 %s 文件", platformLabel(platform), ext))
	}

	exist, err := s.appReleaseRepo.GetByPlatformVersion(ctx, platform, version)
	if err == nil && exist != nil {
		return nil, errs.Conflict(fmt.Sprintf("该平台下版本 %s 已存在", version))
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	objectKey := path.Join(appObjectDir, platform+"-"+version+ext)
	url, err := s.uploadAppPackage(ctx, objectKey, reader, size)
	if err != nil {
		return nil, err
	}

	row := &model.AppRelease{
		Platform:    platform,
		Version:     version,
		DownloadURL: url,
		FileSize:    size,
		Changelog:   changelog,
		IsCurrent:   0,
	}
	if err := s.appReleaseRepo.Create(ctx, row); err != nil {
		s.log.Error(ctx, "保存安装包记录失败", logger.Error(err))
		return nil, fmt.Errorf("保存失败，请稍后重试: %w", err)
	}
	return toAppReleaseItem(row), nil
}

func (s *AppReleaseService) AdminPublishMeta(ctx context.Context, req *types.AppReleasePublishMetaRequest) (*types.AppReleaseItem, error) {
	platform := strings.TrimSpace(req.Platform)
	version := strings.TrimSpace(req.Version)
	downloadURL := strings.TrimSpace(req.DownloadURL)
	changelog := strings.TrimSpace(req.Changelog)
	if platform != repository.AppPlatformH5 {
		return nil, errs.Invalid("仅支持 H5 登记版本")
	}
	if !utils.ValidSemver(version) {
		return nil, errs.Invalid("版本号格式不正确，需为 x.y.z")
	}
	if downloadURL == "" {
		return nil, errs.Invalid("请填写入口地址")
	}
	if !strings.HasPrefix(downloadURL, "https://") && !strings.HasPrefix(downloadURL, "http://") {
		return nil, errs.Invalid("入口地址必须以 http:// 或 https:// 开头")
	}

	exist, err := s.appReleaseRepo.GetByPlatformVersion(ctx, platform, version)
	if err == nil && exist != nil {
		return nil, errs.Conflict(fmt.Sprintf("该平台下版本 %s 已存在", version))
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	row := &model.AppRelease{
		Platform:    platform,
		Version:     version,
		DownloadURL: downloadURL,
		FileSize:    0,
		Changelog:   changelog,
		IsCurrent:   0,
	}
	if err := s.appReleaseRepo.Create(ctx, row); err != nil {
		s.log.Error(ctx, "保存 H5 版本记录失败", logger.Error(err))
		return nil, fmt.Errorf("保存失败，请稍后重试: %w", err)
	}
	return toAppReleaseItem(row), nil
}

func (s *AppReleaseService) AdminSetCurrent(ctx context.Context, id uint) error {
	row, err := s.appReleaseRepo.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NotFound("记录不存在")
	}
	if err != nil {
		return fmt.Errorf("查询失败: %w", err)
	}
	if strings.TrimSpace(row.DownloadURL) == "" && row.Platform != repository.AppPlatformH5 {
		return errs.Invalid("该版本缺少下载地址，不能设为当前对外版本")
	}
	err = database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if err := s.appReleaseRepo.SetCurrent(ctx, row.Platform, id); err != nil {
			return err
		}
		return s.lowerMinSupportedVersion(ctx, row.Platform, row.Version)
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NotFound("记录不存在")
	}
	if err != nil {
		s.log.Error(ctx, "设置当前安装包失败", logger.Error(err))
		return fmt.Errorf("操作失败，请稍后重试: %w", err)
	}
	return nil
}

func (s *AppReleaseService) lowerMinSupportedVersion(ctx context.Context, platform, version string) error {
	policy, err := s.policyRepo.GetByPlatform(ctx, platform)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if !utils.ValidSemver(policy.MinSupportedVersion) || !utils.ValidSemver(version) {
		return nil
	}
	cmp, err := utils.CompareSemver(version, policy.MinSupportedVersion)
	if err != nil || cmp >= 0 {
		return err
	}
	policy.MinSupportedVersion = version
	return s.policyRepo.Upsert(ctx, policy)
}

func (s *AppReleaseService) AdminUnsetCurrent(ctx context.Context, id uint) error {
	if err := s.appReleaseRepo.ClearCurrent(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.Invalid("该版本不是当前对外版本")
		}
		s.log.Error(ctx, "下架安装包失败", logger.Error(err))
		return fmt.Errorf("操作失败，请稍后重试: %w", err)
	}
	return nil
}

// AdminDelete 删除误传或作废的非当前版本，安装包文件一并从存储中删除。
func (s *AppReleaseService) AdminDelete(ctx context.Context, id uint) error {
	row, err := s.appReleaseRepo.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NotFound("记录不存在")
	}
	if err != nil {
		return fmt.Errorf("查询失败: %w", err)
	}
	if row.IsCurrent == 1 {
		return errs.Invalid("当前对外版本不能删除，请先下架或切换到其他版本")
	}
	if err := s.appReleaseRepo.DeleteNotCurrent(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.Invalid("该版本已是当前对外版本或已被删除")
		}
		s.log.Error(ctx, "删除版本记录失败", logger.Error(err))
		return fmt.Errorf("删除失败，请稍后重试: %w", err)
	}
	if row.Platform != repository.AppPlatformH5 {
		objectKey := path.Join(appObjectDir, row.Platform+"-"+row.Version+packageExtFor(row.Platform))
		if err := s.store.Delete(ctx, objectKey); err != nil {
			s.log.Warn(ctx, "删除安装包文件失败", logger.String("key", objectKey), logger.Error(err))
		}
	}
	return nil
}

func validPlatform(platform string) bool {
	_, ok := platformText[platform]
	return ok
}

func platformLabel(platform string) string {
	if t, ok := platformText[platform]; ok {
		return t
	}
	return platform
}

func packageExtFor(platform string) string {
	if platform == repository.AppPlatformIOS {
		return ".ipa"
	}
	return ".apk"
}

const packageMaxSizeMB = 100
const packageMaxBytes = int64(packageMaxSizeMB) * 1024 * 1024

func (s *AppReleaseService) uploadAppPackage(ctx context.Context, objectKey string, reader io.Reader, size int64) (string, error) {
	objectKey = strings.Trim(strings.TrimSpace(objectKey), "/")
	if objectKey == "" || strings.Contains(objectKey, "..") {
		return "", errs.Invalid("非法安装包路径")
	}
	if size <= 0 {
		return "", errs.Invalid("文件不能为空")
	}
	if size > packageMaxBytes {
		return "", errs.Invalid(fmt.Sprintf("安装包大小不能超过 %dMB", packageMaxSizeMB))
	}

	limited := &limitedReader{r: reader, remaining: packageMaxBytes}
	url, err := s.store.Upload(objectKey, limited, "application/octet-stream")
	if err != nil {
		if errors.Is(err, errReaderTooLarge) {
			return "", errs.Invalid(fmt.Sprintf("安装包大小不能超过 %dMB", packageMaxSizeMB))
		}
		s.log.Error(ctx, "上传安装包失败", logger.Error(err))
		return "", fmt.Errorf("上传失败，请稍后重试: %w", err)
	}
	return url, nil
}

func toAppReleaseItem(row *model.AppRelease) *types.AppReleaseItem {
	return &types.AppReleaseItem{
		ID:           row.ID,
		Platform:     row.Platform,
		PlatformText: platformLabel(row.Platform),
		Version:      row.Version,
		DownloadURL:  row.DownloadURL,
		FileSize:     row.FileSize,
		Changelog:    row.Changelog,
		IsCurrent:    row.IsCurrent,
		CreatedAt:    utils.FormatDateTime(row.CreatedAt),
	}
}
