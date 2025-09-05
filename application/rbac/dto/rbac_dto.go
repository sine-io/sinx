package dto

// 通用分页请求
type PageRequest struct {
	PageNum  int `form:"pageNum" json:"pageNum"`
	PageSize int `form:"pageSize" json:"pageSize"`
}

type PageResult[T any] struct {
	Total int64 `json:"total"`
	Data  []T   `json:"data"`
}

// 用户相关
type UserCreateRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
	Avatar   string `json:"avatar"`
}

type UserUpdateRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
	Status   *int16 `json:"status"`
}

type UserDeleteRequest struct {
	ID uint `json:"id" binding:"required"`
}

type ChangePasswordRequest struct {
	UserID      uint   `json:"userId" binding:"required"`
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type BindUserRoleRequest struct {
	UserID  uint   `json:"userId" binding:"required"`
	RoleIDs []uint `json:"roleIds" binding:"required"`
}

type UnbindUserRoleRequest struct {
	UserID  uint   `json:"userId" binding:"required"`
	RoleIDs []uint `json:"roleIds" binding:"required"`
}

// 角色相关
type RoleCreateOrUpdateRequest struct {
	ID     uint   `json:"id"`
	Name   string `json:"name" binding:"required"`
	Remark string `json:"remark"`
	Status int16  `json:"status" binding:"required"`
}

type RoleDeleteRequest struct {
	ID uint `json:"id" binding:"required"`
}

type BindRoleMenuRequest struct {
	RoleID  uint   `json:"roleId" binding:"required"`
	MenuIDs []uint `json:"menuIds" binding:"required"`
}

type UnbindRoleMenuRequest struct {
	RoleID  uint   `json:"roleId" binding:"required"`
	MenuIDs []uint `json:"menuIds" binding:"required"`
}

// 菜单相关
type MenuCreateOrUpdateRequest struct {
	ID        uint   `json:"id"`
	Name      string `json:"name" binding:"required"`
	ParentID  uint   `json:"parentId" binding:"required"`
	OrderNum  int    `json:"orderNum" binding:"required"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Query     string `json:"query"`
	IsFrame   int16  `json:"isFrame" binding:"required"`
	MenuType  string `json:"menuType" binding:"required"`
	IsCatch   int16  `json:"isCatch" binding:"required"`
	IsHidden  int16  `json:"isHidden" binding:"required"`
	Perms     string `json:"perms"`
	Icon      string `json:"icon"`
	Status    int16  `json:"status" binding:"required"`
	Remark    string `json:"remark"`
}

type MenuDeleteRequest struct {
	ID uint `json:"id" binding:"required"`
}

// 输出结构
type UserSimple struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Status   int16  `json:"status"`
}

type RoleSimple struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Remark string `json:"remark"`
	Status int16  `json:"status"`
}

type MenuSimple struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	ParentID  uint   `json:"parentId"`
	OrderNum  int    `json:"orderNum"`
	Path      string `json:"path"`
	Component string `json:"component"`
	MenuType  string `json:"menuType"`
	Icon      string `json:"icon"`
	Status    int16  `json:"status"`
	Perms     string `json:"perms"`
}

type MenuTreeNode struct {
	ID        uint            `json:"id"`
	Name      string          `json:"name"`
	ParentID  uint            `json:"parentId"`
	Path      string          `json:"path,omitempty"`
	Component string          `json:"component,omitempty"`
	Icon      string          `json:"icon,omitempty"`
	Children  []*MenuTreeNode `json:"children"`
}

type RoleMenuTreeResponse struct {
	MenuIDs []uint `json:"menuIds"`
}
