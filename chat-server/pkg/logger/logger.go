package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	TraceIDKey = "request_id"
)

type Config struct {
	Level         string `mapstructure:"level"`
	Encoding      string `mapstructure:"encoding"`
	TimeFormat    string `mapstructure:"time_format"`
	OutputPath    string `mapstructure:"output_path"`
	Filename      string `mapstructure:"filename"`
	MaxSize       int    `mapstructure:"max_size"`
	MaxAge        int    `mapstructure:"max_age"`
	MaxBackups    int    `mapstructure:"max_backups"`
	Compress      bool   `mapstructure:"compress"`
	ConsoleOutput bool   `mapstructure:"console_output"`
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.Level) == "" {
		return fmt.Errorf("log.level 未配置")
	}
	if strings.TrimSpace(c.Encoding) == "" {
		return fmt.Errorf("log.encoding 未配置")
	}
	if strings.TrimSpace(c.TimeFormat) == "" {
		return fmt.Errorf("log.time_format 未配置")
	}
	if !c.ConsoleOutput {
		if strings.TrimSpace(c.OutputPath) == "" {
			return fmt.Errorf("log.output_path 未配置（或启用 log.console_output）")
		}
		if strings.TrimSpace(c.Filename) == "" {
			return fmt.Errorf("log.filename 未配置")
		}
	}
	if c.MaxSize <= 0 {
		return fmt.Errorf("log.max_size 未配置")
	}
	if c.MaxAge <= 0 {
		return fmt.Errorf("log.max_age 未配置")
	}
	if c.MaxBackups <= 0 {
		return fmt.Errorf("log.max_backups 未配置")
	}
	return nil
}

type Logger struct {
	*zap.Logger
}

type Field = zap.Field

var (
	String   = zap.String
	Int      = zap.Int
	Uint     = zap.Uint
	Int64    = zap.Int64
	Float64  = zap.Float64
	Bool     = zap.Bool
	Error    = zap.Error
	Any      = zap.Any
	Duration = zap.Duration
)

func Init(cfg *Config, instanceID string) (*Logger, error) {
	zapConfig := zap.NewProductionConfig()

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("log.level 无效: %w", err)
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	zapConfig.Encoding = cfg.Encoding

	zapConfig.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
	}
	timeFormat := cfg.TimeFormat
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(timeFormat)

	cores := []zapcore.Core{}

	if cfg.ConsoleOutput {
		var consoleEncoder zapcore.Encoder
		if cfg.Encoding == "console" {
			consoleEncoder = zapcore.NewConsoleEncoder(zapConfig.EncoderConfig)
		} else {
			consoleEncoder = zapcore.NewJSONEncoder(zapConfig.EncoderConfig)
		}
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapConfig.Level)
		cores = append(cores, consoleCore)
	}

	if cfg.OutputPath != "" {
		if err := os.MkdirAll(cfg.OutputPath, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %w", err)
		}

		filePath := filepath.Join(cfg.OutputPath, cfg.Filename)
		fileWriter := &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			Compress:   cfg.Compress,
		}

		var fileEncoder zapcore.Encoder
		if cfg.Encoding == "console" {
			fileEncoder = zapcore.NewConsoleEncoder(zapConfig.EncoderConfig)
		} else {
			fileEncoder = zapcore.NewJSONEncoder(zapConfig.EncoderConfig)
		}
		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), zapConfig.Level)
		cores = append(cores, fileCore)
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("log 未配置任何输出（console 或文件）")
	}

	core := zapcore.NewTee(cores...)
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if id := strings.TrimSpace(instanceID); id != "" {
		zapLogger = zapLogger.With(zap.String("instance_id", id))
	}

	return &Logger{zapLogger}, nil
}

func (l *Logger) withContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return l.Logger
	}

	val := ctx.Value(TraceIDKey)
	if valStr, ok := val.(string); ok && valStr != "" {
		return l.Logger.With(zap.String(TraceIDKey, valStr))
	}

	return l.Logger
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.withContext(ctx).Debug(msg, fields...)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...Field) {
	l.withContext(ctx).Info(msg, fields...)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.withContext(ctx).Warn(msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...Field) {
	l.withContext(ctx).Error(msg, fields...)
}
