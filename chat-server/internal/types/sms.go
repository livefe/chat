package types

const (
	SMSCodeLength              = 6
	SMSUserDailyMaxSendCount   = 10
	SMSSystemDailyMaxSendCount = 100
	SMSSendCooldownSeconds     = 60
	SMSVerifyMaxFailCount      = 5
	SMSVerifyFailLockSeconds   = 15 * 60

	SMSMockFixedCode = "111111"

	SMSSignName     = "app-center"
	SMSTemplateCode = "SMS_123456789"

	SourceAppLogin           = "app_login"
	SourceAdminLogin         = "admin_login"
	SourceAppDeactivate      = "app_deactivate"
	SourceAppChangeMobileOld = "app_change_mobile_old"
	SourceAppChangeMobileNew = "app_change_mobile_new"
)

const (
	CacheSMSLoginCodeKey            = cachePrefix + "sms:login_code:%s:%s"
	CacheSMSUserDailySendLimitKey   = cachePrefix + "sms:user:daily_send_limit:%s:%s"
	CacheSMSSystemDailySendLimitKey = cachePrefix + "sms:system:daily_send_limit:%s"
	CacheSMSSendCooldownKey         = cachePrefix + "sms:cooldown:%s:%s"
	CacheSMSVerifyFailKey           = cachePrefix + "sms:verify_fail:%s:%s"
)

type SendCodeRequest struct {
	Mobile string `json:"mobile" binding:"required,mobile"`
	Source string `json:"source" binding:"required,oneof=app_login admin_login app_deactivate app_change_mobile_old app_change_mobile_new"`
}
