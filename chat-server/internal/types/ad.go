package types

const (
	FeedItemTypePost      = "post"
	FeedItemTypeSponsored = "sponsored"
)

type SponsoredItem struct {
	CampaignID      uint    `json:"campaignId"`
	Title           string  `json:"title"`
	Subtitle        string  `json:"subtitle"`
	ImageURL        string  `json:"imageUrl"`
	AdvertiserName  string  `json:"advertiserName"`
	Badge           string  `json:"badge"`
	LandingType     string  `json:"landingType"`
	LandingValue    string  `json:"landingValue"`
	Distance        float64 `json:"distance"`
	ImpressionToken string  `json:"impressionToken"`
	RequestID       string  `json:"requestId"`
	SlotIndex       int     `json:"slotIndex"`
}

type FeedItem struct {
	Type      string         `json:"type"`
	Post      *PostItem      `json:"post,omitempty"`
	Sponsored *SponsoredItem `json:"sponsored,omitempty"`
}

type DiscoverFeedResponse struct {
	Total int64      `json:"total"`
	List  []FeedItem `json:"list"`
}

type AdEventItemRequest struct {
	RequestID       string  `json:"requestId" binding:"required,max=64"`
	CampaignID      uint    `json:"campaignId" binding:"required"`
	EventType       string  `json:"eventType" binding:"required,oneof=impression click"`
	ImpressionToken string  `json:"impressionToken" binding:"required"`
	SlotIndex       int     `json:"slotIndex"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
}

type AdEventsRequest struct {
	Events []AdEventItemRequest `json:"events" binding:"required,min=1,max=50,dive"`
}

type AdminAdCampaignSaveRequest struct {
	ID                   uint    `json:"id"`
	Name                 string  `json:"name" binding:"max=100"`
	AdvertiserName       string  `json:"advertiserName" binding:"required,max=80"`
	Title                string  `json:"title" binding:"required,max=80"`
	Subtitle             string  `json:"subtitle" binding:"max=120"`
	ImageMediaID         uint    `json:"imageMediaId" binding:"required"`
	LandingType          string  `json:"landingType" binding:"required,oneof=url post room"`
	LandingValue         string  `json:"landingValue" binding:"required,max=500"`
	Priority             int     `json:"priority" binding:"required,min=1"`
	StartAt              int64   `json:"startAt" binding:"required"`
	EndAt                int64   `json:"endAt" binding:"required"`
	CPDPrice             float64 `json:"cpdPrice"`
	HasGeo               bool    `json:"hasGeo"`
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	RadiusM              float64 `json:"radiusM"`
	DailyImpressionLimit uint    `json:"dailyImpressionLimit"`
	SubmitAudit          bool    `json:"submitAudit"`
}

type AdminAdCampaignListRequest struct {
	Status   *uint8 `form:"status"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type AdminAdCampaignItem struct {
	ID                   uint    `json:"id"`
	Name                 string  `json:"name"`
	AdvertiserName       string  `json:"advertiserName"`
	Title                string  `json:"title"`
	Subtitle             string  `json:"subtitle"`
	ImageMediaID         uint    `json:"imageMediaId"`
	ImageURL             string  `json:"imageUrl"`
	LandingType          string  `json:"landingType"`
	LandingValue         string  `json:"landingValue"`
	Status               uint8   `json:"status"`
	StatusText           string  `json:"statusText"`
	Priority             int     `json:"priority"`
	StartAt              int64   `json:"startAt"`
	EndAt                int64   `json:"endAt"`
	CPDPrice             float64 `json:"cpdPrice"`
	HasGeo               bool    `json:"hasGeo"`
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	RadiusM              float64 `json:"radiusM"`
	DailyImpressionLimit uint    `json:"dailyImpressionLimit"`
	RejectRemark         string  `json:"rejectRemark"`
	LandingInvalid       bool    `json:"landingInvalid"`
	TodayImpressions     int64   `json:"todayImpressions"`
	TodayClicks          int64   `json:"todayClicks"`
	TodayCTR             float64 `json:"todayCtr"`
	LowCTRWarning        bool    `json:"lowCtrWarning"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
}

type AdminAdCampaignListResponse struct {
	Total int64                 `json:"total"`
	List  []AdminAdCampaignItem `json:"list"`
}

type AdminAdUpdateStatusRequest struct {
	ID           uint   `json:"id" binding:"required"`
	Status       uint8  `json:"status" binding:"required"`
	RejectRemark string `json:"rejectRemark" binding:"max=200"`
}

type AdminAdBatchPauseRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=100"`
}

type AdminAdCopyRequest struct {
	ID uint `json:"id" binding:"required"`
}

type AdminAdDailyStatRequest struct {
	CampaignID uint   `form:"campaignId" binding:"required"`
	FromDate   string `form:"fromDate"`
	ToDate     string `form:"toDate"`
}

type AdminAdDailyStatItem struct {
	StatDate    string  `json:"statDate"`
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	CTR         float64 `json:"ctr"`
}

type AdminAdDailyStatResponse struct {
	List []AdminAdDailyStatItem `json:"list"`
}

type AdminAdSetting struct {
	Enabled                bool    `json:"enabled"`
	FirstSlotAfter         int     `json:"firstSlotAfter"`
	SlotInterval           int     `json:"slotInterval"`
	MaxActivePerAdvertiser int     `json:"maxActivePerAdvertiser"`
	LowCTRThreshold        float64 `json:"lowCtrThreshold"`
}
