package repository

import (
	"context"

	menuEntity "github.com/sine-io/sinx/domain/menu/entity"
	rbacEntity "github.com/sine-io/sinx/domain/rbac/entity"
	roleEntity "github.com/sine-io/sinx/domain/role/entity"
)

type RBACRepository interface {
	BindUserRoles(ctx context.Context, userID uint, roleIDs []uint) error
	UnbindUserRoles(ctx context.Context, userID uint, roleIDs []uint) error
	GetUserRoles(ctx context.Context, userID uint) ([]*roleEntity.Role, error)
	BindRoleMenus(ctx context.Context, roleID uint, menuIDs []uint) error
	UnbindRoleMenus(ctx context.Context, roleID uint, menuIDs []uint) error
	GetRoleMenus(ctx context.Context, roleID uint) ([]*menuEntity.Menu, error)
	GetUserMenus(ctx context.Context, userID uint) ([]*menuEntity.Menu, error)
	GetMenuRoles(ctx context.Context, menuID uint) ([]*roleEntity.Role, error)
	GetRoleUsers(ctx context.Context, roleID uint) ([]uint, error)
	GetMenuIDsByRole(ctx context.Context, roleID uint) ([]uint, error)
}

// 复用实体定义，避免循环引用
type UserRole = rbacEntity.UserRole
type RoleMenu = rbacEntity.RoleMenu
