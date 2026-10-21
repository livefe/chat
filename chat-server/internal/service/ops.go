package service

import (
	"context"
	"fmt"
	"sync"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/logger"
)

type OpsService struct {
	log         *logger.Logger
	permSvc     *PermissionService
	imSvc       *IMService
	supportSvc  *SupportService
	reportSvc   *ReportService
	feedbackSvc *FeedbackService
	roomSvc     *RoomService
	postSvc     *PostService
}

func NewOpsService(
	log *logger.Logger,
	permSvc *PermissionService,
	imSvc *IMService,
	supportSvc *SupportService,
	reportSvc *ReportService,
	feedbackSvc *FeedbackService,
	roomSvc *RoomService,
	postSvc *PostService,
) *OpsService {
	return &OpsService{
		log:         log,
		permSvc:     permSvc,
		imSvc:       imSvc,
		supportSvc:  supportSvc,
		reportSvc:   reportSvc,
		feedbackSvc: feedbackSvc,
		roomSvc:     roomSvc,
		postSvc:     postSvc,
	}
}

func (s *OpsService) PendingCounts(ctx context.Context, employeeID uint) (*types.AdminPendingCountsResponse, error) {
	out := &types.AdminPendingCountsResponse{}
	codes, err := s.menuCodeSet(ctx, employeeID)
	if err != nil {
		s.log.Error(ctx, "查询员工菜单权限失败", logger.Error(err), logger.Uint("employeeId", employeeID))
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	run := func(need bool, fn func() (int64, error), set func(int64)) {
		if !need {
			return
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := fn()
			if err != nil {
				s.log.Warn(ctx, "pending count failed", logger.Error(err))
				return
			}
			mu.Lock()
			set(n)
			mu.Unlock()
		}()
	}

	run(codes["im:chat"], func() (int64, error) {
		sum, err := s.imSvc.UnreadSummary(ctx, employeeID)
		if err != nil || sum == nil {
			return 0, err
		}
		return sum.Total, nil
	}, func(n int64) { out.IMUnread = n })

	run(codes["im:support"] || codes["im:support_overview"], func() (int64, error) {
		return s.supportSvc.CountQueue(ctx)
	}, func(n int64) { out.SupportQueue = n })

	run(codes["user:report"], func() (int64, error) {
		return s.reportSvc.CountByStatus(ctx, model.ReportStatusPending)
	}, func(n int64) { out.ReportPending = n })

	run(codes["user:feedback"], func() (int64, error) {
		return s.feedbackSvc.CountByStatus(ctx, model.FeedbackStatusPending)
	}, func(n int64) { out.FeedbackPending = n })

	run(codes["chat:room"], func() (int64, error) {
		return s.roomSvc.CountByStatus(ctx, model.RoomStatusPending)
	}, func(n int64) { out.RoomAuditPending = n })

	run(codes["post:post"], func() (int64, error) {
		return s.postSvc.CountByStatus(ctx, model.PostStatusPending)
	}, func(n int64) { out.PostAuditPending = n })

	wg.Wait()
	return out, nil
}

func (s *OpsService) menuCodeSet(ctx context.Context, employeeID uint) (map[string]bool, error) {
	isSuper, perms, err := s.permSvc.EmployeeAccess(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, 8)
	if isSuper {
		out["im:chat"] = true
		out["im:support"] = true
		out["im:support_overview"] = true
		out["user:report"] = true
		out["user:feedback"] = true
		out["chat:room"] = true
		out["post:post"] = true
		return out, nil
	}
	for _, p := range perms {
		if p == nil {
			continue
		}
		switch p.Code {
		case "im:chat", "im:support", "im:support_overview", "user:report", "user:feedback", "chat:room", "post:post":
			out[p.Code] = true
		}
	}
	return out, nil
}
