package types

const CacheSensitiveWordReloadChannel = cachePrefix + "sensitive_word:reload"

const SensitiveWordMaxLen = 64

type SensitiveWordListRequest struct {
	Word     string `form:"word"`
	Category string `form:"category"`
	Status   *uint8 `form:"status"`
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type SensitiveWordItem struct {
	ID           uint   `json:"id"`
	Word         string `json:"word"`
	Category     string `json:"category"`
	CategoryText string `json:"categoryText"`
	Status       uint8  `json:"status"`
	StatusText   string `json:"statusText"`
	Remark       string `json:"remark"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type SensitiveWordListResponse struct {
	Total int64               `json:"total"`
	List  []SensitiveWordItem `json:"list"`
}

type CreateSensitiveWordRequest struct {
	Word     string `json:"word" binding:"required,max=64"`
	Category string `json:"category" binding:"required,max=32"`
	Status   uint8  `json:"status" binding:"oneof=0 1"`
	Remark   string `json:"remark" binding:"omitempty,max=255"`
}

type UpdateSensitiveWordRequest struct {
	ID uint `json:"id" binding:"required"`
	CreateSensitiveWordRequest
}

type UpdateSensitiveWordStatusRequest struct {
	ID     uint  `json:"id" binding:"required"`
	Status uint8 `json:"status" binding:"oneof=0 1"`
}

type BatchSensitiveWordRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=500,dive,required"`
}

type BatchSensitiveWordStatusRequest struct {
	IDs    []uint `json:"ids" binding:"required,min=1,max=500,dive,required"`
	Status uint8  `json:"status" binding:"oneof=0 1"`
}

type ImportSensitiveWordsRequest struct {
	Words    string `json:"words" binding:"required"`
	Category string `json:"category" binding:"required,max=32"`
}

// 跳过的词按原因分组返回，便于管理员核对。
type ImportSensitiveWordsResponse struct {
	Created    int      `json:"created"`
	Existing   []string `json:"existing"`
	TooLong    []string `json:"tooLong"`
	Duplicated []string `json:"duplicated"`
	Invalid    []string `json:"invalid"`
}
