package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sine-io/sinx/api/middleware"
	rbacdto "github.com/sine-io/sinx/application/rbac/dto"
	rbacService "github.com/sine-io/sinx/application/rbac/service"
	"github.com/sine-io/sinx/pkg/errorx"
	"github.com/sine-io/sinx/pkg/response"
)

type RBACHandler struct {
	svc *rbacService.RBACApplicationService
}

func NewRBACHandler(s *rbacService.RBACApplicationService) *RBACHandler { return &RBACHandler{svc: s} }

// Service 暴露内部应用服务（供路由层构建权限检查器使用）
func (h *RBACHandler) Service() *rbacService.RBACApplicationService { return h.svc }

// 用户接口
// CreateUser 创建用户
// @Summary 创建用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.UserCreateRequest true "创建用户参数"
// @Success 200 {object} response.Response
// @Router /api/user/create [post]
func (h *RBACHandler) CreateUser(c *gin.Context) {
	var req rbacdto.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	if err := h.svc.CreateUser(c, &req); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.UserUpdateRequest true "更新用户参数"
// @Success 200 {object} response.Response
// @Router /api/user/update [post]
func (h *RBACHandler) UpdateUser(c *gin.Context) {
	var req rbacdto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	if err := h.svc.UpdateUser(c, &req); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.UserDeleteRequest true "删除用户参数"
// @Success 200 {object} response.Response
// @Router /api/user/delete [post]
func (h *RBACHandler) DeleteUser(c *gin.Context) {
	var req rbacdto.UserDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	if err := h.svc.DeleteUser(c, req.ID); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

// UserList 用户列表
// @Summary 获取用户列表
// @Tags 用户管理
// @Produce json
// @Security ApiKeyAuth
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} response.Response
// @Router /api/user/list [get]
func (h *RBACHandler) UserList(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	total, list, err := h.svc.ListUsers(c, pageNum, pageSize)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, gin.H{"total": total, "data": list})
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.ChangePasswordRequest true "修改密码"
// @Success 200 {object} response.Response
// @Router /api/user/changePassword [post]
func (h *RBACHandler) ChangePassword(c *gin.Context) {
	var req rbacdto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	if err := h.svc.ChangePassword(c, &req); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

// BindUserRole 绑定角色
// @Summary 绑定用户角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.BindUserRoleRequest true "绑定角色"
// @Success 200 {object} response.Response
// @Router /api/user/bindRole [post]
func (h *RBACHandler) BindUserRole(c *gin.Context) {
	var req rbacdto.BindUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	added, skipped, err := h.svc.BindUserRoles(c, req.UserID, req.RoleIDs)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, gin.H{"added": added, "skipped": skipped})
}

// UnbindUserRole 解绑角色
// @Summary 解绑用户角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.UnbindUserRoleRequest true "解绑角色"
// @Success 200 {object} response.Response
// @Router /api/user/unbindRole [post]
func (h *RBACHandler) UnbindUserRole(c *gin.Context) {
	var req rbacdto.UnbindUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	if err := h.svc.UnbindUserRoles(c, req.UserID, req.RoleIDs); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

// GetUserRoles 用户角色列表
// @Summary 获取用户角色列表
// @Tags 用户管理
// @Produce json
// @Security ApiKeyAuth
// @Param id query int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/user/roles [get]
func (h *RBACHandler) GetUserRoles(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	id, _ := strconv.Atoi(idStr)
	roles, err := h.svc.GetUserRoles(c, uint(id))
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, roles)
}

// GetUserMenus 用户菜单树
// @Summary 获取用户菜单树
// @Tags 用户管理
// @Produce json
// @Security ApiKeyAuth
// @Param userId query int false "用户ID(可选)"
// @Success 200 {object} response.Response
// @Router /api/user/menus [get]
func (h *RBACHandler) GetUserMenus(c *gin.Context) {
	idStr := c.Query("userId")
	var id uint
	if idStr != "" {
		v, _ := strconv.Atoi(idStr)
		id = uint(v)
	} else {
		if uid, ok := middleware.GetUserID(c); ok {
			id = uid
		}
	}
	menus, err := h.svc.GetUserMenus(c, id)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, menus)
}

// 角色接口
// CreateRole 创建或更新角色
// @Summary 创建角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.RoleCreateOrUpdateRequest true "角色"
// @Success 200 {object} response.Response
// @Router /api/role/create [post]
func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req rbacdto.RoleCreateOrUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	if err := h.svc.CreateOrUpdateRole(c, &req); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.RoleCreateOrUpdateRequest true "角色"
// @Success 200 {object} response.Response
// @Router /api/role/update [post]
func (h *RBACHandler) UpdateRole(c *gin.Context) { h.CreateRole(c) }

// DeleteRole 删除角色
// @Summary 删除角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.RoleDeleteRequest true "删除角色"
// @Success 200 {object} response.Response
// @Router /api/role/delete [post]
func (h *RBACHandler) DeleteRole(c *gin.Context) {
	var req rbacdto.RoleDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	if err := h.svc.DeleteRole(c, req.ID); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

// RoleList 角色列表
// @Summary 获取角色列表
// @Tags 角色管理
// @Produce json
// @Security ApiKeyAuth
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} response.Response
// @Router /api/role/list [get]
func (h *RBACHandler) RoleList(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	total, list, err := h.svc.ListRoles(c, pageNum, pageSize)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, gin.H{"total": total, "data": list})
}

