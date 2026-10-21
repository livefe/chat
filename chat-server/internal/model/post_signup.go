package model

const (
	SignupStatusConfirmed  uint8 = 1
	SignupStatusWaitlisted uint8 = 2
	SignupStatusCancelled  uint8 = 3

	SignupCheckInNotChecked uint8 = 1
	SignupCheckInChecked    uint8 = 2
)

type PostSignup struct {
	Base
	PostID       uint   `gorm:"type:int unsigned;not null;uniqueIndex:idx_post_user,priority:1;comment:帖子ID" json:"postId"`
	UserID       uint   `gorm:"type:int unsigned;not null;uniqueIndex:idx_post_user,priority:2;comment:用户ID" json:"userId"`
	Status       uint8  `gorm:"type:tinyint unsigned;not null;default:1;index;comment:状态(1已报名 2候补 3已取消)" json:"status"`
	CancelledAt  int64  `gorm:"type:bigint;not null;default:0;comment:取消时间(秒)" json:"cancelledAt"`
	CheckInToken string `gorm:"type:varchar(40);not null;default:'';index;comment:核验凭证(已报名时生成)" json:"checkInToken"`
	CheckedInAt  int64  `gorm:"type:bigint;not null;default:0;comment:核验时间(秒)" json:"checkedInAt"`
	HiddenAt     int64  `gorm:"type:bigint;not null;default:0;index;comment:用户从我参与的移除时间(秒)" json:"hiddenAt"`
	Remark       string `gorm:"type:varchar(100);not null;default:'';comment:主办备注" json:"remark"`
}

func (PostSignup) TableName() string { return "post_signup" }
