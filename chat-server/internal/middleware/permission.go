package middleware

import (
	"app/internal/service"
	"app/pkg/errs"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

func Permission(permSvc *service.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		employeeID := CurrentEmployeeID(c)
		if employeeID == 0 {
			response.WriteError(c, errs.Unauthorized("员工未认证"))
			c.Abort()
			return
		}

		ok, err := permSvc.CheckPermission(c.Request.Context(), employeeID, c.FullPath(), c.Request.Method)
		if err != nil {
			response.WriteError(c, err)
			c.Abort()
			return
		}
		if !ok {
			response.WriteError(c, errs.Forbidden("没有访问权限"))
			c.Abort()
			return
		}
		c.Next()
	}
}
