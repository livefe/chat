package service

import (
	"context"
	"fmt"
	mrand "math/rand"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/utils"
)

func (s *AdService) MergeDiscoverFeed(
	ctx context.Context,
	userID uint,
	posts []types.PostItem,
	total int64,
	page, pageSize int,
	lat, lng float64,
	hasLocation bool,
) *types.DiscoverFeedResponse {
	out := &types.DiscoverFeedResponse{Total: total, List: make([]types.FeedItem, 0, len(posts)+4)}
	appendPostsOnly := func() {
		for i := range posts {
			p := posts[i]
			out.List = append(out.List, types.FeedItem{Type: types.FeedItemTypePost, Post: &p})
		}
	}
	rules := s.configSvc.GetAd()
	if !rules.Enabled || pageSize <= 0 || len(posts) == 0 {
		appendPostsOnly()
		return out
	}

	naturalStart := (page - 1) * pageSize
	slotCount := 0
	for i := 0; i < len(posts); i++ {
		naturalCount := naturalStart + i + 1
		if isAdSlot(naturalCount, rules) {
			slotCount++
		}
	}
	if slotCount == 0 {
		appendPostsOnly()
		return out
	}

	picked, err := s.pickCampaigns(ctx, slotCount, lat, lng, hasLocation)
	if err != nil {
		s.log.Warn(ctx, "广告选单失败，降级纯帖子", logger.Error(err))
		appendPostsOnly()
		return out
	}

	pickIdx := 0
	for i := range posts {
		p := posts[i]
		out.List = append(out.List, types.FeedItem{Type: types.FeedItemTypePost, Post: &p})
		naturalCount := naturalStart + i + 1
		if !isAdSlot(naturalCount, rules) || pickIdx >= len(picked) {
			continue
		}
		c := picked[pickIdx]
		pickIdx++
		reqID := newAdRequestID()
		token := s.signToken(c.ID, userID, reqID, time.Now().Unix()+s.tokenTTLSeconds)
		item := types.SponsoredItem{
			CampaignID:      c.ID,
			Title:           c.Title,
			Subtitle:        c.Subtitle,
			AdvertiserName:  c.AdvertiserName,
			Badge:           "广告",
			LandingType:     c.LandingType,
			LandingValue:    c.LandingValue,
			ImpressionToken: token,
			RequestID:       reqID,
			SlotIndex:       len(out.List),
		}
		if s.mediaSvc != nil && c.ImageMediaID > 0 {
			item.ImageURL = s.mediaSvc.ResolveURL(ctx, c.ImageMediaID)
		}
		if hasLocation && c.HasGeo {
			item.Distance = utils.Haversine(lat, lng, c.Latitude, c.Longitude)
		}
		out.List = append(out.List, types.FeedItem{Type: types.FeedItemTypeSponsored, Sponsored: &item})
	}
	return out
}

func isAdSlot(naturalCount int, rules types.AdminAdSetting) bool {
	if rules.SlotInterval < 1 || naturalCount < rules.FirstSlotAfter {
		return false
	}
	return (naturalCount-rules.FirstSlotAfter)%rules.SlotInterval == 0
}

func (s *AdService) pickCampaigns(ctx context.Context, need int, lat, lng float64, hasLocation bool) ([]*model.AdCampaign, error) {
	list, err := s.campaignRepo.ListActiveCandidates(ctx, time.Now().Unix())
	if err != nil {
		return nil, fmt.Errorf("调用 ListActiveCandidates 失败: %w", err)
	}
	today := repository.TodayDate()
	ids := make([]uint, 0, len(list))
	for _, c := range list {
		ids = append(ids, c.ID)
	}
	statMap, err := s.statRepo.MapTodayByCampaignIDs(ctx, ids, today)
	if err != nil {
		return nil, fmt.Errorf("调用 MapTodayByCampaignIDs 失败: %w", err)
	}
	pool := s.buildCampaignPool(ctx, list, statMap, lat, lng, hasLocation)
	if len(pool) == 0 {
		return nil, nil
	}
	return selectCampaigns(pool, need), nil
}

