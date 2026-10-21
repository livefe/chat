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
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

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
	sensitiveLevelText = map[uint8]string{
		model.SensitiveWordLevelHigh:   "高",
		model.SensitiveWordLevelMedium: "中",
		model.SensitiveWordLevelLow:    "低",
	}
	sensitiveActionText = map[uint8]string{
		model.SensitiveWordActionReject: "拒绝",
	}
	sensitiveStatusText = map[uint8]string{
		model.SensitiveWordStatusDisabled: "停用",
		model.SensitiveWordStatusEnabled:  "启用",
	}
)

type SensitiveWordService struct {
	log  *logger.Logger
	repo repository.SensitiveWordRepository

	mu    sync.RWMutex
	words []string
}

func NewSensitiveWordService(log *logger.Logger, repo repository.SensitiveWordRepository) *SensitiveWordService {
	return &SensitiveWordService{log: log, repo: repo}
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

func (s *SensitiveWordService) AdminCreate(ctx context.Context, req *types.CreateSensitiveWordRequest) (uint, error) {
	row, err := s.buildWord(0, req)
	if err != nil {
		return 0, err
	}
	if _, err := s.repo.GetByWord(ctx, row.Word); err == nil {
		return 0, errs.Conflict("屏蔽词已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("操作失败: %w", err)
	}
	if err := s.repo.Create(ctx, row); err != nil {
		s.log.Error(ctx, "创建屏蔽词失败", logger.Error(err))
		return 0, fmt.Errorf("创建失败: %w", err)
	}
	return row.ID, nil
}

func (s *SensitiveWordService) AdminUpdate(ctx context.Context, req *types.UpdateSensitiveWordRequest) error {
	row, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("屏蔽词不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	next, err := s.buildWord(req.ID, &req.CreateSensitiveWordRequest)
	if err != nil {
		return err
	}
	if next.Word != row.Word {
		if exist, err := s.repo.GetByWord(ctx, next.Word); err == nil && exist.ID != row.ID {
			return errs.Conflict("屏蔽词已存在")
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("操作失败: %w", err)
		}
	}
	row.Word = next.Word
	row.Category = next.Category
	row.Level = next.Level
	row.Action = next.Action
	row.Status = next.Status
	row.Remark = next.Remark
	if err := s.repo.Update(ctx, row); err != nil {
		s.log.Error(ctx, "更新屏蔽词失败", logger.Error(err))
		return fmt.Errorf("更新失败: %w", err)
	}
	return nil
}

func (s *SensitiveWordService) AdminUpdateStatus(ctx context.Context, id uint, status uint8) error {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("屏蔽词不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	row.Status = status
	if err := s.repo.Update(ctx, row); err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}
	return nil
}

func (s *SensitiveWordService) AdminDelete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("屏蔽词不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}
	return nil
}

func (s *SensitiveWordService) AdminImport(ctx context.Context, req *types.ImportSensitiveWordsRequest) (*types.ImportSensitiveWordsResponse, error) {
	category, err := normalizeCategory(req.Category)
	if err != nil {
		return nil, err
	}
	if req.Level == 0 {
		return nil, errs.Invalid("请选择等级")
	}
	level := req.Level
	created, skipped := 0, 0
	for _, line := range strings.Split(req.Words, "\n") {
		word := strings.TrimSpace(line)
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}
		if utf8.RuneCountInString(word) > 64 {
			skipped++
			continue
		}
		if _, err := s.repo.GetByWord(ctx, word); err == nil {
			skipped++
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("导入失败: %w", err)
		}
		row := &model.SensitiveWord{
			Word:     word,
			Category: category,
			Level:    level,
			Action:   model.SensitiveWordActionReject,
			Status:   model.SensitiveWordStatusEnabled,
		}
		if err := s.repo.Create(ctx, row); err != nil {
			s.log.Error(ctx, "导入屏蔽词失败", logger.Error(err))
			return nil, fmt.Errorf("导入失败: %w", err)
		}
		created++
	}
	return &types.ImportSensitiveWordsResponse{Created: created, Skipped: skipped}, nil
}

func (s *SensitiveWordService) buildWord(id uint, req *types.CreateSensitiveWordRequest) (*model.SensitiveWord, error) {
	word := strings.TrimSpace(req.Word)
	if word == "" {
		return nil, errs.Invalid("请填写屏蔽词")
	}
	if utf8.RuneCountInString(word) > 64 {
		return nil, errs.Invalid("屏蔽词过长")
	}
	if req.Level == 0 {
		return nil, errs.Invalid("请选择等级")
	}
	if req.Action == 0 {
		return nil, errs.Invalid("请选择处理方式")
	}
	category, err := normalizeCategory(req.Category)
	if err != nil {
		return nil, err
	}
	return &model.SensitiveWord{
		Base:     model.Base{ID: id},
		Word:     word,
		Category: category,
		Level:    req.Level,
		Action:   req.Action,
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
		Level:        row.Level,
		LevelText:    utils.TextOf(sensitiveLevelText, row.Level),
		Action:       row.Action,
		ActionText:   utils.TextOf(sensitiveActionText, row.Action),
		Status:       row.Status,
		StatusText:   utils.TextOf(sensitiveStatusText, row.Status),
		Remark:       row.Remark,
		CreatedAt:    utils.FormatDateTime(row.CreatedAt),
		UpdatedAt:    utils.FormatDateTime(row.UpdatedAt),
	}
}
