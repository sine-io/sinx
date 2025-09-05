package service

import (
	"context"

	"github.com/sine-io/sinx/application/user/dto"
	"github.com/sine-io/sinx/domain/user/entity"
	"github.com/sine-io/sinx/domain/user/service"
	"github.com/sine-io/sinx/pkg/auth"
	"github.com/sine-io/sinx/pkg/errorx"
)

type UserApplicationService struct {
	userDomainService *service.UserDomainService
}

func NewUserApplicationService(userDomainService *service.UserDomainService) *UserApplicationService {
	return &UserApplicationService{
		userDomainService: userDomainService,
	}
}

// Register 用户注册
func (s *UserApplicationService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	user, err := s.userDomainService.CreateUser(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(user), nil
}

// Login 用户登录
func (s *UserApplicationService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userDomainService.AuthenticateUser(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	// 生成JWT令牌
	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, errorx.New(errorx.ErrInternalServer, "failed to generate token")
	}

	return &dto.LoginResponse{
		Token: token,
		User:  *s.entityToResponse(user),
	}, nil
}

// GetProfile 获取用户资料
func (s *UserApplicationService) GetProfile(ctx context.Context, userID uint) (*dto.UserResponse, error) {
	user, err := s.userDomainService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.entityToResponse(user), nil
}

// entityToResponse 将实体转换为响应DTO
func (s *UserApplicationService) entityToResponse(user *entity.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		IsActive: user.IsActive,
	}
}
