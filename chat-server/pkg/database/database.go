package database

import (
	"fmt"
	"strings"
	"time"

	"app/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parse_time"`
	Loc             string `mapstructure:"loc"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	SlowThresholdMS int    `mapstructure:"slow_threshold_ms"`
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.Driver) == "" {
		return fmt.Errorf("database.driver 未配置")
	}
	if strings.TrimSpace(c.Host) == "" {
		return fmt.Errorf("database.host 未配置")
	}
	if c.Port <= 0 {
		return fmt.Errorf("database.port 未配置")
	}
	if strings.TrimSpace(c.Username) == "" {
		return fmt.Errorf("database.username 未配置")
	}
	if c.Password == "" {
		return fmt.Errorf("database.password 未配置")
	}
	if strings.TrimSpace(c.DBName) == "" {
		return fmt.Errorf("database.dbname 未配置")
	}
	if strings.TrimSpace(c.Charset) == "" {
		return fmt.Errorf("database.charset 未配置")
	}
	if strings.TrimSpace(c.Loc) == "" {
		return fmt.Errorf("database.loc 未配置")
	}
	if c.MaxIdleConns <= 0 {
		return fmt.Errorf("database.max_idle_conns 未配置")
	}
	if c.MaxOpenConns <= 0 {
		return fmt.Errorf("database.max_open_conns 未配置")
	}
	if c.ConnMaxLifetime <= 0 {
		return fmt.Errorf("database.conn_max_lifetime 未配置")
	}
	if c.SlowThresholdMS <= 0 {
		return fmt.Errorf("database.slow_threshold_ms 未配置")
	}
	return nil
}

type DB struct {
	*gorm.DB
}

func Init(cfg *Config, appLogger *logger.Logger) (*DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.Charset,
		cfg.ParseTime,
		cfg.Loc,
	)

	gormConfig := &gorm.Config{
		Logger: newGormLogger(appLogger, cfg.SlowThresholdMS),
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	return &DB{db}, nil
}

func (db *DB) Close() error {
	if db == nil || db.DB == nil {
		return nil
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %w", err)
	}

	return nil
}
