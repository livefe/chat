package model

type UserLoginRecord struct {
	Base
	UserID    uint   `gorm:"type:int unsigned;index;comment:用户ID" json:"userId"`
	Mobile    string `gorm:"type:varchar(11);not null;index;comment:手机号" json:"mobile"`
	LoginType string `gorm:"type:varchar(20);not null;comment:登录类型" json:"loginType"`
	IP        string `gorm:"type:varchar(50);not null;default:'';comment:IP" json:"ip"`
	UserAgent string `gorm:"type:varchar(500);not null;default:'';comment:UA" json:"userAgent"`
	Status    uint8  `gorm:"type:tinyint unsigned;not null;default:0;comment:状态(0:失败 1:成功)" json:"status"`
	Remark    string `gorm:"type:varchar(200);not null;default:'';comment:备注" json:"remark"`
}

func (UserLoginRecord) TableName() string { return "user_login_record" }
