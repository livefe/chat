package model

const (
	PunishActionWarn          uint8 = 1
	PunishActionMute          uint8 = 2
	PunishActionUnmute        uint8 = 3
	PunishActionDisable       uint8 = 4
	PunishActionEnable        uint8 = 5
	PunishActionDeactivate    uint8 = 6
	PunishActionDeleteMessage uint8 = 7
	PunishActionDisablePost   uint8 = 8
)

type UserPunishment struct {
	Base
	UserID     uint   `gorm:"type:int unsigned;not null;index;comment:用户ID" json:"userId"`
	Action     uint8  `gorm:"type:tinyint unsigned;not null;comment:动作(1警告 2禁言 3解除禁言 4禁用 5启用 6注销 7删除消息 8下架帖子)" json:"action"`
	MuteUntil  int64  `gorm:"type:bigint;not null;default:0;comment:禁言截止时间(秒),-1永久" json:"muteUntil"`
	TargetID   uint   `gorm:"type:int unsigned;not null;default:0;comment:关联对象ID(消息/帖子)" json:"targetId"`
	Reason     string `gorm:"type:varchar(500);not null;default:'';comment:原因" json:"reason"`
	EmployeeID uint   `gorm:"type:int unsigned;not null;default:0;comment:操作员工ID" json:"employeeId"`
	ReportID   uint   `gorm:"type:int unsigned;not null;default:0;index;comment:来源举报ID" json:"reportId"`
}

func (UserPunishment) TableName() string { return "user_punishment" }
