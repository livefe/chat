package handler

import (
	"app/pkg/errs"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

func bindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		response.WriteError(c, errs.Invalid("参数错误"))
		return false
	}
	return true
}

func bindQuery(c *gin.Context, obj any) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		response.WriteError(c, errs.Invalid("参数错误"))
		return false
	}
	return true
}

func bindURI(c *gin.Context, obj any) bool {
	if err := c.ShouldBindUri(obj); err != nil {
		response.WriteError(c, errs.Invalid("参数错误"))
		return false
	}
	return true
}
