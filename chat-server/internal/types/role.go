package types

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Status      uint8  `json:"status" binding:"oneof=0 1"`
}

type UpdateRoleRequest struct {
	ID uint `json:"id" binding:"required"`
	CreateRoleRequest
}

type RoleListRequest struct {
	Name     string `form:"name"`
	Code     string `form:"code"`
	Status   *uint8 `form:"status"`
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=200"`
}

type RoleListResponse struct {
	Total int64      `json:"total"`
	List  []RoleItem `json:"list"`
}

type RoleItem struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      uint8  `json:"status"`
	StatusText  string `json:"statusText"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type AssignPermissionsRequest struct {
	RoleID        uint   `json:"roleId" binding:"required"`
	PermissionIDs []uint `json:"permissionIds" binding:"required"`
}
