package service

import (
	"context"

	menuEntity "github.com/sine-io/sinx/domain/menu/entity"
	menuRepo "github.com/sine-io/sinx/domain/menu/repository"
	rbacRepo "github.com/sine-io/sinx/domain/rbac/repository"
	roleEntity "github.com/sine-io/sinx/domain/role/entity"
	roleRepo "github.com/sine-io/sinx/domain/role/repository"
	userEntity "github.com/sine-io/sinx/domain/user/entity"
	userRepo "github.com/sine-io/sinx/domain/user/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// user repo mock
type mockUserRepo struct{ db *gorm.DB }

func NewMockUserRepo(db *gorm.DB) userRepo.UserRepository { return &mockUserRepo{db: db} }
func (m *mockUserRepo) Create(ctx context.Context, u *userEntity.User) error {
	return m.db.WithContext(ctx).Create(u).Error
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uint) (*userEntity.User, error) {
	var u userEntity.User
	err := m.db.WithContext(ctx).First(&u, id).Error
	return &u, err
}
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*userEntity.User, error) {
	var u userEntity.User
	err := m.db.WithContext(ctx).Where("username=?", username).First(&u).Error
	return &u, err
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	var u userEntity.User
	err := m.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	return &u, err
}
func (m *mockUserRepo) Update(ctx context.Context, u *userEntity.User) error {
	return m.db.WithContext(ctx).Save(u).Error
}
func (m *mockUserRepo) Delete(ctx context.Context, id uint) error {
	return m.db.WithContext(ctx).Delete(&userEntity.User{}, id).Error
}
func (m *mockUserRepo) List(ctx context.Context, offset, limit int) ([]*userEntity.User, error) {
	var list []*userEntity.User
	err := m.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&list).Error
	return list, err
}
func (m *mockUserRepo) Count(ctx context.Context) (int64, error) {
	var c int64
	err := m.db.WithContext(ctx).Model(&userEntity.User{}).Count(&c).Error
	return c, err
}

// role repo mock
type mockRoleRepo struct{ db *gorm.DB }

func NewMockRoleRepo(db *gorm.DB) roleRepo.RoleRepository { return &mockRoleRepo{db: db} }
func (m *mockRoleRepo) Create(ctx context.Context, r *roleEntity.Role) error {
	return m.db.WithContext(ctx).Create(r).Error
}
func (m *mockRoleRepo) GetByID(ctx context.Context, id uint) (*roleEntity.Role, error) {
	var r roleEntity.Role
	err := m.db.WithContext(ctx).First(&r, id).Error
	return &r, err
}
func (m *mockRoleRepo) Update(ctx context.Context, r *roleEntity.Role) error {
	return m.db.WithContext(ctx).Save(r).Error
}
func (m *mockRoleRepo) Delete(ctx context.Context, id uint) error {
	return m.db.WithContext(ctx).Delete(&roleEntity.Role{}, id).Error
}
func (m *mockRoleRepo) List(ctx context.Context, offset, limit int) ([]*roleEntity.Role, error) {
	var list []*roleEntity.Role
	err := m.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&list).Error
	return list, err
}
func (m *mockRoleRepo) Count(ctx context.Context) (int64, error) {
	var c int64
	err := m.db.WithContext(ctx).Model(&roleEntity.Role{}).Count(&c).Error
	return c, err
}

// menu repo mock
type mockMenuRepo struct{ db *gorm.DB }

