package service

import (
	"context"
	"testing"

	menuEntity "github.com/sine-io/sinx/domain/menu/entity"
	menuRepo "github.com/sine-io/sinx/domain/menu/repository"
	rbacRepo "github.com/sine-io/sinx/domain/rbac/repository"
	roleEntity "github.com/sine-io/sinx/domain/role/entity"
	roleRepo "github.com/sine-io/sinx/domain/role/repository"
	userEntity "github.com/sine-io/sinx/domain/user/entity"
	userRepo "github.com/sine-io/sinx/domain/user/repository"
	"github.com/sine-io/sinx/pkg/config"
	"github.com/sine-io/sinx/pkg/logger"
)

// 简单内存自增ID
type idGen struct{ next uint }

func (g *idGen) nextID() uint { g.next++; return g.next }

// In-memory implementations -------------------------------------------------

type memUserRepo struct {
	idg  idGen
	data map[uint]*userEntity.User
}

func newMemUserRepo() userRepo.UserRepository { return &memUserRepo{data: map[uint]*userEntity.User{}} }
func (m *memUserRepo) Create(_ context.Context, u *userEntity.User) error {
	u.ID = m.idg.nextID()
	m.data[u.ID] = u
	return nil
}
func (m *memUserRepo) GetByID(_ context.Context, id uint) (*userEntity.User, error) {
	return m.data[id], nil
}
func (m *memUserRepo) GetByUsername(_ context.Context, username string) (*userEntity.User, error) {
	for _, u := range m.data {
		if u.Username == username {
			return u, nil
		}
	}
	return &userEntity.User{}, nil
}
func (m *memUserRepo) GetByEmail(_ context.Context, email string) (*userEntity.User, error) {
	for _, u := range m.data {
		if u.Email == email {
			return u, nil
		}
	}
	return &userEntity.User{}, nil
}
func (m *memUserRepo) Update(_ context.Context, u *userEntity.User) error {
	m.data[u.ID] = u
	return nil
}
func (m *memUserRepo) Delete(_ context.Context, id uint) error { delete(m.data, id); return nil }
func (m *memUserRepo) List(_ context.Context, offset, limit int) ([]*userEntity.User, error) {
	res := []*userEntity.User{}
	for _, u := range m.data {
		res = append(res, u)
	}
	if offset > len(res) {
		return []*userEntity.User{}, nil
	}
	end := offset + limit
	if end > len(res) {
		end = len(res)
	}
	return res[offset:end], nil
}
func (m *memUserRepo) Count(_ context.Context) (int64, error) { return int64(len(m.data)), nil }

type memRoleRepo struct {
	idg  idGen
	data map[uint]*roleEntity.Role
}

func newMemRoleRepo() roleRepo.RoleRepository { return &memRoleRepo{data: map[uint]*roleEntity.Role{}} }
func (m *memRoleRepo) Create(_ context.Context, r *roleEntity.Role) error {
	r.ID = m.idg.nextID()
	m.data[r.ID] = r
	return nil
}
func (m *memRoleRepo) Update(_ context.Context, r *roleEntity.Role) error {
	m.data[r.ID] = r
	return nil
}
func (m *memRoleRepo) Delete(_ context.Context, id uint) error { delete(m.data, id); return nil }
func (m *memRoleRepo) GetByID(_ context.Context, id uint) (*roleEntity.Role, error) {
	return m.data[id], nil
}
func (m *memRoleRepo) List(_ context.Context, offset, limit int) ([]*roleEntity.Role, error) {
	res := []*roleEntity.Role{}
	for _, r := range m.data {
		res = append(res, r)
	}
	if offset > len(res) {
		return []*roleEntity.Role{}, nil
	}
	end := offset + limit
	if end > len(res) {
		end = len(res)
	}
	return res[offset:end], nil
}
func (m *memRoleRepo) Count(_ context.Context) (int64, error) { return int64(len(m.data)), nil }

