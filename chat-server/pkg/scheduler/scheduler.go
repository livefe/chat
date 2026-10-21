package scheduler

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"app/pkg/logger"

	"github.com/robfig/cron/v3"
)

type JobFunc func(ctx context.Context) error

type Locker interface {
	TryLock(ctx context.Context, key string, ttl time.Duration) (unlock func(context.Context) error, acquired bool, err error)
}

type Scheduler struct {
	cron    *cron.Cron
	log     *logger.Logger
	locker  Locker
	rootCtx context.Context
	cfg     *Config
	runners map[string]JobFunc
	jobs    map[string]cron.EntryID
	metrics *Metrics
	mu      sync.Mutex
	started bool
}

func New(log *logger.Logger, locker Locker, ctx context.Context) (*Scheduler, error) {
	if log == nil {
		return nil, fmt.Errorf("定时任务日志未初始化")
	}
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation(strings.TrimSpace(cfg.Timezone))
	if err != nil {
		return nil, fmt.Errorf("解析定时任务时区失败: %w", err)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	c := cron.New(
		cron.WithLocation(loc),
		cron.WithChain(
			cron.Recover(newCronLogger(log)),
			cron.SkipIfStillRunning(newCronLogger(log)),
		),
	)

	return &Scheduler{
		cron:    c,
		log:     log,
		locker:  locker,
		rootCtx: ctx,
		cfg:     cfg,
		runners: make(map[string]JobFunc),
		jobs:    make(map[string]cron.EntryID),
		metrics: NewMetrics(),
	}, nil
}

func (s *Scheduler) Metrics() *Metrics {
	if s == nil {
		return nil
	}
	return s.metrics
}

func (s *Scheduler) Register(runners map[string]JobFunc) error {
	if s == nil || s.cfg == nil {
		return fmt.Errorf("定时任务调度器未初始化")
	}

	enabled := 0
	for _, j := range s.cfg.Jobs {
		run, ok := runners[j.Name]
		if !j.Enabled {
			s.log.Info(context.Background(), "跳过已禁用定时任务",
				logger.String("job", j.Name),
				logger.String("desc", j.Desc),
			)
			if ok {
				s.runners[j.Name] = run
			}
			continue
		}
		if !ok || run == nil {
			return fmt.Errorf("未实现的定时任务: %s", j.Name)
		}

		s.log.Info(context.Background(), "注册定时任务",
			logger.String("job", j.Name),
			logger.String("spec", j.Spec),
			logger.String("desc", j.Desc),
			logger.Duration("timeout", j.Timeout),
			logger.Duration("lock_ttl", s.cfg.LockTTL),
		)
		if err := s.addCron(j); err != nil {
			return err
		}
		s.runners[j.Name] = run
		enabled++
	}
	if enabled == 0 {
		return fmt.Errorf("没有启用的定时任务")
	}
	return nil
}

func (s *Scheduler) addCron(job JobConfig) error {
	if _, exists := s.jobs[job.Name]; exists {
		return fmt.Errorf("定时任务 %q 已存在", job.Name)
	}
	name := job.Name
	entryID, err := s.cron.AddFunc(job.Spec, func() {
		_, _ = s.run(s.rootCtx, name)
	})
	if err != nil {
		return fmt.Errorf("注册定时任务 %q 失败: %w", name, err)
	}
	s.jobs[name] = entryID
	return nil
}

func (s *Scheduler) RunOnce(ctx context.Context, name string) error {
	if s == nil {
		return fmt.Errorf("定时任务调度器未初始化")
	}
	if _, ok := s.runners[name]; !ok {
		return fmt.Errorf("未找到定时任务: %s", name)
	}
	status, err := s.run(ctx, name)
	if status == StatusSkipped {
		return fmt.Errorf("任务 %s 正在其他实例执行", name)
	}
	return err
}

func (s *Scheduler) run(parent context.Context, name string) (string, error) {
	job, ok := s.jobByName(name)
	if !ok || job.Timeout <= 0 {
		return StatusError, fmt.Errorf("定时任务 %s 未配置超时", name)
	}
	fn := s.runners[name]

	ctx := parent
	if ctx == nil {
		ctx = context.Background()
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, job.Timeout)
	defer cancel()

	start := time.Now()
	s.log.Info(ctx, "定时任务开始执行", logger.String("job", name))

	status, runErr := s.execute(ctx, name, s.cfg.LockTTL, fn)
	duration := time.Since(start)
	s.metrics.Record(name, status, duration)

	fields := []logger.Field{
		logger.String("job", name),
		logger.String("status", status),
		logger.Duration("duration", duration),
	}
	switch status {
	case StatusSkipped:
		s.log.Info(ctx, "定时任务已跳过", fields...)
	case StatusSuccess:
		s.log.Info(ctx, "定时任务执行完成", fields...)
	default:
		fields = append(fields, logger.Error(runErr))
		s.log.Error(ctx, "定时任务执行失败", fields...)
	}
	return status, runErr
}

func (s *Scheduler) jobByName(name string) (JobConfig, bool) {
	if s == nil || s.cfg == nil {
		return JobConfig{}, false
	}
	for _, job := range s.cfg.Jobs {
		if job.Name == name {
			return job, true
		}
	}
	return JobConfig{}, false
}

func (s *Scheduler) execute(ctx context.Context, name string, lockTTL time.Duration, fn JobFunc) (status string, err error) {
	if s.locker != nil && lockTTL > 0 {
		unlock, acquired, lockErr := s.locker.TryLock(ctx, lockKey(name), lockTTL)
		if lockErr != nil {
			return StatusError, fmt.Errorf("获取任务锁失败: %w", lockErr)
		}
		if !acquired {
			return StatusSkipped, nil
		}
		defer func() {
			if uerr := unlock(context.Background()); uerr != nil {
				s.log.Warn(ctx, "释放任务锁失败", logger.String("job", name), logger.Error(uerr))
			}
		}()
	}

	if err := fn(ctx); err != nil {
		return StatusError, err
	}
	return StatusSuccess, nil
}

func lockKey(name string) string {
	return "scheduler:lock:" + name
}

func (s *Scheduler) Start() {
	if s == nil || s.cron == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return
	}
	s.cron.Start()
	s.started = true
	s.log.Info(context.Background(), "定时任务调度器已启动", logger.Int("jobs", len(s.jobs)))
}

