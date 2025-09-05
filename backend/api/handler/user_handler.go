package handler

import (
	"github.com/sine-io/sinx/api/middleware"
	"github.com/sine-io/sinx/application/user/dto"
	"github.com/sine-io/sinx/application/user/service"
	"github.com/sine-io/sinx/pkg/errorx"
	"github.com/sine-io/sinx/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userAppService *service.UserApplicationService
}

func NewUserHandler(userAppService *service.UserApplicationService) *UserHandler {
	return &UserHandler{
		userAppService: userAppService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册信息"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Router /api/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}

	user, err := h.userAppService.Register(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errorx.Error); ok {
			response.Error(c, appErr)
		} else {
			response.InternalError(c, err)
		}
		return
	}

	response.Success(c, user)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取令牌
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Failure 400 {object} response.Response
// @Router /api/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}

	loginResp, err := h.userAppService.Login(c.Request.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errorx.Error); ok {
			response.Error(c, appErr)
		} else {
			response.InternalError(c, err)
		}
		return
	}

	response.Success(c, loginResp)
}

// GetProfile 获取用户资料
// @Summary 获取用户资料
// @Description 获取当前登录用户的资料信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 401 {object} response.Response
// @Router /api/user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.ErrorWithCode(c, errorx.ErrUnauthorized)
		return
	}

	user, err := h.userAppService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*errorx.Error); ok {
			response.Error(c, appErr)
		} else {
			response.InternalError(c, err)
		}
		return
	}

	response.Success(c, user)
}
