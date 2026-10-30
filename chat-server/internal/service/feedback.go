package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

const (
	feedbackContentMinLen = 10
	feedbackContentMaxLen = 500
	feedbackContactMaxLen = 64
	feedbackReplyMaxLen   = 500
	feedbackDailyLimit    = 3
	feedbackMaxImages     = 3
)

var (
	feedbackTypeText = map[uint8]string{
		model.FeedbackTypeProductSuggestion: "产品建议",
		model.FeedbackTypeFeatureBug:        "功能异常",
		model.FeedbackTypeExperienceIssue:   "体验问题",
		model.FeedbackTypeAccountIssue:      "账号问题",
		model.FeedbackTypeOther:             "其他",
	}
	feedbackStatusText = map[uint8]string{
		model.FeedbackStatusPending:  "待处理",
		model.FeedbackStatusResolved: "已解决",
		model.FeedbackStatusClosed:   "已关闭",
	}
	validFeedbackTypes = map[uint8]struct{}{
		model.FeedbackTypeProductSuggestion: {},
		model.FeedbackTypeFeatureBug:        {},
		model.FeedbackTypeExperienceIssue:   {},
		model.FeedbackTypeAccountIssue:      {},
		model.FeedbackTypeOther:             {},
	}
)

type FeedbackService struct {
	log          *logger.Logger
	feedbackRepo repository.FeedbackRepository
	userRepo     repository.UserRepository
	employeeRepo repository.EmployeeRepository
	mediaSvc     *MediaService
}

func NewFeedbackService(
	log *logger.Logger,
	feedbackRepo repository.FeedbackRepository,
	userRepo repository.UserRepository,
	employeeRepo repository.EmployeeRepository,
	mediaSvc *MediaService,
) *FeedbackService {
	return &FeedbackService{
		log:          log,
		feedbackRepo: feedbackRepo,
		userRepo:     userRepo,
		employeeRepo: employeeRepo,
		mediaSvc:     mediaSvc,
	}
}

func (s *FeedbackService) CountByStatus(ctx context.Context, status uint8) (int64, error) {
	return s.feedbackRepo.CountByStatus(ctx, status)
}

