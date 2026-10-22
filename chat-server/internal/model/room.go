package model

import "gorm.io/datatypes"

const (
	RoomShapeCircle  uint8 = 1
	RoomShapePolygon uint8 = 2

	RoomSourceAdmin uint8 = 1
	RoomSourceUser  uint8 = 2

	RoomStatusPending  uint8 = 0
	RoomStatusEnabled  uint8 = 1
	RoomStatusRejected uint8 = 2
	RoomStatusDisabled uint8 = 3

	RoomPlaceCommunity uint8 = 1
	RoomPlaceSchool    uint8 = 2
	RoomPlaceBusiness  uint8 = 3
	RoomPlaceScenic    uint8 = 4
	RoomPlaceCampus    uint8 = 5
	RoomPlaceTransit   uint8 = 6
	RoomPlaceOther     uint8 = 7
)

type Room struct {
	Base
	UserID          uint           `gorm:"type:int unsigned;not null;default:0;index;comment:创建用户ID(用户来源时)" json:"userId"`
	EmployeeID      uint           `gorm:"type:int unsigned;not null;default:0;index;comment:创建员工ID(后台来源时)" json:"employeeId"`
	Source          uint8          `gorm:"type:tinyint unsigned;not null;default:1;comment:来源(1:admin 2:user)" json:"source"`
	Status          uint8          `gorm:"type:tinyint unsigned;not null;default:0;index;comment:状态(0:待审核 1:正常 2:驳回 3:禁用)" json:"status"`
	Name            string         `gorm:"type:varchar(64);not null;comment:名称" json:"name"`
	Description     string         `gorm:"type:varchar(200);not null;default:'';comment:描述" json:"description"`
	PlaceType       uint8          `gorm:"type:tinyint unsigned;not null;default:7;index;comment:场所类型(1小区 2学校 3商圈 4景点 5园区 6交通枢纽 7其他)" json:"placeType"`
	ShapeType       uint8          `gorm:"type:tinyint unsigned;not null;default:1;comment:形状(1:圆 2:多边形)" json:"shapeType"`
	CenterLat       float64        `gorm:"type:decimal(10,6);not null;default:0;comment:中心纬度" json:"centerLat"`
	CenterLng       float64        `gorm:"type:decimal(10,6);not null;default:0;comment:中心经度" json:"centerLng"`
	Radius          uint           `gorm:"type:int unsigned;not null;default:0;comment:半径(米),多边形=外接圆半径" json:"radius"`
	Polygon         datatypes.JSON `gorm:"type:json;comment:多边形顶点[[lng,lat],...]" json:"polygon,omitempty"`
	RequireGeofence bool           `gorm:"type:tinyint(1);not null;default:0;comment:是否限制仅在地理围栏内可聊天" json:"requireGeofence"`
	AuditRemark     string         `gorm:"type:varchar(200);not null;default:'';comment:驳回原因" json:"auditRemark"`
	AuditedBy       uint           `gorm:"type:int unsigned;not null;default:0;comment:最近审核员工ID" json:"auditedBy"`
	AuditedAt       int64          `gorm:"type:bigint;not null;default:0;comment:最近审核时间(秒)" json:"auditedAt"`
}

func (Room) TableName() string { return "chat_room" }
