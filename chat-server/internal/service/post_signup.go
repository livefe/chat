package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var signupUnlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

var signupStatusText = map[uint8]string{
	model.SignupStatusConfirmed:  "已报名",
	model.SignupStatusWaitlisted: "候补",
	model.SignupStatusCancelled:  "已取消",
}

const signupRemarkMaxLen = 100

func (s *PostService) withSignupLock(ctx context.Context, postID uint, failMsg string, fn func() error) error {
	lockKey := fmt.Sprintf(types.CachePostSignupLockKey, postID)
	token := uuid.NewString()
	ok, err := s.cache.SetNX(ctx, lockKey, token, 5*time.Second).Result()
	if err != nil {
		return errs.Unavailable(failMsg, err)
	}
	if !ok {
		return errs.RateLimited("操作太频繁，请稍后重试")
	}
	defer func() {
		if err := signupUnlockScript.Run(ctx, s.cache, []string{lockKey}, token).Err(); err != nil && !errors.Is(err, redis.Nil) {
			s.log.Warn(ctx, "释放报名锁失败", logger.Uint("post_id", postID), logger.Error(err))
		}
	}()
	return fn()
}

func assignCheckInToken(item *model.PostSignup) {
	if item == nil {
		return
	}
	if item.Status != model.SignupStatusConfirmed {
		item.CheckInToken = ""
		item.CheckedInAt = 0
		return
	}
	if item.CheckInToken == "" {
		item.CheckInToken = utils.RandomToken(18)
	}
}

func clearCheckInCredential(item *model.PostSignup) {
	item.CheckInToken = ""
	item.CheckedInAt = 0
}

func buildCheckInPayload(postID uint, token string) string {
	return fmt.Sprintf("%s%d:%s", types.CheckInPayloadPrefix, postID, token)
}

func parseCheckInPayload(raw string) (postID uint, token string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, "", false
	}
	if !strings.HasPrefix(raw, types.CheckInPayloadPrefix) {
		return 0, "", false
	}
	rest := strings.TrimPrefix(raw, types.CheckInPayloadPrefix)
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 {
		return 0, "", false
	}
	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil || id == 0 {
		return 0, "", false
	}
	token = strings.TrimSpace(parts[1])
	if token == "" {
		return 0, "", false
	}
	return uint(id), token, true
}

func (s *PostService) Signup(ctx context.Context, userID uint, postID uint) (*types.PostSignupResult, error) {
	var result *types.PostSignupResult
	err := s.withSignupLock(ctx, postID, "报名失败", func() error {
		p, err := s.postRepo.GetByID(ctx, postID)
		if err != nil {
			return errs.NotFound("动态不存在")
		}
		if p.Status != model.PostStatusPublished {
			return errs.Forbidden("动态不可报名")
		}
		extras, err := s.extrasRepo.GetByPostID(ctx, postID)
		if err != nil || !extras.NeedSignup {
			return errs.Invalid("该动态无需报名")
		}
		now := time.Now().Unix()
		if isSignupDeadlinePassed(extras.SignupDeadlineAt, now) {
			return errs.Invalid("报名已截止")
		}
		if p.EndAt > 0 && now > p.EndAt {
			return errs.Invalid("已结束")
		}

		existed, err := s.signupRepo.GetByPostAndUser(ctx, postID, userID)
		if err == nil {
			if existed.Status != model.SignupStatusCancelled {
				result = &types.PostSignupResult{Status: existed.Status}
				return nil
			}
			targetStatus, resolveErr := s.resolveSignupStatus(ctx, postID, extras)
			if resolveErr != nil {
				return resolveErr
			}
			existed.Status = targetStatus
			existed.CancelledAt = 0
			clearCheckInCredential(existed)
			assignCheckInToken(existed)
			if err := s.signupRepo.Update(ctx, existed); err != nil {
				return fmt.Errorf("报名失败: %w", err)
			}
			result = &types.PostSignupResult{Status: targetStatus}
			return nil
		}

		targetStatus, err := s.resolveSignupStatus(ctx, postID, extras)
		if err != nil {
			return err
		}
		item := &model.PostSignup{
			PostID: postID,
			UserID: userID,
			Status: targetStatus,
		}
		assignCheckInToken(item)
		if err := s.signupRepo.Create(ctx, item); err != nil {
			return fmt.Errorf("报名失败: %w", err)
		}
		result = &types.PostSignupResult{Status: targetStatus}
		return nil
	})
	return result, err
}

