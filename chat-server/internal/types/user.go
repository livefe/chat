package types

const (
	CacheUserByMobileKey       = cachePrefix + "user:mobile:%s"
	CacheUserByIDKey           = cachePrefix + "user:id:%d"
	CacheChangeMobileTicketKey = cachePrefix + "user:change_mobile_ticket:%d"
)

type UserInfo struct {
	ID        uint   `json:"id"`
	Mobile    string `json:"mobile"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
	Status    uint8  `json:"status"`
}

type PublicUserProfile struct {
	ID          uint   `json:"id"`
	Nickname    string `json:"nickname"`
	AvatarURL   string `json:"avatarUrl"`
	IsSelf      bool   `json:"isSelf"`
	Deactivated bool   `json:"deactivated"`
	BlockedByMe bool   `json:"blockedByMe"`
	CanMessage  bool   `json:"canMessage"`
	CreatedAt   string `json:"createdAt"`
}

type UserProfile struct {
	ID        uint
	Mobile    string
	Nickname  string
	AvatarURL string
	Status    uint8
}

type UpdateProfileRequest struct {
	Nickname      string `json:"nickname" binding:"omitempty,max=32"`
	AvatarMediaID *uint  `json:"avatarMediaId"`
}

type UserListRequest struct {
	ID        uint   `form:"id"`
	Mobile    string `form:"mobile"`
	Nickname  string `form:"nickname"`
	Status    *uint8 `form:"status"`
	Muted     *bool  `form:"muted"`
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
	Page      int    `form:"page" binding:"required,min=1"`
	PageSize  int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type UserListResponse struct {
	Total int64      `json:"total"`
	List  []UserItem `json:"list"`
}

type UserItem struct {
	ID          uint   `json:"id"`
	Mobile      string `json:"mobile"`
	Nickname    string `json:"nickname"`
	AvatarURL   string `json:"avatarUrl"`
	Status      uint8  `json:"status"`
	StatusText  string `json:"statusText"`
	MutedUntil  int64  `json:"mutedUntil"`
	MuteText    string `json:"muteText"`
	Deactivated bool   `json:"deactivated"`
	LastLoginAt string `json:"lastLoginAt"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type CreateUserRequest struct {
	Mobile   string `json:"mobile" binding:"required,mobile"`
	Nickname string `json:"nickname" binding:"required,max=32"`
	Status   uint8  `json:"status" binding:"oneof=0 1"`
}

// UpdateUserRequest Mobile 为空表示不修改；账号状态只能走状态接口，以便记录原因。
type UpdateUserRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Mobile   string `json:"mobile" binding:"omitempty,mobile"`
	Nickname string `json:"nickname" binding:"required,max=32"`
}

type UpdateUserStatusRequest struct {
	ID     uint   `json:"id" binding:"required"`
	Status uint8  `json:"status" binding:"oneof=0 1"`
	Reason string `json:"reason" binding:"max=200"`
}

type BatchUserStatusRequest struct {
	IDs    []uint `json:"ids" binding:"required,min=1,max=100"`
	Status uint8  `json:"status" binding:"oneof=0 1"`
	Reason string `json:"reason" binding:"max=200"`
}

type BatchUserMuteRequest struct {
	IDs      []uint `json:"ids" binding:"required,min=1,max=100"`
	Duration int    `json:"duration" binding:"min=-1,max=87600"`
	Reason   string `json:"reason" binding:"max=200"`
}

type BatchUserResult struct {
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
}

type RevealMobileResponse struct {
	Mobile string `json:"mobile"`
}

type AdminUserSubListRequest struct {
	UserID   uint `form:"userId" binding:"required"`
	Page     int  `form:"page" binding:"required,min=1"`
	PageSize int  `form:"pageSize" binding:"required,min=1,max=100"`
}

type DeactivationRequest struct {
	Code string `json:"code" binding:"required"`
}

type SendNewMobileCodeRequest struct {
	Mobile string `json:"mobile" binding:"required,mobile"`
	Ticket string `json:"ticket" binding:"required"`
}

type VerifyOldMobileCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

type VerifyOldMobileCodeResponse struct {
	Ticket string `json:"ticket"`
}

type ChangeMobileRequest struct {
	Ticket    string `json:"ticket" binding:"required"`
	NewMobile string `json:"newMobile" binding:"required,mobile"`
	NewCode   string `json:"newCode" binding:"required"`
}

type AdminDeactivateUserRequest struct {
	ID     uint   `json:"id" binding:"required"`
	Reason string `json:"reason" binding:"required,max=200"`
}

type AdminUserMuteRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Duration int    `json:"duration" binding:"min=-1,max=87600"`
	Reason   string `json:"reason" binding:"max=200"`
}

type LoginRecordItem struct {
	ID         uint   `json:"id"`
	LoginType  string `json:"loginType"`
	IP         string `json:"ip"`
	UserAgent  string `json:"userAgent"`
	Status     uint8  `json:"status"`
	StatusText string `json:"statusText"`
	Remark     string `json:"remark"`
	CreatedAt  string `json:"createdAt"`
}

type LoginRecordListResponse struct {
	Total int64             `json:"total"`
	List  []LoginRecordItem `json:"list"`
}

type UserPunishmentItem struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"userId"`
	Action       uint8  `json:"action"`
	ActionText   string `json:"actionText"`
	MuteUntil    int64  `json:"muteUntil"`
	MuteText     string `json:"muteText"`
	TargetID     uint   `json:"targetId"`
	Reason       string `json:"reason"`
	EmployeeID   uint   `json:"employeeId"`
	EmployeeName string `json:"employeeName"`
	ReportID     uint   `json:"reportId"`
	CreatedAt    string `json:"createdAt"`
}

type UserPunishmentListResponse struct {
	Total int64                `json:"total"`
	List  []UserPunishmentItem `json:"list"`
}

// UserGovernanceSummary 举报审核时展示的被举报用户现状。
type UserGovernanceSummary struct {
	UserID          uint   `json:"userId"`
	Nickname        string `json:"nickname"`
	Mobile          string `json:"mobile"`
	AvatarURL       string `json:"avatarUrl"`
	Status          uint8  `json:"status"`
	StatusText      string `json:"statusText"`
	MuteText        string `json:"muteText"`
	Deactivated     bool   `json:"deactivated"`
	ReportCount     int64  `json:"reportCount"`
	PunishmentCount int64  `json:"punishmentCount"`
}

type UserDeactivationLogItem struct {
	ID         uint   `json:"id"`
	Mobile     string `json:"mobile"`
	Source     uint8  `json:"source"`
	SourceText string `json:"sourceText"`
	EmployeeID uint   `json:"employeeId"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"createdAt"`
}

type AdminUserDetailResponse struct {
	UserItem
	DeactivationLogs []UserDeactivationLogItem `json:"deactivationLogs"`
	Rooms            []AdminUserRoomSummary    `json:"rooms"`
}
