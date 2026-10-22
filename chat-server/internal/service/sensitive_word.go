package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	sensitiveCategoryText = map[string]string{
		model.SensitiveWordCategoryAbuse:    "辱骂",
		model.SensitiveWordCategoryPorn:     "色情",
		model.SensitiveWordCategoryPolitics: "政治",
		model.SensitiveWordCategoryAd:       "广告",
		model.SensitiveWordCategoryOther:    "其他",
	}
	sensitiveStatusText = map[uint8]string{
		model.SensitiveWordStatusDisabled: "停用",
		model.SensitiveWordStatusEnabled:  "启用",
	}
)

type SensitiveWordService struct {
	log        *logger.Logger
	db         *database.DB
	cache      *cache.Cache
	instanceID string
	repo       repository.SensitiveWordRepository

	mu    sync.RWMutex
	words []string

	sub    *redis.PubSub
	cancel context.CancelFunc
}

func NewSensitiveWordService(
	log *logger.Logger,
	db *database.DB,
	c *cache.Cache,
	instanceID string,
	repo repository.SensitiveWordRepository,
) *SensitiveWordService {
	return &SensitiveWordService{log: log, db: db, cache: c, instanceID: instanceID, repo: repo}
}

// Start 加载词库并订阅其它实例发出的重载通知。
func (s *SensitiveWordService) Start(ctx context.Context) error {
	if err := s.Load(ctx); err != nil {
		return err
	}
	subCtx, cancel := context.WithCancel(context.Background())
	sub := s.cache.Subscribe(subCtx, types.CacheSensitiveWordReloadChannel)
	if _, err := sub.Receive(subCtx); err != nil {
		cancel()
		_ = sub.Close()
		return fmt.Errorf("订阅屏蔽词重载通知失败: %w", err)
	}
	s.sub = sub
	s.cancel = cancel
	go func() {
		for msg := range sub.Channel() {
			if msg.Payload == s.instanceID {
				continue
			}
			if err := s.Load(subCtx); err != nil {
				s.log.Error(subCtx, "屏蔽词重载失败", logger.Error(err))
			}
		}
	}()
	return nil
}

func (s *SensitiveWordService) Close() error {
	if s.cancel != nil {
		s.cancel()
	}
	if s.sub != nil {
		return s.sub.Close()
	}
	return nil
}

