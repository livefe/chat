package model

type IMSupportState struct {
	Base
	ConversationID     uint   `gorm:"type:int unsigned;not null;uniqueIndex;comment:会话ID" json:"conversationId"`
	UserID             uint   `gorm:"type:int unsigned;not null;index;comment:用户ID" json:"userId"`
	AssigneeEmployeeID uint   `gorm:"type:int unsigned;not null;default:0;index;comment:当前坐席员工ID" json:"assigneeEmployeeId"`
	QueueAt            int64  `gorm:"type:bigint;not null;default:0;index;comment:进入排队时间" json:"queueAt"`
	AcceptAt           int64  `gorm:"type:bigint;not null;default:0;comment:接入时间" json:"acceptAt"`
	CloseAt            int64  `gorm:"type:bigint;not null;default:0;comment:结案时间" json:"closeAt"`
	Source             string `gorm:"type:varchar(32);not null;default:'app';comment:来源" json:"source"`
	CloseReason        string `gorm:"type:varchar(64);not null;default:'';comment:结案原因" json:"closeReason"`
}

func (IMSupportState) TableName() string { return "im_support_state" }
