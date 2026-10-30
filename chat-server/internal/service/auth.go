package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/auth"
	"app/pkg/cache"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var consumeRefreshJTIScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then
  return 0
end
redis.call('DEL', KEYS[1])
redis.call('SREM', KEYS[2], ARGV[1])
return 1
`)

type AuthService struct {
	log               *logger.Logger
	cache             *cache.Cache
	tokenAuth         *auth.Auth
	smsSvc            *SMSService
	userRepo          repository.UserRepository
	userLoginRepo     repository.UserLoginRecordRepository
	employeeRepo      repository.EmployeeRepository
	employeeRoleRepo  repository.EmployeeRoleRepository
	employeeLoginRepo repository.EmployeeLoginRecordRepository
	deptRepo          repository.DepartmentRepository
	mediaSvc          *MediaService
}

func NewAuthService(
	log *logger.Logger,
	c *cache.Cache,
	tokenAuth *auth.Auth,
	smsSvc *SMSService,
	userRepo repository.UserRepository,
	userLoginRepo repository.UserLoginRecordRepository,
	employeeRepo repository.EmployeeRepository,
	employeeRoleRepo repository.EmployeeRoleRepository,
	employeeLoginRepo repository.EmployeeLoginRecordRepository,
	deptRepo repository.DepartmentRepository,
	mediaSvc *MediaService,
) *AuthService {
	return &AuthService{
		log:               log,
		cache:             c,
		tokenAuth:         tokenAuth,
		smsSvc:            smsSvc,
		userRepo:          userRepo,
		userLoginRepo:     userLoginRepo,
		employeeRepo:      employeeRepo,
		employeeRoleRepo:  employeeRoleRepo,
		employeeLoginRepo: employeeLoginRepo,
		deptRepo:          deptRepo,
		mediaSvc:          mediaSvc,
	}
}

func (s *AuthService) userAvatarURL(ctx context.Context, mediaID uint) string {
	if s.mediaSvc == nil || mediaID == 0 {
		return ""
	}
	return s.mediaSvc.ResolveURL(ctx, mediaID)
}

func (s *AuthService) Revoke(ctx context.Context, subjectType string, subjectID uint) {
	if subjectID == 0 {
		return
	}
	key := fmt.Sprintf(types.CacheAuthTokenEpochKey, subjectType, subjectID)
	if err := s.cache.Set(ctx, key, time.Now().Unix(), s.tokenAuth.RefreshTTL()).Err(); err != nil {
		s.log.Error(ctx, "写入令牌吊销时间失败",
			logger.String("subject_type", subjectType), logger.Uint("subject_id", subjectID), logger.Error(err))
	}
	s.revokeAllRefreshJTI(ctx, subjectType, subjectID)
}

func (s *AuthService) storeRefreshJTI(ctx context.Context, subjectType string, subjectID uint, jti string) error {
	if jti == "" || subjectID == 0 {
		return nil
	}
	ttl := s.tokenAuth.RefreshTTL()
	key := fmt.Sprintf(types.CacheAuthRefreshJTIKey, subjectType, subjectID, jti)
	setKey := fmt.Sprintf(types.CacheAuthRefreshJTISetKey, subjectType, subjectID)
	pipe := s.cache.TxPipeline()
	pipe.Set(ctx, key, 1, ttl)
	pipe.SAdd(ctx, setKey, jti)
	pipe.Expire(ctx, setKey, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *AuthService) restoreRefreshJTI(ctx context.Context, claims *auth.Claims) {
	if err := s.storeRefreshJTI(ctx, claims.SubjectType, claims.SubjectID, claims.ID); err != nil {
		s.log.Error(ctx, "恢复原刷新令牌失败",
			logger.String("subject_type", claims.SubjectType), logger.Uint("subject_id", claims.SubjectID), logger.Error(err))
	}
}

func (s *AuthService) consumeRefreshJTI(ctx context.Context, subjectType string, subjectID uint, jti string) error {
	if jti == "" {
		return errs.Unauthorized("无效的刷新令牌")
	}
	key := fmt.Sprintf(types.CacheAuthRefreshJTIKey, subjectType, subjectID, jti)
	setKey := fmt.Sprintf(types.CacheAuthRefreshJTISetKey, subjectType, subjectID)
	n, err := consumeRefreshJTIScript.Run(ctx, s.cache, []string{key, setKey}, jti).Int64()
	if err != nil {
		return fmt.Errorf("登录状态校验失败，请稍后重试: %w", err)
	}
	if n == 0 {
		return errs.Unauthorized("刷新令牌已失效，请重新登录")
	}
	return nil
}

func (s *AuthService) revokeAllRefreshJTI(ctx context.Context, subjectType string, subjectID uint) {
	setKey := fmt.Sprintf(types.CacheAuthRefreshJTISetKey, subjectType, subjectID)
	jtis, err := s.cache.SMembers(ctx, setKey).Result()
	if err != nil || len(jtis) == 0 {
		_ = s.cache.Del(ctx, setKey).Err()
		return
	}
	keys := make([]string, 0, len(jtis)+1)
	for _, jti := range jtis {
		keys = append(keys, fmt.Sprintf(types.CacheAuthRefreshJTIKey, subjectType, subjectID, jti))
	}
	keys = append(keys, setKey)
	if err := s.cache.Del(ctx, keys...).Err(); err != nil {
		s.log.Warn(ctx, "批量吊销刷新令牌失败",
			logger.String("subject_type", subjectType), logger.Uint("subject_id", subjectID), logger.Error(err))
	}
}

func (s *AuthService) AssertNotRevoked(ctx context.Context, claims *auth.Claims) error {
	if claims == nil || claims.SubjectID == 0 {
		return errs.Unauthorized("无效的令牌")
	}
	key := fmt.Sprintf(types.CacheAuthTokenEpochKey, claims.SubjectType, claims.SubjectID)
	epoch, err := s.cache.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return fmt.Errorf("登录状态校验失败，请稍后重试: %w", err)
	}
	if claims.IssuedAt != nil && claims.IssuedAt.Unix() < epoch {
		return errs.Unauthorized("登录已失效，请重新登录")
	}
	return nil
}

func (s *AuthService) refreshTokenPair(
	ctx context.Context,
	refreshToken string,
	expectedSubjectType string,
	ensureActive func(ctx context.Context, subjectID uint) error,
	afterRevoke func(ctx context.Context, subjectID uint) error,
) (*types.RefreshTokenResponse, error) {
	claims, err := s.tokenAuth.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, errs.Unauthorized("无效的刷新令牌")
	}
	if claims.SubjectType != expectedSubjectType {
		return nil, errs.Unauthorized("无效的刷新令牌")
	}
	if err := ensureActive(ctx, claims.SubjectID); err != nil {
		return nil, err
	}
	if err := s.AssertNotRevoked(ctx, claims); err != nil {
		return nil, err
	}
	if afterRevoke != nil {
		if err := afterRevoke(ctx, claims.SubjectID); err != nil {
			return nil, err
		}
	}
	if err := s.consumeRefreshJTI(ctx, claims.SubjectType, claims.SubjectID, claims.ID); err != nil {
		return nil, err
	}

	pair, err := s.tokenAuth.GenerateTokenPair(claims.SubjectType, claims.SubjectID)
	if err != nil {
		s.log.Error(ctx, "生成令牌失败",
			logger.String("subject_type", claims.SubjectType), logger.Uint("subject_id", claims.SubjectID), logger.Error(err))
		s.restoreRefreshJTI(ctx, claims)
		return nil, fmt.Errorf("登录状态更新失败，请稍后重试: %w", err)
	}
	if err := s.storeRefreshJTI(ctx, claims.SubjectType, claims.SubjectID, pair.RefreshJTI); err != nil {
		s.log.Error(ctx, "存储刷新令牌失败",
			logger.String("subject_type", claims.SubjectType), logger.Uint("subject_id", claims.SubjectID), logger.Error(err))
		s.restoreRefreshJTI(ctx, claims)
		return nil, fmt.Errorf("登录状态更新失败，请稍后重试: %w", err)
	}
	return &types.RefreshTokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}

func (s *AuthService) LoginUser(ctx context.Context, req *types.LoginRequest, ip, ua string) (*types.LoginResponse, error) {
	record := &model.UserLoginRecord{
		Mobile:    req.Mobile,
		IP:        ip,
		UserAgent: ua,
		LoginType: types.LoginTypeSMS,
	}
	defer func() {
		if err := s.userLoginRepo.Create(ctx, record); err != nil {
			s.log.Error(ctx, "创建登录记录失败", logger.String("mobile", req.Mobile), logger.Error(err))
		}
	}()

	if err := s.smsSvc.VerifyCode(ctx, req.Mobile, req.Code, req.Source); err != nil {
		record.Remark = err.Error()
		s.log.Warn(ctx, "用户登录失败", logger.String("reason", err.Error()))
		return nil, err
	}

	user, err := s.userRepo.GetByMobile(ctx, req.Mobile)
	isFirstLogin := false
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			record.Remark = fmt.Sprintf("查询用户失败: %v", err)
			s.log.Error(ctx, "用户登录失败", logger.String("reason", "查询用户失败"), logger.Error(err))
			return nil, fmt.Errorf("登录失败，请稍后重试: %w", err)
		}
		isFirstLogin = true
		user = &model.User{
			Mobile:   req.Mobile,
			Nickname: utils.RandomNickname(),
			Status:   model.UserStatusEnabled,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			record.Remark = fmt.Sprintf("用户创建失败: %v", err)
			s.log.Error(ctx, "用户登录失败", logger.String("reason", "用户创建失败"), logger.Error(err))
			return nil, fmt.Errorf("登录失败，请稍后重试: %w", err)
		}
	}

	if user.Status != model.UserStatusEnabled {
		record.Remark = "用户已被禁用"
		s.log.Warn(ctx, "用户登录失败", logger.String("reason", "用户已被禁用"), logger.Uint("user_id", user.ID))
		return nil, errs.Forbidden("用户已被禁用")
	}
	record.UserID = user.ID
	record.Status = model.SMSStatusSuccess

	pair, err := s.tokenAuth.GenerateTokenPair(auth.SubjectTypeUser, user.ID)
	if err != nil {
		s.log.Error(ctx, "用户登录失败",
			logger.String("reason", "生成令牌失败"),
			logger.Uint("user_id", user.ID),
			logger.Error(err),
		)
		return nil, fmt.Errorf("登录失败，请稍后重试: %w", err)
	}
	if err := s.storeRefreshJTI(ctx, auth.SubjectTypeUser, user.ID, pair.RefreshJTI); err != nil {
		s.log.Error(ctx, "用户登录失败",
			logger.String("reason", "存储刷新令牌失败"),
			logger.Uint("user_id", user.ID),
			logger.Error(err),
		)
		return nil, fmt.Errorf("登录失败，请稍后重试: %w", err)
	}

	return &types.LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		IsFirstLogin: isFirstLogin,
		User:         toUserInfo(user, s.userAvatarURL(ctx, user.AvatarMediaID)),
	}, nil
}

func (s *AuthService) RefreshUser(ctx context.Context, refreshToken string) (*types.RefreshTokenResponse, error) {
	return s.refreshTokenPair(ctx, refreshToken, auth.SubjectTypeUser,
		func(ctx context.Context, subjectID uint) error {
			user, err := s.userRepo.GetByID(ctx, subjectID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errs.Unauthorized("登录已失效，请重新登录")
				}
				return fmt.Errorf("系统错误，请稍后重试: %w", err)
			}
			if user.Status != model.UserStatusEnabled {
				return errs.Forbidden("用户已被禁用")
			}
			return nil
		},
		nil,
	)
}
