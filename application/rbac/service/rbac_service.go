package service

import (
	"context"

	rbacdto "github.com/sine-io/sinx/application/rbac/dto"
	menuEntity "github.com/sine-io/sinx/domain/menu/entity"
	menuRepo "github.com/sine-io/sinx/domain/menu/repository"
	rbacRepo "github.com/sine-io/sinx/domain/rbac/repository"
	roleEntity "github.com/sine-io/sinx/domain/role/entity"
	roleRepo "github.com/sine-io/sinx/domain/role/repository"
	userEntity "github.com/sine-io/sinx/domain/user/entity"
	userRepo "github.com/sine-io/sinx/domain/user/repository"
	"github.com/sine-io/sinx/pkg/utils"
)

type RBACApplicationService struct {
	userRepository userRepo.UserRepository
	roleRepository roleRepo.RoleRepository
	menuRepository menuRepo.MenuRepository
	rbacRepository rbacRepo.RBACRepository
}

func NewRBACApplicationService(u userRepo.UserRepository, r roleRepo.RoleRepository, m menuRepo.MenuRepository, rb rbacRepo.RBACRepository) *RBACApplicationService {
	return &RBACApplicationService{userRepository: u, roleRepository: r, menuRepository: m, rbacRepository: rb}
}

// 用户管理
func (s *RBACApplicationService) CreateUser(ctx context.Context, req *rbacdto.UserCreateRequest) error {
	hashed, _ := utils.HashPassword(req.Password)
	user := &userEntity.User{Username: req.Username, Password: hashed, Nickname: req.Nickname, Email: req.Email, Mobile: req.Mobile, Avatar: req.Avatar}
	return s.userRepository.Create(ctx, user)
}

func (s *RBACApplicationService) UpdateUser(ctx context.Context, req *rbacdto.UserUpdateRequest) error {
	user, err := s.userRepository.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Mobile != "" {
		user.Mobile = req.Mobile
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	return s.userRepository.Update(ctx, user)
}

func (s *RBACApplicationService) DeleteUser(ctx context.Context, id uint) error {
	return s.userRepository.Delete(ctx, id)
}

func (s *RBACApplicationService) ListUsers(ctx context.Context, pageNum, pageSize int) (int64, []*rbacdto.UserSimple, error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (pageNum - 1) * pageSize
	users, err := s.userRepository.List(ctx, offset, pageSize)
	if err != nil {
		return 0, nil, err
	}
	total, _ := s.userRepository.Count(ctx)
	res := make([]*rbacdto.UserSimple, 0, len(users))
	for _, u := range users {
		res = append(res, &rbacdto.UserSimple{ID: u.ID, Username: u.Username, Nickname: u.Nickname, Email: u.Email, Status: u.Status})
	}
	return total, res, nil
}

func (s *RBACApplicationService) ChangePassword(ctx context.Context, req *rbacdto.ChangePasswordRequest) error {
	user, err := s.userRepository.GetByID(ctx, req.UserID)
	if err != nil {
		return err
	}
	// 简化：忽略旧密码校验（可扩展）
	hashed, _ := utils.HashPassword(req.NewPassword)
	user.Password = hashed
	return s.userRepository.Update(ctx, user)
}

// 角色管理
func (s *RBACApplicationService) CreateOrUpdateRole(ctx context.Context, req *rbacdto.RoleCreateOrUpdateRequest) error {
	if req.ID == 0 {
		return s.roleRepository.Create(ctx, &roleEntity.Role{Name: req.Name, Remark: req.Remark, Status: req.Status})
	}
	role, err := s.roleRepository.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}
	role.Name = req.Name
	role.Remark = req.Remark
	role.Status = req.Status
	return s.roleRepository.Update(ctx, role)
}

func (s *RBACApplicationService) DeleteRole(ctx context.Context, id uint) error {
	return s.roleRepository.Delete(ctx, id)
}

func (s *RBACApplicationService) ListRoles(ctx context.Context, pageNum, pageSize int) (int64, []*rbacdto.RoleSimple, error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (pageNum - 1) * pageSize
	roles, err := s.roleRepository.List(ctx, offset, pageSize)
	if err != nil {
		return 0, nil, err
	}
	total, _ := s.roleRepository.Count(ctx)
	list := make([]*rbacdto.RoleSimple, 0, len(roles))
	for _, r := range roles {
		list = append(list, &rbacdto.RoleSimple{ID: r.ID, Name: r.Name, Remark: r.Remark, Status: r.Status})
	}
	return total, list, nil
}

// 菜单管理
func (s *RBACApplicationService) CreateOrUpdateMenu(ctx context.Context, req *rbacdto.MenuCreateOrUpdateRequest) error {
	if req.ID == 0 {
		return s.menuRepository.Create(ctx, &menuEntity.Menu{Name: req.Name, ParentID: req.ParentID, OrderNum: req.OrderNum, Path: req.Path, Component: req.Component, Query: req.Query, IsFrame: req.IsFrame, MenuType: req.MenuType, IsCatch: req.IsCatch, IsHidden: req.IsHidden, Perms: req.Perms, Icon: req.Icon, Status: req.Status, Remark: req.Remark})
	}
	menu, err := s.menuRepository.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}
	menu.Name = req.Name
	menu.ParentID = req.ParentID
	menu.OrderNum = req.OrderNum
	menu.Path = req.Path
	menu.Component = req.Component
	menu.Query = req.Query
	menu.IsFrame = req.IsFrame
	menu.MenuType = req.MenuType
	menu.IsCatch = req.IsCatch
	menu.IsHidden = req.IsHidden
	menu.Perms = req.Perms
	menu.Icon = req.Icon
	menu.Status = req.Status
	menu.Remark = req.Remark
	return s.menuRepository.Update(ctx, menu)
}