// BindRoleMenu 绑定菜单
// @Summary 绑定角色菜单
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.BindRoleMenuRequest true "绑定菜单"
// @Success 200 {object} response.Response
// @Router /api/role/bindMenu [post]
func (h *RBACHandler) BindRoleMenu(c *gin.Context) {
	var req rbacdto.BindRoleMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	added, skipped, err := h.svc.BindRoleMenus(c, req.RoleID, req.MenuIDs)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, gin.H{"added": added, "skipped": skipped})
}

// UnbindRoleMenu 解绑菜单
// @Summary 解绑角色菜单
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.UnbindRoleMenuRequest true "解绑菜单"
// @Success 200 {object} response.Response
// @Router /api/role/unbindMenu [post]
func (h *RBACHandler) UnbindRoleMenu(c *gin.Context) {
	var req rbacdto.UnbindRoleMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	if err := h.svc.UnbindRoleMenus(c, req.RoleID, req.MenuIDs); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

// GetRoleMenus 角色菜单列表
// @Summary 获取角色菜单列表
// @Tags 角色管理
// @Produce json
// @Security ApiKeyAuth
// @Param id query int true "角色ID"
// @Success 200 {object} response.Response
// @Router /api/role/menus [get]
func (h *RBACHandler) GetRoleMenus(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	id, _ := strconv.Atoi(idStr)
	menus, err := h.svc.GetRoleMenus(c, uint(id))
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, menus)
}

// GetRoleUsers 拥有该角色的用户列表(占位)
// @Summary 获取拥有该角色的用户列表
// @Tags 角色管理
// @Produce json
// @Security ApiKeyAuth
// @Param roleId query int true "角色ID"
// @Success 200 {object} response.Response
// @Router /api/role/users [get]
func (h *RBACHandler) GetRoleUsers(c *gin.Context) { response.Success(c, []any{}) }

// GetRoleMenuTree 角色菜单ID集合
// @Summary 获取角色菜单树(返回菜单ID集合)
// @Tags 菜单管理
// @Produce json
// @Security ApiKeyAuth
// @Param roleId query int true "角色ID"
// @Success 200 {object} response.Response
// @Router /api/menu/roleMenuTree [get]
func (h *RBACHandler) GetRoleMenuTree(c *gin.Context) {
	roleIdStr := c.Query("roleId")
	if roleIdStr == "" {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	id, _ := strconv.Atoi(roleIdStr)
	tree, err := h.svc.GetRoleMenuTree(c, uint(id))
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, tree)
}

// 菜单接口
// CreateMenu 创建菜单
// @Summary 创建菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.MenuCreateOrUpdateRequest true "菜单"
// @Success 200 {object} response.Response
// @Router /api/menu/create [post]
func (h *RBACHandler) CreateMenu(c *gin.Context) {
	var req rbacdto.MenuCreateOrUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	if err := h.svc.CreateOrUpdateMenu(c, &req); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

// UpdateMenu 更新菜单
// @Summary 更新菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.MenuCreateOrUpdateRequest true "菜单"
// @Success 200 {object} response.Response
// @Router /api/menu/update [post]
func (h *RBACHandler) UpdateMenu(c *gin.Context) { h.CreateMenu(c) }

// DeleteMenu 删除菜单
// @Summary 删除菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body rbacdto.MenuDeleteRequest true "删除菜单"
// @Success 200 {object} response.Response
// @Router /api/menu/delete [post]
func (h *RBACHandler) DeleteMenu(c *gin.Context) {
	var req rbacdto.MenuDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	if err := h.svc.DeleteMenu(c, req.ID); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

// MenuList 菜单列表
// @Summary 获取菜单列表
// @Tags 菜单管理
// @Produce json
// @Security ApiKeyAuth
// @Param pageNum query int false "页码"
// @Param pageSize query int false "每页"
// @Param name query string false "名称模糊"
// @Param status query int false "状态"
// @Success 200 {object} response.Response
// @Router /api/menu/list [get]
func (h *RBACHandler) MenuList(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	name := c.Query("name")
	var statusPtr *int
	if s := c.Query("status"); s != "" {
		v, _ := strconv.Atoi(s)
		statusPtr = &v
	}
	total, list, err := h.svc.ListMenus(c, pageNum, pageSize, name, statusPtr)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, gin.H{"total": total, "data": list})
}

// MenuTree 菜单树
// @Summary 获取菜单树
// @Tags 菜单管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response
// @Router /api/menu/tree [get]
func (h *RBACHandler) MenuTree(c *gin.Context) {
	tree, err := h.svc.MenuTree(c)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, tree)
}

// MenuRoles 拥有该菜单的角色列表
// @Summary 获取拥有该菜单的角色列表
// @Tags 菜单管理
// @Produce json
// @Security ApiKeyAuth
// @Param menuId query int true "菜单ID"
// @Success 200 {object} response.Response
// @Router /api/menu/roles [get]
func (h *RBACHandler) MenuRoles(c *gin.Context) {
	menuStr := c.Query("menuId")
	if menuStr == "" {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	id, _ := strconv.Atoi(menuStr)
	roles, err := h.svc.GetMenuRoles(c, uint(id))
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, roles)
}
