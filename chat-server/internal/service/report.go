package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/datatypes"
)

const (
	reportContentMaxLen  = 500
	reportDailyLimit     = 10
	reportMaxImages      = 3
	reportDuplicateHours = 24
)

var (
	reportTargetTypeText = map[uint8]string{
		model.ReportTargetUser:    "用户",
		model.ReportTargetMessage: "消息",
		model.ReportTargetPost:    "动态",
	}
	reportReasonText = map[uint8]string{
		model.ReportReasonSpam:         "垃圾广告",
		model.ReportReasonPorn:         "色情低俗",
		model.ReportReasonHarassment:   "辱骂骚扰",
		model.ReportReasonIllegal:      "违法违规",
		model.ReportReasonFake:         "虚假信息",
		model.ReportReasonInfringement: "侵权",
		model.ReportReasonOther:        "其他",
	}
	reportStatusText = map[uint8]string{
		model.ReportStatusPending:  "待处理",
		model.ReportStatusResolved: "已处理",
		model.ReportStatusRejected: "已驳回",
	}
	reportHandleActionText = map[uint8]string{
		model.ReportHandleActionNone:          "无",
		model.ReportHandleActionWarn:          "警告",
		model.ReportHandleActionDeleteContent: "删除内容",
		model.ReportHandleActionMute:          "禁言用户",
		model.ReportHandleActionDisableUser:   "禁用账号",
		model.ReportHandleActionDisablePost:   "下架动态",
	}
	validReportReasons = map[uint8]struct{}{
		model.ReportReasonSpam:         {},
		model.ReportReasonPorn:         {},
		model.ReportReasonHarassment:   {},
		model.ReportReasonIllegal:      {},
		model.ReportReasonFake:         {},
		model.ReportReasonInfringement: {},
		model.ReportReasonOther:        {},
	}
	validReportTargetTypes = map[uint8]struct{}{
		model.ReportTargetUser:    {},
		model.ReportTargetMessage: {},
		model.ReportTargetPost:    {},
	}
)

type ReportService struct {
	log          *logger.Logger
	reportRepo   repository.ReportRepository
	employeeRepo repository.EmployeeRepository
	userSvc      *UserService
	messageSvc   *MessageService
	postSvc      *PostService
	mediaSvc     *MediaService
}

func NewReportService(
	log *logger.Logger,
	reportRepo repository.ReportRepository,
	employeeRepo repository.EmployeeRepository,
	userSvc *UserService,
	messageSvc *MessageService,
	postSvc *PostService,
	mediaSvc *MediaService,
) *ReportService {
	return &ReportService{
		log:          log,
		reportRepo:   reportRepo,
		employeeRepo: employeeRepo,
		userSvc:      userSvc,
		messageSvc:   messageSvc,
		postSvc:      postSvc,
		mediaSvc:     mediaSvc,
	}
}

func (s *ReportService) CountByStatus(ctx context.Context, status uint8) (int64, error) {
	return s.reportRepo.CountByStatus(ctx, status)
}

func (s *ReportService) Submit(ctx context.Context, reporterID uint, req *types.SubmitReportRequest) error {
	if _, ok := validReportTargetTypes[req.TargetType]; !ok {
		return errs.Invalid("举报类型无效")
	}
	if _, ok := validReportReasons[req.Reason]; !ok {
		return errs.Invalid("举报原因无效")
	}

	content := strings.TrimSpace(req.Content)
	if n := utf8.RuneCountInString(content); n > reportContentMaxLen {
		return errs.Invalid("补充说明不能超过 500 个字符")
	}

	images, err := normalizeMediaIDs(req.Images, reportMaxImages)
	if err != nil {
		return err
	}

	targetUserID, snapshot, err := s.resolveTarget(ctx, reporterID, req.TargetType, req.TargetID)
	if err != nil {
		return err
	}

	now := time.Now()
	since := now.Add(-reportDuplicateHours * time.Hour)
	dup, err := s.reportRepo.CountByReporterTargetSince(ctx, reporterID, req.TargetType, req.TargetID, since)
	if err != nil {
		s.log.Error(ctx, "统计重复举报失败", logger.Error(err))
		return fmt.Errorf("提交失败，请稍后重试: %w", err)
	}
	if dup > 0 {
		return errs.Conflict("24 小时内已举报过该内容")
	}

	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)
	dayCount, err := s.reportRepo.CountByReporterBetween(ctx, reporterID, dayStart, dayEnd)
	if err != nil {
		s.log.Error(ctx, "统计用户举报次数失败", logger.Error(err))
		return fmt.Errorf("提交失败，请稍后重试: %w", err)
	}
	if dayCount >= reportDailyLimit {
		return errs.Conflict("今日举报次数已达上限，请明日再试")
	}

	ctxJSON, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("提交失败，请稍后重试: %w", err)
	}
	imagesJSON, err := encodeMediaIDs(images)
	if err != nil {
		return errs.Invalid("图片参数错误")
	}

	item := &model.Report{
		ReporterID:   reporterID,
		TargetType:   req.TargetType,
		TargetID:     req.TargetID,
		TargetUserID: targetUserID,
		Reason:       req.Reason,
		Content:      content,
		Images:       imagesJSON,
		Context:      datatypes.JSON(ctxJSON),
		Status:       model.ReportStatusPending,
	}
	if err := s.reportRepo.Create(ctx, item); err != nil {
		s.log.Error(ctx, "创建举报失败", logger.Error(err))
		return fmt.Errorf("提交失败，请稍后重试: %w", err)
	}
	if len(images) > 0 && s.mediaSvc != nil {
		if _, err := s.mediaSvc.BindMany(ctx, reporterID, images, model.MediaKindImage, model.MediaBizReport, item.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *ReportService) resolveTarget(ctx context.Context, reporterID uint, targetType uint8, targetID uint) (uint, types.ReportContextSnapshot, error) {
	switch targetType {
	case model.ReportTargetUser:
		return s.userSvc.ReportTarget(ctx, reporterID, targetID)
	case model.ReportTargetMessage:
		return s.messageSvc.ReportTarget(ctx, reporterID, targetID)
	case model.ReportTargetPost:
		return s.postSvc.ReportTarget(ctx, reporterID, targetID)
	default:
		return 0, types.ReportContextSnapshot{}, errs.Invalid("举报类型无效")
	}
}
