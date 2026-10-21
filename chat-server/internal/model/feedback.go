package model

import "gorm.io/datatypes"

const (
	FeedbackTypeProductSuggestion uint8 = 1
	FeedbackTypeFeatureBug        uint8 = 2
	FeedbackTypeExperienceIssue   uint8 = 3
	FeedbackTypeAccountIssue      uint8 = 4
	FeedbackTypeOther             uint8 = 5

	FeedbackStatusPending  uint8 = 0
	FeedbackStatusResolved uint8 = 1
	FeedbackStatusClosed   uint8 = 2
)

type Feedback struct {
	Base
	UserID       uint           `gorm:"type:int unsigned;index;not null;comment:提交用户ID" json:"userId"`
	Type         uint8          `gorm:"type:tinyint unsigned;not null;index;comment:反馈类型" json:"type"`
	Content      string         `gorm:"type:varchar(500);not null;comment:反馈内容" json:"content"`
	Contact      string         `gorm:"type:varchar(64);not null;default:'';comment:联系方式" json:"contact"`
	Images       datatypes.JSON `gorm:"type:json;comment:图片媒体ID列表" json:"images,omitempty"`
	Status       uint8          `gorm:"type:tinyint unsigned;not null;default:0;index;comment:状态(0待处理 1已处理 2已关闭)" json:"status"`
	HandleResult string         `gorm:"type:varchar(500);not null;default:'';comment:处理说明" json:"handleResult"`
	EmployeeID   uint           `gorm:"type:int unsigned;not null;default:0;comment:处理人员工ID" json:"employeeId"`
	HandledAt    int64          `gorm:"type:bigint;not null;default:0;comment:处理时间(秒)" json:"handledAt"`
}

func (Feedback) TableName() string { return "feedback" }
