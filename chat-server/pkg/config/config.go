package config

import (
	"fmt"
	"os"
	"strings"

	"app/pkg/auth"
	"app/pkg/cache"
	"app/pkg/database"
	"app/pkg/logger"
	"app/pkg/push"
	"app/pkg/sms"
	"app/pkg/storage"
	"app/pkg/websocket"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Env       string           `mapstructure:"env"`
	Server    Server           `mapstructure:"server"`
	Log       logger.Config    `mapstructure:"log"`
	DB        database.Config  `mapstructure:"database"`
	Cache     cache.Config     `mapstructure:"cache"`
	SMS       sms.Config       `mapstructure:"sms"`
	Auth      auth.Config      `mapstructure:"auth"`
	WebSocket websocket.Config `mapstructure:"websocket"`
	Storage   storage.Config   `mapstructure:"storage"`
	Media     MediaConfig      `mapstructure:"media"`
	Push      PushConfig       `mapstructure:"push"`
	Ad        AdConfig         `mapstructure:"ad"`
}

const (
	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"
)

func validateEnv(env string) error {
	switch env {
	case EnvDev, EnvTest, EnvProd:
		return nil
	case "":
		return fmt.Errorf("env 未配置")
	default:
		return fmt.Errorf("env 无效: %s（允许 dev / test / prod）", env)
	}
}

type Server struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

func (s *Server) Validate() error {
	if strings.TrimSpace(s.Host) == "" {
		return fmt.Errorf("server.host 未配置")
	}
	if s.Port <= 0 {
		return fmt.Errorf("server.port 未配置")
	}
	if s.ReadTimeout <= 0 {
		return fmt.Errorf("server.read_timeout 未配置")
	}
	if s.WriteTimeout <= 0 {
		return fmt.Errorf("server.write_timeout 未配置")
	}
	if s.IdleTimeout <= 0 {
		return fmt.Errorf("server.idle_timeout 未配置")
	}
	return nil
}

type PushConfig struct {
	Xiaomi push.XiaomiConfig `mapstructure:"xiaomi"`
}

type MediaConfig struct {
	MaxImageMB     int    `mapstructure:"max_image_mb"`
	MaxVideoMB     int    `mapstructure:"max_video_mb"`
	MaxAudioMB     int    `mapstructure:"max_audio_mb"`
	MaxFileMB      int    `mapstructure:"max_file_mb"`
	PathPrefix     string `mapstructure:"path_prefix"`
	OrphanTTLHours int    `mapstructure:"orphan_ttl_hours"`
}

func (m *MediaConfig) Validate() error {
	if m.MaxImageMB <= 0 {
		return fmt.Errorf("media.max_image_mb 未配置")
	}
	if m.MaxVideoMB <= 0 {
		return fmt.Errorf("media.max_video_mb 未配置")
	}
	if m.MaxAudioMB <= 0 {
		return fmt.Errorf("media.max_audio_mb 未配置")
	}
	if m.MaxFileMB <= 0 {
		return fmt.Errorf("media.max_file_mb 未配置")
	}
	prefix := strings.Trim(strings.TrimSpace(m.PathPrefix), "/")
	if prefix == "" || strings.Contains(prefix, "..") {
		return fmt.Errorf("media.path_prefix 未配置或非法")
	}
	if m.OrphanTTLHours <= 0 {
		return fmt.Errorf("media.orphan_ttl_hours 未配置")
	}
	return nil
}

type AdConfig struct {
	TokenSecret     string `mapstructure:"token_secret"`
	TokenTTLSeconds int64  `mapstructure:"token_ttl_seconds"`
}

func (a *AdConfig) Validate() error {
	if strings.TrimSpace(a.TokenSecret) == "" {
		return fmt.Errorf("ad.token_secret 未配置")
	}
	if a.TokenTTLSeconds <= 0 {
		return fmt.Errorf("ad.token_ttl_seconds 未配置")
	}
	return nil
}

func Init() (*Config, error) {
	envFile := "./configs/.env"
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("加载环境变量文件失败: %w", err)
		}
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取 config.yaml 失败: %w", err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析 config.yaml 失败: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("配置为空")
	}

	var errs []string
	appendErr := func(err error) {
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	appendErr(validateEnv(c.Env))
	appendErr(c.Server.Validate())
	appendErr(c.Log.Validate())
	appendErr(c.DB.Validate())
	appendErr(c.Cache.Validate())
	appendErr(c.Auth.Validate())
	appendErr(c.WebSocket.Validate())
	appendErr(c.SMS.Validate())
	appendErr(c.Storage.Validate())
	appendErr(c.Media.Validate())
	appendErr(c.Push.Xiaomi.Validate())
	appendErr(c.Ad.Validate())

	if len(errs) > 0 {
		return fmt.Errorf("配置校验失败:\n- %s", strings.Join(errs, "\n- "))
	}
	return nil
}