type memMenuRepo struct {
	idg  idGen
	data map[uint]*menuEntity.Menu
}

func newMemMenuRepo() menuRepo.MenuRepository { return &memMenuRepo{data: map[uint]*menuEntity.Menu{}} }
func (m *memMenuRepo) Create(_ context.Context, e *menuEntity.Menu) error {
	e.ID = m.idg.nextID()
	m.data[e.ID] = e
	return nil
}
func (m *memMenuRepo) Update(_ context.Context, e *menuEntity.Menu) error {
	m.data[e.ID] = e
	return nil
}
func (m *memMenuRepo) Delete(_ context.Context, id uint) error { delete(m.data, id); return nil }
func (m *memMenuRepo) GetByID(_ context.Context, id uint) (*menuEntity.Menu, error) {
	return m.data[id], nil
}
func (m *memMenuRepo) List(_ context.Context, offset, limit int, name string, status *int) ([]*menuEntity.Menu, error) {
	res := []*menuEntity.Menu{}
	for _, o := range m.data {
		if name != "" && !contains(o.Name, name) {
			continue
		}
		if status != nil && o.Status != int16(*status) {
			continue
		}
		res = append(res, o)
	}
	if offset > len(res) {
		return []*menuEntity.Menu{}, nil
	}
	end := offset + limit
	if end > len(res) {
		end = len(res)
	}
	return res[offset:end], nil
}
func (m *memMenuRepo) Count(_ context.Context, name string, status *int) (int64, error) {
	var c int64
	for _, o := range m.data {
		if name != "" && !contains(o.Name, name) {
			continue
		}
		if status != nil && o.Status != int16(*status) {
			continue
		}
		c++
	}
	return c, nil
}
func (m *memMenuRepo) ListAll(_ context.Context) ([]*menuEntity.Menu, error) {
	res := []*menuEntity.Menu{}
	for _, o := range m.data {
		res = append(res, o)
	}
	return res, nil
}
func (m *memMenuRepo) HasChildren(_ context.Context, id uint) (bool, error) {
	for _, o := range m.data {
		if o.ParentID == id {
			return true, nil
		}
	}
	return false, nil
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && indexOf(s, sub) >= 0))
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

type memRBACRepo struct {
	userRoles map[uint]map[uint]struct{}
	roleMenus map[uint]map[uint]struct{}
	roles     map[uint]*roleEntity.Role
	menus     map[uint]*menuEntity.Menu
}

func newMemRBACRepo(rr *memRoleRepo, mr *memMenuRepo) rbacRepo.RBACRepository {
	return &memRBACRepo{userRoles: map[uint]map[uint]struct{}{}, roleMenus: map[uint]map[uint]struct{}{}, roles: rr.data, menus: mr.data}
}

