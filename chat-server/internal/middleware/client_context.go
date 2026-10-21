package middleware

import (
	"strings"
	"unicode/utf8"

	"app/pkg/errs"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	headerClientPlatform = "X-Client-Platform"
	headerClientVersion  = "X-Client-Version"
	headerDeviceID       = "X-Device-ID"

	platformAndroid = "android"
	platformIOS     = "ios"
	platformH5      = "h5"
	platformAdmin   = "admin"

	ctxKeyClientPlatform = "client_platform"
	ctxKeyClientVersion  = "client_version"
	ctxKeyDeviceID       = "client_device_id"
	ctxKeyRequestID      = "client_request_id"
)

func ClientContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestID, platform, version, deviceID string
		if isWebSocketUpgrade(c) {

			requestID = strings.TrimSpace(c.Query("requestId"))
			platform = strings.ToLower(strings.TrimSpace(c.Query("clientPlatform")))
			version = strings.TrimSpace(c.Query("clientVersion"))
			deviceID = strings.TrimSpace(c.Query("deviceId"))
		} else {
			requestID = strings.TrimSpace(c.GetHeader(headerXRequestID))
			platform = strings.ToLower(strings.TrimSpace(c.GetHeader(headerClientPlatform)))
			version = strings.TrimSpace(c.GetHeader(headerClientVersion))
			deviceID = strings.TrimSpace(c.GetHeader(headerDeviceID))
		}

		if requestID == "" || utf8.RuneCountInString(requestID) > 64 {
			response.WriteError(c, errs.Invalid("缺少或非法的请求 ID"))
			c.Abort()
			return
		}
		if !validClientPlatform(platform) {
			response.WriteError(c, errs.Invalid("缺少或非法的客户端平台"))
			c.Abort()
			return
		}
		if version == "" || utf8.RuneCountInString(version) > 32 {
			response.WriteError(c, errs.Invalid("缺少或非法的客户端版本"))
			c.Abort()
			return
		}
		if deviceID == "" || utf8.RuneCountInString(deviceID) > 64 {
			response.WriteError(c, errs.Invalid("缺少或非法的设备标识"))
			c.Abort()
			return
		}

		c.Set(ctxKeyRequestID, requestID)
		c.Set(headerXRequestID, requestID)
		c.Header(headerXRequestID, requestID)
		c.Set(ctxKeyClientPlatform, platform)
		c.Set(ctxKeyClientVersion, version)
		c.Set(ctxKeyDeviceID, deviceID)
		c.Next()
	}
}

func isWebSocketUpgrade(c *gin.Context) bool {
	return strings.EqualFold(c.GetHeader("Upgrade"), "websocket")
}

func ClientPlatform(c *gin.Context) string {
	v, _ := c.Get(ctxKeyClientPlatform)
	s, _ := v.(string)
	return s
}

func ClientVersion(c *gin.Context) string {
	v, _ := c.Get(ctxKeyClientVersion)
	s, _ := v.(string)
	return s
}

func ClientDeviceID(c *gin.Context) string {
	v, _ := c.Get(ctxKeyDeviceID)
	s, _ := v.(string)
	return s
}

func ClientRequestID(c *gin.Context) string {
	v, _ := c.Get(ctxKeyRequestID)
	s, _ := v.(string)
	return s
}

func validClientPlatform(p string) bool {
	switch p {
	case platformAndroid, platformIOS, platformH5, platformAdmin:
		return true
	default:
		return false
	}
}

func ValidAppPlatform(p string) bool {
	switch p {
	case platformAndroid, platformIOS, platformH5:
		return true
	default:
		return false
	}
}

func ValidDevicePlatform(p string) bool {
	switch p {
	case platformAndroid, platformIOS:
		return true
	default:
		return false
	}
}
