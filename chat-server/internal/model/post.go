package model

import "gorm.io/datatypes"

const (
	PostStatusPending   uint8 = 0
	PostStatusPublished uint8 = 1
	PostStatusRejected  uint8 = 2
	PostStatusDisabled  uint8 = 3

	PostCategoryEvent  = "event"
	PostCategoryNotice = "notice"
	PostCategoryIdle   = "idle"
	PostCategoryHelp   = "help"
)

type Post struct {
	Base
	UserID              uint           `gorm:"type:int unsigned;not null;default:0;index;comment:发布用户ID" json:"userId"`
	RoomID              uint           `gorm:"type:int unsigned;not null;default:0;index;comment:聊天区ID(0=独立发现帖)" json:"roomId"`
	Category            string         `gorm:"type:varchar(20);not null;default:'idle';index;comment:分类" json:"category"`
	Pinned              bool           `gorm:"type:tinyint(1);not null;default:0;index;comment:区内置顶" json:"pinned"`
	Title               string         `gorm:"type:varchar(100);not null;comment:标题" json:"title"`
	Content             string         `gorm:"type:text;not null;comment:内容" json:"content"`
	Images              datatypes.JSON `gorm:"type:json;comment:图片媒体ID列表" json:"images,omitempty"`
	PlaceName           string         `gorm:"type:varchar(120);not null;default:'';comment:地点名称" json:"placeName"`
	Latitude            float64        `gorm:"type:decimal(10,6);not null;default:0;comment:纬度" json:"latitude"`
	Longitude           float64        `gorm:"type:decimal(10,6);not null;default:0;comment:经度" json:"longitude"`
	StartAt             int64          `gorm:"type:bigint;not null;default:0;index;comment:开始/出发时间(秒)" json:"startAt"`
	EndAt               int64          `gorm:"type:bigint;not null;default:0;index;comment:结束/截止时间(秒)" json:"endAt"`
	Status              uint8          `gorm:"type:tinyint unsigned;not null;default:0;index;comment:状态(0待审核 1发布中 2驳回 3下架)" json:"status"`
	DiscussMessageCount uint           `gorm:"type:int unsigned;not null;default:0;comment:可见讨论消息数" json:"discussMessageCount"`
	AuditRemark         string         `gorm:"type:varchar(200);not null;default:'';comment:审核/下架说明" json:"auditRemark"`
	AuditedBy           uint           `gorm:"type:int unsigned;not null;default:0;comment:最近审核员工ID" json:"auditedBy"`
	AuditedAt           int64          `gorm:"type:bigint;not null;default:0;comment:最近审核时间(秒)" json:"auditedAt"`
}

func (Post) TableName() string { return "post" }
