package types

const (
	CachePostSignupLockKey = cachePrefix + "post:signup:lock:%d"

	CheckInPayloadPrefix = "jdcheckin:1:"
)

type PostExtrasPayload struct {
	Amount           float64 `json:"amount"`
	Negotiable       bool    `json:"negotiable"`
	NeedSignup       bool    `json:"needSignup"`
	NeedCheckIn      bool    `json:"needCheckIn"`
	MaxParticipants  uint    `json:"maxParticipants"`
	AllowWaitlist    bool    `json:"allowWaitlist"`
	SignupDeadlineAt int64   `json:"signupDeadlineAt"`
	FeeType          uint8   `json:"feeType"`
	FeeDescription   string  `json:"feeDescription"`
}

type CreatePostRequest struct {
	RoomID    uint               `json:"roomId"`
	Category  string             `json:"category" binding:"required,oneof=event notice idle help"`
	Pinned    bool               `json:"pinned"`
	Title     string             `json:"title" binding:"required,max=100"`
	Content   string             `json:"content"`
	Images    []uint             `json:"images"`
	Extras    *PostExtrasPayload `json:"extras"`
	PlaceName string             `json:"placeName" binding:"max=120"`
	Latitude  float64            `json:"latitude"`
	Longitude float64            `json:"longitude"`
	StartAt   int64              `json:"startAt"`
	EndAt     int64              `json:"endAt"`
}

type PostListRequest struct {
	Latitude  float64 `form:"latitude"`
	Longitude float64 `form:"longitude"`
	RoomID    *uint   `form:"roomId"`
	Category  string  `form:"category"`
	Page      int     `form:"page" binding:"required,min=1"`
	PageSize  int     `form:"pageSize" binding:"required,min=1,max=100"`
}

type MyPostListRequest struct {
	Latitude  float64 `form:"latitude"`
	Longitude float64 `form:"longitude"`
	Page      int     `form:"page" binding:"required,min=1"`
	PageSize  int     `form:"pageSize" binding:"required,min=1,max=100"`
}

type UserPostListRequest struct {
	UserID    uint    `form:"userId" binding:"required"`
	Latitude  float64 `form:"latitude"`
	Longitude float64 `form:"longitude"`
	Page      int     `form:"page" binding:"required,min=1"`
	PageSize  int     `form:"pageSize" binding:"required,min=1,max=100"`
}