func (r *memRBACRepo) BindUserRoles(_ context.Context, userID uint, roleIDs []uint) (int, int, error) {
	if _, ok := r.userRoles[userID]; !ok {
		r.userRoles[userID] = map[uint]struct{}{}
	}
	added, skipped := 0, 0
	for _, id := range roleIDs {
		if _, exists := r.userRoles[userID][id]; exists {
			skipped++
			continue
		}
		r.userRoles[userID][id] = struct{}{}
		added++
	}
	return added, skipped, nil
}
func (r *memRBACRepo) UnbindUserRoles(_ context.Context, userID uint, roleIDs []uint) error {
	if m, ok := r.userRoles[userID]; ok {
		for _, id := range roleIDs {
			delete(m, id)
		}
	}
	return nil
}
func (r *memRBACRepo) GetUserRoles(_ context.Context, userID uint) ([]*roleEntity.Role, error) {
	res := []*roleEntity.Role{}
	for id := range r.userRoles[userID] {
		if rl, ok := r.roles[id]; ok {
			res = append(res, rl)
		}
	}
	return res, nil
}
func (r *memRBACRepo) BindRoleMenus(_ context.Context, roleID uint, menuIDs []uint) (int, int, error) {
	if _, ok := r.roleMenus[roleID]; !ok {
		r.roleMenus[roleID] = map[uint]struct{}{}
	}
	added, skipped := 0, 0
	for _, id := range menuIDs {
		if _, ex := r.roleMenus[roleID][id]; ex {
			skipped++
			continue
		}
		r.roleMenus[roleID][id] = struct{}{}
		added++
	}
	return added, skipped, nil
}
func (r *memRBACRepo) UnbindRoleMenus(_ context.Context, roleID uint, menuIDs []uint) error {
	if m, ok := r.roleMenus[roleID]; ok {
		for _, id := range menuIDs {
			delete(m, id)
		}
	}
	return nil
}
func (r *memRBACRepo) GetRoleMenus(_ context.Context, roleID uint) ([]*menuEntity.Menu, error) {
	res := []*menuEntity.Menu{}
	for id := range r.roleMenus[roleID] {
		if m, ok := r.menus[id]; ok {
			res = append(res, m)
		}
	}
	return res, nil
}
func (r *memRBACRepo) GetUserMenus(_ context.Context, userID uint) ([]*menuEntity.Menu, error) {
	res := []*menuEntity.Menu{}
	for rid := range r.userRoles[userID] {
		for mid := range r.roleMenus[rid] {
			if m, ok := r.menus[mid]; ok {
				res = append(res, m)
			}
		}
	}
	return res, nil
}
func (r *memRBACRepo) GetMenuRoles(_ context.Context, menuID uint) ([]*roleEntity.Role, error) {
	res := []*roleEntity.Role{}
	for rid, mids := range r.roleMenus {
		if _, ok := mids[menuID]; ok {
			if rl, ok2 := r.roles[rid]; ok2 {
				res = append(res, rl)
			}
		}
	}
	return res, nil
}
func (r *memRBACRepo) GetRoleUsers(_ context.Context, roleID uint) ([]uint, error) {
	res := []uint{}
	for uid, rids := range r.userRoles {
		if _, ok := rids[roleID]; ok {
			res = append(res, uid)
		}
	}
	return res, nil
}
func (r *memRBACRepo) GetMenuIDsByRole(_ context.Context, roleID uint) ([]uint, error) {
	res := []uint{}
	for mid := range r.roleMenus[roleID] {
		res = append(res, mid)
	}
	return res, nil
}

// Test ----------------------------------------------------------------------
func TestBindAndPerms_InMemory(t *testing.T) {
	// init config & logger once
	_ = config.LoadEnv()
	_ = logger.Init()
	ctx := context.Background()
	ur := newMemUserRepo().(*memUserRepo)
	rr := newMemRoleRepo().(*memRoleRepo)
	mr := newMemMenuRepo().(*memMenuRepo)
	rb := newMemRBACRepo(rr, mr)
	svc := NewRBACApplicationService(ur, rr, mr, rb)

	_ = ur.Create(ctx, &userEntity.User{Username: "u1", Password: "p"})                                                        // id=1
	_ = rr.Create(ctx, &roleEntity.Role{Name: "r1", Status: 1})                                                                // id=1
	_ = mr.Create(ctx, &menuEntity.Menu{Name: "m1", ParentID: 0, OrderNum: 1, MenuType: "B", Perms: "user:create", Status: 1}) // id=1
	_, _, _ = svc.BindRoleMenus(ctx, 1, []uint{1})
	_, _, _ = svc.BindUserRoles(ctx, 1, []uint{1})
	perms, err := svc.GetUserPerms(ctx, 1)
	if err != nil {
		t.Fatalf("get perms: %v", err)
	}
	if _, ok := perms["user:create"]; !ok {
		t.Fatalf("expected perm not found")
	}
	// 再次获取应命中缓存
	perms2, _ := svc.GetUserPerms(ctx, 1)
	if len(perms2) != len(perms) {
		t.Fatalf("cache mismatch")
	}
}
