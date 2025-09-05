package repository

import (
	"context"

	roleEntity "github.com/sine-io/sinx/domain/role/entity"
	roleRepo "github.com/sine-io/sinx/domain/role/repository"
	"gorm.io/gorm"
)

type roleRepositoryImpl struct{ db *gorm.DB }

func NewRoleRepository(db *gorm.DB) roleRepo.RoleRepository { return &roleRepositoryImpl{db: db} }

func (r *roleRepositoryImpl) Create(ctx context.Context, role *roleEntity.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}
func (r *roleRepositoryImpl) Update(ctx context.Context, role *roleEntity.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}
func (r *roleRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&roleEntity.Role{}, id).Error
}
func (r *roleRepositoryImpl) GetByID(ctx context.Context, id uint) (*roleEntity.Role, error) {
	var role roleEntity.Role
	if err := r.db.WithContext(ctx).First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
func (r *roleRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*roleEntity.Role, error) {
	var roles []*roleEntity.Role
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("id DESC").Find(&roles).Error
	return roles, err
}
func (r *roleRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var c int64
	err := r.db.WithContext(ctx).Model(&roleEntity.Role{}).Count(&c).Error
	return c, err
}
