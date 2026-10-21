package model

const (
	AdEventImpression = "impression"
	AdEventClick      = "click"
)

type AdEvent struct {
	Base
	RequestID  string  `gorm:"type:varchar(64);not null;uniqueIndex:uk_ad_event_req_type,priority:1;comment:请求幂等ID" json:"requestId"`
	EventType  string  `gorm:"type:varchar(16);not null;uniqueIndex:uk_ad_event_req_type,priority:2;index;comment:impression|click" json:"eventType"`
	CampaignID uint    `gorm:"type:int unsigned;not null;index;comment:计划ID" json:"campaignId"`
	UserID     uint    `gorm:"type:int unsigned;not null;default:0;index;comment:用户ID" json:"userId"`
	SlotIndex  int     `gorm:"type:int;not null;default:0;comment:Feed线性位置" json:"slotIndex"`
	Latitude   float64 `gorm:"type:decimal(10,6);not null;default:0;comment:上报纬度" json:"latitude"`
	Longitude  float64 `gorm:"type:decimal(10,6);not null;default:0;comment:上报经度" json:"longitude"`
	Billable   bool    `gorm:"type:tinyint(1);not null;default:1;comment:是否计费口径" json:"billable"`
}

func (AdEvent) TableName() string { return "ad_event" }
