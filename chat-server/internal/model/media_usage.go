package model

const (
	MediaBizAvatar     = "avatar"
	MediaBizMessage    = "message"
	MediaBizPost       = "post"
	MediaBizFeedback   = "feedback"
	MediaBizReport     = "report"
	MediaBizAdCampaign = "ad_campaign"
	MediaBizIMMessage  = "im_message"
)

type MediaUsage struct {
	Base

	MediaID uint   `gorm:"type:int unsigned;not null;uniqueIndex:uk_media_biz,priority:3;index;comment:媒体ID" json:"mediaId"`
	BizType string `gorm:"type:varchar(32);not null;uniqueIndex:uk_media_biz,priority:1;comment:业务类型" json:"bizType"`
	BizID   uint   `gorm:"type:int unsigned;not null;uniqueIndex:uk_media_biz,priority:2;comment:业务主键" json:"bizId"`
	UserID  uint   `gorm:"type:int unsigned;not null;index;comment:绑定操作者" json:"userId"`
}

func (MediaUsage) TableName() string { return "media_usage" }