type adCandidate struct {
	c      *model.AdCampaign
	weight float64
}

func (s *AdService) buildCampaignPool(ctx context.Context, list []*model.AdCampaign, statMap map[uint]*model.AdDailyStat, lat, lng float64, hasLocation bool) []adCandidate {
	var pool []adCandidate
	for _, c := range list {
		if c.HasGeo {
			if !hasLocation || !utils.InCircle(lat, lng, c.Latitude, c.Longitude, c.RadiusM) {
				continue
			}
		}
		impr := int64(0)
		if st := statMap[c.ID]; st != nil {
			impr = st.Impressions
			if c.DailyImpressionLimit > 0 && uint(st.Impressions) >= c.DailyImpressionLimit {
				continue
			}
		}
		if !s.landingLooksOK(ctx, c) {
			if err := s.campaignRepo.UpdateStatus(ctx, c.ID, model.AdStatusPaused, "落地失效自动暂停", true); err != nil {
				s.log.Warn(ctx, "自动暂停广告失败", logger.Uint("campaign_id", c.ID), logger.Error(err))
			}
			continue
		}

		prio := float64(c.Priority)
		if prio < 1 {
			prio = 1
		}
		w := prio * (1.0 + 1.0/float64(impr+1)) * (0.5 + mrand.Float64())
		pool = append(pool, adCandidate{c: c, weight: w})
	}
	return pool
}

func selectCampaigns(pool []adCandidate, need int) []*model.AdCampaign {
	picked := make([]*model.AdCampaign, 0, need)
	usedAdv := map[string]int{}
	for len(picked) < need && len(pool) > 0 {
		best := bestCandidate(pool, usedAdv, true)
		if best < 0 {
			best = bestCandidate(pool, usedAdv, false)
		}
		if best < 0 {
			break
		}
		c := pool[best].c
		picked = append(picked, c)
		usedAdv[strings.ToLower(c.AdvertiserName)]++
		pool = append(pool[:best], pool[best+1:]...)
	}
	return picked
}

func bestCandidate(pool []adCandidate, usedAdv map[string]int, distinctAdvertiser bool) int {
	best := -1
	bestW := -1.0
	for i, it := range pool {
		if distinctAdvertiser && usedAdv[strings.ToLower(it.c.AdvertiserName)] >= 1 {
			continue
		}
		if it.weight > bestW {
			bestW = it.weight
			best = i
		}
	}
	return best
}

func (s *AdService) ReportEvents(ctx context.Context, userID uint, req *types.AdEventsRequest) error {
	for _, ev := range req.Events {
		if ev.EventType != model.AdEventImpression && ev.EventType != model.AdEventClick {
			continue
		}
		if !s.verifyToken(ev.CampaignID, userID, ev.RequestID, ev.ImpressionToken) {
			continue
		}
		item := &model.AdEvent{
			RequestID:  ev.RequestID,
			EventType:  ev.EventType,
			CampaignID: ev.CampaignID,
			UserID:     userID,
			SlotIndex:  ev.SlotIndex,
			Latitude:   ev.Latitude,
			Longitude:  ev.Longitude,
			Billable:   true,
		}
		inserted, err := s.eventRepo.CreateIgnoreDuplicate(ctx, item)
		if err != nil {
			s.log.Warn(ctx, "写入广告事件失败", logger.Error(err))
			continue
		}
		if !inserted {
			continue
		}
		var impr, click int64
		if ev.EventType == model.AdEventImpression {
			impr = 1
		} else {
			click = 1
		}
		if err := s.statRepo.Incr(ctx, ev.CampaignID, repository.TodayDate(), impr, click); err != nil {
			s.log.Warn(ctx, "更新广告日统计失败", logger.Error(err))
		}
	}
	return nil
}
