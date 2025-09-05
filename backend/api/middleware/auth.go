package middleware

import (
	"strings"

	"github.com/sine-io/sinx/pkg/auth"
	"github.com/sine-io/sinx/pkg/errorx"
	"github.com/sine-io/sinx/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey           = "user_id"
	UsernameKey         = "username"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			response.ErrorWithCode(c, errorx.ErrUnauthorized)
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			response.ErrorWithCode(c, errorx.ErrUnauthorized)
			c.Abort()
			return
		}

		tokenString := authHeader[len(BearerPrefix):]
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			response.ErrorWithCode(c, errorx.ErrUserInvalidToken)
			c.Abort()
			return
		}

		// 将用户信息设置到上下文中
		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)

		c.Next()
	})
}

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}

// GetUsername 从上下文中获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get(UsernameKey)
	if !exists {
		return "", false
	}

	name, ok := username.(string)
	return name, ok
}