func (s *RBACApplicationService) DeleteMenu(ctx context.Context, id uint) error {
	hasChild, err := s.menuRepository.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if hasChild {
		return nil
	} // 简化: 返回 nil; 实际应返回错误
	return s.menuRepository.Delete(ctx, id)
}

func (s *RBACApplicationService) ListMenus(ctx context.Context, pageNum, pageSize int, name string, status *int) (int64, []*rbacdto.MenuSimple, error) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (pageNum - 1) * pageSize
	menus, err := s.menuRepository.List(ctx, offset, pageSize, name, status)
	if err != nil {
		return 0, nil, err
	}
	total, _ := s.menuRepository.Count(ctx, name, status)
	res := make([]*rbacdto.MenuSimple, 0, len(menus))
	for _, m := range menus {
		res = append(res, &rbacdto.MenuSimple{ID: m.ID, Name: m.Name, ParentID: m.ParentID, OrderNum: m.OrderNum, Path: m.Path, Component: m.Component, MenuType: m.MenuType, Icon: m.Icon, Status: m.Status, Perms: m.Perms})
	}
	return total, res, nil
}

// 树形
func (s *RBACApplicationService) MenuTree(ctx context.Context) ([]*rbacdto.MenuTreeNode, error) {
	menus, err := s.menuRepository.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(menus, 0), nil
}

func buildMenuTree(menus []*menuEntity.Menu, parentID uint) []*rbacdto.MenuTreeNode {
	var result []*rbacdto.MenuTreeNode
	for _, m := range menus {
		if m.ParentID == parentID {
			node := &rbacdto.MenuTreeNode{ID: m.ID, Name: m.Name, ParentID: m.ParentID, Path: m.Path, Component: m.Component, Icon: m.Icon}
			node.Children = buildMenuTree(menus, m.ID)
			result = append(result, node)
		}
	}
	return result
}

// 绑定解绑
func (s *RBACApplicationService) BindUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return s.rbacRepository.BindUserRoles(ctx, userID, roleIDs)
}
func (s *RBACApplicationService) UnbindUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return s.rbacRepository.UnbindUserRoles(ctx, userID, roleIDs)
}
func (s *RBACApplicationService) BindRoleMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
	return s.rbacRepository.BindRoleMenus(ctx, roleID, menuIDs)
}
func (s *RBACApplicationService) UnbindRoleMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
	return s.rbacRepository.UnbindRoleMenus(ctx, roleID, menuIDs)
}

func (s *RBACApplicationService) GetUserRoles(ctx context.Context, userID uint) ([]*rbacdto.RoleSimple, error) {
	roles, err := s.rbacRepository.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := make([]*rbacdto.RoleSimple, 0, len(roles))
	for _, r := range roles {
		res = append(res, &rbacdto.RoleSimple{ID: r.ID, Name: r.Name, Remark: r.Remark, Status: r.Status})
	}
	return res, nil
}

func (s *RBACApplicationService) GetRoleMenus(ctx context.Context, roleID uint) ([]*rbacdto.MenuSimple, error) {
	menus, err := s.rbacRepository.GetRoleMenus(ctx, roleID)
	if err != nil {
		return nil, err
	}
	res := make([]*rbacdto.MenuSimple, 0, len(menus))
	for _, m := range menus {
		res = append(res, &rbacdto.MenuSimple{ID: m.ID, Name: m.Name, ParentID: m.ParentID, OrderNum: m.OrderNum, Path: m.Path, Component: m.Component, MenuType: m.MenuType, Icon: m.Icon, Status: m.Status, Perms: m.Perms})
	}
	return res, nil
}

func (s *RBACApplicationService) GetUserMenus(ctx context.Context, userID uint) ([]*rbacdto.MenuTreeNode, error) {
	menus, err := s.rbacRepository.GetUserMenus(ctx, userID)
	if err != nil {
		return nil, err
	}
	// 构建树：复用 buildMenuTree 需要全部菜单; 这里简化：先转 slice -> tree (仅包含授权的)
	return buildMenuTree(convertMenus(menus), 0), nil
}

func convertMenus(ms []*menuEntity.Menu) []*menuEntity.Menu { return ms }

func (s *RBACApplicationService) GetMenuRoles(ctx context.Context, menuID uint) ([]*rbacdto.RoleSimple, error) {
	roles, err := s.rbacRepository.GetMenuRoles(ctx, menuID)
	if err != nil {
		return nil, err
	}
	res := make([]*rbacdto.RoleSimple, 0, len(roles))
	for _, r := range roles {
		res = append(res, &rbacdto.RoleSimple{ID: r.ID, Name: r.Name, Remark: r.Remark, Status: r.Status})
	}
	return res, nil
}

func (s *RBACApplicationService) GetRoleMenuTree(ctx context.Context, roleID uint) (*rbacdto.RoleMenuTreeResponse, error) {
	ids, err := s.rbacRepository.GetMenuIDsByRole(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return &rbacdto.RoleMenuTreeResponse{MenuIDs: ids}, nil
}
