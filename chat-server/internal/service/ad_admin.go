package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

var adStatusText = map[uint8]string{
	model.AdStatusDraft:    "草稿",
	model.AdStatusPending:  "待审核",
	model.AdStatusActive:   "投放中",
	model.AdStatusRejected: "已驳回",
	model.AdStatusPaused:   "已暂停",
	model.AdStatusEnded:    "已结束",
}

func (s *AdService) AdminList(ctx context.Context, req *types.AdminAdCampaignListRequest) (*types.AdminAdCampaignListResponse, error) {
	list, total, err := s.campaignRepo.List(ctx, req.Status, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	ids := make([]uint, len(list))
	for i, c := range list {
		ids[i] = c.ID
	}
	statMap, _ := s.statRepo.MapTodayByCampaignIDs(ctx, ids, repository.TodayDate())
	items := make([]types.AdminAdCampaignItem, 0, len(list))
	for _, c := range list {
		it := s.toAdminItem(ctx, c)
		if st := statMap[c.ID]; st != nil {
			it.TodayImpressions = st.Impressions
			it.TodayClicks = st.Clicks
			it.TodayCTR = repository.AdStatCTR(st.Impressions, st.Clicks)
			it.LowCTRWarning = st.Impressions >= 100 && it.TodayCTR < s.configSvc.GetAd().LowCTRThreshold
		}
		items = append(items, it)
	}
	return &types.AdminAdCampaignListResponse{Total: total, List: items}, nil
}

func (s *AdService) AdminGet(ctx context.Context, id uint) (*types.AdminAdCampaignItem, error) {
	c, err := s.campaignRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("计划不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	it := s.toAdminItem(ctx, c)
	if st, err := s.statRepo.Get(ctx, id, repository.TodayDate()); err == nil && st != nil {
		it.TodayImpressions = st.Impressions
		it.TodayClicks = st.Clicks
		it.TodayCTR = repository.AdStatCTR(st.Impressions, st.Clicks)
		it.LowCTRWarning = st.Impressions >= 100 && it.TodayCTR < s.configSvc.GetAd().LowCTRThreshold
	}
	return &it, nil
}

func (s *AdService) toAdminItem(ctx context.Context, c *model.AdCampaign) types.AdminAdCampaignItem {
	it := types.AdminAdCampaignItem{
		ID:                   c.ID,
		Name:                 c.Name,
		AdvertiserName:       c.AdvertiserName,
		Title:                c.Title,
		Subtitle:             c.Subtitle,
		ImageMediaID:         c.ImageMediaID,
		LandingType:          c.LandingType,
		LandingValue:         c.LandingValue,
		Status:               c.Status,
		StatusText:           utils.TextOf(adStatusText, c.Status),
		Priority:             c.Priority,
		StartAt:              c.StartAt,
		EndAt:                c.EndAt,
		CPDPrice:             c.CPDPrice,
		HasGeo:               c.HasGeo,
		Latitude:             c.Latitude,
		Longitude:            c.Longitude,
		RadiusM:              c.RadiusM,
		DailyImpressionLimit: c.DailyImpressionLimit,
		RejectRemark:         c.RejectRemark,
		LandingInvalid:       c.LandingInvalid,
		CreatedAt:            c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:            c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if s.mediaSvc != nil && c.ImageMediaID > 0 {
		it.ImageURL = s.mediaSvc.ResolveURL(ctx, c.ImageMediaID)
	}
	return it
}

func (s *AdService) AdminSave(ctx context.Context, operatorID uint, req *types.AdminAdCampaignSaveRequest) (uint, error) {
	if req.EndAt <= req.StartAt {
		return 0, errs.Invalid("结束时间必须大于开始时间")
	}
	if req.HasGeo && (req.RadiusM <= 0 || math.Abs(req.Latitude) < 1e-6 && math.Abs(req.Longitude) < 1e-6) {
		return 0, errs.Invalid("地理定向需填写有效圆心与半径")
	}
	tmp := &model.AdCampaign{
		LandingType:  req.LandingType,
		LandingValue: req.LandingValue,
	}
	if !s.landingLooksOK(ctx, tmp) {
		return 0, errs.Invalid("落地无效，请检查链接/帖子/聊天区")
	}

	var item *model.AdCampaign
	var err error
	if req.ID > 0 {
		item, err = s.campaignRepo.GetByID(ctx, req.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, errs.NotFound("计划不存在")
			}
			return 0, fmt.Errorf("查询失败: %w", err)
		}
		if item.Status == model.AdStatusActive {
			return 0, errs.Invalid("投放中请先暂停再编辑")
		}
	} else {
		item = &model.AdCampaign{}
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = strings.TrimSpace(req.Title)
	}
	item.Name = name
	item.AdvertiserName = strings.TrimSpace(req.AdvertiserName)
	item.Title = strings.TrimSpace(req.Title)
	item.Subtitle = strings.TrimSpace(req.Subtitle)
	item.ImageMediaID = req.ImageMediaID
	item.LandingType = req.LandingType
	item.LandingValue = strings.TrimSpace(req.LandingValue)
	item.Priority = req.Priority
	item.StartAt = req.StartAt
	item.EndAt = req.EndAt
	item.CPDPrice = req.CPDPrice
	item.HasGeo = req.HasGeo
	item.Latitude = req.Latitude
	item.Longitude = req.Longitude
	item.RadiusM = req.RadiusM
	item.DailyImpressionLimit = req.DailyImpressionLimit
	item.LandingInvalid = false
	item.RejectRemark = ""

	if req.SubmitAudit {
		item.Status = model.AdStatusPending
	} else if item.ID == 0 {
		item.Status = model.AdStatusDraft
	}

	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if item.ID == 0 {
			if err := s.campaignRepo.Create(ctx, item); err != nil {
				return err
			}
		} else if err := s.campaignRepo.Update(ctx, item); err != nil {
			return err
		}
		if s.mediaSvc != nil && item.ImageMediaID > 0 {
			if _, err := s.mediaSvc.ReplaceBizMedia(ctx, operatorID, []uint{item.ImageMediaID}, model.MediaKindImage, model.MediaBizAdCampaign, item.ID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		var svcErr *errs.Error
		if errors.As(err, &svcErr) {
			return 0, svcErr
		}
		s.log.Error(ctx, "保存广告失败", logger.Error(err))
		return 0, fmt.Errorf("保存失败: %w", err)
	}
	return item.ID, nil
}

func (s *AdService) AdminUpdateStatus(ctx context.Context, req *types.AdminAdUpdateStatusRequest) error {
	c, err := s.campaignRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("计划不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	now := time.Now().Unix()
	switch req.Status {
	case model.AdStatusActive:
		if c.Status != model.AdStatusPending && c.Status != model.AdStatusPaused {
			return errs.Invalid("仅待审核或已暂停可上架")
		}
		if c.EndAt < now {
			return errs.Invalid("已过期，请先修改排期")
		}
		if !s.landingLooksOK(ctx, c) {
			return errs.Invalid("落地无效，无法上架")
		}
		n, err := s.campaignRepo.CountActiveByAdvertiser(ctx, c.AdvertiserName, c.ID)
		if err != nil {
			return fmt.Errorf("校验失败: %w", err)
		}
		limit := s.configSvc.GetAd().MaxActivePerAdvertiser
		if limit < 1 {
			return errs.Invalid("投放设置未配置")
		}
		if int(n) >= limit {
			return errs.Invalid(fmt.Sprintf("同一商家同时投放中最多 %d 条", limit))
		}
		return s.updateCampaignStatus(ctx, c.ID, model.AdStatusActive, "", false)
	case model.AdStatusRejected:
		if c.Status != model.AdStatusPending {
			return errs.Invalid("仅待审核可驳回")
		}
		return s.updateCampaignStatus(ctx, c.ID, model.AdStatusRejected, strings.TrimSpace(req.RejectRemark), false)
	case model.AdStatusPaused:
		if c.Status != model.AdStatusActive {
			return errs.Invalid("仅投放中可暂停")
		}
		return s.updateCampaignStatus(ctx, c.ID, model.AdStatusPaused, "", c.LandingInvalid)
	case model.AdStatusPending:
		if c.Status != model.AdStatusDraft && c.Status != model.AdStatusRejected && c.Status != model.AdStatusPaused {
			return errs.Invalid("当前状态不可提交审核")
		}
		return s.updateCampaignStatus(ctx, c.ID, model.AdStatusPending, "", false)
	case model.AdStatusEnded:
		return s.updateCampaignStatus(ctx, c.ID, model.AdStatusEnded, "", c.LandingInvalid)
	case model.AdStatusDraft:
		return s.updateCampaignStatus(ctx, c.ID, model.AdStatusDraft, "", false)
	default:
		return errs.Invalid("不支持的状态")
	}
}

func (s *AdService) updateCampaignStatus(ctx context.Context, id uint, status uint8, remark string, landingInvalid bool) error {
	if err := s.campaignRepo.UpdateStatus(ctx, id, status, remark, landingInvalid); err != nil {
		s.log.Error(ctx, "更新广告状态失败", logger.Error(err), logger.Uint("id", id))
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

func (s *AdService) AdminBatchPause(ctx context.Context, ids []uint) error {
	if err := s.campaignRepo.BatchUpdateStatus(ctx, ids, model.AdStatusPaused); err != nil {
		s.log.Error(ctx, "批量暂停广告失败", logger.Error(err))
		return fmt.Errorf("操作失败: %w", err)
	}
	return nil
}

func (s *AdService) AdminCopy(ctx context.Context, operatorID, id uint) (uint, error) {
	c, err := s.campaignRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errs.NotFound("计划不存在")
		}
		return 0, fmt.Errorf("查询失败: %w", err)
	}
	cp := *c
	cp.ID = 0
	cp.CreatedAt = time.Time{}
	cp.UpdatedAt = time.Time{}
	cp.DeletedAt = gorm.DeletedAt{}
	cp.Name = c.Name + " (复制)"
	cp.Status = model.AdStatusDraft
	cp.RejectRemark = ""
	cp.LandingInvalid = false
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if err := s.campaignRepo.Create(ctx, &cp); err != nil {
			return err
		}
		if s.mediaSvc != nil && cp.ImageMediaID > 0 {
			if _, err := s.mediaSvc.ReplaceBizMedia(ctx, operatorID, []uint{cp.ImageMediaID}, model.MediaKindImage, model.MediaBizAdCampaign, cp.ID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		var svcErr *errs.Error
		if errors.As(err, &svcErr) {
			return 0, svcErr
		}
		s.log.Error(ctx, "复制广告失败", logger.Error(err))
		return 0, fmt.Errorf("复制失败: %w", err)
	}
	return cp.ID, nil
}

func (s *AdService) AdminListDailyStats(ctx context.Context, req *types.AdminAdDailyStatRequest) (*types.AdminAdDailyStatResponse, error) {
	if _, err := s.campaignRepo.GetByID(ctx, req.CampaignID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("计划不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	list, err := s.statRepo.ListByCampaign(ctx, req.CampaignID, req.FromDate, req.ToDate)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items := make([]types.AdminAdDailyStatItem, 0, len(list))
	for _, st := range list {
		items = append(items, types.AdminAdDailyStatItem{
			StatDate:    st.StatDate,
			Impressions: st.Impressions,
			Clicks:      st.Clicks,
			CTR:         repository.AdStatCTR(st.Impressions, st.Clicks),
		})
	}
	return &types.AdminAdDailyStatResponse{List: items}, nil
}

func (s *AdService) AdminGetSetting(ctx context.Context) types.AdminAdSetting {
	return s.configSvc.GetAd()
}

func (s *AdService) AdminUpdateSetting(ctx context.Context, req types.AdminAdSetting) error {
	return s.configSvc.SaveAd(ctx, req)
}
