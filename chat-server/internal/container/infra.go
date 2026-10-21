package container

import (
	"context"
	"fmt"

	"app/pkg/auth"
	"app/pkg/cache"
	"app/pkg/config"
	"app/pkg/database"
	"app/pkg/logger"
	"app/pkg/utils"
)

type closerFn func() error

type Infra struct {
	Cfg        *config.Config
	InstanceID string
	Log        *logger.Logger
	DB         *database.DB
	Cache      *cache.Cache
	Auth       *auth.Auth

	closers []closerFn
}

func NewInfra() (in *Infra, err error) {
	in = &Infra{}

	defer func() {
		if err != nil {
			in.Close()
			in = nil
		}
	}()

	cfg, e := config.Init()
	if e != nil {
		err = fmt.Errorf("配置初始化失败: %w", e)
		return
	}
	in.Cfg = cfg
	in.InstanceID = utils.InstanceID()

	log, e := logger.Init(&cfg.Log, in.InstanceID)
	if e != nil {
		err = fmt.Errorf("日志初始化失败: %w", e)
		return
	}
	in.Log = log
	in.addCloser(func() error {
		_ = log.Sync()
		return nil
	})
	log.Info(context.Background(), "实例身份已确定")

	db, e := database.Init(&cfg.DB, log)
	if e != nil {
		err = fmt.Errorf("数据库初始化失败: %w", e)
		return
	}
	in.DB = db
	in.addCloser(func() error { return db.Close() })

	c, e := cache.Init(&cfg.Cache)
	if e != nil {
		err = fmt.Errorf("缓存初始化失败: %w", e)
		return
	}
	in.Cache = c
	in.addCloser(func() error { return c.Close() })

	in.Auth = auth.NewAuth(&cfg.Auth)

	return in, nil
}

func (in *Infra) addCloser(fn closerFn) {
	in.closers = append(in.closers, fn)
}

func (in *Infra) Close() {
	if in == nil {
		return
	}

	for i := len(in.closers) - 1; i >= 0; i-- {
		if err := in.closers[i](); err != nil && in.Log != nil {
			in.Log.Error(context.Background(), "资源关闭失败", logger.Error(err))
		}
	}
	in.closers = nil
}
