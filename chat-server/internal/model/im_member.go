package model

const (
	IMPrincipalUser     uint8 = 1
	IMPrincipalEmployee uint8 = 2
	IMPrincipalSystem   uint8 = 3

	IMMemberRoleMember uint8 = 1
	IMMemberRoleAgent  uint8 = 3
)

type IMMember struct {
	Base
	ConversationID uint  `gorm:"type:int unsigned;not null;uniqueIndex:uk_im_member,priority:1;index:idx_im_member_principal,priority:3;comment:会话ID" json:"conversationId"`
	PrincipalKind  uint8 `gorm:"type:tinyint unsigned;not null;uniqueIndex:uk_im_member,priority:2;index:idx_im_member_principal,priority:1;comment:主体类型(1:user 2:employee 3:system)" json:"principalKind"`
	PrincipalID    uint  `gorm:"type:int unsigned;not null;uniqueIndex:uk_im_member,priority:3;index:idx_im_member_principal,priority:2;comment:主体ID" json:"principalId"`
	Role           uint8 `gorm:"type:tinyint unsigned;not null;default:1;comment:角色(1:member 3:agent)" json:"role"`
	LastReadMsgID  uint  `gorm:"type:int unsigned;not null;default:0;comment:已读到的消息ID" json:"lastReadMsgId"`
	Pinned         uint8 `gorm:"type:tinyint unsigned;not null;default:0;comment:是否置顶" json:"pinned"`
	Muted          uint8 `gorm:"type:tinyint unsigned;not null;default:0;comment:是否免打扰" json:"muted"`
	JoinedAt       int64 `gorm:"type:bigint;not null;default:0;comment:加入时间(秒)" json:"joinedAt"`
	LeftAt         int64 `gorm:"type:bigint;not null;default:0;comment:离开时间(秒,0=在席)" json:"leftAt"`
}

func (IMMember) TableName() string { return "im_member" }
