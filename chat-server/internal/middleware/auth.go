package middleware

import (
	"context"

	"app/internal/service"
	"app/pkg/auth"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID     = "ctx:user_id"
	ContextEmployeeID = "ctx:employee_id"

	TokenErrMissing = "missing"
	TokenErrTooLong = "too_long"
	TokenErrExpired = "expired"
	TokenErrInvalid = "invalid"
	TokenErrRevoked = "revoked"
)

type AccessTokenError struct {
	Kind    string
	Message string
}

func ValidateAccessToken(a *auth.Auth, tokenString string) (*auth.Claims, *AccessTokenError) {
	if tokenString == "" {
		return nil, &AccessTokenError{Kind: TokenErrMissing, Message: "缺少认证令牌"}
	}
	if len(tokenString) > auth.MaxTokenLen {
		return nil, &AccessTokenError{Kind: TokenErrTooLong, Message: "令牌格式错误"}
	}
	claims, err := a.ParseAccessToken(tokenString)
	if err != nil {
		switch err {
		case auth.ErrExpiredToken:
			return nil, &AccessTokenError{Kind: TokenErrExpired, Message: "令牌已过期"}
		case auth.ErrInvalidToken, auth.ErrInvalidTokenType, auth.ErrInvalidSubjectType:
			return nil, &AccessTokenError{Kind: TokenErrInvalid, Message: "无效的令牌"}
		default:
			return nil, &AccessTokenError{Kind: TokenErrInvalid, Message: "认证失败"}
		}
	}
	return claims, nil
}

func AssertAccessTokenNotRevoked(ctx context.Context, authSvc *service.AuthService, claims *auth.Claims) *AccessTokenError {
	if err := authSvc.AssertNotRevoked(ctx, claims); err != nil {
		msg := err.Error()
		if msg == "" {
			msg = "令牌已失效"
		}
		return &AccessTokenError{Kind: TokenErrRevoked, Message: msg}
	}
	return nil
}

func CurrentUserID(c *gin.Context) uint {
	if v, ok := c.Get(ContextUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

func CurrentEmployeeID(c *gin.Context) uint {
	if v, ok := c.Get(ContextEmployeeID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

func authReject(log *logger.Logger, c *gin.Context, reason string, fields ...logger.Field) {
	all := append([]logger.Field{
		logger.String("reason", reason),
		logger.String("path", c.Request.URL.Path),
		logger.String("ip", c.ClientIP()),
	}, fields...)
	log.Warn(c.Request.Context(), "认证拒绝", all...)
}

func parseAccessToken(log *logger.Logger, a *auth.Auth, c *gin.Context) (*auth.Claims, bool) {
	tokenString := auth.ExtractTokenFromHeader(c.GetHeader(auth.AuthorizationHeader))
	claims, fail := ValidateAccessToken(a, tokenString)
	if fail != nil {
		fields := []logger.Field{}
		if fail.Kind == TokenErrTooLong {
			fields = append(fields, logger.Int("token_len", len(tokenString)))
		}
		authReject(log, c, fail.Message, fields...)
		response.WriteError(c, errs.Unauthorized(fail.Message))
		c.Abort()
		return nil, false
	}
	return claims, true
}

func rejectIfRevoked(log *logger.Logger, authSvc *service.AuthService, c *gin.Context, claims *auth.Claims) bool {
	if fail := AssertAccessTokenNotRevoked(c.Request.Context(), authSvc, claims); fail != nil {
		authReject(log, c, "令牌已失效", logger.Uint("subject_id", claims.SubjectID))
		response.WriteError(c, errs.Unauthorized(fail.Message))
		c.Abort()
		return true
	}
	return false
}

type ErrorWriter func(c *gin.Context, err error)

func UserAuth(log *logger.Logger, a *auth.Auth, authSvc *service.AuthService, userSvc *service.UserService, writeError ErrorWriter) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := parseAccessToken(log, a, c)
		if !ok {
			return
		}
		if claims.SubjectType != auth.SubjectTypeUser {
			authReject(log, c, "令牌主体类型错误", logger.String("subject_type", claims.SubjectType))
			response.WriteError(c, errs.Unauthorized("无效的令牌"))
			c.Abort()
			return
		}
		if claims.SubjectID == 0 {
			authReject(log, c, "用户ID无效")
			response.WriteError(c, errs.Unauthorized("用户ID无效"))
			c.Abort()
			return
		}
		if rejectIfRevoked(log, authSvc, c, claims) {
			return
		}

		if err := userSvc.AssertAuthenticatedUser(c.Request.Context(), claims.SubjectID); err != nil {
			authReject(log, c, "用户认证失败", logger.Uint("user_id", claims.SubjectID), logger.Error(err))
			writeError(c, err)
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.SubjectID)
		c.Next()
	}
}

func EmployeeAuth(log *logger.Logger, a *auth.Auth, authSvc *service.AuthService, employeeSvc *service.EmployeeService, writeError ErrorWriter) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := parseAccessToken(log, a, c)
		if !ok {
			return
		}
		if claims.SubjectType != auth.SubjectTypeEmployee {
			authReject(log, c, "令牌主体类型错误", logger.String("subject_type", claims.SubjectType))
			response.WriteError(c, errs.Unauthorized("无效的令牌"))
			c.Abort()
			return
		}
		if claims.SubjectID == 0 {
			authReject(log, c, "员工ID无效")
			response.WriteError(c, errs.Unauthorized("员工ID无效"))
			c.Abort()
			return
		}
		if rejectIfRevoked(log, authSvc, c, claims) {
			return
		}

		if err := employeeSvc.EnsureActive(c.Request.Context(), claims.SubjectID); err != nil {
			authReject(log, c, "员工认证失败", logger.Uint("employee_id", claims.SubjectID), logger.Error(err))
			writeError(c, err)
			c.Abort()
			return
		}

		c.Set(ContextEmployeeID, claims.SubjectID)
		c.Next()
	}
}
