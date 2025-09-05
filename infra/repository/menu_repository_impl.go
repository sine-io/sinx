package repository

import (
	"context"

	menuEntity "github.com/sine-io/sinx/domain/menu/entity"
	menuRepo "github.com/sine-io/sinx/domain/menu/repository"
	"gorm.io/gorm"
)

type menuRepositoryImpl struct{ db *gorm.DB }

func NewMenuRepository(db *gorm.DB) menuRepo.MenuRepository { return &menuRepositoryImpl{db: db} }

func (r *menuRepositoryImpl) Create(ctx context.Context, menu *menuEntity.Menu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}
func (r *menuRepositoryImpl) Update(ctx context.Context, menu *menuEntity.Menu) error {
	return r.db.WithContext(ctx).Save(menu).Error
}
func (r *menuRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&menuEntity.Menu{}, id).Error
}
func (r *menuRepositoryImpl) GetByID(ctx context.Context, id uint) (*menuEntity.Menu, error) {
	var m menuEntity.Menu
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *menuRepositoryImpl) List(ctx context.Context, offset, limit int, name string, status *int) ([]*menuEntity.Menu, error) {
	var menus []*menuEntity.Menu
	q := r.db.WithContext(ctx).Model(&menuEntity.Menu{})
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	err := q.Offset(offset).Limit(limit).Order("order_num ASC").Find(&menus).Error
	return menus, err
}
func (r *menuRepositoryImpl) Count(ctx context.Context, name string, status *int) (int64, error) {
	var c int64
	q := r.db.WithContext(ctx).Model(&menuEntity.Menu{})
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	err := q.Count(&c).Error
	return c, err
}
func (r *menuRepositoryImpl) ListAll(ctx context.Context) ([]*menuEntity.Menu, error) {
	var menus []*menuEntity.Menu
	err := r.db.WithContext(ctx).Order("order_num ASC").Find(&menus).Error
	return menus, err
}
func (r *menuRepositoryImpl) HasChildren(ctx context.Context, id uint) (bool, error) {
	var c int64
	if err := r.db.WithContext(ctx).Model(&menuEntity.Menu{}).Where("parent_id = ?", id).Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}
