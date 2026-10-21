package model

type AdDailyStat struct {
	Base
	CampaignID  uint   `gorm:"type:int unsigned;not null;uniqueIndex:uk_ad_daily,priority:1;comment:计划ID" json:"campaignId"`
	StatDate    string `gorm:"type:char(10);not null;uniqueIndex:uk_ad_daily,priority:2;index;comment:日期YYYY-MM-DD" json:"statDate"`
	Impressions int64  `gorm:"type:bigint;not null;default:0;comment:曝光" json:"impressions"`
	Clicks      int64  `gorm:"type:bigint;not null;default:0;comment:点击" json:"clicks"`
}

func (AdDailyStat) TableName() string { return "ad_daily_stat" }
