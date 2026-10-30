package service

import (
	"context"

	"app/internal/model"
	"app/internal/repository"
	"app/pkg/cache"
	"app/pkg/logger"
)

type PostService struct {
	log          *logger.Logger
	cache        *cache.Cache
	postRepo     repository.PostRepository
	extrasRepo   repository.PostExtraRepository
	signupRepo   repository.PostSignupRepository
	msgRepo      repository.MessageRepository
	employeeRepo repository.EmployeeRepository
	userSvc      *UserService
	roomSvc      *RoomService
	sensitive    *SensitiveWordService
	mediaSvc     *MediaService
	adSvc        *AdService
	channels     postChannelNotifier
}

type postChannelNotifier interface {
	ClosePost(ctx context.Context, postID uint, reason string)
}

func NewPostService(
	log *logger.Logger,
	c *cache.Cache,
	postRepo repository.PostRepository,
	extrasRepo repository.PostExtraRepository,
	signupRepo repository.PostSignupRepository,
	msgRepo repository.MessageRepository,
	employeeRepo repository.EmployeeRepository,
	userSvc *UserService,
	roomSvc *RoomService,
	mediaSvc *MediaService,
	sensitive *SensitiveWordService,
	adSvc *AdService,
	channels postChannelNotifier,
) *PostService {
	return &PostService{
		log:          log,
		cache:        c,
		postRepo:     postRepo,
		extrasRepo:   extrasRepo,
		signupRepo:   signupRepo,
		msgRepo:      msgRepo,
		employeeRepo: employeeRepo,
		userSvc:      userSvc,
		roomSvc:      roomSvc,
		mediaSvc:     mediaSvc,
		sensitive:    sensitive,
		adSvc:        adSvc,
		channels:     channels,
	}
}

var postStatusText = map[uint8]string{
	model.PostStatusPending:   "待审核",
	model.PostStatusPublished: "发布中",
	model.PostStatusRejected:  "驳回",
	model.PostStatusDisabled:  "下架",
}

func postPublished(p *model.Post) bool {
	return p != nil && p.Status == model.PostStatusPublished
}

func ownerAuditRemark(p *model.Post) string {
	if p.Status == model.PostStatusRejected || p.Status == model.PostStatusDisabled {
		return p.AuditRemark
	}
	return ""
}

func (s *PostService) closePostChannel(ctx context.Context, postID uint, reason string) {
	if postID == 0 {
		return
	}
	s.channels.ClosePost(ctx, postID, reason)
}

func (s *PostService) CountByStatus(ctx context.Context, status uint8) (int64, error) {
	return s.postRepo.CountByStatus(ctx, status)
}
