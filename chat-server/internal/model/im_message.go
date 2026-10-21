package model

type IMMessage struct {
	Base
	ConversationID uint   `gorm:"type:int unsigned;not null;index:idx_im_msg_conv,priority:1;comment:会话ID" json:"conversationId"`
	SenderKind     uint8  `gorm:"type:tinyint unsigned;not null;index:idx_im_client_msg,priority:1;comment:发送者类型" json:"senderKind"`
	SenderID       uint   `gorm:"type:int unsigned;not null;index:idx_im_client_msg,priority:2;index;comment:发送者ID" json:"senderId"`
	ContentType    uint8  `gorm:"type:tinyint unsigned;not null;default:1;comment:内容类型(1:text 2:image 3:voice 4:video 5:file 6:location 7:post)" json:"contentType"`
	Content        string `gorm:"type:varchar(1000);not null;comment:内容" json:"content"`
	ReplyToID      uint   `gorm:"type:int unsigned;not null;default:0;index;comment:回复消息ID" json:"replyToId"`
	Recalled       uint8  `gorm:"type:tinyint unsigned;not null;default:0;comment:是否撤回" json:"recalled"`
	ClientMsgID    string `gorm:"type:varchar(64);not null;default:'';index:idx_im_client_msg,priority:3;comment:客户端消息ID" json:"clientMsgId"`
}

func (IMMessage) TableName() string { return "im_message" }
