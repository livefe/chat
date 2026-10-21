package service

import (
	"context"
	"fmt"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/sms"
	"app/pkg/utils"

	"github.com/google/uuid"
)

type SMSService struct {
	log           *logger.Logger
	smsClient     sms.SMS
	cache         *cache.Cache
	smsRecordRepo repository.SMSRecordRepository
	employeeRepo  repository.EmployeeRepository
	mock          bool
}

func NewSMSService(
	log *logger.Logger,
	cfg sms.Config,
	cache *cache.Cache,
	smsRecordRepo repository.SMSRecordRepository,
	employeeRepo repository.EmployeeRepository,
) (*SMSService, error) {
	provider, err := sms.NewSMS(cfg)
	if err != nil {
		return nil, fmt.Errorf("短信客户端初始化失败: %w", err)
	}
	svc := &SMSService{
		log:           log,
		smsClient:     provider,
		cache:         cache,
		smsRecordRepo: smsRecordRepo,
		employeeRepo:  employeeRepo,
		mock:          cfg.Provider == "mock",
	}
	if svc.mock {
		log.Warn(context.Background(), "短信服务运行在 mock 模式，验证码固定，仅可用于开发环境")
	}
	return svc, nil
}

func (s *SMSService) SendLoginCode(ctx context.Context, req *types.SendCodeRequest, ip, ua string) error {
	cooldownKey := fmt.Sprintf(types.CacheSMSSendCooldownKey, req.Mobile, req.Source)
	if ok, err := s.cache.SetNX(ctx, cooldownKey, "1", time.Duration(types.SMSSendCooldownSeconds)*time.Second).Result(); err != nil {
		s.log.Error(ctx, "短信冷却锁失败", logger.Error(err))
		return errs.Unavailable("系统繁忙，请稍后重试", err)
	} else if !ok {
		return errs.RateLimited("发送过于频繁，请稍后再试")
	}

	systemKey := fmt.Sprintf(types.CacheSMSSystemDailySendLimitKey, utils.Today())
	if n, err := s.cache.Get(ctx, systemKey).Int(); err == nil && n >= types.SMSSystemDailyMaxSendCount {
		return errs.RateLimited("系统今日发送量已达上限，请明天再试")
	}

	userKey := fmt.Sprintf(types.CacheSMSUserDailySendLimitKey, req.Mobile, req.Source)
	if n, err := s.cache.Get(ctx, userKey).Int(); err == nil && n >= types.SMSUserDailyMaxSendCount {
		return errs.RateLimited("今日发送次数过多，请明天再试")
	}

	if req.Source == types.SourceAdminLogin && s.employeeRepo != nil {
		emp, err := s.employeeRepo.GetByMobile(ctx, req.Mobile)
		if err != nil || emp == nil || emp.Status != model.EmployeeStatusEnabled {
			return errs.Invalid("该手机号未注册为可用员工")
		}
	}

	code := utils.GenerateCode(types.SMSCodeLength)
	if s.mock {
		code = types.SMSMockFixedCode
	}
	codeKey := fmt.Sprintf(types.CacheSMSLoginCodeKey, req.Mobile, req.Source)
	if err := s.cache.Set(ctx, codeKey, code, types.Expire5m).Err(); err != nil {
		s.log.Error(ctx, "缓存验证码失败", logger.String("mobile", req.Mobile), logger.Error(err))
		return errs.Unavailable("系统繁忙，请稍后重试", err)
	}

	if err := s.sendVerificationSMS(ctx, req.Mobile, code, req.Source, ip, ua); err != nil {
		s.log.Error(ctx, "发送验证码失败", logger.String("mobile", req.Mobile), logger.Error(err))
		_ = s.cache.Del(ctx, cooldownKey).Err()
		return errs.Unavailable("短信发送失败，请稍后重试", err)
	}

	s.incrDaily(ctx, systemKey)
	s.incrDaily(ctx, userKey)
	return nil
}