func (s *PostService) resolveSignupStatus(ctx context.Context, postID uint, extras *model.PostExtra) (uint8, error) {
	targetStatus := uint8(model.SignupStatusConfirmed)
	if extras.MaxParticipants > 0 {
		confirmed, err := s.signupRepo.CountByPostAndStatus(ctx, postID, model.SignupStatusConfirmed)
		if err != nil {
			return 0, fmt.Errorf("报名失败: %w", err)
		}
		if uint(confirmed) >= extras.MaxParticipants {
			if extras.AllowWaitlist {
				targetStatus = model.SignupStatusWaitlisted
			} else {
				return 0, errs.Conflict("名额已满")
			}
		}
	}
	return targetStatus, nil
}

func (s *PostService) CancelSignup(ctx context.Context, userID uint, postID uint) error {
	return s.withSignupLock(ctx, postID, "取消失败", func() error {
		p, err := s.postRepo.GetByID(ctx, postID)
		if err != nil {
			return errs.NotFound("动态不存在")
		}
		if p.Status != model.PostStatusPublished {
			return errs.Forbidden("动态不可取消报名")
		}
		item, err := s.signupRepo.GetByPostAndUser(ctx, postID, userID)
		if err != nil {
			return errs.NotFound("未找到报名记录")
		}
		if item.Status == model.SignupStatusCancelled {
			return errs.Conflict("报名已取消")
		}
		prev := item.Status
		item.Status = model.SignupStatusCancelled
		item.CancelledAt = time.Now().Unix()
		clearCheckInCredential(item)
		if err := s.signupRepo.Update(ctx, item); err != nil {
			return fmt.Errorf("取消失败: %w", err)
		}
		if prev == model.SignupStatusConfirmed {
			s.promoteWaitlistedIfNeeded(ctx, postID)
		}
		return nil
	})
}

func (s *PostService) promoteWaitlistedIfNeeded(ctx context.Context, postID uint) {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil || p.Status != model.PostStatusPublished {
		return
	}
	now := time.Now().Unix()
	if p.EndAt > 0 && now > p.EndAt {
		return
	}
	extras, err := s.extrasRepo.GetByPostID(ctx, postID)
	if err != nil || !extras.AllowWaitlist {
		return
	}
	if isSignupDeadlinePassed(extras.SignupDeadlineAt, now) {
		return
	}
	waiting, err := s.signupRepo.GetFirstWaitlisted(ctx, postID)
	if err != nil {
		return
	}
	waiting.Status = model.SignupStatusConfirmed
	clearCheckInCredential(waiting)
	assignCheckInToken(waiting)
	if updErr := s.signupRepo.Update(ctx, waiting); updErr != nil {
		s.log.Error(ctx, "候补晋升失败", logger.Uint("post_id", postID), logger.Error(updErr))
	}
}

func isSignupDeadlinePassed(deadline int64, now int64) bool {
	if deadline <= 0 {
		return false
	}
	return now > deadline+59
}

func extrasAllowCheckIn(extras *model.PostExtra) bool {
	return extras != nil && extras.NeedSignup && extras.NeedCheckIn
}

func checkInWindowError(p *model.Post, now int64) error {
	if p.StartAt > 0 && now < p.StartAt {
		return errs.Invalid("活动未开始")
	}
	if p.EndAt > 0 && now > p.EndAt {
		return errs.Invalid("活动已结束")
	}
	return nil
}

func formatSignupUnix(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return utils.FormatDateTime(time.Unix(ts, 0))
}

