package middleware

import (
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error(c.Request.Context(), "HTTP请求异常",
					logger.Any("error", err),
					logger.String("method", c.Request.Method),
					logger.String("path", c.Request.URL.Path),
					logger.String("ip", c.ClientIP()),
				)

				response.InternalError(c, "服务器内部错误", nil)
				c.Abort()
			}
		}()

		c.Next()
	}
}
