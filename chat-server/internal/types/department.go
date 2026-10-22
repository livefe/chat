package types

type DepartmentItem struct {
	ID            uint   `json:"id"`
	ParentID      uint   `json:"parentId"`
	ParentName    string `json:"parentName"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	Level         uint8  `json:"level"`
	Sort          int    `json:"sort"`
	Remark        string `json:"remark"`
	Status        uint8  `json:"status"`
	StatusText    string `json:"statusText"`
	EmployeeCount int64  `json:"employeeCount"`
	// 含全部下级部门的员工数
	TotalEmployeeCount int64  `json:"totalEmployeeCount"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

type DepartmentTreeItem struct {
	DepartmentItem
	Children []DepartmentTreeItem `json:"children"`
}

type DepartmentTreeResponse struct {
	List []DepartmentTreeItem `json:"list"`
}

type DepartmentOption struct {
	ID       uint               `json:"id"`
	Name     string             `json:"name"`
	Children []DepartmentOption `json:"children,omitempty"`
}

type CreateDepartmentRequest struct {
	ParentID uint   `json:"parentId"`
	Name     string `json:"name" binding:"required,max=64"`
	Status   uint8  `json:"status" binding:"oneof=0 1"`
	Sort     int    `json:"sort"`
	Remark   string `json:"remark" binding:"max=255"`
}

type UpdateDepartmentRequest struct {
	ID uint `json:"id" binding:"required"`
	CreateDepartmentRequest
}

type UpdateDepartmentStatusRequest struct {
	ID     uint  `json:"id" binding:"required"`
	Status uint8 `json:"status" binding:"oneof=0 1"`
}

const DepartmentMaxLevel uint8 = 5
