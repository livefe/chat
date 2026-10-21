package model

import "gorm.io/datatypes"

const (
	ReportTargetUser    uint8 = 1
	ReportTargetMessage uint8 = 2
	ReportTargetPost    uint8 = 3

	ReportReasonSpam         uint8 = 1
	ReportReasonPorn         uint8 = 2
	ReportReasonHarassment   uint8 = 3
	ReportReasonIllegal      uint8 = 4
	ReportReasonFake         uint8 = 5
	ReportReasonInfringement uint8 = 6
	ReportReasonOther        uint8 = 7

	ReportStatusPending  uint8 = 0
	ReportStatusResolved uint8 = 1
	ReportStatusRejected uint8 = 2

	ReportHandleActionNone          uint8 = 0
	ReportHandleActionWarn          uint8 = 1
	ReportHandleActionDeleteContent uint8 = 2
	ReportHandleActionMute          uint8 = 3
	ReportHandleActionDisableUser   uint8 = 4
	ReportHandleActionDisablePost   uint8 = 5
)

type Report struct {
	Base
	ReporterID   uint           `gorm:"type:int unsigned;index:idx_reporter_target,priority:1;not null;comment:举报人ID" json:"reporterId"`
	TargetType   uint8          `gorm:"type:tinyint unsigned;not null;index:idx_target,priority:1;index:idx_reporter_target,priority:2;comment:目标类型" json:"targetType"`
	TargetID     uint           `gorm:"type:int unsigned;not null;index:idx_target,priority:2;index:idx_reporter_target,priority:3;comment:目标ID" json:"targetId"`
	TargetUserID uint           `gorm:"type:int unsigned;not null;default:0;index;comment:被举报用户ID" json:"targetUserId"`
	Reason       uint8          `gorm:"type:tinyint unsigned;not null;index;comment:举报原因" json:"reason"`
	Content      string         `gorm:"type:varchar(500);not null;default:'';comment:补充说明" json:"content"`
	Images       datatypes.JSON `gorm:"type:json;comment:图片媒体ID列表" json:"images,omitempty"`
	Context      datatypes.JSON `gorm:"type:json;comment:上下文快照" json:"context,omitempty"`
	Status       uint8          `gorm:"type:tinyint unsigned;not null;default:0;index;comment:状态(0待处理 1已处理 2已驳回)" json:"status"`
	HandleResult string         `gorm:"type:varchar(500);not null;default:'';comment:处理说明" json:"handleResult"`
	HandleAction uint8          `gorm:"type:tinyint unsigned;not null;default:0;comment:处置动作" json:"handleAction"`
	EmployeeID   uint           `gorm:"type:int unsigned;not null;default:0;comment:处理人员工ID" json:"employeeId"`
	HandledAt    int64          `gorm:"type:bigint;not null;default:0;comment:处理时间(秒)" json:"handledAt"`
}

func (Report) TableName() string { return "report" }
