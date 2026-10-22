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
	ID           *uint  `form:"id"`
	Mobile       string `form:"mobile"`
	Name         string `form:"name"`
	EmployeeNo   string `form:"employeeNo"`
	DepartmentID *uint  `form:"departmentId"`
	RoleID       *uint  `form:"roleId"`
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
	LastLoginAt  string              `json:"lastLoginAt"`
	LastLoginIP  string              `json:"lastLoginIp"`
	CreatedAt    string              `json:"createdAt"`
	UpdatedAt    string              `json:"updatedAt"`
}

type EmployeeRoleBrief struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Status uint8  `json:"status"`
}

type CreateEmployeeRequest struct {
	Mobile       string `json:"mobile" binding:"required,mobile"`
	Name         string `json:"name" binding:"required,max=32"`
	EmployeeNo   string `json:"employeeNo" binding:"max=32"`
	DepartmentID uint   `json:"departmentId" binding:"required"`
	Status       uint8  `json:"status" binding:"oneof=0 1"`
	Remark       string `json:"remark" binding:"max=255"`
}

// 头像只能由员工本人修改，状态只能通过状态接口修改。
type UpdateEmployeeRequest struct {
	ID           uint   `json:"id" binding:"required"`
	Mobile       string `json:"mobile" binding:"required,mobile"`
	Name         string `json:"name" binding:"required,max=32"`
	EmployeeNo   string `json:"employeeNo" binding:"max=32"`
	DepartmentID uint   `json:"departmentId" binding:"required"`
	Remark       string `json:"remark" binding:"max=255"`
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

// 不传 employeeId 时只返回启用角色（用于筛选）。
type EmployeeRoleOptionsRequest struct {
	EmployeeID uint `form:"employeeId"`
}

// Options 为全部启用角色，加上该员工已分配但已禁用的角色。
type EmployeeRoleOptionsResponse struct {
	Options         []EmployeeRoleBrief `json:"options"`
	AssignedIDs     []uint              `json:"assignedIds"`
	OperatorIsSuper bool                `json:"operatorIsSuper"`
	TargetIsSuper   bool                `json:"targetIsSuper"`
}

type EmployeeLoginRecordListRequest struct {
	EmployeeID uint `form:"employeeId" binding:"required"`
	Page       int  `form:"page" binding:"required,min=1"`
	PageSize   int  `form:"pageSize" binding:"required,min=1,max=100"`
}

type EmployeeLoginRecordItem struct {
	ID         uint   `json:"id"`
	LoginType  string `json:"loginType"`
	IP         string `json:"ip"`
	UserAgent  string `json:"userAgent"`
	Status     uint8  `json:"status"`
	StatusText string `json:"statusText"`
	Remark     string `json:"remark"`
	CreatedAt  string `json:"createdAt"`
}

type EmployeeLoginRecordListResponse struct {
	Total int64                     `json:"total"`
	List  []EmployeeLoginRecordItem `json:"list"`
}
