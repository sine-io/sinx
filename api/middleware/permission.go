package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sine-io/sinx/pkg/errorx"
	"github.com/sine-io/sinx/pkg/response"
)

// PermissionChecker 定义一个函数类型，从上下文和 perms 判断是否允许
type PermissionChecker func(c *gin.Context, required string) bool

// PermissionMiddleware 基于路由元数据(自定义)的权限中间件
func PermissionMiddleware(required string, checker PermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.TrimSpace(required) == "" {
			c.Next()
			return
		}
		if checker != nil && checker(c, required) {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, response.Response{Code: errorx.ErrForbidden, Message: errorx.GetErrorMessage(errorx.ErrForbidden)})
	}
}
