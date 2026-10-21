package middleware

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"app/pkg/logger"

	"github.com/gin-gonic/gin"
)

const (
	maxBodyLogLen    = 1024
	headerXRequestID = "X-Request-ID"
)

func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := ClientRequestID(c)
		c.Set(logger.TraceIDKey, requestID)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), logger.TraceIDKey, requestID))

		fields := []logger.Field{
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("query", redactQuery(c.Request.URL.RawQuery)),
			logger.String("ip", c.ClientIP()),
			logger.String("user_agent", c.Request.UserAgent()),
			logger.String("client_platform", ClientPlatform(c)),
			logger.String("client_version", ClientVersion(c)),
			logger.String("device_id", ClientDeviceID(c)),
		}
		if body := readBody(c); body != "" {
			fields = append(fields, logger.String("body", redactBody(body)))
		}
		log.Info(c.Request.Context(), "HTTP请求", fields...)

		c.Next()

		resp := []logger.Field{
			logger.Int("status", c.Writer.Status()),
			logger.Duration("latency", time.Since(start)),
			logger.Int("size", c.Writer.Size()),
		}
		switch {
		case len(c.Errors) > 0:
			log.Error(c.Request.Context(), "HTTP响应", append(resp, logger.String("errors", c.Errors.String()))...)
		case c.Writer.Status() >= 400:
			log.Warn(c.Request.Context(), "HTTP响应", resp...)
		default:
			log.Info(c.Request.Context(), "HTTP响应", resp...)
		}
	}
}

func readBody(c *gin.Context) string {
	m := c.Request.Method
	if m != "POST" && m != "PUT" && m != "PATCH" {
		return ""
	}
	if ct := c.GetHeader("Content-Type"); strings.HasPrefix(strings.ToLower(ct), "multipart/") {
		return "[multipart]"
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	if len(body) > maxBodyLogLen {
		return string(body[:maxBodyLogLen]) + "...(truncated)"
	}
	return string(body)
}

func redactQuery(raw string) string {
	if raw == "" {
		return ""
	}
	parts := strings.Split(raw, "&")
	for i, p := range parts {
		eq := strings.IndexByte(p, '=')
		if eq <= 0 {
			continue
		}
		key := strings.ToLower(p[:eq])
		if key == "token" || key == "access_token" || key == "refresh_token" || key == "authorization" {
			parts[i] = p[:eq+1] + "***"
		}
	}
	return strings.Join(parts, "&")
}

func redactBody(body string) string {
	lower := strings.ToLower(body)
	sensitive := []string{`"code"`, `"password"`, `"refreshToken"`, `"accessToken"`, `"token"`}
	for _, key := range sensitive {
		if strings.Contains(lower, strings.ToLower(key)) {
			return "[redacted]"
		}
	}
	return body
}
