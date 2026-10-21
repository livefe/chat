package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"app/internal/service"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type captureResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *captureResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func AuditMiddleware(log *logger.Logger, auditSvc *service.AuditLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
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

		success := false
		if writer.Status() == http.StatusOK {
			var resp response.Response
			if err := json.Unmarshal(writer.body.Bytes(), &resp); err == nil && resp.Code == response.CodeSuccess {
				success = true
			}
		}

		input := service.AuditLogInput{
			EmployeeID:    employeeID,
			RequestPath:   c.FullPath(),
			RequestMethod: c.Request.Method,
			RequestBody:   append([]byte(nil), bodyBytes...),
			IP:            c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
			Success:       success,
		}
		go auditSvc.Record(context.Background(), input)
	}
}