func (s *SensitiveWordService) Load(ctx context.Context) error {
	list, err := s.repo.ListEnabled(ctx)
	if err != nil {
		return err
	}
	words := make([]string, 0, len(list))
	seen := make(map[string]struct{}, len(list))
	for _, row := range list {
		n := normalizeSensitiveText(row.Word)
		if n == "" {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		words = append(words, n)
	}
	s.mu.Lock()
	s.words = words
	s.mu.Unlock()
	s.log.Info(ctx, "敏感词已加载", logger.Int("count", len(words)))
	return nil
}

// reload 在词库变更后刷新本实例，并通知其它实例刷新。
func (s *SensitiveWordService) reload(ctx context.Context) {
	if err := s.Load(ctx); err != nil {
		s.log.Error(ctx, "屏蔽词重载失败", logger.Error(err))
	}
	if err := s.cache.Publish(ctx, types.CacheSensitiveWordReloadChannel, s.instanceID).Err(); err != nil {
		s.log.Error(ctx, "广播屏蔽词重载失败", logger.Error(err))
	}
}

func (s *SensitiveWordService) Contains(text string) bool {
	n := normalizeSensitiveText(text)
	if n == "" {
		return false
	}
	s.mu.RLock()
	words := s.words
	s.mu.RUnlock()
	for _, w := range words {
		if w != "" && strings.Contains(n, w) {
			return true
		}
	}
	return false
}

func (s *SensitiveWordService) AssertClean(texts ...string) error {
	for _, text := range texts {
		if strings.TrimSpace(text) == "" {
			continue
		}
		if s.Contains(text) {
			return errs.Invalid("内容包含敏感词")
		}
	}
	return nil
}

func assertClean(sensitive *SensitiveWordService, texts ...string) error {
	if sensitive == nil {
		return nil
	}
	return sensitive.AssertClean(texts...)
}

func normalizeSensitiveText(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsSpace(r) || r == 0x200B || r == 0xFEFF {
			continue
		}
		switch r {
		case '.', '·', '*', '_', '-', '|', '/', '\\', '`', '~', ',', '，', '。', '、', '!', '！', '?', '？':
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func (s *SensitiveWordService) AdminList(ctx context.Context, req *types.SensitiveWordListRequest) (*types.SensitiveWordListResponse, error) {
	list, total, err := s.repo.List(ctx, repository.SensitiveWordListFilter{
		Word:     strings.TrimSpace(req.Word),
		Category: strings.TrimSpace(req.Category),
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items := make([]types.SensitiveWordItem, 0, len(list))
	for _, row := range list {
		items = append(items, toSensitiveWordItem(row))
	}
	return &types.SensitiveWordListResponse{Total: total, List: items}, nil
}

// normalizedWordIndex 返回「归一化后的词 → 原词」，按匹配规则判重。
func (s *SensitiveWordService) normalizedWordIndex(ctx context.Context, excludeID uint) (map[string]string, error) {
	all, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	index := make(map[string]string, len(all))
	for _, row := range all {
		if row.ID == excludeID {
			continue
		}
		if n := normalizeSensitiveText(row.Word); n != "" {
			index[n] = row.Word
		}
	}
	return index, nil
}

func (s *SensitiveWordService) assertWordUnique(ctx context.Context, word string, excludeID uint) error {
	index, err := s.normalizedWordIndex(ctx, excludeID)
	if err != nil {
		return fmt.Errorf("查询失败: %w", err)
	}
	if exist, ok := index[normalizeSensitiveText(word)]; ok {
		if exist == word {
			return errs.Conflict("屏蔽词已存在")
		}
		return errs.Conflict(fmt.Sprintf("与已有屏蔽词「%s」等价", exist))
	}
	return nil
}

func (s *SensitiveWordService) AdminCreate(ctx context.Context, req *types.CreateSensitiveWordRequest) (uint, error) {
	row, err := buildSensitiveWord(req)
	if err != nil {
		return 0, err
	}
	if err := s.assertWordUnique(ctx, row.Word, 0); err != nil {
		return 0, err
	}
	if err := s.repo.Create(ctx, row); err != nil {
		s.log.Error(ctx, "创建屏蔽词失败", logger.Error(err))
		return 0, fmt.Errorf("创建失败: %w", err)
	}
	s.reload(ctx)
	return row.ID, nil
}

func (s *SensitiveWordService) AdminUpdate(ctx context.Context, req *types.UpdateSensitiveWordRequest) error {
	row, err := s.getByID(ctx, req.ID)
	if err != nil {
		return err
	}
	next, err := buildSensitiveWord(&req.CreateSensitiveWordRequest)
	if err != nil {
		return err
	}
	if next.Word != row.Word {
		if err := s.assertWordUnique(ctx, next.Word, row.ID); err != nil {
			return err
		}
	}
	row.Word = next.Word
	row.Category = next.Category
	row.Status = next.Status
	row.Remark = next.Remark
	if err := s.repo.Update(ctx, row); err != nil {
		s.log.Error(ctx, "更新屏蔽词失败", logger.Error(err))
		return fmt.Errorf("更新失败: %w", err)
	}
	s.reload(ctx)
	return nil
}

func (s *SensitiveWordService) AdminUpdateStatus(ctx context.Context, id uint, status uint8) error {
	row, err := s.getByID(ctx, id)
	if err != nil {
		return err
	}
	row.Status = status
	if err := s.repo.Update(ctx, row); err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}
	s.reload(ctx)
	return nil
}

func (s *SensitiveWordService) AdminDelete(ctx context.Context, id uint) error {
	if _, err := s.getByID(ctx, id); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}
	s.reload(ctx)
	return nil
}

func (s *SensitiveWordService) AdminBatchDelete(ctx context.Context, ids []uint) error {
	if err := s.repo.DeleteByIDs(ctx, ids); err != nil {
		s.log.Error(ctx, "批量删除屏蔽词失败", logger.Error(err))
		return fmt.Errorf("删除失败: %w", err)
	}
	s.reload(ctx)
	return nil
}

func (s *SensitiveWordService) AdminBatchUpdateStatus(ctx context.Context, req *types.BatchSensitiveWordStatusRequest) error {
	if err := s.repo.UpdateStatusByIDs(ctx, req.IDs, req.Status); err != nil {
		s.log.Error(ctx, "批量更新屏蔽词状态失败", logger.Error(err))
		return fmt.Errorf("更新失败: %w", err)
	}
	s.reload(ctx)
	return nil
}

func (s *SensitiveWordService) AdminImport(ctx context.Context, req *types.ImportSensitiveWordsRequest) (*types.ImportSensitiveWordsResponse, error) {
	category, err := normalizeCategory(req.Category)
	if err != nil {
		return nil, err
	}
	out := &types.ImportSensitiveWordsResponse{
		Existing:   []string{},
		TooLong:    []string{},
		Duplicated: []string{},
		Invalid:    []string{},
	}
	err = database.Transaction(ctx, s.db, func(ctx context.Context) error {
		index, err := s.normalizedWordIndex(ctx, 0)
		if err != nil {
			return err
		}
		batch := make(map[string]struct{})
		rows := make([]*model.SensitiveWord, 0)
		for _, line := range strings.Split(req.Words, "\n") {
			word := strings.TrimSpace(line)
			if word == "" {
				continue
			}
			if utf8.RuneCountInString(word) > types.SensitiveWordMaxLen {
				out.TooLong = append(out.TooLong, word)
				continue
			}
			n := normalizeSensitiveText(word)
			if n == "" {
				out.Invalid = append(out.Invalid, word)
				continue
			}
			if _, ok := index[n]; ok {
				out.Existing = append(out.Existing, word)
				continue
			}
			if _, ok := batch[n]; ok {
				out.Duplicated = append(out.Duplicated, word)
				continue
			}
			batch[n] = struct{}{}
			rows = append(rows, &model.SensitiveWord{
				Word:     word,
				Category: category,
				Status:   model.SensitiveWordStatusEnabled,
			})
		}
		if err := s.repo.CreateBatch(ctx, rows); err != nil {
			return err
		}
		out.Created = len(rows)
		return nil
	})
	if err != nil {
		s.log.Error(ctx, "导入屏蔽词失败", logger.Error(err))
		return nil, fmt.Errorf("导入失败: %w", err)
	}
	if out.Created > 0 {
		s.reload(ctx)
	}
	return out, nil
}

func (s *SensitiveWordService) getByID(ctx context.Context, id uint) (*model.SensitiveWord, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("屏蔽词不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return row, nil
}

func buildSensitiveWord(req *types.CreateSensitiveWordRequest) (*model.SensitiveWord, error) {
	word := strings.TrimSpace(req.Word)
	if word == "" {
		return nil, errs.Invalid("请填写屏蔽词")
	}
	if utf8.RuneCountInString(word) > types.SensitiveWordMaxLen {
		return nil, errs.Invalid(fmt.Sprintf("屏蔽词最多 %d 个字", types.SensitiveWordMaxLen))
	}
	if normalizeSensitiveText(word) == "" {
		return nil, errs.Invalid("屏蔽词不能只包含空格或分隔符")
	}
	category, err := normalizeCategory(req.Category)
	if err != nil {
		return nil, err
	}
	return &model.SensitiveWord{
		Word:     word,
		Category: category,
		Status:   req.Status,
		Remark:   strings.TrimSpace(req.Remark),
	}, nil
}

func normalizeCategory(c string) (string, error) {
	c = strings.TrimSpace(c)
	switch c {
	case model.SensitiveWordCategoryAbuse,
		model.SensitiveWordCategoryPorn,
		model.SensitiveWordCategoryPolitics,
		model.SensitiveWordCategoryAd,
		model.SensitiveWordCategoryOther:
		return c, nil
	default:
		return "", errs.Invalid("请选择分类")
	}
}

func toSensitiveWordItem(row *model.SensitiveWord) types.SensitiveWordItem {
	return types.SensitiveWordItem{
		ID:           row.ID,
		Word:         row.Word,
		Category:     row.Category,
		CategoryText: utils.TextOf(sensitiveCategoryText, row.Category),
		Status:       row.Status,
		StatusText:   utils.TextOf(sensitiveStatusText, row.Status),
		Remark:       row.Remark,
		CreatedAt:    utils.FormatDateTime(row.CreatedAt),
		UpdatedAt:    utils.FormatDateTime(row.UpdatedAt),
	}
}
