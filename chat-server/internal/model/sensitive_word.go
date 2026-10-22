package model

const (
	SensitiveWordStatusDisabled uint8 = 0
	SensitiveWordStatusEnabled  uint8 = 1

	SensitiveWordCategoryAbuse    = "abuse"
	SensitiveWordCategoryPorn     = "porn"
	SensitiveWordCategoryPolitics = "politics"
	SensitiveWordCategoryAd       = "ad"
	SensitiveWordCategoryOther    = "other"
)

// 命中任一启用的屏蔽词即拒绝发布。
type SensitiveWord struct {
	Base
	Word     string `gorm:"type:varchar(64);not null;uniqueIndex;comment:屏蔽词" json:"word"`
	Category string `gorm:"type:varchar(32);not null;default:'other';index;comment:分类" json:"category"`
	Status   uint8  `gorm:"type:tinyint unsigned;not null;default:1;index;comment:状态(0停用 1启用)" json:"status"`
	Remark   string `gorm:"type:varchar(255);not null;default:'';comment:备注" json:"remark"`
}

func (SensitiveWord) TableName() string { return "sensitive_word" }
