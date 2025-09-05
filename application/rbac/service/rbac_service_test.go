//go:build cgo
// +build cgo

package service

import (
	"context"
	"testing"

	menuEntity "github.com/sine-io/sinx/domain/menu/entity"
	roleEntity "github.com/sine-io/sinx/domain/role/entity"
	userEntity "github.com/sine-io/sinx/domain/user/entity"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// initTestService 初始化内存数据库及服务
func initTestService(t *testing.T) *RBACApplicationService {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&userEntity.User{}, &roleEntity.Role{}, &menuEntity.Menu{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// 关联表手动迁移
	type UserRole struct {
		UserID uint
		RoleID uint
	}
	type RoleMenu struct {
		RoleID uint
		MenuID uint
	}
	_ = db.AutoMigrate(&UserRole{}, &RoleMenu{})

	userRepo := NewMockUserRepo(db)
	roleRepo := NewMockRoleRepo(db)
	menuRepo := NewMockMenuRepo(db)
	rbacRepo := NewMockRBACRepo(db)
	return NewRBACApplicationService(userRepo, roleRepo, menuRepo, rbacRepo)
}

func TestBindAndPerms(t *testing.T) {
	svc := initTestService(t)
	ctx := context.Background()
	// 创建用户/角色/菜单
	_ = svc.userRepository.Create(ctx, &userEntity.User{Username: "u1", Password: "p"})
	_ = svc.roleRepository.Create(ctx, &roleEntity.Role{Name: "r1", Status: 1})
	_ = svc.menuRepository.Create(ctx, &menuEntity.Menu{Name: "m1", ParentID: 0, OrderNum: 1, MenuType: "B", Perms: "user:create", Status: 1})
	// 绑定角色菜单
	_ = svc.BindRoleMenus(ctx, 1, []uint{1})
	// 绑定用户角色
	_ = svc.BindUserRoles(ctx, 1, []uint{1})
	perms, err := svc.GetUserPerms(ctx, 1)
	if err != nil {
		t.Fatalf("get perms err: %v", err)
	}
	if _, ok := perms["user:create"]; !ok {
		t.Fatalf("expected perm not found")
	}
	// 缓存命中再取一次
	perms2, _ := svc.GetUserPerms(ctx, 1)
	if len(perms2) != len(perms) {
		t.Fatalf("cache mismatch")
	}
}
