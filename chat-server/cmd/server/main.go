package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"app/internal/container"
	"app/internal/router"
	"app/pkg/config"
	"app/pkg/logger"
	"app/pkg/utils"

	"github.com/gin-gonic/gin"
)

const shutdownTimeout = 10 * time.Second

func main() {
	c, err := container.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "应用初始化失败: %v\n", err)
		os.Exit(1)
	}
	os.Exit(serve(c))
}

func serve(c *container.Container) int {
	defer c.Close()
	log := c.Infra.Log

	server, serveErr := runHTTPServer(c)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serveErr:
		log.Error(context.Background(), "服务器异常退出", logger.Error(err))
		return 1
	case <-quit:
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error(ctx, "服务器优雅关闭失败", logger.Error(err))
		return 1
	}
	return 0
}

func runHTTPServer(c *container.Container) (*http.Server, <-chan error) {
	cfg, log := c.Infra.Cfg, c.Infra.Log

	if cfg.Env == config.EnvProd {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	utils.InitCustomValidators()

	engine := gin.New()
	router.Register(engine, c.Handlers)

	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	log.Info(context.Background(), "服务器启动", logger.String("address", server.Addr))
	serveErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	return server, serveErr
}