func NewMockMenuRepo(db *gorm.DB) menuRepo.MenuRepository { return &mockMenuRepo{db: db} }
func (m *mockMenuRepo) Create(ctx context.Context, e *menuEntity.Menu) error {
	return m.db.WithContext(ctx).Create(e).Error
}
func (m *mockMenuRepo) GetByID(ctx context.Context, id uint) (*menuEntity.Menu, error) {
	var o menuEntity.Menu
	err := m.db.WithContext(ctx).First(&o, id).Error
	return &o, err
}
func (m *mockMenuRepo) Update(ctx context.Context, e *menuEntity.Menu) error {
	return m.db.WithContext(ctx).Save(e).Error
}
func (m *mockMenuRepo) Delete(ctx context.Context, id uint) error {
	return m.db.WithContext(ctx).Delete(&menuEntity.Menu{}, id).Error
}
func (m *mockMenuRepo) List(ctx context.Context, offset, limit int, name string, status *int) ([]*menuEntity.Menu, error) {
	var list []*menuEntity.Menu
	q := m.db.WithContext(ctx)
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if status != nil {
		q = q.Where("status=?", *status)
	}
	err := q.Offset(offset).Limit(limit).Find(&list).Error
	return list, err
}
func (m *mockMenuRepo) Count(ctx context.Context, name string, status *int) (int64, error) {
	q := m.db.WithContext(ctx).Model(&menuEntity.Menu{})
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if status != nil {
		q = q.Where("status=?", *status)
	}
	var c int64
	err := q.Count(&c).Error
	return c, err
}
func (m *mockMenuRepo) ListAll(ctx context.Context) ([]*menuEntity.Menu, error) {
	var list []*menuEntity.Menu
	err := m.db.WithContext(ctx).Find(&list).Error
	return list, err
}
func (m *mockMenuRepo) HasChildren(ctx context.Context, id uint) (bool, error) {
	var c int64
	if err := m.db.WithContext(ctx).Model(&menuEntity.Menu{}).Where("parent_id=?", id).Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}

// rbac repo mock
type mockRBACRepo struct{ db *gorm.DB }

func NewMockRBACRepo(db *gorm.DB) rbacRepo.RBACRepository { return &mockRBACRepo{db: db} }

type UserRole struct {
	UserID uint
	RoleID uint `gorm:"primaryKey"`
}
type RoleMenu struct {
	RoleID uint
	MenuID uint `gorm:"primaryKey"`
}

func (r *mockRBACRepo) BindUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range roleIDs {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&UserRole{UserID: userID, RoleID: id}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *mockRBACRepo) UnbindUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return r.db.WithContext(ctx).Where("user_id=? AND role_id IN ?", userID, roleIDs).Delete(&UserRole{}).Error
}
func (r *mockRBACRepo) GetUserRoles(ctx context.Context, userID uint) ([]*roleEntity.Role, error) {
	var roles []*roleEntity.Role
	err := r.db.WithContext(ctx).Table("roles r").Select("r.*").Joins("JOIN user_roles ur ON ur.role_id=r.id").Where("ur.user_id=?", userID).Scan(&roles).Error
	return roles, err
}
func (r *mockRBACRepo) BindRoleMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range menuIDs {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&RoleMenu{RoleID: roleID, MenuID: id}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *mockRBACRepo) UnbindRoleMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
	return r.db.WithContext(ctx).Where("role_id=? AND menu_id IN ?", roleID, menuIDs).Delete(&RoleMenu{}).Error
}
func (r *mockRBACRepo) GetRoleMenus(ctx context.Context, roleID uint) ([]*menuEntity.Menu, error) {
	var menus []*menuEntity.Menu
	err := r.db.WithContext(ctx).Table("menus m").Select("m.*").Joins("JOIN role_menus rm ON rm.menu_id=m.id").Where("rm.role_id=?", roleID).Scan(&menus).Error
	return menus, err
}
func (r *mockRBACRepo) GetUserMenus(ctx context.Context, userID uint) ([]*menuEntity.Menu, error) {
	var menus []*menuEntity.Menu
	err := r.db.WithContext(ctx).Table("menus m").Select("DISTINCT m.*").Joins("JOIN role_menus rm ON rm.menu_id=m.id").Joins("JOIN user_roles ur ON ur.role_id=rm.role_id").Where("ur.user_id=?", userID).Scan(&menus).Error
	return menus, err
}
func (r *mockRBACRepo) GetMenuRoles(ctx context.Context, menuID uint) ([]*roleEntity.Role, error) {
	var roles []*roleEntity.Role
	err := r.db.WithContext(ctx).Table("roles r").Select("r.*").Joins("JOIN role_menus rm ON rm.role_id=r.id").Where("rm.menu_id=?", menuID).Scan(&roles).Error
	return roles, err
}
func (r *mockRBACRepo) GetRoleUsers(ctx context.Context, roleID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&UserRole{}).Where("role_id=?", roleID).Pluck("user_id", &ids).Error
	return ids, err
}
func (r *mockRBACRepo) GetMenuIDsByRole(ctx context.Context, roleID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&RoleMenu{}).Where("role_id=?", roleID).Pluck("menu_id", &ids).Error
	return ids, err
}
