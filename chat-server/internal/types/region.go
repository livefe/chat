package types

type RegionItem struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	ShortName  string `json:"shortName"`
	ParentCode string `json:"parentCode"`
	Level      uint8  `json:"level"`
	FullPath   string `json:"fullPath"`
	TiandituGB string `json:"tiandituGb"`
	AmapAdcode string `json:"amapAdcode"`
}

type RegionListResponse struct {
	List []RegionItem `json:"list"`
}

type RegionSearchResponse struct {
	List      []RegionItem `json:"list"`
	Truncated bool         `json:"truncated"`
}

type RegionChildrenRequest struct {
	ParentCode string `form:"parentCode"`
}

type RegionSearchRequest struct {
	Keyword string `form:"keyword" binding:"required"`
	Level   uint8  `form:"level"`
	Limit   int    `form:"limit" binding:"required,min=1"`
}

type RegionDetailRequest struct {
	Code string `form:"code" binding:"required"`
}

type RegionAdminItem struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	ShortName  string `json:"shortName"`
	ParentCode string `json:"parentCode"`
	Level      uint8  `json:"level"`
	LevelText  string `json:"levelText"`
	FullPath   string `json:"fullPath"`
	TiandituGB string `json:"tiandituGb"`
	AmapAdcode string `json:"amapAdcode"`
	Status     uint8  `json:"status"`
	StatusText string `json:"statusText"`
	DataYear   string `json:"dataYear"`
}
