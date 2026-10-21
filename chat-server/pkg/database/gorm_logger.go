package database

import (
	"context"
	"errors"
	"time"

	"app/pkg/logger"

	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type gormLoggerImpl struct {
	logger        *logger.Logger
	level         gormLogger.LogLevel
	slowThreshold time.Duration
}

func newGormLogger(logger *logger.Logger, slowThresholdMS int) gormLogger.Interface {
	return &gormLoggerImpl{
		logger:        logger,
		level:         gormLogger.Info,
		slowThreshold: time.Duration(slowThresholdMS) * time.Millisecond,
	}
}

func (l *gormLoggerImpl) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

func (l *gormLoggerImpl) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logger != nil && l.level >= gormLogger.Info {
		l.logger.Info(ctx, "GORM: "+msg, logger.Any("data", data))
	}
}

func (l *gormLoggerImpl) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logger != nil && l.level >= gormLogger.Warn {
		l.logger.Warn(ctx, "GORM: "+msg, logger.Any("data", data))
	}
}

func (l *gormLoggerImpl) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logger != nil && l.level >= gormLogger.Error {
		l.logger.Error(ctx, "GORM: "+msg, logger.Any("data", data))
	}
}

func (l *gormLoggerImpl) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logger == nil || l.level == gormLogger.Silent {
		return
	}

	sql, rows := fc()
	duration := time.Since(begin)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if l.level < gormLogger.Warn {
				return
			}
			l.logger.Warn(ctx, "SQL执行",
				logger.String("sql", sql),
				logger.Int64("rows", rows),
				logger.Duration("duration", duration))
		} else {
			if l.level < gormLogger.Error {
				return
			}
			l.logger.Error(ctx, "SQL执行错误",
				logger.String("sql", sql),
				logger.Int64("rows", rows),
				logger.Duration("duration", duration),
				logger.Error(err))
		}
	} else {
		if duration >= l.slowThreshold {
			if l.level >= gormLogger.Warn {
				l.logger.Warn(ctx, "SQL慢查询",
					logger.String("sql", sql),
					logger.Int64("rows", rows),
					logger.Duration("duration", duration),
					logger.Duration("slow_threshold", l.slowThreshold))
			}
			return
		}

		if l.level < gormLogger.Info {
			return
		}
		l.logger.Info(ctx, "SQL执行",
			logger.String("sql", sql),
			logger.Int64("rows", rows),
			logger.Duration("duration", duration))
	}
}
