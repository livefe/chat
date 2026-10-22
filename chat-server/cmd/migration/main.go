package main

import (
	"context"
	"fmt"
	"os"

	"app/internal/container"
	"app/internal/model"
	"app/pkg/logger"
)

func main() {
	in, err := container.NewInfra()
	if err != nil {
		fmt.Fprintf(os.Stderr, "基础设施初始化失败: %v\n", err)
		os.Exit(1)
	}
	os.Exit(migrate(in))
}

func migrate(in *container.Infra) int {
	defer in.Close()
	ctx := context.Background()
	log := in.Log

	log.Info(ctx, "开始执行数据库迁移")

	models := []any{
		&model.User{},
		&model.Employee{},
		&model.Department{},
		&model.SMSRecord{},
		&model.UserLoginRecord{},
		&model.EmployeeLoginRecord{},
		&model.Role{},
		&model.Permission{},
		&model.EmployeeRole{},
		&model.RolePermission{},
		&model.Room{},
		&model.Message{},
		&model.Conversation{},
		&model.ConversationDeletion{},
		&model.ConversationPin{},
		&model.UserBlock{},
		&model.UserBlockEvent{},
		&model.SensitiveWord{},
		&model.Feedback{},
		&model.Report{},
		&model.UserDeactivationLog{},
		&model.UserPunishment{},
		&model.AuditLog{},
		&model.Post{},
		&model.PostExtra{},
		&model.PostSignup{},
		&model.Region{},
		&model.Device{},
		&model.AppRelease{},
		&model.AppUpdatePolicy{},
		&model.MediaAsset{},
		&model.MediaUsage{},
		&model.AdCampaign{},
		&model.AdEvent{},
		&model.AdDailyStat{},
		&model.Config{},
		&model.IMConversation{},
		&model.IMMember{},
		&model.IMMessage{},
		&model.IMStaffDMPair{},
		&model.IMSupportState{},
		&model.IMSupportConfig{},
	}

	for _, m := range models {
		table := fmt.Sprintf("%T", m)
		if err := in.DB.AutoMigrate(m); err != nil {
			log.Error(ctx, "迁移表失败", logger.String("model", table), logger.Error(err))
			return 1
		}
		log.Info(ctx, "迁移表完成", logger.String("model", table))
	}

	log.Info(ctx, "数据库迁移完成")
	return 0
}
