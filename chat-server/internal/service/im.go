package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

const maxIMTextLen = 1000
const maxIMClientMsgIDLen = 64

type imDeliverer interface {
	DeliverMessage(ctx context.Context, employeeIDs []uint, item *types.IMMessageItem)
	DeliverRecall(ctx context.Context, employeeIDs []uint, conversationID, messageID uint)
	OnlineEmployees(ctx context.Context, ids []uint) map[uint]struct{}
}

type IMService struct {
	log             *logger.Logger
	cache           *cache.Cache
	db              *database.DB
	convRepo        repository.IMConversationRepository
	memberRepo      repository.IMMemberRepository
	messageRepo     repository.IMMessageRepository
	staffDMPairRepo repository.IMStaffDMPairRepository
	employeeRepo    repository.EmployeeRepository
	departmentRepo  repository.DepartmentRepository
	mediaSvc        *MediaService
	sensitiveWord   *SensitiveWordService
	deliverer       imDeliverer
}

func NewIMService(
	log *logger.Logger,
	cache *cache.Cache,
	db *database.DB,
	convRepo repository.IMConversationRepository,
	memberRepo repository.IMMemberRepository,
	messageRepo repository.IMMessageRepository,
	staffDMPairRepo repository.IMStaffDMPairRepository,
	employeeRepo repository.EmployeeRepository,
	departmentRepo repository.DepartmentRepository,
	mediaSvc *MediaService,
	sensitiveWord *SensitiveWordService,
	deliverer imDeliverer,
) *IMService {
	return &IMService{
		log:             log,
		cache:           cache,
		db:              db,
		convRepo:        convRepo,
		memberRepo:      memberRepo,
		messageRepo:     messageRepo,
		staffDMPairRepo: staffDMPairRepo,
		employeeRepo:    employeeRepo,
		departmentRepo:  departmentRepo,
		mediaSvc:        mediaSvc,
		sensitiveWord:   sensitiveWord,
		deliverer:       deliverer,
	}
}

func (s *IMService) transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return database.Transaction(ctx, s.db, fn)
}

func (s *IMService) requireActiveMember(ctx context.Context, conversationID, employeeID uint) (*model.IMMember, error) {
	m, err := s.memberRepo.GetActive(ctx, conversationID, model.IMPrincipalEmployee, employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.Forbidden("无权访问该会话")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return m, nil
}

func (s *IMService) departmentNames(ctx context.Context, ids []uint) map[uint]string {
	out := make(map[uint]string)
	if len(ids) == 0 {
		return out
	}
	uniq := make(map[uint]struct{}, len(ids))
	list := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := uniq[id]; ok {
			continue
		}
		uniq[id] = struct{}{}
		list = append(list, id)
	}
	deps, err := s.departmentRepo.GetByIDs(ctx, list)
	if err != nil {
		return out
	}
	for _, d := range deps {
		out[d.ID] = d.Name
	}
	return out
}

func (s *IMService) UnreadMap(ctx context.Context, employeeID uint) (map[string]int64, error) {
	raw, err := s.cache.HGetAll(ctx, formatIMUnreadKey(employeeID)).Result()
	if err != nil {
		return map[string]int64{}, err
	}
	out := make(map[string]int64, len(raw))
	for k, v := range raw {
		n, _ := strconv.ParseInt(v, 10, 64)
		out[k] = n
	}
	return out, nil
}

func (s *IMService) IncrUnread(ctx context.Context, employeeID, conversationID uint) error {
	key := formatIMUnreadKey(employeeID)
	field := strconv.FormatUint(uint64(conversationID), 10)
	pipe := s.cache.Pipeline()
	pipe.HIncrBy(ctx, key, field, 1)
	pipe.Expire(ctx, key, types.Expire1d*31)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *IMService) ClearUnread(ctx context.Context, employeeID, conversationID uint) error {
	field := strconv.FormatUint(uint64(conversationID), 10)
	return s.cache.HDel(ctx, formatIMUnreadKey(employeeID), field).Err()
}

func formatIMUnreadKey(employeeID uint) string {
	return fmt.Sprintf(types.CacheIMUnreadKey, employeeID)
}

func orderPair(a, b uint) (uint, uint) {
	if a < b {
		return a, b
	}
	return b, a
}

func isAllDigits(s string) bool {
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
