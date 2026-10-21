package model

const (
	AdStatusDraft    uint8 = 0
	AdStatusPending  uint8 = 1
	AdStatusActive   uint8 = 2
	AdStatusRejected uint8 = 3
	AdStatusPaused   uint8 = 4
	AdStatusEnded    uint8 = 5

	AdLandingURL  = "url"
	AdLandingPost = "post"
	AdLandingRoom = "room"
)

type AdCampaign struct {
	Base
	Name                 string  `gorm:"type:varchar(100);not null;default:'';comment:内部名称" json:"name"`
	AdvertiserName       string  `gorm:"type:varchar(80);not null;index;comment:商家展示名/广告主名" json:"advertiserName"`
	Title                string  `gorm:"type:varchar(80);not null;comment:卡片标题" json:"title"`
	Subtitle             string  `gorm:"type:varchar(120);not null;default:'';comment:副文案" json:"subtitle"`
	ImageMediaID         uint    `gorm:"type:int unsigned;not null;default:0;comment:封面媒体ID" json:"imageMediaId"`
	LandingType          string  `gorm:"type:varchar(16);not null;default:'url';comment:落地类型 url|post|room" json:"landingType"`
	LandingValue         string  `gorm:"type:varchar(500);not null;default:'';comment:落地值" json:"landingValue"`
	Status               uint8   `gorm:"type:tinyint unsigned;not null;default:0;index;comment:0草稿 1待审 2投放中 3驳回 4暂停 5结束" json:"status"`
	Priority             int     `gorm:"type:int;not null;default:100;index;comment:优先级(越大越易出)" json:"priority"`
	StartAt              int64   `gorm:"type:bigint;not null;default:0;index;comment:开始时间(秒)" json:"startAt"`
	EndAt                int64   `gorm:"type:bigint;not null;default:0;index;comment:结束时间(秒)" json:"endAt"`
	CPDPrice             float64 `gorm:"type:decimal(10,2);not null;default:0;comment:日价格备忘" json:"cpdPrice"`
	HasGeo               bool    `gorm:"type:tinyint(1);not null;default:0;index;comment:是否地理定向" json:"hasGeo"`
	Latitude             float64 `gorm:"type:decimal(10,6);not null;default:0;comment:圆心纬度" json:"latitude"`
	Longitude            float64 `gorm:"type:decimal(10,6);not null;default:0;comment:圆心经度" json:"longitude"`
	RadiusM              float64 `gorm:"type:decimal(12,2);not null;default:0;comment:半径米" json:"radiusM"`
	DailyImpressionLimit uint    `gorm:"type:int unsigned;not null;default:0;comment:日曝光上限(0不限)" json:"dailyImpressionLimit"`
	RejectRemark         string  `gorm:"type:varchar(200);not null;default:'';comment:驳回原因" json:"rejectRemark"`
	LandingInvalid       bool    `gorm:"type:tinyint(1);not null;default:0;comment:落地失效标记" json:"landingInvalid"`
}

func (AdCampaign) TableName() string { return "ad_campaign" }
