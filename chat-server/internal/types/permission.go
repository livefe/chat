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
	Type        uint8  `json:"type" binding:"required,oneof=1 2 3"`
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

type PermissionTreeRequest struct {
	// 为 true 时只返回可分配给角色的启用菜单与按钮
	Assignable bool `form:"assignable"`
}

type PermissionTreeItem struct {
	PermissionItem
	Children []*PermissionTreeItem `json:"children,omitempty"`
}