func (s *PostService) GetCheckInCode(ctx context.Context, userID uint, postID uint) (*types.PostSignupCheckInCodeResponse, error) {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, errs.NotFound("动态不存在")
	}
	extras, err := s.extrasRepo.GetByPostID(ctx, postID)
	if err != nil || !extrasAllowCheckIn(extras) {
		return nil, errs.Invalid("该活动无需核验")
	}
	item, err := s.signupRepo.GetByPostAndUser(ctx, postID, userID)
	if err != nil {
		return nil, errs.NotFound("未找到报名记录")
	}
	if item.Status == model.SignupStatusCancelled {
		return nil, errs.Invalid("报名已取消")
	}
	if item.Status != model.SignupStatusConfirmed {
		return nil, errs.Invalid("候补不能出示核验码")
	}
	alreadyChecked := item.CheckedInAt > 0
	if !alreadyChecked {
		if err := checkInWindowError(p, time.Now().Unix()); err != nil {
			return nil, err
		}
	}
	if item.CheckInToken == "" {
		return nil, errs.Invalid("核验码不可用")
	}
	profiles, err := s.userSvc.GetProfiles(ctx, []uint{userID})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	nickname := ""
	if user, ok := profiles[userID]; ok {
		nickname = utils.DisplayName(user.Nickname)
	}
	return &types.PostSignupCheckInCodeResponse{
		PostID:          postID,
		PostTitle:       p.Title,
		PlaceName:       p.PlaceName,
		StartAt:         p.StartAt,
		EndAt:           p.EndAt,
		Payload:         buildCheckInPayload(postID, item.CheckInToken),
		CheckedIn:       item.CheckedInAt > 0,
		CheckedInAt:     item.CheckedInAt,
		CheckedInAtText: formatSignupUnix(item.CheckedInAt),
		UserNickname:    nickname,
		SignupAt:        utils.FormatDateTime(item.CreatedAt),
	}, nil
}

func (s *PostService) CheckInSignup(ctx context.Context, ownerID uint, req *types.PostSignupCheckInRequest) (*types.PostSignupCheckInResult, error) {
	target, err := resolveCheckInTarget(req)
	if err != nil {
		return nil, err
	}

	var result *types.PostSignupCheckInResult
	err = s.withSignupLock(ctx, target.postID, "核验失败", func() error {
		p, err := s.checkInPost(ctx, ownerID, target.postID)
		if err != nil {
			return err
		}
		item, err := s.findCheckInSignup(ctx, target)
		if err != nil {
			return err
		}

		already := item.CheckedInAt > 0
		if !already {
			item.CheckedInAt = time.Now().Unix()
			if err := s.signupRepo.Update(ctx, item); err != nil {
				return fmt.Errorf("核验失败: %w", err)
			}
		}

		result, err = s.checkInResult(ctx, ownerID, p, item, already)
		return err
	})
	return result, err
}

type checkInTarget struct {
	postID uint
	token  string
	userID uint
}

func resolveCheckInTarget(req *types.PostSignupCheckInRequest) (checkInTarget, error) {
	if req == nil {
		return checkInTarget{}, errs.Invalid("参数无效")
	}
	target := checkInTarget{
		postID: req.PostID,
		userID: req.UserID,
	}
	if payload := strings.TrimSpace(req.Payload); payload != "" {
		pid, t, ok := parseCheckInPayload(payload)
		if !ok {
			return checkInTarget{}, errs.Invalid("无效的核验码")
		}
		target = checkInTarget{postID: pid, token: t}
	}
	if target.postID == 0 {
		return checkInTarget{}, errs.Invalid("请指定活动")
	}
	if target.token == "" && target.userID == 0 {
		return checkInTarget{}, errs.Invalid("请提供核验码或选择用户")
	}
	return target, nil
}

func (s *PostService) checkInPost(ctx context.Context, ownerID, postID uint) (*model.Post, error) {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, errs.NotFound("动态不存在")
	}
	if p.UserID != ownerID {
		return nil, errs.Forbidden("无权核验")
	}
	if !postPublished(p) {
		return nil, errs.Forbidden("动态不可核验")
	}
	extras, err := s.extrasRepo.GetByPostID(ctx, postID)
	if err != nil || !extrasAllowCheckIn(extras) {
		return nil, errs.Invalid("该活动无需核验")
	}
	if err := checkInWindowError(p, time.Now().Unix()); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PostService) findCheckInSignup(ctx context.Context, target checkInTarget) (*model.PostSignup, error) {
	var item *model.PostSignup
	var err error
	if target.token != "" {
		item, err = s.signupRepo.GetByCheckInToken(ctx, target.token)
		if err != nil {
			return nil, errs.NotFound("核验码无效")
		}
		if item.PostID != target.postID {
			return nil, errs.Invalid("核验码与当前活动不符")
		}
	} else {
		item, err = s.signupRepo.GetByPostAndUser(ctx, target.postID, target.userID)
		if err != nil {
			return nil, errs.NotFound("未找到报名记录")
		}
	}
	if item.Status != model.SignupStatusConfirmed {
		return nil, errs.Invalid("只能为已报名用户核验")
	}
	return item, nil
}

