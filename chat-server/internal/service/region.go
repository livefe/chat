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

	"gorm.io/gorm"
)

type RegionService struct {
	log        *logger.Logger
	regionRepo repository.RegionRepository
}

func NewRegionService(
	log *logger.Logger,
	regionRepo repository.RegionRepository,
) *RegionService {
	return &RegionService{
		log:        log,
		regionRepo: regionRepo,
	}
}

func (s *RegionService) ListChildren(ctx context.Context, parentCode string) (*types.RegionListResponse, error) {
	list, err := s.regionRepo.ListChildren(ctx, parentCode)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return &types.RegionListResponse{List: toRegionItems(list)}, nil
}

func (s *RegionService) Search(ctx context.Context, req *types.RegionSearchRequest) (*types.RegionSearchResponse, error) {
	keyword := strings.TrimSpace(req.Keyword)
	byCode := isDigits(keyword)
	list, err := s.regionRepo.Search(ctx, keyword, byCode, req.Level, req.Limit+1)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	truncated := len(list) > req.Limit
	if truncated {
		list = list[:req.Limit]
	}
	return &types.RegionSearchResponse{List: toRegionItems(list), Truncated: truncated}, nil
}

// Trail 返回从省级到该区划自身的完整路径。
func (s *RegionService) Trail(ctx context.Context, code string) (*types.RegionListResponse, error) {
	trail := make([]*model.Region, 0, 4)
	next := normalizeRegionCode(code)
	for next != "" && len(trail) <= int(model.RegionLevelTownship) {
		item, err := s.regionRepo.GetByCode(ctx, next)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errs.NotFound("区划不存在或已撤销")
			}
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		trail = append([]*model.Region{item}, trail...)
		next = item.ParentCode
	}
	return &types.RegionListResponse{List: toRegionItems(trail)}, nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (s *RegionService) Get(ctx context.Context, code string) (*types.RegionItem, error) {
	item, err := s.regionRepo.GetByCode(ctx, normalizeRegionCode(code))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("区划不存在或已撤销")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	if item.Status != model.RegionStatusActive {
		return nil, errs.NotFound("区划不存在或已撤销")
	}
	row := toRegionItem(item)
	return &row, nil
}

func (s *RegionService) AdminGet(ctx context.Context, code string) (*types.RegionAdminItem, error) {
	item, err := s.regionRepo.GetByCode(ctx, normalizeRegionCode(code))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("区划不存在或已撤销")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	row := toRegionAdminItem(item)
	return &row, nil
}

func toRegionItems(list []*model.Region) []types.RegionItem {
	items := make([]types.RegionItem, 0, len(list))
	for _, row := range list {
		items = append(items, toRegionItem(row))
	}
	return items
}

func toRegionItem(row *model.Region) types.RegionItem {
	return types.RegionItem{
		Code:       row.Code,
		Name:       row.Name,
		ShortName:  row.ShortName,
		ParentCode: row.ParentCode,
		Level:      row.Level,
		FullPath:   row.FullPath,
		TiandituGB: row.TiandituGB,
		AmapAdcode: row.AmapAdcode,
	}
}

func toRegionAdminItem(row *model.Region) types.RegionAdminItem {
	return types.RegionAdminItem{
		Code:       row.Code,
		Name:       row.Name,
		ShortName:  row.ShortName,
		ParentCode: row.ParentCode,
		Level:      row.Level,
		LevelText:  regionLevelText(row.Level),
		FullPath:   row.FullPath,
		TiandituGB: row.TiandituGB,
		AmapAdcode: row.AmapAdcode,
		Status:     row.Status,
		StatusText: regionStatusText(row.Status),
		DataYear:   row.DataYear,
	}
}

func regionLevelText(level uint8) string {
	switch level {
	case model.RegionLevelProvince:
		return "省级"
	case model.RegionLevelCity:
		return "市级"
	case model.RegionLevelDistrict:
		return "区县级"
	case model.RegionLevelTownship:
		return "乡镇级"
	default:
		return "未知"
	}
}

func regionStatusText(status uint8) string {
	if status == model.RegionStatusActive {
		return "有效"
	}
	return "已撤销"
}

func normalizeRegionCode(code string) string {
	code = strings.TrimSpace(code)
	if len(code) > 12 {
		return code[:12]
	}
	return code + strings.Repeat("0", 12-len(code))
}