func (s *SMSService) VerifyCode(ctx context.Context, mobile, code, source string) error {
	failKey := fmt.Sprintf(types.CacheSMSVerifyFailKey, mobile, source)
	if n, err := s.cache.Get(ctx, failKey).Int(); err == nil && n >= types.SMSVerifyMaxFailCount {
		return errs.RateLimited("验证码错误次数过多，请稍后再试")
	}

	codeKey := fmt.Sprintf(types.CacheSMSLoginCodeKey, mobile, source)
	cached, err := s.cache.Get(ctx, codeKey).Result()
	if err != nil {
		return errs.Invalid("验证码已过期")
	}
	if cached != code {
		count, incrErr := s.cache.Incr(ctx, failKey).Result()
		if incrErr == nil && count == 1 {
			s.cache.Expire(ctx, failKey, time.Duration(types.SMSVerifyFailLockSeconds)*time.Second)
		}
		if count >= int64(types.SMSVerifyMaxFailCount) {
			_ = s.cache.Del(ctx, codeKey).Err()
			return errs.RateLimited("验证码错误次数过多，请稍后再试")
		}
		return errs.Invalid("验证码错误")
	}
	_ = s.cache.Del(ctx, codeKey).Err()
	_ = s.cache.Del(ctx, failKey).Err()
	return nil
}

func (s *SMSService) IssueChangeMobileTicket(ctx context.Context, userID uint) (string, error) {
	ticket := uuid.NewString()
	key := fmt.Sprintf(types.CacheChangeMobileTicketKey, userID)
	if err := s.cache.Set(ctx, key, ticket, types.Expire10m).Err(); err != nil {
		return "", fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	return ticket, nil
}

func (s *SMSService) CheckChangeMobileTicket(ctx context.Context, userID uint, ticket string) error {
	if ticket == "" {
		return errs.Invalid("请先验证当前手机号")
	}
	key := fmt.Sprintf(types.CacheChangeMobileTicketKey, userID)
	cached, err := s.cache.Get(ctx, key).Result()
	if err != nil || cached == "" || cached != ticket {
		return errs.Invalid("请先验证当前手机号")
	}
	return nil
}

func (s *SMSService) ClearChangeMobileTicket(ctx context.Context, userID uint) {
	_ = s.cache.Del(ctx, fmt.Sprintf(types.CacheChangeMobileTicketKey, userID)).Err()
}

func (s *SMSService) sendVerificationSMS(ctx context.Context, mobile, code, source, ip, userAgent string) error {
	templateParam := fmt.Sprintf(`{"code":"%s"}`, code)
	result, err := s.smsClient.SendSMS(ctx, mobile, types.SMSSignName, types.SMSTemplateCode, templateParam)

	record := &model.SMSRecord{
		Mobile:        mobile,
		Content:       "登录验证码",
		Source:        source,
		IP:            ip,
		UserAgent:     userAgent,
		SignName:      types.SMSSignName,
		TemplateCode:  types.SMSTemplateCode,
		TemplateParam: `{"code":"******"}`,
	}

	if err != nil {
		record.Status = model.SMSStatusFailed
		record.Remark = err.Error()
	} else {
		record.Status = model.SMSStatusSuccess
		record.BizID = result.BizID
		record.RequestID = result.RequestID
		record.Remark = fmt.Sprintf("Code: %s, Message: %s", result.Code, result.Message)
	}

	if dbErr := s.smsRecordRepo.Create(ctx, record); dbErr != nil {
		s.log.Error(ctx, "保存短信记录失败", logger.String("mobile", mobile), logger.Error(dbErr))
	}

	if err != nil {
		return fmt.Errorf("短信发送失败: %w", err)
	}
	return nil
}

func (s *SMSService) incrDaily(ctx context.Context, key string) {
	count, err := s.cache.Incr(ctx, key).Result()
	if err != nil {
		s.log.Warn(ctx, "发送量计数失败", logger.String("key", key), logger.Error(err))
		return
	}
	if count == 1 {
		s.cache.Expire(ctx, key, types.Expire1d)
	}
}
