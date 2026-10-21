package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/pkg/config"
	"app/pkg/database"
	"app/pkg/logger"
)

type AdService struct {
	log             *logger.Logger
	db              *database.DB
	tokenSecret     string
	tokenTTLSeconds int64
	configSvc       *ConfigService
	campaignRepo    repository.AdCampaignRepository
	eventRepo       repository.AdEventRepository
	statRepo        repository.AdDailyStatRepository
	postRepo        repository.PostRepository
	roomRepo        repository.RoomRepository
	mediaSvc        *MediaService
}

func NewAdService(
	log *logger.Logger,
	db *database.DB,
	cfg config.AdConfig,
	configSvc *ConfigService,
	campaignRepo repository.AdCampaignRepository,
	eventRepo repository.AdEventRepository,
	statRepo repository.AdDailyStatRepository,
	postRepo repository.PostRepository,
	roomRepo repository.RoomRepository,
	mediaSvc *MediaService,
) *AdService {
	return &AdService{
		log:             log,
		db:              db,
		tokenSecret:     cfg.TokenSecret,
		tokenTTLSeconds: cfg.TokenTTLSeconds,
		configSvc:       configSvc,
		campaignRepo:    campaignRepo,
		eventRepo:       eventRepo,
		statRepo:        statRepo,
		postRepo:        postRepo,
		roomRepo:        roomRepo,
		mediaSvc:        mediaSvc,
	}
}

func (s *AdService) landingLooksOK(ctx context.Context, c *model.AdCampaign) bool {
	switch c.LandingType {
	case model.AdLandingURL:
		v := strings.TrimSpace(c.LandingValue)
		return strings.HasPrefix(v, "https://") || strings.HasPrefix(v, "http://")
	case model.AdLandingPost:
		id, err := strconv.ParseUint(strings.TrimSpace(c.LandingValue), 10, 64)
		if err != nil || id == 0 {
			return false
		}
		p, err := s.postRepo.GetByID(ctx, uint(id))
		return err == nil && postPublished(p)
	case model.AdLandingRoom:
		id, err := strconv.ParseUint(strings.TrimSpace(c.LandingValue), 10, 64)
		if err != nil || id == 0 {
			return false
		}
		room, err := s.roomRepo.GetByID(ctx, uint(id))
		return err == nil && roomEnabled(room)
	default:
		return false
	}
}

func (s *AdService) signToken(campaignID, userID uint, requestID string, exp int64) string {
	payload := fmt.Sprintf("%d:%d:%s:%d", campaignID, userID, requestID, exp)
	mac := hmac.New(sha256.New, []byte(s.tokenSecret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil)) + "." + strconv.FormatInt(exp, 10)
}

func (s *AdService) verifyToken(campaignID, userID uint, requestID, token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	exp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || exp < time.Now().Unix() {
		return false
	}
	expect := s.signToken(campaignID, userID, requestID, exp)
	return hmac.Equal([]byte(expect), []byte(token))
}

func newAdRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
