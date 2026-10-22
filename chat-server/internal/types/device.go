package types

const (
	DevicePlatformAndroid = "android"
	DevicePlatformIOS     = "ios"

	PushVendorHuawei = "huawei"
	PushVendorHonor  = "honor"
	PushVendorXiaomi = "xiaomi"
	PushVendorOppo   = "oppo"
	PushVendorVivo   = "vivo"
	PushVendorFCM    = "fcm"
	PushVendorOther  = "other"
)

type DeviceRegisterRequest struct {
	Vendor    string `json:"vendor" binding:"required,oneof=huawei honor xiaomi oppo vivo fcm other"`
	PushToken string `json:"pushToken" binding:"required,max=512"`
	OSVersion string `json:"osVersion" binding:"omitempty,max=32"`
}

type AdminDeviceItem struct {
	ID           uint   `json:"id"`
	DeviceID     string `json:"deviceId"`
	Platform     string `json:"platform"`
	Vendor       string `json:"vendor"`
	AppVersion   string `json:"appVersion"`
	OSVersion    string `json:"osVersion"`
	PushEnabled  bool   `json:"pushEnabled"`
	LastActiveAt int64  `json:"lastActiveAt"`
	CreatedAt    string `json:"createdAt"`
}

type AdminDeviceListResponse struct {
	Total int64             `json:"total"`
	List  []AdminDeviceItem `json:"list"`
}

type DeviceRegisterCommand struct {
	DeviceID   string
	Platform   string
	Vendor     string
	PushToken  string
	AppVersion string
	OSVersion  string
}
