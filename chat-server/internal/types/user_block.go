package types

type BlockUserRequest struct {
	UserID uint `json:"userId" binding:"required"`
}

type BlockListRequest struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

type BlockListItem struct {
	UserID    uint   `json:"userId"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
	CreatedAt string `json:"createdAt"`
}

type BlockListResponse struct {
	Total int64           `json:"total"`
	List  []BlockListItem `json:"list"`
}
