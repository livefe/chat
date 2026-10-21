package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"app/internal/container"
	"app/pkg/logger"
)

const shutdownTimeout = 30 * time.Second

func main() {
	once := flag.String("once", "", "立即执行指定定时任务后退出，如 region.sync")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	c, err := container.NewWorker(ctx)
	if err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "Worker 初始化失败: %v\n", err)
		os.Exit(1)
	}

	if *once != "" {
		cancel()
		os.Exit(runOnce(c, *once))
	}
	os.Exit(runScheduler(c, cancel))
}

func runOnce(c *container.WorkerContainer, name string) int {
	defer c.Close()
	if err := c.Scheduler.RunOnce(context.Background(), name); err != nil {
		return 1
	}
	return 0
}

func runScheduler(c *container.WorkerContainer, cancel context.CancelFunc) int {
	defer c.Close()
	log := c.Infra.Log
	c.Scheduler.Start()
	log.Info(context.Background(), "Worker 已启动")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info(context.Background(), "Worker 正在停止")
	cancel()

	ctx, stopCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer stopCancel()
	err := c.Scheduler.Stop(ctx)
	if snap := c.Scheduler.Metrics().Snapshot(); len(snap) > 0 {
		log.Info(context.Background(), "定时任务运行统计", logger.Any("jobs", snap))
	}
	if err != nil {
		log.Error(context.Background(), "Worker 停止超时或失败", logger.Error(err))
		return 1
	}
	log.Info(context.Background(), "Worker 已停止")
	return 0
}
