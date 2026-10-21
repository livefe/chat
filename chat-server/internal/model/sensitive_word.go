package model

const (
	SensitiveWordStatusDisabled uint8 = 0
	SensitiveWordStatusEnabled  uint8 = 1

	SensitiveWordLevelHigh   uint8 = 1
	SensitiveWordLevelMedium uint8 = 2
	SensitiveWordLevelLow    uint8 = 3

	SensitiveWordActionReject uint8 = 1

	SensitiveWordCategoryAbuse    = "abuse"
	SensitiveWordCategoryPorn     = "porn"
	SensitiveWordCategoryPolitics = "politics"
	SensitiveWordCategoryAd       = "ad"
	SensitiveWordCategoryOther    = "other"
)

type SensitiveWord struct {
	Base
	Word     string `gorm:"type:varchar(64);not null;uniqueIndex;comment:屏蔽词" json:"word"`
	Category string `gorm:"type:varchar(32);not null;default:'other';index;comment:分类" json:"category"`
	Level    uint8  `gorm:"type:tinyint unsigned;not null;default:1;comment:等级(1高 2中 3低)" json:"level"`
	Action   uint8  `gorm:"type:tinyint unsigned;not null;default:1;comment:动作(1拒绝)" json:"action"`
	Status   uint8  `gorm:"type:tinyint unsigned;not null;default:1;index;comment:状态(0停用 1启用)" json:"status"`
	Remark   string `gorm:"type:varchar(255);not null;default:'';comment:备注" json:"remark"`
}

func (SensitiveWord) TableName() string { return "sensitive_word" }
