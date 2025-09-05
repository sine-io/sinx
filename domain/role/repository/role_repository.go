package repository

import (
	"context"

	"github.com/sine-io/sinx/domain/role/entity"
)

type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.Role, error)
	List(ctx context.Context, offset, limit int) ([]*entity.Role, error)
	Count(ctx context.Context) (int64, error)
}
