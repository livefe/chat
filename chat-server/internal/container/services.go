package container

import (
	"context"
	"fmt"

	"app/internal/service"
	"app/pkg/push"
	"app/pkg/storage"
)

type Services struct {
	Auth          *service.AuthService
	SMS           *service.SMSService
	User          *service.UserService
	Employee      *service.EmployeeService
	Department    *service.DepartmentService
	Device        *service.DeviceService
	Permission    *service.PermissionService
	Role          *service.RoleService
	Media         *service.MediaService
	Presence      *service.PresenceService
	Room          *service.RoomService
	Message       *service.MessageService
	Conversation  *service.ConversationService
	Realtime      *service.RealtimeService
	Post          *service.PostService
	Config        *service.ConfigService
	Ad            *service.AdService
	Report        *service.ReportService
	SensitiveWord *service.SensitiveWordService
	Feedback      *service.FeedbackService
	IM            *service.IMService
	IMRealtime    *service.IMRealtimeService
	Support       *service.SupportService
	Region        *service.RegionService
	AuditLog      *service.AuditLogService
	AppRelease    *service.AppReleaseService
	Ops           *service.OpsService
}

func buildServices(in *Infra, r *Repos) (*Services, error) {
	smsSvc, err := service.NewSMSService(in.Log, in.Cfg.SMS, in.Cache, r.SMSRecord, r.Employee)
	if err != nil {
		return nil, fmt.Errorf("短信服务初始化失败: %w", err)
	}
	store, err := storage.New(&in.Cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("对象存储初始化失败: %w", err)
	}
	mediaSvc, err := service.NewMediaService(in.Log, store, &in.Cfg.Storage, r.MediaAsset, r.MediaUsage)
	if err != nil {
		return nil, fmt.Errorf("媒体服务初始化失败: %w", err)
	}

	sensitiveWordSvc := service.NewSensitiveWordService(in.Log, r.SensitiveWord)
	if err := sensitiveWordSvc.Load(context.Background()); err != nil {
		return nil, fmt.Errorf("敏感词加载失败: %w", err)
	}

	userStore := service.NewUserStore(r.User, in.Cache)
	authSvc := service.NewAuthService(
		in.Log, in.Cache, in.Auth, smsSvc,
		userStore, r.UserLoginRecord, r.Employee, r.EmployeeRole, r.EmployeeLoginRecord, r.Department,
		mediaSvc,
	)

	deviceSvc := service.NewDeviceService(in.Log, in.Cache, r.Device, push.BuildDispatchers(in.Cfg.Push.Xiaomi))
	presenceSvc := service.NewPresenceService(in.Log, in.Cache, userStore, in.InstanceID)
	chatHub, err := service.NewChatHub(in.Log, in.Cache, &in.Cfg.WebSocket, in.InstanceID, presenceSvc, deviceSvc)
	if err != nil {
		return nil, fmt.Errorf("Realtime 服务初始化失败: %w", err)
	}
	userSvc := service.NewUserService(
		in.Log, in.DB, authSvc, smsSvc,
		userStore, r.UserLoginRecord, r.ConversationDeletion, r.UserBlock, r.UserDeactivationLog, r.Message, r.Room,
		mediaSvc, sensitiveWordSvc, deviceSvc, chatHub,
	)

	permSvc := service.NewPermissionService(in.Log, in.Cache, in.DB, r.Permission, r.Role, r.EmployeeRole, r.RolePermission)
	roleSvc := service.NewRoleService(in.Log, in.DB, r.Role, r.EmployeeRole, r.RolePermission, r.Permission, permSvc)
	departmentSvc := service.NewDepartmentService(in.Log, r.Employee, r.Department)
	employeeSvc := service.NewEmployeeService(in.Log, authSvc, permSvc, r.Employee, r.EmployeeRole, r.Role, r.Department)

	configSvc := service.NewConfigService(in.Log, in.Cache, r.Config)
	adSvc := service.NewAdService(in.Log, in.DB, in.Cfg.Ad, configSvc, r.AdCampaign, r.AdEvent, r.AdDailyStat, r.Post, r.Room, mediaSvc)

	roomSvc := service.NewRoomService(in.Log, in.Cache, r.Room, presenceSvc, sensitiveWordSvc, chatHub)
	postSvc := service.NewPostService(
		in.Log, in.Cache, r.Post, r.PostExtra, r.PostSignup, r.Message, userSvc, roomSvc,
		mediaSvc, sensitiveWordSvc, adSvc, chatHub,
	)

	conversationSvc := service.NewConversationService(in.Log, in.Cache, r.Conversation, r.ConversationDeletion, r.ConversationPin, r.Message, userSvc)
	messageSvc := service.NewMessageService(in.Log, in.DB, r.Message, roomSvc, conversationSvc, userSvc, postSvc, mediaSvc, sensitiveWordSvc, chatHub)
	roomSvc.SetMessageReader(messageSvc)
	realtimeSvc := service.NewRealtimeService(chatHub, roomSvc, messageSvc, userSvc, postSvc)

	auditLogSvc := service.NewAuditLogService(in.Log, r.AuditLog, r.Employee, r.Permission)
	reportSvc := service.NewReportService(in.Log, r.Report, r.Employee, userSvc, messageSvc, postSvc, mediaSvc)
	feedbackSvc := service.NewFeedbackService(in.Log, r.Feedback, userStore, r.Employee, mediaSvc)
	regionSvc := service.NewRegionService(in.Log, r.Region)
	appReleaseSvc := service.NewAppReleaseService(in.Log, in.DB, r.AppRelease, r.AppUpdatePolicy, store)

	imHub, err := service.NewIMHub(in.Log, in.Cache, &in.Cfg.WebSocket, in.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("IM Realtime 服务初始化失败: %w", err)
	}
	imSvc := service.NewIMService(
		in.Log, in.Cache, in.DB, r.IMConversation, r.IMMember, r.IMMessage, r.IMStaffDMPair, r.Employee, r.Department,
		mediaSvc, sensitiveWordSvc, imHub,
	)
	imRealtimeSvc := service.NewIMRealtimeService(imHub, imSvc, permSvc, auditLogSvc)
	supportSvc := service.NewSupportService(
		in.Log, in.Cache, in.DB, r.IMSupportState, r.IMSupportConfig,
		userStore, r.Employee, imSvc, imHub, chatHub, permSvc,
	)

	opsSvc := service.NewOpsService(in.Log, permSvc, imSvc, supportSvc, reportSvc, feedbackSvc, roomSvc, postSvc)

	return &Services{
		Auth:          authSvc,
		SMS:           smsSvc,
		User:          userSvc,
		Employee:      employeeSvc,
		Department:    departmentSvc,
		Device:        deviceSvc,
		Permission:    permSvc,
		Role:          roleSvc,
		Media:         mediaSvc,
		Presence:      presenceSvc,
		Room:          roomSvc,
		Message:       messageSvc,
		Conversation:  conversationSvc,
		Realtime:      realtimeSvc,
		Post:          postSvc,
		Config:        configSvc,
		Ad:            adSvc,
		Report:        reportSvc,
		SensitiveWord: sensitiveWordSvc,
		Feedback:      feedbackSvc,
		IM:            imSvc,
		IMRealtime:    imRealtimeSvc,
		Support:       supportSvc,
		Region:        regionSvc,
		AuditLog:      auditLogSvc,
		AppRelease:    appReleaseSvc,
		Ops:           opsSvc,
	}, nil
}