func (s *PostService) checkInResult(ctx context.Context, ownerID uint, p *model.Post, item *model.PostSignup, already bool) (*types.PostSignupCheckInResult, error) {
	profiles, err := s.userSvc.GetProfiles(ctx, []uint{item.UserID, ownerID})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	nickname, avatar := checkInProfile(profiles, item.UserID)
	operatorNickname, operatorAvatar := checkInProfile(profiles, ownerID)
	return &types.PostSignupCheckInResult{
		PostID:            p.ID,
		PostTitle:         p.Title,
		UserID:            item.UserID,
		UserNickname:      nickname,
		UserAvatarURL:     avatar,
		SignupAt:          utils.FormatDateTime(item.CreatedAt),
		CheckedInAt:       item.CheckedInAt,
		CheckedInAtText:   formatSignupUnix(item.CheckedInAt),
		AlreadyCheckedIn:  already,
		OperatorID:        ownerID,
		OperatorNickname:  operatorNickname,
		OperatorAvatarURL: operatorAvatar,
	}, nil
}

func checkInProfile(profiles map[uint]types.UserProfile, userID uint) (string, string) {
	if user, ok := profiles[userID]; ok {
		return utils.DisplayName(user.Nickname), user.AvatarURL
	}
	return deactivatedDisplayName(), ""
}

