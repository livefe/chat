package types

const (
	CacheRoomGeoIndex  = cachePrefix + "geo:room"
	NearbyRadiusMeters = 3000
)

type CreateRoomRequest struct {
	Name            string       `json:"name" binding:"required,max=64"`
	Description     string       `json:"description" binding:"max=200"`
	PlaceType       uint8        `json:"placeType" binding:"required,oneof=1 2 3 4 5 6 7"`
	ShapeType       uint8        `json:"shapeType" binding:"required,oneof=1 2"`
	CenterLat       float64      `json:"centerLat"`
	CenterLng       float64      `json:"centerLng"`
	Radius          uint         `json:"radius"`
	Polygon         [][2]float64 `json:"polygon"`
	RequireGeofence bool         `json:"requireGeofence"`
}

type RoomItem struct {
	ID              uint    `json:"id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	PlaceType       uint8   `json:"placeType"`
	PlaceTypeText   string  `json:"placeTypeText"`
	ShapeType       uint8   `json:"shapeType"`
	CenterLat       float64 `json:"centerLat"`
	CenterLng       float64 `json:"centerLng"`
	Radius          uint    `json:"radius"`
	UserID          uint    `json:"userId"`
	EmployeeID      uint    `json:"employeeId"`
	Source          uint8   `json:"source"`
	SourceText      string  `json:"sourceText"`
	Status          uint8   `json:"status"`
	StatusText      string  `json:"statusText"`
	RequireGeofence bool    `json:"requireGeofence"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

type RoomDetail struct {
	RoomItem
	Polygon [][2]float64 `json:"polygon,omitempty"`
}

type NearbyRoomItem struct {
	ID              uint    `json:"id"`
	Name            string  `json:"name"`
	PlaceType       uint8   `json:"placeType"`
	PlaceTypeText   string  `json:"placeTypeText"`
	ShapeType       uint8   `json:"shapeType"`
	CenterLat       float64 `json:"centerLat"`
	CenterLng       float64 `json:"centerLng"`
	Radius          uint    `json:"radius"`
	Distance        float64 `json:"distance"`
	RequireGeofence bool    `json:"requireGeofence"`
}

type NearbyBootstrapQuery struct {
	Lat          float64 `form:"lat" binding:"required"`
	Lng          float64 `form:"lng" binding:"required"`
	Radius       float64 `form:"radius" binding:"required,min=1"`
	Limit        int     `form:"limit" binding:"required,min=1,max=100"`
	WithMessages bool    `form:"withMessages"`
}

type NearbyBootstrapResponse struct {
	Rooms       []NearbyRoomItem `json:"rooms"`
	OnlineCount int              `json:"onlineCount"`
	Messages    []MessageItem    `json:"messages,omitempty"`
}

type RoomBootstrapQuery struct {
	RoomID       uint    `form:"roomId" binding:"required"`
	Limit        int     `form:"limit" binding:"required,min=1,max=100"`
	WithMessages bool    `form:"withMessages"`
	Lat          float64 `form:"lat"`
	Lng          float64 `form:"lng"`
}

type RoomBootstrapResponse struct {
	Room        RoomDetail    `json:"room"`
	OnlineCount int           `json:"onlineCount"`
	Messages    []MessageItem `json:"messages,omitempty"`
}

type RoomListRequest struct {
	ID        *uint  `form:"id"`
	Name      string `form:"name"`
	Status    *uint8 `form:"status"`
	Source    *uint8 `form:"source"`
	PlaceType *uint8 `form:"placeType"`
	Page      int    `form:"page" binding:"required,min=1"`
	PageSize  int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type RoomListResponse struct {
	Total int64      `json:"total"`
	List  []RoomItem `json:"list"`
}

type UpdateRoomRequest struct {
	ID uint `json:"id" binding:"required"`
	CreateRoomRequest
}

type AuditRoomRequest struct {
	ID      uint `json:"id" binding:"required"`
	Approve bool `json:"approve"`
}

type UpdateRoomStatusRequest struct {
	ID     uint  `json:"id" binding:"required"`
	Status uint8 `json:"status" binding:"oneof=0 1 2 3"`
}

type AdminUserRoomSummary struct {
	RoomID        uint   `json:"roomId"`
	RoomName      string `json:"roomName"`
	MessageCount  int64  `json:"messageCount"`
	LastMessageAt string `json:"lastMessageAt"`
}

type MyRoomApplicationsQuery struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

type AdminRoomNeighborsQuery struct {
	RoomID    uint    `form:"roomId"`
	Lat       float64 `form:"lat"`
	Lng       float64 `form:"lng"`
	RadiusM   float64 `form:"radiusM" binding:"required,gt=0"`
	ExcludeID uint    `form:"excludeId"`
	Limit     int     `form:"limit" binding:"required,min=1"`
}

type AdminRoomNeighborItem struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	PlaceType     uint8   `json:"placeType"`
	PlaceTypeText string  `json:"placeTypeText"`
	ShapeType     uint8   `json:"shapeType"`
	CenterLat     float64 `json:"centerLat"`
	CenterLng     float64 `json:"centerLng"`
	Radius        uint    `json:"radius"`
	Status        uint8   `json:"status"`
	StatusText    string  `json:"statusText"`
	Distance      float64 `json:"distance"`
	Suspect       bool    `json:"suspect"`
	SuspectReason string  `json:"suspectReason"`
}

type AdminRoomNeighborsResponse struct {
	List []AdminRoomNeighborItem `json:"list"`
}

type AdminRoomMapQuery struct {
	Status    *uint8 `form:"status"`
	PlaceType *uint8 `form:"placeType"`
	Source    *uint8 `form:"source"`
	Limit     int    `form:"limit" binding:"required,min=1,max=500"`
}

type AdminRoomMapItem struct {
	ID            uint         `json:"id"`
	Name          string       `json:"name"`
	PlaceType     uint8        `json:"placeType"`
	PlaceTypeText string       `json:"placeTypeText"`
	ShapeType     uint8        `json:"shapeType"`
	CenterLat     float64      `json:"centerLat"`
	CenterLng     float64      `json:"centerLng"`
	Radius        uint         `json:"radius"`
	Polygon       [][2]float64 `json:"polygon,omitempty"`
	Status        uint8        `json:"status"`
	StatusText    string       `json:"statusText"`
	Source        uint8        `json:"source"`
	SourceText    string       `json:"sourceText"`
}

type AdminRoomMapResponse struct {
	List []AdminRoomMapItem `json:"list"`
}