type AdminPostListRequest struct {
	ID               *uint  `form:"id"`
	PublisherID      *uint  `form:"publisherId"`
	PublisherKeyword string `form:"publisherKeyword" binding:"max=32"`
	RoomID           *uint  `form:"roomId"`
	Category         string `form:"category"`
	Status           *uint8 `form:"status"`
	Keyword          string `form:"keyword" binding:"max=100"`
	StartTime        string `form:"startTime" binding:"omitempty,datetime=2006-01-02 15:04:05"`
	EndTime          string `form:"endTime" binding:"omitempty,datetime=2006-01-02 15:04:05"`
	Sort             string `form:"sort" binding:"omitempty,oneof=asc desc"`
	IncludeDeleted   bool   `form:"includeDeleted"`
	Page             int    `form:"page" binding:"required,min=1"`
	PageSize         int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type AdminPostItem struct {
	PostItem
	PublisherMobile string `json:"publisherMobile"`
	RoomName        string `json:"roomName"`
	Deleted         bool   `json:"deleted"`
	AuditedByName   string `json:"auditedByName"`
	AuditedAt       string `json:"auditedAt"`
}

type AdminPostListResponse struct {
	Total int64           `json:"total"`
	List  []AdminPostItem `json:"list"`
}

type PostPublisher struct {
	ID        uint   `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
}

type PostItem struct {
	ID                  uint                `json:"id"`
	PublisherID         uint                `json:"publisherId"`
	RoomID              uint                `json:"roomId"`
	Category            string              `json:"category"`
	Pinned              bool                `json:"pinned"`
	Title               string              `json:"title"`
	Content             string              `json:"content"`
	Images              []string            `json:"images"`
	ImageIDs            []uint              `json:"imageIds,omitempty"`
	Extras              *PostExtrasPayload  `json:"extras,omitempty"`
	PlaceName           string              `json:"placeName"`
	Latitude            float64             `json:"latitude"`
	Longitude           float64             `json:"longitude"`
	StartAt             int64               `json:"startAt"`
	EndAt               int64               `json:"endAt"`
	Status              uint8               `json:"status"`
	Publisher           *PostPublisher      `json:"publisher,omitempty"`
	DiscussMessageCount uint                `json:"discussMessageCount"`
	DiscussPreview      *PostDiscussPreview `json:"discussPreview,omitempty"`
	Distance            float64             `json:"distance"`
	SignupSummary       *PostSignupSummary  `json:"signupSummary,omitempty"`
	MySignupStatus      uint8               `json:"mySignupStatus,omitempty"`
	MyCheckedInAt       int64               `json:"myCheckedInAt,omitempty"`
	AuditRemark         string              `json:"auditRemark,omitempty"`
	CreatedAt           string              `json:"createdAt"`
	UpdatedAt           string              `json:"updatedAt"`
}

type PostDiscussPreview struct {
	FromName    string `json:"fromName"`
	ContentType uint8  `json:"contentType"`
	Content     string `json:"content"`
}

type PostDetail struct {
	PostItem
}

type PostListResponse struct {
	Total int64      `json:"total"`
	List  []PostItem `json:"list"`
}

type PostSignupRequest struct {
	PostID uint `json:"postId" binding:"required"`
}

type PostSignupCheckInRequest struct {
	Payload string `json:"payload"`
	PostID  uint   `json:"postId"`
	UserID  uint   `json:"userId"`
}

type PostSignupCheckInCodeRequest struct {
	PostID uint `form:"postId" binding:"required"`
}

type PostSignupCheckInCodeResponse struct {
	PostID          uint   `json:"postId"`
	PostTitle       string `json:"postTitle"`
	PlaceName       string `json:"placeName"`
	StartAt         int64  `json:"startAt"`
	EndAt           int64  `json:"endAt"`
	Payload         string `json:"payload"`
	CheckedIn       bool   `json:"checkedIn"`
	CheckedInAt     int64  `json:"checkedInAt"`
	CheckedInAtText string `json:"checkedInAtText"`
	UserNickname    string `json:"userNickname"`
	SignupAt        string `json:"signupAt"`
}

type PostSignupCheckInResult struct {
	PostID            uint   `json:"postId"`
	PostTitle         string `json:"postTitle"`
	UserID            uint   `json:"userId"`
	UserNickname      string `json:"userNickname"`
	UserAvatarURL     string `json:"userAvatarUrl"`
	SignupAt          string `json:"signupAt"`
	CheckedInAt       int64  `json:"checkedInAt"`
	CheckedInAtText   string `json:"checkedInAtText"`
	AlreadyCheckedIn  bool   `json:"alreadyCheckedIn"`
	OperatorID        uint   `json:"operatorId"`
	OperatorNickname  string `json:"operatorNickname"`
	OperatorAvatarURL string `json:"operatorAvatarUrl"`
}

type PostSignupSummary struct {
	Confirmed    int64 `json:"confirmed"`
	Waitlisted   int64 `json:"waitlisted"`
	Cancelled    int64 `json:"cancelled"`
	NotCheckedIn int64 `json:"notCheckedIn"`
	CheckedIn    int64 `json:"checkedIn"`
}

type PostSignupResult struct {
	Status uint8 `json:"status"`
}

type AdminPostSignupListRequest struct {
	PostID    uint   `form:"postId" binding:"required"`
	Status    *uint8 `form:"status"`
	CheckIn   *uint8 `form:"checkIn"`
	UserID    *uint  `form:"userId"`
	Keyword   string `form:"keyword" binding:"max=32"`
	SortField string `form:"sortField" binding:"omitempty,oneof=createdAt checkedInAt"`
	SortOrder string `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
	Page      int    `form:"page" binding:"required,min=1"`
	PageSize  int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type AdminUserSignupListRequest struct {
	UserID   uint `form:"userId" binding:"required"`
	Page     int  `form:"page" binding:"required,min=1"`
	PageSize int  `form:"pageSize" binding:"required,min=1,max=100"`
}

type AdminUserSignupItem struct {
	ID               uint   `json:"id"`
	PostID           uint   `json:"postId"`
	PostTitle        string `json:"postTitle"`
	PostStatus       uint8  `json:"postStatus"`
	PostStatusText   string `json:"postStatusText"`
	SignupStatus     uint8  `json:"signupStatus"`
	SignupStatusText string `json:"signupStatusText"`
	CheckedIn        bool   `json:"checkedIn"`
	CheckedInAt      string `json:"checkedInAt"`
	CancelledAt      string `json:"cancelledAt"`
	CreatedAt        string `json:"createdAt"`
}

type AdminUserSignupListResponse struct {
	Total int64                 `json:"total"`
	List  []AdminUserSignupItem `json:"list"`
}

type AdminPostSignupItem struct {
	ID               uint   `json:"id"`
	UserID           uint   `json:"userId"`
	UserNickname     string `json:"userNickname"`
	UserMobile       string `json:"userMobile"`
	SignupStatus     uint8  `json:"signupStatus"`
	SignupStatusText string `json:"signupStatusText"`
	CheckedInAt      string `json:"checkedInAt"`
	CreatedAt        string `json:"createdAt"`
	CancelledAt      string `json:"cancelledAt"`
	Remark           string `json:"remark"`
}

type UpdateSignupRemarkRequest struct {
	PostID uint   `json:"postId" binding:"required"`
	UserID uint   `json:"userId" binding:"required"`
	Remark string `json:"remark"`
}

type PostSignupListCounts struct {
	Total        int64 `json:"total"`
	Confirmed    int64 `json:"confirmed"`
	Waitlisted   int64 `json:"waitlisted"`
	Cancelled    int64 `json:"cancelled"`
	NotCheckedIn int64 `json:"notCheckedIn"`
	CheckedIn    int64 `json:"checkedIn"`
}

type AdminPostSignupListResponse struct {
	Total       int64                 `json:"total"`
	List        []AdminPostSignupItem `json:"list"`
	NeedCheckIn bool                  `json:"needCheckIn"`
	Counts      PostSignupListCounts  `json:"counts"`
}

type UpdatePostStatusRequest struct {
	ID     uint   `json:"id" binding:"required"`
	Status uint8  `json:"status" binding:"oneof=1 2 3"`
	Reason string `json:"reason" binding:"max=200"`
}

type BatchUpdatePostStatusRequest struct {
	IDs    []uint `json:"ids" binding:"required,min=1,max=100"`
	Status uint8  `json:"status" binding:"oneof=1 2 3"`
	Reason string `json:"reason" binding:"max=200"`
}

type ContentBatchResult struct {
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
}

type UpdatePostRequest struct {
	ID        uint               `json:"id" binding:"required"`
	Category  string             `json:"category" binding:"required,oneof=event notice idle help"`
	Pinned    bool               `json:"pinned"`
	Title     string             `json:"title" binding:"required,max=100"`
	Content   string             `json:"content"`
	Images    []uint             `json:"images"`
	Extras    *PostExtrasPayload `json:"extras"`
	PlaceName string             `json:"placeName" binding:"max=120"`
	Latitude  float64            `json:"latitude"`
	Longitude float64            `json:"longitude"`
	StartAt   int64              `json:"startAt"`
	EndAt     int64              `json:"endAt"`
}

type AdminPostMapQuery struct {
	Status           *uint8 `form:"status"`
	Category         string `form:"category"`
	IndependentLimit int    `form:"independentLimit" binding:"required,min=1,max=500"`
}

type AdminPostMapRoomItem struct {
	RoomID         uint         `json:"roomId"`
	Name           string       `json:"name"`
	PlaceType      uint8        `json:"placeType"`
	PlaceTypeText  string       `json:"placeTypeText"`
	ShapeType      uint8        `json:"shapeType"`
	CenterLat      float64      `json:"centerLat"`
	CenterLng      float64      `json:"centerLng"`
	Radius         uint         `json:"radius"`
	Polygon        [][2]float64 `json:"polygon,omitempty"`
	RoomStatus     uint8        `json:"roomStatus"`
	RoomStatusText string       `json:"roomStatusText"`
	PostCount      int64        `json:"postCount"`
}

type AdminPostMapIndependentItem struct {
	ID        uint    `json:"id"`
	Title     string  `json:"title"`
	Category  string  `json:"category"`
	PlaceName string  `json:"placeName"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    uint8   `json:"status"`
	CreatedAt string  `json:"createdAt"`
}

type AdminPostMapResponse struct {
	Rooms       []AdminPostMapRoomItem        `json:"rooms"`
	Independent []AdminPostMapIndependentItem `json:"independent"`
}
