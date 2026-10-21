package container

import (
	"app/internal/handler"
	"app/internal/middleware"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Auth          *handler.AuthHandler
	User          *handler.UserHandler
	Employee      *handler.EmployeeHandler
	Department    *handler.DepartmentHandler
	Device        *handler.DeviceHandler
	Permission    *handler.PermissionHandler
	Role          *handler.RoleHandler
	Media         *handler.MediaHandler
	Room          *handler.RoomHandler
	Message       *handler.MessageHandler
	Conversation  *handler.ConversationHandler
	Realtime      *handler.RealtimeHandler
	Post          *handler.PostHandler
	Ad            *handler.AdHandler
	Report        *handler.ReportHandler
	SensitiveWord *handler.SensitiveWordHandler
	Feedback      *handler.FeedbackHandler
	IM            *handler.IMHandler
	Support       *handler.SupportHandler
	Region        *handler.RegionHandler
	AuditLog      *handler.AuditLogHandler
	AppRelease    *handler.AppReleaseHandler
	Ops           *handler.OpsHandler

	UserAuth          gin.HandlerFunc
	EmployeeAuth      gin.HandlerFunc
	RequirePermission gin.HandlerFunc
	Recovery          gin.HandlerFunc
	ClientContext     gin.HandlerFunc
	Logger            gin.HandlerFunc
	AuditMiddleware   gin.HandlerFunc
}

func buildHandlers(in *Infra, s *Services) *Handlers {
	return &Handlers{
		Auth:          handler.NewAuthHandler(in.Log, s.Auth, s.SMS),
		User:          handler.NewUserHandler(in.Log, s.User),
		Employee:      handler.NewEmployeeHandler(in.Log, s.Employee),
		Department:    handler.NewDepartmentHandler(in.Log, s.Department),
		Device:        handler.NewDeviceHandler(in.Log, s.Device),
		Permission:    handler.NewPermissionHandler(in.Log, s.Permission),
		Role:          handler.NewRoleHandler(in.Log, s.Role),
		Media:         handler.NewMediaHandler(in.Log, s.Media),
		Room:          handler.NewRoomHandler(in.Log, s.Room),
		Message:       handler.NewMessageHandler(in.Log, s.Message),
		Conversation:  handler.NewConversationHandler(in.Log, s.Conversation),
		Realtime:      handler.NewRealtimeHandler(in.Log, in.Auth, s.Auth, s.Realtime, in.Cfg.WebSocket),
		Post:          handler.NewPostHandler(in.Log, s.Post),
		Ad:            handler.NewAdHandler(in.Log, s.Ad),
		Report:        handler.NewReportHandler(in.Log, s.Report),
		SensitiveWord: handler.NewSensitiveWordHandler(in.Log, s.SensitiveWord),
		Feedback:      handler.NewFeedbackHandler(in.Log, s.Feedback),
		IM:            handler.NewIMHandler(in.Log, in.Auth, s.Auth, s.Employee, s.IM, s.IMRealtime, in.Cfg.WebSocket),
		Support:       handler.NewSupportHandler(in.Log, s.Support),
		Region:        handler.NewRegionHandler(in.Log, s.Region),
		AuditLog:      handler.NewAuditLogHandler(in.Log, s.AuditLog),
		AppRelease:    handler.NewAppReleaseHandler(in.Log, s.AppRelease),
		Ops:           handler.NewOpsHandler(in.Log, s.Ops),

		UserAuth:          middleware.UserAuth(in.Log, in.Auth, s.Auth, s.User, response.WriteError),
		EmployeeAuth:      middleware.EmployeeAuth(in.Log, in.Auth, s.Auth, s.Employee, response.WriteError),
		RequirePermission: middleware.Permission(s.Permission),
		Recovery:          middleware.Recovery(in.Log),
		ClientContext:     middleware.ClientContext(),
		Logger:            middleware.Logger(in.Log),
		AuditMiddleware:   middleware.AuditMiddleware(in.Log, s.AuditLog),
	}
}
