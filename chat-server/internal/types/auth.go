package types

const (
	CacheAuthTokenEpochKey = cachePrefix + "auth:epoch:%s:%d"

	CacheAuthRefreshJTIKey = cachePrefix + "auth:rjti:%s:%d:%s"

	CacheAuthRefreshJTISetKey = cachePrefix + "auth:rjti-set:%s:%d"

	LoginTypeSMS = "sms"
)

type LoginRequest struct {
	Mobile string `json:"mobile" binding:"required,mobile"`
	Source string `json:"source" binding:"required,oneof=app_login"`
	Code   string `json:"code" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	IsFirstLogin bool     `json:"isFirstLogin"`
	User         UserInfo `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type EmployeeLoginRequest struct {
	Mobile string `json:"mobile" binding:"required,mobile"`
	Code   string `json:"code" binding:"required"`
}

type EmployeeLoginResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	Employee     EmployeeInfo `json:"employee"`
}
