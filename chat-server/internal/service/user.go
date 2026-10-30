package service

import (
	"context"
	"errors"
	"fmt"
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

var userStatusText = map[uint8]string{
	model.UserStatusDisabled: "禁用",
	model.UserStatusEnabled:  "启用",
}

type userSessionNotifier interface {
	DisconnectUser(userID uint)
	RemoveOnlineGeo(ctx context.Context, userID uint)
}

type userNoticeSender interface {
	NotifyUser(ctx context.Context, userID uint, text string) error
}

type UserService struct {
	log              *logger.Logger
	db               *database.DB
	authSvc          *AuthService
	smsSvc           *SMSService
	userRepo         repository.UserRepository
	loginRepo        repository.UserLoginRecordRepository
	convDelRepo      repository.ConversationDeletionRepository
	blockRepo        repository.UserBlockRepository
	deactivationRepo repository.UserDeactivationLogRepository
	punishRepo       repository.UserPunishmentRepository
	employeeRepo     repository.EmployeeRepository
	msgRepo          repository.MessageRepository
	roomRepo         repository.RoomRepository
	sensitive        *SensitiveWordService
	deviceSvc        *DeviceService
	mediaSvc         *MediaService
	sessions         userSessionNotifier
	notice           userNoticeSender
}

func NewUserService(
	log *logger.Logger,
	db *database.DB,
	authSvc *AuthService,
	smsSvc *SMSService,
	userRepo repository.UserRepository,
	loginRepo repository.UserLoginRecordRepository,
	convDelRepo repository.ConversationDeletionRepository,
	blockRepo repository.UserBlockRepository,
	deactivationRepo repository.UserDeactivationLogRepository,
	punishRepo repository.UserPunishmentRepository,
	employeeRepo repository.EmployeeRepository,
	msgRepo repository.MessageRepository,
	roomRepo repository.RoomRepository,
	mediaSvc *MediaService,
	sensitive *SensitiveWordService,
	deviceSvc *DeviceService,
	sessions userSessionNotifier,
) *UserService {
	return &UserService{
		log:              log,
		db:               db,
		authSvc:          authSvc,
		smsSvc:           smsSvc,
		userRepo:         userRepo,
		loginRepo:        loginRepo,
		convDelRepo:      convDelRepo,
		blockRepo:        blockRepo,
		deactivationRepo: deactivationRepo,
		punishRepo:       punishRepo,
		employeeRepo:     employeeRepo,
		msgRepo:          msgRepo,
		roomRepo:         roomRepo,
		mediaSvc:         mediaSvc,
		sensitive:        sensitive,
		deviceSvc:        deviceSvc,
		sessions:         sessions,
	}
}

func (s *UserService) SetNoticeSender(n userNoticeSender) {
	s.notice = n
}

func (s *UserService) endSession(ctx context.Context, userID uint) {
	if userID == 0 {
		return
	}
	s.sessions.RemoveOnlineGeo(ctx, userID)
	s.sessions.DisconnectUser(userID)
}

func toUserItem(u *model.User, avatarURL string) types.UserItem {
	now := time.Now()
	mobile := ""
	if !u.DeletedAt.Valid {
		mobile = utils.DesensitizeMobile(u.Mobile)
	}
	return types.UserItem{
		ID:          u.ID,
		Mobile:      mobile,
		Nickname:    u.Nickname,
		AvatarURL:   avatarURL,
		Status:      u.Status,
		StatusText:  utils.TextOf(userStatusText, u.Status),
		MutedUntil:  u.MutedUntil,
		MuteText:    muteText(u, now),
		Deactivated: u.DeletedAt.Valid,
		CreatedAt:   utils.FormatDateTime(u.CreatedAt),
		UpdatedAt:   utils.FormatDateTime(u.UpdatedAt),
	}
}

func toUserInfo(u *model.User, avatarURL string) types.UserInfo {
	return types.UserInfo{
		ID:        u.ID,
		Mobile:    utils.DesensitizeMobile(u.Mobile),
		Nickname:  u.Nickname,
		AvatarURL: avatarURL,
		Status:    u.Status,
	}
}

func (s *UserService) avatarURL(ctx context.Context, mediaID uint) string {
	if s.mediaSvc == nil || mediaID == 0 {
		return ""
	}
	return s.mediaSvc.ResolveURL(ctx, mediaID)
}

func (s *UserService) avatarURLs(ctx context.Context, users []*model.User) map[uint]string {
	ids := make([]uint, 0, len(users))
	for _, u := range users {
		if u != nil && u.AvatarMediaID > 0 {
			ids = append(ids, u.AvatarMediaID)
		}
	}
	if s.mediaSvc == nil {
		return map[uint]string{}
	}
	return s.mediaSvc.ResolveURLs(ctx, ids)
}

func (s *UserService) EnsureActive(ctx context.Context, userID uint) error {
	if userID == 0 {
		return errs.NotFound("用户不存在")
	}
	user, err := s.userRepo.GetCached(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if user.Status != model.UserStatusEnabled {
		return errs.Forbidden("对方账号不可用")
	}
	return nil
}

func (s *UserService) AssertAuthenticatedUser(ctx context.Context, userID uint) error {
	if userID == 0 {
		return errs.Unauthorized("用户ID无效")
	}
	user, err := s.userRepo.GetCached(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.Unauthorized("用户不存在")
		}
		return errs.Unauthorized("认证失败")
	}
	if user.Status != model.UserStatusEnabled {
		return errs.Unauthorized("用户已被禁用")
	}
	return nil
}

func isUserMuted(u *model.User, now time.Time) bool {
	if u.MutedUntil == 0 {
		return false
	}
	if u.MutedUntil == model.PermanentMuteUntil {
		return true
	}
	return u.MutedUntil > now.Unix()
}

func muteText(u *model.User, now time.Time) string {
	if !isUserMuted(u, now) {
		return ""
	}
	if u.MutedUntil == model.PermanentMuteUntil {
		return "永久禁言"
	}
	return "禁言至 " + time.Unix(u.MutedUntil, 0).Format("2006-01-02 15:04")
}

func calcMutedUntil(duration int, now time.Time) int64 {
	switch {
	case duration == 0:
		return 0
	case duration < 0:
		return model.PermanentMuteUntil
	default:
		return now.Add(time.Duration(duration) * time.Hour).Unix()
	}
}

func deactivatedDisplayName() string {
	return model.DeactivatedNickname
}

func anonymizedMobile(id uint) string {
	return fmt.Sprintf("9%010d", id)
}