func (s *Scheduler) Stop(ctx context.Context) error {
	if s == nil || s.cron == nil {
		return nil
	}

	s.mu.Lock()
	started := s.started
	s.mu.Unlock()
	if !started {
		return nil
	}

	stopCtx := s.cron.Stop()
	select {
	case <-stopCtx.Done():
		s.mu.Lock()
		s.started = false
		s.mu.Unlock()
		s.log.Info(ctx, "定时任务调度器已停止")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

const (
	StatusSuccess = "success"
	StatusError   = "error"
	StatusSkipped = "skipped"
)

type jobCounters struct {
	success    atomic.Int64
	error      atomic.Int64
	skipped    atomic.Int64
	durationNs atomic.Int64
}

type Metrics struct {
	mu   sync.Mutex
	jobs map[string]*jobCounters
}

func NewMetrics() *Metrics {
	return &Metrics{jobs: make(map[string]*jobCounters)}
}

func (m *Metrics) Record(job, status string, duration time.Duration) {
	if m == nil {
		return
	}
	c := m.counter(job)
	switch status {
	case StatusSuccess:
		c.success.Add(1)
	case StatusSkipped:
		c.skipped.Add(1)
	default:
		c.error.Add(1)
	}
	c.durationNs.Add(duration.Nanoseconds())
}

func (m *Metrics) counter(job string) *jobCounters {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.jobs[job]
	if !ok {
		c = &jobCounters{}
		m.jobs[job] = c
	}
	return c
}

func (m *Metrics) Snapshot() map[string]map[string]int64 {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]map[string]int64, len(m.jobs))
	for name, c := range m.jobs {
		out[name] = map[string]int64{
			"success":     c.success.Load(),
			"error":       c.error.Load(),
			"skipped":     c.skipped.Load(),
			"duration_ms": c.durationNs.Load() / int64(time.Millisecond),
		}
	}
	return out
}

type cronLogger struct {
	log *logger.Logger
}

func newCronLogger(log *logger.Logger) cron.Logger {
	return &cronLogger{log: log}
}

func (l *cronLogger) Info(msg string, keysAndValues ...interface{}) {
	l.log.Info(context.Background(), msg, toFields(keysAndValues...)...)
}

func (l *cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	fields := toFields(keysAndValues...)
	fields = append(fields, logger.Error(err))
	l.log.Error(context.Background(), msg, fields...)
}

func toFields(keysAndValues ...interface{}) []logger.Field {
	if len(keysAndValues) == 0 {
		return nil
	}

	fields := make([]logger.Field, 0, len(keysAndValues)/2)
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		fields = append(fields, logger.Any(key, keysAndValues[i+1]))
	}
	return fields
}
