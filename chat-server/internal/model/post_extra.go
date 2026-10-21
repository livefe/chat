package model

const (
	PostFeeNone   uint8 = 0
	PostFeeCharge uint8 = 1
	PostFeeReward uint8 = 2
)

type PostExtra struct {
	Base
	PostID           uint    `gorm:"type:int unsigned;not null;uniqueIndex;comment:帖子ID" json:"postId"`
	Amount           float64 `gorm:"type:decimal(10,2);not null;default:0;comment:金额(闲置价/费用/报酬)" json:"amount"`
	Negotiable       bool    `gorm:"type:tinyint(1);not null;default:0;comment:是否可议价(闲置)" json:"negotiable"`
	NeedSignup       bool    `gorm:"type:tinyint(1);not null;default:0;comment:是否需要报名" json:"needSignup"`
	NeedCheckIn      bool    `gorm:"type:tinyint(1);not null;default:0;comment:是否需要核验(需先报名)" json:"needCheckIn"`
	MaxParticipants  uint    `gorm:"type:int unsigned;not null;default:0;comment:人数上限(0=不限)" json:"maxParticipants"`
	AllowWaitlist    bool    `gorm:"type:tinyint(1);not null;default:0;comment:是否允许候补" json:"allowWaitlist"`
	SignupDeadlineAt int64   `gorm:"type:bigint;not null;default:0;comment:报名截止时间(秒)" json:"signupDeadlineAt"`
	FeeType          uint8   `gorm:"type:tinyint unsigned;not null;default:0;comment:费用类型(0无 1收费 2报酬)" json:"feeType"`
	FeeDescription   string  `gorm:"type:varchar(200);not null;default:'';comment:费用/报酬说明" json:"feeDescription"`
}

func (PostExtra) TableName() string { return "post_extra" }
