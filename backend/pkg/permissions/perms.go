package permissions

// 定义所有权限常量，便于集中管理与对外导出
const (
	// 用户相关
	PermUserCreate     = "user:create"
	PermUserList       = "user:list"
	PermUserUpdate     = "user:update"
	PermUserDelete     = "user:delete"
	PermUserBindRole   = "user:bindRole"
	PermUserUnbindRole = "user:unbindRole"
	PermUserRoles      = "user:roles"

	// 角色相关
	PermRoleCreate     = "role:create"
	PermRoleList       = "role:list"
	PermRoleUpdate     = "role:update"
	PermRoleDelete     = "role:delete"
	PermRoleBindMenu   = "role:bindMenu"
	PermRoleUnbindMenu = "role:unbindMenu"
	PermRoleMenus      = "role:menus"
	PermRoleUsers      = "role:users"

	// 菜单相关
	PermMenuCreate       = "menu:create"
	PermMenuList         = "menu:list"
	PermMenuUpdate       = "menu:update"
	PermMenuDelete       = "menu:delete"
	PermMenuRoles        = "menu:roles"
	PermMenuRoleMenuTree = "menu:roleMenuTree"
)

// AllPerms 导出全部权限列表（用于前端获取 / 同步 / 测试）
var AllPerms = []string{
	PermUserCreate, PermUserList, PermUserUpdate, PermUserDelete, PermUserBindRole, PermUserUnbindRole, PermUserRoles,
	PermRoleCreate, PermRoleList, PermRoleUpdate, PermRoleDelete, PermRoleBindMenu, PermRoleUnbindMenu, PermRoleMenus, PermRoleUsers,
	PermMenuCreate, PermMenuList, PermMenuUpdate, PermMenuDelete, PermMenuRoles, PermMenuRoleMenuTree,
}
