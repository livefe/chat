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

type DeviceRegisterCommand struct {
	DeviceID   string
	Platform   string
	Vendor     string
	PushToken  string
	AppVersion string
	OSVersion  string
}