func (s *PostService) UpdateSignupRemark(ctx context.Context, operatorID uint, req *types.UpdateSignupRemarkRequest) error {
	if req == nil || req.PostID == 0 || req.UserID == 0 {
		return errs.Invalid("参数无效")
	}
	remark := strings.TrimSpace(req.Remark)
	if utf8.RuneCountInString(remark) > signupRemarkMaxLen {
		return errs.Invalid(fmt.Sprintf("备注不能超过%d字", signupRemarkMaxLen))
	}
	if err := assertClean(s.sensitive, remark); err != nil {
		return err
	}

	p, err := s.postRepo.GetByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("动态不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	if p.UserID != operatorID {
		return errs.Forbidden("无权修改备注")
	}
	extras, err := s.extrasRepo.GetByPostID(ctx, req.PostID)
	if err != nil || extras == nil || !extras.NeedSignup {
		return errs.Invalid("该动态无需报名")
	}

	item, err := s.signupRepo.GetByPostAndUser(ctx, req.PostID, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("未找到报名记录")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	item.Remark = remark
	if err := s.signupRepo.Update(ctx, item); err != nil {
		return fmt.Errorf("保存备注失败: %w", err)
	}
	return nil
}

func (s *PostService) ListSignups(ctx context.Context, userID uint, req *types.AdminPostSignupListRequest) (*types.AdminPostSignupListResponse, error) {
	p, err := s.postRepo.GetByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("动态不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	if p.UserID != userID {
		return nil, errs.Forbidden("无权查看报名名单")
	}
	return s.buildSignupList(ctx, p, req, false)
}

func (s *PostService) AdminListSignups(ctx context.Context, req *types.AdminPostSignupListRequest) (*types.AdminPostSignupListResponse, error) {
	p, err := s.postRepo.GetByIDUnscoped(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("动态不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.buildSignupList(ctx, p, req, true)
}

func (s *PostService) AdminListUserSignups(ctx context.Context, req *types.AdminUserSignupListRequest) (*types.AdminUserSignupListResponse, error) {
	list, total, err := s.signupRepo.ListByUser(ctx, req.UserID, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	postIDs := make([]uint, 0, len(list))
	for _, item := range list {
		postIDs = append(postIDs, item.PostID)
	}
	posts, err := s.postRepo.GetByIDsUnscoped(ctx, uniqueNonZeroIDs(postIDs))
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	postMap := make(map[uint]*model.Post, len(posts))
	for _, p := range posts {
		postMap[p.ID] = p
	}

	items := make([]types.AdminUserSignupItem, 0, len(list))
	for _, item := range list {
		row := types.AdminUserSignupItem{
			ID:               item.ID,
			PostID:           item.PostID,
			SignupStatus:     item.Status,
			SignupStatusText: utils.TextOf(signupStatusText, item.Status),
			CheckedIn:        item.CheckedInAt > 0,
			CheckedInAt:      formatSignupUnix(item.CheckedInAt),
			CancelledAt:      formatSignupUnix(item.CancelledAt),
			CreatedAt:        utils.FormatDateTime(item.CreatedAt),
		}
		if p, ok := postMap[item.PostID]; ok {
			row.PostTitle = p.Title
			row.PostStatus = p.Status
			row.PostStatusText = utils.TextOf(postStatusText, p.Status)
			if p.DeletedAt.Valid {
				row.PostStatusText = "已删除"
			}
		} else {
			row.PostStatusText = "已删除"
		}
		items = append(items, row)
	}
	return &types.AdminUserSignupListResponse{Total: total, List: items}, nil
}

// buildSignupList 后台视角可按手机号检索并返回脱敏手机号；发布者视角只能按昵称检索、不返回手机号。
func (s *PostService) buildSignupList(ctx context.Context, p *model.Post, req *types.AdminPostSignupListRequest, admin bool) (*types.AdminPostSignupListResponse, error) {
	extras, _ := s.extrasRepo.GetByPostID(ctx, p.ID)
	if extras == nil || !extras.NeedSignup {
		return nil, errs.Invalid("该动态无需报名")
	}

	needCheckIn := extrasAllowCheckIn(extras)
	checkIn := req.CheckIn
	if !needCheckIn {
		checkIn = nil
	}

	list, total, err := s.signupRepo.ListByPost(ctx, repository.SignupListFilter{
		PostID:             req.PostID,
		Status:             req.Status,
		CheckIn:            checkIn,
		UserID:             req.UserID,
		Keyword:            strings.TrimSpace(req.Keyword),
		KeywordMatchMobile: admin,
		SortField:          req.SortField,
		SortDesc:           req.SortOrder == "desc",
		Page:               req.Page,
		PageSize:           req.PageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	counts, err := s.signupRepo.CountListByPost(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	userIDs := make([]uint, 0, len(list))
	for _, item := range list {
		if item.UserID > 0 {
			userIDs = append(userIDs, item.UserID)
		}
	}
	userMap, err := s.userSvc.GetProfiles(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	items := make([]types.AdminPostSignupItem, 0, len(list))
	for _, item := range list {
		statusText := utils.TextOf(signupStatusText, item.Status)
		row := types.AdminPostSignupItem{
			ID:               item.ID,
			UserID:           item.UserID,
			SignupStatus:     item.Status,
			SignupStatusText: statusText,
			CheckedInAt:      formatSignupUnix(item.CheckedInAt),
			CreatedAt:        utils.FormatDateTime(item.CreatedAt),
			Remark:           item.Remark,
		}
		if item.CancelledAt > 0 {
			row.CancelledAt = formatSignupUnix(item.CancelledAt)
		}
		if user, ok := userMap[item.UserID]; ok {
			row.UserNickname = utils.DisplayName(user.Nickname)
			if admin {
				row.UserMobile = utils.DesensitizeMobile(user.Mobile)
			}
		} else {
			row.UserNickname = deactivatedDisplayName()
		}
		items = append(items, row)
	}
	return &types.AdminPostSignupListResponse{
		Total:       total,
		List:        items,
		NeedCheckIn: needCheckIn,
		Counts:      toSignupListCounts(counts),
	}, nil
}

func toSignupListCounts(c repository.SignupListCounts) types.PostSignupListCounts {
	return types.PostSignupListCounts{
		Total:        c.Total,
		Confirmed:    c.Confirmed,
		Waitlisted:   c.Waitlisted,
		Cancelled:    c.Cancelled,
		NotCheckedIn: c.NotCheckedIn,
		CheckedIn:    c.CheckedIn,
	}
}
