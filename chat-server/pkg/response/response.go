package response

import (
	"net/http"

	"app/pkg/errs"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess = 0

	CodeBadRequest      = 40001
	CodeUnauthorized    = 40101
	CodeForbidden       = 40301
	CodeNotFound        = 40401
	CodeConflict        = 40901
	CodeTooManyRequests = 42901
	CodeInternal        = 50001
	CodeUnavailable     = 50301
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func write(c *gin.Context, httpStatus, code int, message string, data interface{}) {
	c.JSON(httpStatus, Response{Code: code, Message: message, Data: data})
}

func Fail(c *gin.Context, httpStatus, code int, message string) {
	write(c, httpStatus, code, message, nil)
}

func WriteError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	e, ok := errs.As(err)
	if !ok {
		_ = c.Error(err)
		InternalError(c, "服务器内部错误", nil)
		return
	}
	status, code, ok := statusOf(e.Code)
	if !ok {
		_ = c.Error(err)
		InternalError(c, "服务器内部错误", nil)
		return
	}
	Fail(c, status, code, e.Message)
}

func statusOf(code errs.Code) (int, int, bool) {
	switch code {
	case errs.CodeInvalid:
		return http.StatusBadRequest, CodeBadRequest, true
	case errs.CodeUnauthorized:
		return http.StatusUnauthorized, CodeUnauthorized, true
	case errs.CodeForbidden:
		return http.StatusForbidden, CodeForbidden, true
	case errs.CodeNotFound:
		return http.StatusNotFound, CodeNotFound, true
	case errs.CodeConflict:
		return http.StatusConflict, CodeConflict, true
	case errs.CodeRateLimited:
		return http.StatusTooManyRequests, CodeTooManyRequests, true
	case errs.CodeUnavailable:
		return http.StatusServiceUnavailable, CodeUnavailable, true
	default:
		return 0, 0, false
	}
}

func Success(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = "操作成功"
	}
	write(c, http.StatusOK, CodeSuccess, message, data)
}

func InternalError(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = "服务器内部错误"
	}
	write(c, http.StatusInternalServerError, CodeInternal, message, data)
}
