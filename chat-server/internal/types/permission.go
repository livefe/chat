package types

const (
	SuperAdminRoleCode = "super_admin"

	CacheEmployeePermKey     = cachePrefix + "perm:employee:%d"
	CacheEmployeePermPattern = cachePrefix + "perm:employee:*"
)

type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Type        uint8  `json:"type" binding:"required,oneof=1 2 3 4"`
	ParentID    uint   `json:"parentId"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	Sort        uint   `json:"sort"`
	Status      uint8  `json:"status" binding:"oneof=0 1"`
	Icon        string `json:"icon"`
	Component   string `json:"component"`
	HideInMenu  bool   `json:"hideInMenu"`
}

type UpdatePermissionRequest struct {
	ID uint `json:"id" binding:"required"`
	CreatePermissionRequest
}

type PermissionListRequest struct {
	Name     string `form:"name"`
	Code     string `form:"code"`
	Type     *uint8 `form:"type"`
	Status   *uint8 `form:"status"`
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type PermissionListResponse struct {
	Total int64            `json:"total"`
	List  []PermissionItem `json:"list"`
}

type PermissionItem struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Type        uint8  `json:"type"`
	TypeText    string `json:"typeText"`
	ParentID    uint   `json:"parentId"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	Sort        uint   `json:"sort"`
	Status      uint8  `json:"status"`
	StatusText  string `json:"statusText"`
	Icon        string `json:"icon"`
	Component   string `json:"component"`
	HideInMenu  bool   `json:"hideInMenu"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type PermissionTreeItem struct {
	ID       uint                  `json:"id"`
	Name     string                `json:"name"`
	Code     string                `json:"code"`
	Type     uint8                 `json:"type"`
	ParentID uint                  `json:"parentId"`
	Path     string                `json:"path,omitempty"`
	Method   string                `json:"method,omitempty"`
	Sort     uint                  `json:"sort"`
	Children []*PermissionTreeItem `json:"children,omitempty"`
}
