package types

type EmployeeInfo struct {
	ID         uint   `json:"id"`
	Mobile     string `json:"mobile"`
	Name       string `json:"name"`
	EmployeeNo string `json:"employeeNo"`
	AvatarURL  string `json:"avatarUrl"`
	Department string `json:"department"`
	Status     uint8  `json:"status"`
	Remark     string `json:"remark"`
}

type EmployeeListRequest struct {
	Mobile       string `form:"mobile"`
	Name         string `form:"name"`
	DepartmentID *uint  `form:"departmentId"`
	Status       *uint8 `form:"status"`
	StartTime    string `form:"startTime"`
	EndTime      string `form:"endTime"`
	Page         int    `form:"page" binding:"required,min=1"`
	PageSize     int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type EmployeeListResponse struct {
	Total int64          `json:"total"`
	List  []EmployeeItem `json:"list"`
}

type EmployeeItem struct {
	ID           uint                `json:"id"`
	Mobile       string              `json:"mobile"`
	Name         string              `json:"name"`
	EmployeeNo   string              `json:"employeeNo"`
	AvatarURL    string              `json:"avatarUrl"`
	DepartmentID uint                `json:"departmentId"`
	Department   string              `json:"department"`
	Roles        []EmployeeRoleBrief `json:"roles"`
	Status       uint8               `json:"status"`
	StatusText   string              `json:"statusText"`
	Remark       string              `json:"remark"`
	CreatedAt    string              `json:"createdAt"`
	UpdatedAt    string              `json:"updatedAt"`
}

type EmployeeRoleBrief struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type CreateEmployeeRequest struct {
	Mobile       string `json:"mobile" binding:"required,mobile"`
	Name         string `json:"name" binding:"required,max=32"`
	EmployeeNo   string `json:"employeeNo" binding:"max=32"`
	AvatarURL    string `json:"avatarUrl" binding:"max=255"`
	DepartmentID uint   `json:"departmentId" binding:"required"`
	Status       uint8  `json:"status" binding:"oneof=0 1"`
	Remark       string `json:"remark" binding:"max=255"`
}

type UpdateEmployeeRequest struct {
	ID uint `json:"id" binding:"required"`
	CreateEmployeeRequest
}

type UpdateEmployeeStatusRequest struct {
	ID     uint  `json:"id" binding:"required"`
	Status uint8 `json:"status" binding:"oneof=0 1"`
}

type UpdateMyProfileRequest struct {
	AvatarURL string `json:"avatarUrl" binding:"max=255"`
}

type AssignEmployeeRolesRequest struct {
	EmployeeID uint   `json:"employeeId" binding:"required"`
	RoleIDs    []uint `json:"roleIds" binding:"required"`
}
