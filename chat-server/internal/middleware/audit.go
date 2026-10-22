package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"app/internal/service"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

// 高频聊天类接口不记录操作审计，格式为 "METHOD 路由"。
var auditExcludedRoutes = []string{
	"POST /api/admin/im/conversations/dm",
	"POST /api/admin/im/messages",
	"POST /api/admin/im/read",
	"POST /api/admin/support/messages",
	"POST /api/admin/support/read",
}

type captureResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *captureResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func AuditMiddleware(log *logger.Logger, auditSvc *service.AuditLogService) gin.HandlerFunc {
	auditSvc.ExcludeRoutes(auditExcludedRoutes...)
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost || auditSvc.Excluded(c.Request.Method, c.FullPath()) {
			c.Next()
			return
		}

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		writer := &captureResponseWriter{ResponseWriter: c.Writer, body: bytes.NewBuffer(nil)}
		c.Writer = writer
		c.Next()

		employeeID := CurrentEmployeeID(c)
		if employeeID == 0 {
			return
		}

		input := service.AuditLogInput{
			EmployeeID:    employeeID,
			RequestPath:   c.FullPath(),
			RequestMethod: c.Request.Method,
			RequestBody:   append([]byte(nil), bodyBytes...),
			IP:            c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
		}
		var resp response.Response
		if err := json.Unmarshal(writer.body.Bytes(), &resp); err == nil {
			input.Success = writer.Status() == http.StatusOK && resp.Code == response.CodeSuccess
			input.ResultCode = resp.Code
			input.ResultMessage = resp.Message
		} else {
			input.ResultCode = writer.Status()
			input.ResultMessage = "HTTP " + strconv.Itoa(writer.Status())
		}
		go auditSvc.Record(context.Background(), input)
	}
}
