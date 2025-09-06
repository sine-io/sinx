package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sine-io/sinx/api/handler"
	"github.com/sine-io/sinx/api/middleware"
	"github.com/sine-io/sinx/pkg/permissions"
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
		// 简单权限检查器（每次实时查询，可后续增加缓存）
		permChecker := func(c *gin.Context, required string) bool {
			uid, ok := middleware.GetUserID(c)
			if !ok {
				return false
			}
			perms, err := rbacHandler.Service().GetUserPerms(c, uid)
			if err != nil {
				return false
			}
			_, exist := perms[required]
			return exist
		}

		user := api.Group("/user")
		user.Use(middleware.AuthMiddleware())
		{
			user.GET("/profile", userHandler.GetProfile)
			user.POST("/create", middleware.PermissionMiddleware("user:create", permChecker), rbacHandler.CreateUser)
			user.GET("/list", middleware.PermissionMiddleware("user:list", permChecker), rbacHandler.UserList)
			user.POST("/update", middleware.PermissionMiddleware("user:update", permChecker), rbacHandler.UpdateUser)
			user.POST("/delete", middleware.PermissionMiddleware("user:delete", permChecker), rbacHandler.DeleteUser)
			user.POST("/changePassword", rbacHandler.ChangePassword)
			user.POST("/bindRole", middleware.PermissionMiddleware("user:bindRole", permChecker), rbacHandler.BindUserRole)
			user.POST("/unbindRole", middleware.PermissionMiddleware("user:unbindRole", permChecker), rbacHandler.UnbindUserRole)
			user.GET("/roles", middleware.PermissionMiddleware("user:roles", permChecker), rbacHandler.GetUserRoles)
			user.GET("/menus", rbacHandler.GetUserMenus)
		}

		role := api.Group("/role").Use(middleware.AuthMiddleware())
		{
			role.POST("/create", middleware.PermissionMiddleware("role:create", permChecker), rbacHandler.CreateRole)
			role.GET("/list", middleware.PermissionMiddleware("role:list", permChecker), rbacHandler.RoleList)
			role.POST("/update", middleware.PermissionMiddleware("role:update", permChecker), rbacHandler.UpdateRole)
			role.POST("/delete", middleware.PermissionMiddleware("role:delete", permChecker), rbacHandler.DeleteRole)
			role.POST("/bindMenu", middleware.PermissionMiddleware("role:bindMenu", permChecker), rbacHandler.BindRoleMenu)
			role.POST("/unbindMenu", middleware.PermissionMiddleware("role:unbindMenu", permChecker), rbacHandler.UnbindRoleMenu)
			role.GET("/menus", middleware.PermissionMiddleware("role:menus", permChecker), rbacHandler.GetRoleMenus)
			role.GET("/users", middleware.PermissionMiddleware("role:users", permChecker), rbacHandler.GetRoleUsers)
		}

		menu := api.Group("/menu").Use(middleware.AuthMiddleware())
		{
			menu.POST("/create", middleware.PermissionMiddleware("menu:create", permChecker), rbacHandler.CreateMenu)
			menu.GET("/list", middleware.PermissionMiddleware("menu:list", permChecker), rbacHandler.MenuList)
			menu.POST("/update", middleware.PermissionMiddleware("menu:update", permChecker), rbacHandler.UpdateMenu)
			menu.POST("/delete", middleware.PermissionMiddleware("menu:delete", permChecker), rbacHandler.DeleteMenu)
			menu.GET("/roles", middleware.PermissionMiddleware("menu:roles", permChecker), rbacHandler.MenuRoles)
			menu.GET("/tree", rbacHandler.MenuTree)
			menu.GET("/roleMenuTree", middleware.PermissionMiddleware("menu:roleMenuTree", permChecker), rbacHandler.GetRoleMenuTree)
		}

		// 导出所有权限（需登录，便于前端动态渲染）
		api.GET("/perms/all", middleware.AuthMiddleware(), func(c *gin.Context) {
			c.JSON(200, gin.H{"code": 0, "data": permissions.AllPerms})
		})

		// 返回当前用户拥有的权限标识集合（前端可用于按钮/接口按需请求）
		api.GET("/perms/me", middleware.AuthMiddleware(), func(c *gin.Context) {
			uid, ok := middleware.GetUserID(c)
			if !ok {
				c.JSON(401, gin.H{"code": 10003, "message": "未认证"})
				return
			}
			perms, err := rbacHandler.Service().GetUserPerms(c, uid)
			if err != nil {
				c.JSON(500, gin.H{"code": 1, "message": err.Error()})
				return
			}
			// 转为 slice
			list := make([]string, 0, len(perms))
			for k := range perms {
				list = append(list, k)
			}
			c.JSON(200, gin.H{"code": 0, "data": list})
		})
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
