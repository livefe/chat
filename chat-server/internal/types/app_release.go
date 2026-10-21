package types

type AppReleasePublicItem struct {
	Platform     string `json:"platform"`
	PlatformText string `json:"platformText"`
	Version      string `json:"version"`
	DownloadURL  string `json:"downloadUrl"`
	FileSize     int64  `json:"fileSize"`
	Changelog    string `json:"changelog"`
}

type AppReleasePublicResponse struct {
	List []*AppReleasePublicItem `json:"list"`
}

type AppReleaseItem struct {
	ID           uint   `json:"id"`
	Platform     string `json:"platform"`
	PlatformText string `json:"platformText"`
	Version      string `json:"version"`
	DownloadURL  string `json:"downloadUrl"`
	FileSize     int64  `json:"fileSize"`
	Changelog    string `json:"changelog"`
	IsCurrent    uint8  `json:"isCurrent"`
	CreatedAt    string `json:"createdAt"`
}

type AppReleaseListResponse struct {
	List []*AppReleaseItem `json:"list"`
}

type AppReleaseAdminListRequest struct {
	Platform string `form:"platform" binding:"omitempty,oneof=android ios h5"`
}

type AppUpdateCheckRequest struct {
	Platform string
	Version  string
}

type AppUpdateCheckResponse struct {
	LatestVersion       string `json:"latestVersion"`
	MinSupportedVersion string `json:"minSupportedVersion"`
	DownloadURL         string `json:"downloadUrl"`
	FileSize            int64  `json:"fileSize"`
	HasUpdate           bool   `json:"hasUpdate"`
	ForceUpdate         bool   `json:"forceUpdate"`
	Title               string `json:"title"`
	Message             string `json:"message"`
}

type AppUpdatePolicyItem struct {
	Platform            string `json:"platform"`
	PlatformText        string `json:"platformText"`
	MinSupportedVersion string `json:"minSupportedVersion"`
	LatestVersion       string `json:"latestVersion"`
	DownloadURL         string `json:"downloadUrl"`
	Title               string `json:"title"`
	Message             string `json:"message"`
	UpdatedAt           string `json:"updatedAt"`
}

type AppUpdatePolicyListResponse struct {
	List []*AppUpdatePolicyItem `json:"list"`
}

type AppUpdatePolicySaveRequest struct {
	Platform            string `json:"platform" binding:"required,oneof=android ios h5"`
	MinSupportedVersion string `json:"minSupportedVersion" binding:"required,max=32"`
	Title               string `json:"title" binding:"max=80"`
	Message             string `json:"message" binding:"max=500"`
}

type AppReleasePublishMetaRequest struct {
	Platform    string `json:"platform" binding:"required,oneof=h5"`
	Version     string `json:"version" binding:"required,max=32"`
	DownloadURL string `json:"downloadUrl" binding:"required,max=512"`
	Changelog   string `json:"changelog" binding:"max=2000"`
}
