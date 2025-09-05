package router

import (
	"github.com/sine-io/sinx/api/handler"
	"github.com/sine-io/sinx/api/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, userHandler *handler.UserHandler) {
	// 设置全局中间件
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware())

	// API路由组
	api := r.Group("/api")
	{
		// 认证相关路由（不需要JWT验证）
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		// 用户相关路由（需要JWT验证）
		user := api.Group("/user")
		user.Use(middleware.AuthMiddleware())
		{
			user.GET("/profile", userHandler.GetProfile)
			// TODO: /api/user/create /api/user/list 等 RBAC 用户管理接口
		}

		// 预留角色、菜单管理路由组
		// role := api.Group("/role").Use(middleware.AuthMiddleware())
		// menu := api.Group("/menu").Use(middleware.AuthMiddleware())
		// TODO: 添加角色与菜单处理器后注册接口
	}

	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Service is running",
		})
	})

	// Swagger 文档路由 (/swagger/index.html)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
