package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	rbacdto "github.com/sine-io/sinx/application/rbac/dto"
	rbacService "github.com/sine-io/sinx/application/rbac/service"
	"github.com/sine-io/sinx/pkg/errorx"
	"github.com/sine-io/sinx/pkg/response"
)

type RBACHandler struct {
	svc *rbacService.RBACApplicationService
}

func NewRBACHandler(s *rbacService.RBACApplicationService) *RBACHandler { return &RBACHandler{svc: s} }

// 用户接口
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
func (h *RBACHandler) BindUserRole(c *gin.Context) {
	var req rbacdto.BindUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	if err := h.svc.BindUserRoles(c, req.UserID, req.RoleIDs); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}
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
func (h *RBACHandler) GetUserMenus(c *gin.Context) {
	idStr := c.Query("userId")
	var id uint
	if idStr != "" {
		v, _ := strconv.Atoi(idStr)
		id = uint(v)
	} // 若为空，可结合token获取当前用户，这里简化
	menus, err := h.svc.GetUserMenus(c, id)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, menus)
}

// 角色接口
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
func (h *RBACHandler) UpdateRole(c *gin.Context) { h.CreateRole(c) }
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
func (h *RBACHandler) BindRoleMenu(c *gin.Context) {
	var req rbacdto.BindRoleMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, errorx.ErrInvalidParam)
		return
	}
	if err := h.svc.BindRoleMenus(c, req.RoleID, req.MenuIDs); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}
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
func (h *RBACHandler) GetRoleUsers(c *gin.Context) { /* 可扩展: 查询拥有该角色的用户 */
	response.Success(c, []any{})
}
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
func (h *RBACHandler) UpdateMenu(c *gin.Context) { h.CreateMenu(c) }
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
func (h *RBACHandler) MenuTree(c *gin.Context) {
	tree, err := h.svc.MenuTree(c)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, tree)
}
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
