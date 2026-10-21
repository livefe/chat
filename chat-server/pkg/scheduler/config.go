package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Timezone string        `mapstructure:"timezone"`
	LockTTL  time.Duration `mapstructure:"lock_ttl"`
	Jobs     []JobConfig   `mapstructure:"jobs"`
}

type JobConfig struct {
	Name    string        `mapstructure:"name"`
	Spec    string        `mapstructure:"spec"`
	Desc    string        `mapstructure:"desc"`
	Timeout time.Duration `mapstructure:"timeout"`
	Enabled bool          `mapstructure:"enabled"`
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("scheduler.yaml 未配置")
	}
	if strings.TrimSpace(c.Timezone) == "" {
		return fmt.Errorf("timezone 未配置")
	}
	if _, err := time.LoadLocation(strings.TrimSpace(c.Timezone)); err != nil {
		return fmt.Errorf("timezone 无效: %w", err)
	}
	if c.LockTTL <= 0 {
		return fmt.Errorf("lock_ttl 未配置")
	}

	seen := make(map[string]struct{}, len(c.Jobs))
	for i, job := range c.Jobs {
		prefix := fmt.Sprintf("jobs[%d]", i)
		name := strings.TrimSpace(job.Name)
		if name == "" {
			return fmt.Errorf("%s.name 未配置", prefix)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("%s.name 重复: %s", prefix, name)
		}
		seen[name] = struct{}{}
		if !job.Enabled {
			continue
		}
		if strings.TrimSpace(job.Spec) == "" {
			return fmt.Errorf("%s.spec 未配置", prefix)
		}
		if job.Timeout <= 0 {
			return fmt.Errorf("%s.timeout 未配置", prefix)
		}
	}
	return nil
}

func loadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("scheduler")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.SetEnvPrefix("APP_SCHEDULER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取 scheduler.yaml 失败: %w", err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析 scheduler.yaml 失败: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
