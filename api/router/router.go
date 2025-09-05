package router

import (
	"github.com/sine-io/sinx/api/handler"
	"github.com/sine-io/sinx/api/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, userHandler *handler.UserHandler, rbacHandler *handler.RBACHandler) {
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
			user.POST("/create", rbacHandler.CreateUser)
			user.GET("/list", rbacHandler.UserList)
			user.POST("/update", rbacHandler.UpdateUser)
			user.POST("/delete", rbacHandler.DeleteUser)
			user.POST("/changePassword", rbacHandler.ChangePassword)
			user.POST("/bindRole", rbacHandler.BindUserRole)
			user.POST("/unbindRole", rbacHandler.UnbindUserRole)
			user.GET("/roles", rbacHandler.GetUserRoles)
			user.GET("/menus", rbacHandler.GetUserMenus)
		}

		role := api.Group("/role").Use(middleware.AuthMiddleware())
		{
			role.POST("/create", rbacHandler.CreateRole)
			role.GET("/list", rbacHandler.RoleList)
			role.POST("/update", rbacHandler.UpdateRole)
			role.POST("/delete", rbacHandler.DeleteRole)
			role.POST("/bindMenu", rbacHandler.BindRoleMenu)
			role.POST("/unbindMenu", rbacHandler.UnbindRoleMenu)
			role.GET("/menus", rbacHandler.GetRoleMenus)
			role.GET("/users", rbacHandler.GetRoleUsers)
		}

		menu := api.Group("/menu").Use(middleware.AuthMiddleware())
		{
			menu.POST("/create", rbacHandler.CreateMenu)
			menu.GET("/list", rbacHandler.MenuList)
			menu.POST("/update", rbacHandler.UpdateMenu)
			menu.POST("/delete", rbacHandler.DeleteMenu)
			menu.GET("/roles", rbacHandler.MenuRoles)
			menu.GET("/tree", rbacHandler.MenuTree)
			menu.GET("/roleMenuTree", rbacHandler.GetRoleMenuTree)
		}
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