func (s *FeedbackService) Submit(ctx context.Context, userID uint, req *types.SubmitFeedbackRequest) error {
	if _, ok := validFeedbackTypes[req.Type]; !ok {
		return errs.Invalid("反馈类型无效")
	}

	content := strings.TrimSpace(req.Content)
	n := utf8.RuneCountInString(content)
	if n < feedbackContentMinLen {
		return errs.Invalid("反馈内容至少 10 个字符")
	}
	if n > feedbackContentMaxLen {
		return errs.Invalid("反馈内容不能超过 500 个字符")
	}

	contact := strings.TrimSpace(req.Contact)
	if utf8.RuneCountInString(contact) > feedbackContactMaxLen {
		return errs.Invalid("联系方式过长")
	}

	images, err := normalizeMediaIDs(req.Images, feedbackMaxImages)
	if err != nil {
		return err
	}

	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)
	count, err := s.feedbackRepo.CountByUserBetween(ctx, userID, dayStart, dayEnd)
	if err != nil {
		s.log.Error(ctx, "统计用户反馈次数失败", logger.Error(err))
		return fmt.Errorf("提交失败，请稍后重试: %w", err)
	}
	if count >= feedbackDailyLimit {
		return errs.Conflict("今日反馈次数已达上限，请明日再试")
	}

	imagesJSON, err := encodeMediaIDs(images)
	if err != nil {
		return errs.Invalid("图片参数错误")
	}
	fb := &model.Feedback{
		UserID:  userID,
		Type:    req.Type,
		Content: content,
		Contact: contact,
		Images:  imagesJSON,
		Status:  model.FeedbackStatusPending,
	}
	if err := s.feedbackRepo.Create(ctx, fb); err != nil {
		s.log.Error(ctx, "创建反馈失败", logger.Error(err))
		return fmt.Errorf("提交失败，请稍后重试: %w", err)
	}
	if len(images) > 0 && s.mediaSvc != nil {
		if err := s.mediaSvc.BindMany(ctx, UserUploader(userID), images, model.MediaKindImage, model.MediaBizFeedback, fb.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *FeedbackService) ListMine(ctx context.Context, userID uint, req *types.MyFeedbackListRequest) (*types.FeedbackListResponse, error) {
	list, total, err := s.feedbackRepo.List(ctx, repository.FeedbackListFilter{
		UserID: userID, Page: req.Page, PageSize: req.PageSize,
	})
	if err != nil {
		s.log.Error(ctx, "查询我的反馈失败", logger.Error(err))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items := make([]types.FeedbackItem, 0, len(list))
	for _, fb := range list {
		items = append(items, s.toFeedbackItem(ctx, fb, nil, nil))
	}
	return &types.FeedbackListResponse{Total: total, List: items}, nil
}

func (s *FeedbackService) GetMine(ctx context.Context, userID, id uint) (*types.FeedbackItem, error) {
	fb, err := s.feedbackRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("反馈不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	if fb.UserID != userID {
		return nil, errs.NotFound("反馈不存在")
	}
	item := s.toFeedbackItem(ctx, fb, nil, nil)
	return &item, nil
}

func (s *FeedbackService) AdminList(ctx context.Context, req *types.AdminFeedbackListRequest) (*types.FeedbackListResponse, error) {
	list, total, err := s.feedbackRepo.List(ctx, repository.FeedbackListFilter{
		Status: req.Status, Type: req.Type, UserID: req.UserID, Keyword: req.Keyword,
		StartTime: req.StartTime, EndTime: req.EndTime,
		Page: req.Page, PageSize: req.PageSize,
	})
	if err != nil {
		s.log.Error(ctx, "查询反馈列表失败", logger.Error(err))
		return nil, fmt.Errorf("查询反馈列表失败: %w", err)
	}

	userIDs := make([]uint, 0, len(list))
	employeeIDs := make([]uint, 0, len(list))
	for _, fb := range list {
		userIDs = append(userIDs, fb.UserID)
		employeeIDs = append(employeeIDs, fb.EmployeeID)
	}
	userMap := make(map[uint]*model.User)
	if ids := uniqueUints(userIDs); len(ids) > 0 {
		users, err := s.userRepo.GetByIDs(ctx, ids)
		if err != nil {
			return nil, fmt.Errorf("查询反馈列表失败: %w", err)
		}
		for _, u := range users {
			userMap[u.ID] = u
		}
	}
	employeeMap := make(map[uint]*model.Employee)
	if ids := uniqueUints(employeeIDs); len(ids) > 0 {
		employees, err := s.employeeRepo.GetByIDs(ctx, ids)
		if err != nil {
			return nil, fmt.Errorf("查询反馈列表失败: %w", err)
		}
		for _, e := range employees {
			employeeMap[e.ID] = e
		}
	}

	items := make([]types.FeedbackItem, 0, len(list))
	for _, fb := range list {
		items = append(items, s.toFeedbackItem(ctx, fb, userMap[fb.UserID], employeeMap[fb.EmployeeID]))
	}
	return &types.FeedbackListResponse{Total: total, List: items}, nil
}

func (s *FeedbackService) AdminGet(ctx context.Context, id uint) (*types.FeedbackItem, error) {
	fb, err := s.feedbackRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("反馈不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	var user *model.User
	if u, err := s.userRepo.GetByID(ctx, fb.UserID); err == nil {
		user = u
	}
	var employee *model.Employee
	if fb.EmployeeID > 0 {
		if e, err := s.employeeRepo.GetByID(ctx, fb.EmployeeID); err == nil {
			employee = e
		}
	}
	item := s.toFeedbackItem(ctx, fb, user, employee)
	return &item, nil
}

func (s *FeedbackService) AdminUpdateStatus(ctx context.Context, employeeID uint, req *types.AdminFeedbackUpdateStatusRequest) error {
	fb, err := s.feedbackRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("反馈不存在")
		}
		return fmt.Errorf("更新失败: %w", err)
	}
	if !canTransitionFeedbackStatus(fb.Status, req.Status) {
		return errs.Conflict("不允许的状态变更")
	}

	handleResult := strings.TrimSpace(req.HandleResult)
	if utf8.RuneCountInString(handleResult) > feedbackReplyMaxLen {
		return errs.Invalid("处理说明过长")
	}

	fb.Status = req.Status
	fb.HandleResult = handleResult
	fb.EmployeeID = employeeID
	fb.HandledAt = time.Now().Unix()

	if err := s.feedbackRepo.Update(ctx, fb); err != nil {
		s.log.Error(ctx, "更新反馈状态失败", logger.Error(err))
		return fmt.Errorf("更新失败: %w", err)
	}
	return nil
}

func (s *FeedbackService) AdminBatchClose(ctx context.Context, employeeID uint, req *types.AdminFeedbackBatchCloseRequest) (*types.AdminFeedbackBatchResult, error) {
	ids := uniqueUints(req.IDs)
	n, err := s.feedbackRepo.UpdatePendingByIDs(ctx, ids, map[string]any{
		"status":        model.FeedbackStatusClosed,
		"handle_result": strings.TrimSpace(req.HandleResult),
		"employee_id":   employeeID,
		"handled_at":    time.Now().Unix(),
	})
	if err != nil {
		s.log.Error(ctx, "批量关闭反馈失败", logger.Error(err))
		return nil, fmt.Errorf("批量关闭失败: %w", err)
	}
	return &types.AdminFeedbackBatchResult{Updated: n, Skipped: int64(len(ids)) - n}, nil
}

func canTransitionFeedbackStatus(from, to uint8) bool {
	if from == to {
		return true
	}
	switch from {
	case model.FeedbackStatusPending:
		return to == model.FeedbackStatusResolved || to == model.FeedbackStatusClosed
	case model.FeedbackStatusResolved, model.FeedbackStatusClosed:
		return false
	default:
		return false
	}
}

func (s *FeedbackService) toFeedbackItem(ctx context.Context, fb *model.Feedback, user *model.User, employee *model.Employee) types.FeedbackItem {
	ids := parseMediaIDs(fb.Images)
	images := []string{}
	if s.mediaSvc != nil {
		images = s.mediaSvc.ResolveURLList(ctx, ids)
	}
	item := types.FeedbackItem{
		ID:           fb.ID,
		UserID:       fb.UserID,
		Type:         fb.Type,
		TypeText:     utils.TextOf(feedbackTypeText, fb.Type),
		Content:      fb.Content,
		Contact:      fb.Contact,
		Images:       images,
		Status:       fb.Status,
		StatusText:   utils.TextOf(feedbackStatusText, fb.Status),
		HandleResult: fb.HandleResult,
		EmployeeID:   fb.EmployeeID,
		CreatedAt:    utils.FormatDateTime(fb.CreatedAt),
		UpdatedAt:    utils.FormatDateTime(fb.UpdatedAt),
	}
	if fb.HandledAt > 0 {
		item.HandledAt = utils.FormatDateTime(time.Unix(fb.HandledAt, 0))
	}
	if user != nil {
		item.UserNickname = utils.DisplayName(user.Nickname)
		item.UserMobile = utils.DesensitizeMobile(user.Mobile)
	}
	if employee != nil {
		item.EmployeeName = employee.Name
	}
	return item
}
