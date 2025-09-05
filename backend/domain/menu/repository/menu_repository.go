package repository

import (
	"context"

	"github.com/sine-io/sinx/domain/menu/entity"
)

type MenuRepository interface {
	Create(ctx context.Context, menu *entity.Menu) error
	Update(ctx context.Context, menu *entity.Menu) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.Menu, error)
	List(ctx context.Context, offset, limit int, name string, status *int) ([]*entity.Menu, error)
	Count(ctx context.Context, name string, status *int) (int64, error)
	ListAll(ctx context.Context) ([]*entity.Menu, error)
	HasChildren(ctx context.Context, id uint) (bool, error)
}
