package model

type IMStaffDMPair struct {
	Base
	EmpMinID       uint `gorm:"type:int unsigned;not null;uniqueIndex:uk_im_staff_dm,priority:1;comment:较小员工ID" json:"empMinId"`
	EmpMaxID       uint `gorm:"type:int unsigned;not null;uniqueIndex:uk_im_staff_dm,priority:2;comment:较大员工ID" json:"empMaxId"`
	ConversationID uint `gorm:"type:int unsigned;not null;uniqueIndex;comment:会话ID" json:"conversationId"`
}

func (IMStaffDMPair) TableName() string { return "im_staff_dm_pair" }
