package repository

import (
	"context"
	"errors"

	menuEntity "github.com/sine-io/sinx/domain/menu/entity"
	rbacEntity "github.com/sine-io/sinx/domain/rbac/entity"
	rbacRepo "github.com/sine-io/sinx/domain/rbac/repository"
	roleEntity "github.com/sine-io/sinx/domain/role/entity"
	"gorm.io/gorm"
)

type rbacRepositoryImpl struct{ db *gorm.DB }

func NewRBACRepository(db *gorm.DB) rbacRepo.RBACRepository { return &rbacRepositoryImpl{db: db} }

func (r *rbacRepositoryImpl) BindUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rid := range roleIDs {
			ur := &rbacEntity.UserRole{UserID: userID, RoleID: rid}
			if err := tx.Create(ur).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *rbacRepositoryImpl) UnbindUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id IN ?", userID, roleIDs).Delete(&rbacEntity.UserRole{}).Error
}

func (r *rbacRepositoryImpl) GetUserRoles(ctx context.Context, userID uint) ([]*roleEntity.Role, error) {
	var roles []*roleEntity.Role
	err := r.db.WithContext(ctx).Table("roles r").Select("r.*").Joins("JOIN user_roles ur ON ur.role_id = r.id").Where("ur.user_id = ?", userID).Scan(&roles).Error
	return roles, err
}

func (r *rbacRepositoryImpl) BindRoleMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, mid := range menuIDs {
			rm := &rbacEntity.RoleMenu{RoleID: roleID, MenuID: mid}
			if err := tx.Create(rm).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *rbacRepositoryImpl) UnbindRoleMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
	return r.db.WithContext(ctx).Where("role_id = ? AND menu_id IN ?", roleID, menuIDs).Delete(&rbacEntity.RoleMenu{}).Error
}

func (r *rbacRepositoryImpl) GetRoleMenus(ctx context.Context, roleID uint) ([]*menuEntity.Menu, error) {
	var menus []*menuEntity.Menu
	err := r.db.WithContext(ctx).Table("menus m").Select("m.*").Joins("JOIN role_menus rm ON rm.menu_id = m.id").Where("rm.role_id = ?", roleID).Scan(&menus).Error
	return menus, err
}

func (r *rbacRepositoryImpl) GetUserMenus(ctx context.Context, userID uint) ([]*menuEntity.Menu, error) {
	var menus []*menuEntity.Menu
	err := r.db.WithContext(ctx).Table("menus m").Select("DISTINCT m.*").Joins("JOIN role_menus rm ON rm.menu_id = m.id").Joins("JOIN user_roles ur ON ur.role_id = rm.role_id").Where("ur.user_id = ?", userID).Scan(&menus).Error
	return menus, err
}

func (r *rbacRepositoryImpl) GetMenuRoles(ctx context.Context, menuID uint) ([]*roleEntity.Role, error) {
	var roles []*roleEntity.Role
	err := r.db.WithContext(ctx).Table("roles r").Select("r.*").Joins("JOIN role_menus rm ON rm.role_id = r.id").Where("rm.menu_id = ?", menuID).Scan(&roles).Error
	return roles, err
}

func (r *rbacRepositoryImpl) GetRoleUsers(ctx context.Context, roleID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&rbacEntity.UserRole{}).Where("role_id = ?", roleID).Pluck("user_id", &ids).Error
	return ids, err
}

func (r *rbacRepositoryImpl) GetMenuIDsByRole(ctx context.Context, roleID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&rbacEntity.RoleMenu{}).Where("role_id = ?", roleID).Pluck("menu_id", &ids).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []uint{}, nil
	}
	return ids, err
}
