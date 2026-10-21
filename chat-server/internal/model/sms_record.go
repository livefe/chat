package model

const (
	SMSStatusFailed  = 0
	SMSStatusSuccess = 1
)

type SMSRecord struct {
	Base
	Mobile        string `gorm:"type:varchar(11);not null;index;comment:手机号" json:"mobile"`
	Content       string `gorm:"type:varchar(500);not null;comment:内容" json:"content"`
	Source        string `gorm:"type:varchar(20);not null;default:'';comment:业务来源" json:"source"`
	IP            string `gorm:"type:varchar(50);not null;default:'';comment:IP" json:"ip"`
	UserAgent     string `gorm:"type:varchar(500);not null;default:'';comment:UA" json:"userAgent"`
	Status        uint8  `gorm:"type:tinyint unsigned;not null;default:0;comment:状态(0:失败 1:成功)" json:"status"`
	Remark        string `gorm:"type:varchar(500);not null;default:'';comment:备注" json:"remark"`
	BizID         string `gorm:"type:varchar(100);not null;default:'';comment:业务ID" json:"bizId"`
	RequestID     string `gorm:"type:varchar(100);not null;default:'';comment:请求ID" json:"requestId"`
	SignName      string `gorm:"type:varchar(50);not null;default:'';comment:签名" json:"signName"`
	TemplateCode  string `gorm:"type:varchar(50);not null;default:'';comment:模板编码" json:"templateCode"`
	TemplateParam string `gorm:"type:varchar(500);not null;default:'';comment:模板参数" json:"templateParam"`
}

func (SMSRecord) TableName() string { return "sms_record" }
