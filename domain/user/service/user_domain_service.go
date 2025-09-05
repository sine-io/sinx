package service

import (
	"context"

	"github.com/sine-io/sinx/domain/user/entity"
	"github.com/sine-io/sinx/domain/user/repository"
	"github.com/sine-io/sinx/pkg/errorx"
	"github.com/sine-io/sinx/pkg/utils"
)

type UserDomainService struct {
	userRepo repository.UserRepository
}

func NewUserDomainService(userRepo repository.UserRepository) *UserDomainService {
	return &UserDomainService{
		userRepo: userRepo,
	}
}

// CreateUser 创建用户
func (s *UserDomainService) CreateUser(ctx context.Context, username, email, password string) (*entity.User, error) {
	// 检查用户名是否已存在
	existingUser, _ := s.userRepo.GetByUsername(ctx, username)
	if existingUser != nil {
		return nil, errorx.NewWithCode(errorx.ErrUserAlreadyExists)
	}

	// 检查邮箱是否已存在
	existingUser, _ = s.userRepo.GetByEmail(ctx, email)
	if existingUser != nil {
		return nil, errorx.NewWithCode(errorx.ErrUserAlreadyExists)
	}

	// 对密码进行哈希处理
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, errorx.New(errorx.ErrInternalServer, "failed to hash password")
	}

	user := &entity.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		IsActive: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser 验证用户登录
func (s *UserDomainService) AuthenticateUser(ctx context.Context, username, password string) (*entity.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errorx.NewWithCode(errorx.ErrUserNotFound)
	}

	if !user.IsActive {
		return nil, errorx.NewWithCode(errorx.ErrUserNotFound)
	}

	if !utils.CheckPassword(password, user.Password) {
		return nil, errorx.NewWithCode(errorx.ErrUserInvalidPassword)
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserDomainService) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errorx.NewWithCode(errorx.ErrUserNotFound)
	}

	if !user.IsActive {
		return nil, errorx.NewWithCode(errorx.ErrUserNotFound)
	}

	return user, nil
}
