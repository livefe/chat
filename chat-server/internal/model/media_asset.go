package model

import "time"

const (
	MediaKindImage uint8 = 1
	MediaKindVideo uint8 = 2
	MediaKindAudio uint8 = 3
	MediaKindFile  uint8 = 4

	MediaUploaderUser     uint8 = 1
	MediaUploaderEmployee uint8 = 2

	MediaStatusUploaded uint8 = 1
	MediaStatusBound    uint8 = 2
	MediaStatusDeleted  uint8 = 3
	MediaStatusPending  uint8 = 4
)

type MediaAsset struct {
	Base

	UploaderKind uint8      `gorm:"type:tinyint unsigned;not null;comment:上传者类型(1:用户 2:员工)" json:"uploaderKind"`
	UploaderID   uint       `gorm:"type:int unsigned;not null;comment:上传者ID" json:"uploaderId"`
	Kind         uint8      `gorm:"type:tinyint unsigned;not null;comment:类型(1:image 2:video 3:audio 4:file)" json:"kind"`
	ObjectKey    string     `gorm:"type:varchar(512);not null;uniqueIndex;comment:存储对象键" json:"objectKey"`
	MimeType     string     `gorm:"type:varchar(128);not null;default:'';comment:MIME" json:"mimeType"`
	Ext          string     `gorm:"type:varchar(16);not null;default:'';comment:扩展名" json:"ext"`
	SizeBytes    int64      `gorm:"type:bigint;not null;default:0;comment:字节数" json:"sizeBytes"`
	Status       uint8      `gorm:"type:tinyint unsigned;not null;default:1;index:idx_media_status_created,priority:1;comment:状态(1:已上传 2:已绑定 3:已删除 4:上传中)" json:"status"`
	BoundAt      *time.Time `gorm:"type:datetime;comment:首次绑定时间" json:"boundAt,omitempty"`
	PurgedAt     *time.Time `gorm:"type:datetime;comment:物理文件删除时间" json:"purgedAt,omitempty"`
}

func (MediaAsset) TableName() string { return "media_asset" }
