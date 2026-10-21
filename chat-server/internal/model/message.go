package model

const (
	MessageScopeNearby  uint8 = 1
	MessageScopeRoom    uint8 = 2
	MessageScopePrivate uint8 = 3
	MessageScopePost    uint8 = 4

	MessageContentTypeText     uint8 = 1
	MessageContentTypeImage    uint8 = 2
	MessageContentTypeVoice    uint8 = 3
	MessageContentTypeVideo    uint8 = 4
	MessageContentTypeFile     uint8 = 5
	MessageContentTypeLocation uint8 = 6
	MessageContentTypePost     uint8 = 7
)

type Message struct {
	Base

	Scope       uint8   `gorm:"type:tinyint unsigned;not null;index:idx_scope_room_id,priority:1;index:idx_scope_post_id,priority:1;index:idx_scope_nearby,priority:1;index:idx_private_ab,priority:1;index:idx_private_ba,priority:1;comment:范围(1:nearby 2:room 3:private 4:post)" json:"scope"`
	RoomID      uint    `gorm:"type:int unsigned;not null;default:0;index:idx_scope_room_id,priority:2;comment:房间ID" json:"roomId"`
	PostID      uint    `gorm:"type:int unsigned;not null;default:0;index:idx_scope_post_id,priority:2;comment:帖子ID(scope=post)" json:"postId"`
	FromUserID  uint    `gorm:"type:int unsigned;not null;index:idx_from;index:idx_private_ab,priority:2;index:idx_private_ba,priority:3;comment:发送者ID" json:"fromUserId"`
	ToUserID    uint    `gorm:"type:int unsigned;not null;default:0;index:idx_to;index:idx_private_ab,priority:3;index:idx_private_ba,priority:2;comment:接收者ID(私聊)" json:"toUserId"`
	ContentType uint8   `gorm:"type:tinyint unsigned;not null;default:1;comment:内容类型(1:text 2:image 3:voice 4:video 5:file 6:location 7:post)" json:"contentType"`
	Content     string  `gorm:"type:varchar(1000);not null;comment:内容" json:"content"`
	ReplyToID   uint    `gorm:"type:int unsigned;not null;default:0;index;comment:回复的消息ID" json:"replyToId"`
	Latitude    float64 `gorm:"type:decimal(10,6);not null;default:0;index:idx_scope_nearby,priority:2;comment:发送时纬度" json:"latitude"`
	Longitude   float64 `gorm:"type:decimal(10,6);not null;default:0;index:idx_scope_nearby,priority:3;comment:发送时经度" json:"longitude"`
	Recalled    uint8   `gorm:"type:tinyint unsigned;not null;default:0;comment:是否撤回" json:"recalled"`
}

func (Message) TableName() string { return "message" }
