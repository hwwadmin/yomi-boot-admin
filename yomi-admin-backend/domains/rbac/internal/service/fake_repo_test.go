package service

import (
	"context"
	"sort"
	"strings"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"
)

// fakeRepo 是 RbacRepo 的内存实现，供 service 单测使用
type fakeRepo struct {
	roles            map[int64]*entity.Role
	permissionGroups map[int64]*entity.PermissionGroup
	permissions      map[int64]*entity.Permission
	userRoles        map[int64][]int64
	roleGroups       map[int64][]int64
	groupPermissions map[int64][]int64
}

var _ repo_contract.RbacRepo = (*fakeRepo)(nil)

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		roles:            make(map[int64]*entity.Role),
		permissionGroups: make(map[int64]*entity.PermissionGroup),
		permissions:      make(map[int64]*entity.Permission),
		userRoles:        make(map[int64][]int64),
		roleGroups:       make(map[int64][]int64),
		groupPermissions: make(map[int64][]int64),
	}
}

func (r *fakeRepo) FindRoleByID(_ context.Context, id int64) (*entity.Role, bool, error) {
	role, ok := r.roles[id]
	return role, ok, nil
}

func (r *fakeRepo) FindRoleByCode(_ context.Context, code string) (*entity.Role, bool, error) {
	for _, role := range r.roles {
		if role.Code == code {
			return role, true, nil
		}
	}
	return nil, false, nil
}

func (r *fakeRepo) FindRolesByIDs(_ context.Context, ids []int64) ([]*entity.Role, error) {
	roles := make([]*entity.Role, 0, len(ids))
	for _, id := range ids {
		if role, ok := r.roles[id]; ok {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

func (r *fakeRepo) ListAllRoles(_ context.Context) ([]*entity.Role, error) {
	roles := make([]*entity.Role, 0, len(r.roles))
	for _, role := range r.roles {
		roles = append(roles, role)
	}
	sort.Slice(roles, func(i, j int) bool { return roles[i].ID < roles[j].ID })
	return roles, nil
}

func (r *fakeRepo) PageRoles(_ context.Context, keyword string, status enums.Status, page, pageSize int) ([]*entity.Role, int64, error) {
	roles := make([]*entity.Role, 0)
	for _, role := range r.roles {
		if keyword != "" && !strings.Contains(role.Code, keyword) && !strings.Contains(role.Name, keyword) {
			continue
		}
		if status != "" && role.Status != status {
			continue
		}
		roles = append(roles, role)
	}
	sort.Slice(roles, func(i, j int) bool { return roles[i].ID < roles[j].ID })
	return paginateRoles(roles, page, pageSize), int64(len(roles)), nil
}

func (r *fakeRepo) CreateRole(_ context.Context, role *entity.Role, _ int64) error {
	r.roles[role.ID] = role
	return nil
}

func (r *fakeRepo) UpdateRole(_ context.Context, role *entity.Role, _ int64) error {
	r.roles[role.ID] = role
	return nil
}

func (r *fakeRepo) FindPermissionGroupByID(_ context.Context, id int64) (*entity.PermissionGroup, bool, error) {
	group, ok := r.permissionGroups[id]
	return group, ok, nil
}

func (r *fakeRepo) FindPermissionGroupByCode(_ context.Context, code string) (*entity.PermissionGroup, bool, error) {
	for _, group := range r.permissionGroups {
		if group.Code == code {
			return group, true, nil
		}
	}
	return nil, false, nil
}

func (r *fakeRepo) FindPermissionGroupsByIDs(_ context.Context, ids []int64) ([]*entity.PermissionGroup, error) {
	groups := make([]*entity.PermissionGroup, 0, len(ids))
	for _, id := range ids {
		if group, ok := r.permissionGroups[id]; ok {
			groups = append(groups, group)
		}
	}
	return groups, nil
}

func (r *fakeRepo) PagePermissionGroups(_ context.Context, keyword string, status enums.Status, page, pageSize int) ([]*entity.PermissionGroup, int64, error) {
	groups := make([]*entity.PermissionGroup, 0)
	for _, group := range r.permissionGroups {
		if keyword != "" && !strings.Contains(group.Code, keyword) && !strings.Contains(group.Name, keyword) {
			continue
		}
		if status != "" && group.Status != status {
			continue
		}
		groups = append(groups, group)
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].ID < groups[j].ID })
	return paginateGroups(groups, page, pageSize), int64(len(groups)), nil
}

func (r *fakeRepo) CreatePermissionGroup(_ context.Context, group *entity.PermissionGroup, _ int64) error {
	r.permissionGroups[group.ID] = group
	return nil
}

func (r *fakeRepo) UpdatePermissionGroup(_ context.Context, group *entity.PermissionGroup, _ int64) error {
	r.permissionGroups[group.ID] = group
	return nil
}

func (r *fakeRepo) FindPermissionByID(_ context.Context, id int64) (*entity.Permission, bool, error) {
	permission, ok := r.permissions[id]
	return permission, ok, nil
}

func (r *fakeRepo) FindPermissionsByIDs(_ context.Context, ids []int64) ([]*entity.Permission, error) {
	permissions := make([]*entity.Permission, 0, len(ids))
	for _, id := range ids {
		if permission, ok := r.permissions[id]; ok {
			permissions = append(permissions, permission)
		}
	}
	return permissions, nil
}

func (r *fakeRepo) PagePermissions(_ context.Context, keyword string, permissionType enums.PermissionType, status enums.Status, page, pageSize int) ([]*entity.Permission, int64, error) {
	permissions := make([]*entity.Permission, 0)
	for _, permission := range r.permissions {
		if keyword != "" && !strings.Contains(permission.Code, keyword) && !strings.Contains(permission.Name, keyword) {
			continue
		}
		if permissionType != "" && permission.Type != permissionType {
			continue
		}
		if status != "" && permission.Status != status {
			continue
		}
		permissions = append(permissions, permission)
	}
	sort.Slice(permissions, func(i, j int) bool { return permissions[i].ID < permissions[j].ID })
	return paginatePermissions(permissions, page, pageSize), int64(len(permissions)), nil
}

func (r *fakeRepo) UpdatePermissionStatus(_ context.Context, id int64, status enums.Status, _ int64) error {
	if permission, ok := r.permissions[id]; ok {
		permission.Status = status
	}
	return nil
}

func (r *fakeRepo) ListGroupIDsByRoleID(_ context.Context, roleID int64) ([]int64, error) {
	return append([]int64(nil), r.roleGroups[roleID]...), nil
}

func (r *fakeRepo) ReplaceRoleGroups(_ context.Context, roleID int64, rels []*entity.RoleGroupRel, _ int64) error {
	ids := make([]int64, 0, len(rels))
	for _, rel := range rels {
		ids = append(ids, rel.GroupID)
	}
	r.roleGroups[roleID] = ids
	return nil
}

func (r *fakeRepo) ListPermissionIDsByRoleID(_ context.Context, roleID int64) ([]int64, error) {
	seen := make(map[int64]struct{})
	ids := make([]int64, 0)
	for _, groupID := range r.roleGroups[roleID] {
		for _, permissionID := range r.groupPermissions[groupID] {
			if _, ok := seen[permissionID]; ok {
				continue
			}
			seen[permissionID] = struct{}{}
			ids = append(ids, permissionID)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}

func (r *fakeRepo) ListPermissionIDsByGroupID(_ context.Context, groupID int64) ([]int64, error) {
	return append([]int64(nil), r.groupPermissions[groupID]...), nil
}

func (r *fakeRepo) ReplaceGroupPermissions(_ context.Context, groupID int64, rels []*entity.GroupPermissionRel, _ int64) error {
	ids := make([]int64, 0, len(rels))
	for _, rel := range rels {
		ids = append(ids, rel.PermissionID)
	}
	r.groupPermissions[groupID] = ids
	return nil
}

func (r *fakeRepo) ListRoleIDsByUserID(_ context.Context, userID int64) ([]int64, error) {
	return append([]int64(nil), r.userRoles[userID]...), nil
}

func (r *fakeRepo) ListRolesByUserIDs(_ context.Context, userIDs []int64) (map[int64][]*entity.Role, error) {
	result := make(map[int64][]*entity.Role, len(userIDs))
	for _, userID := range userIDs {
		for _, roleID := range r.userRoles[userID] {
			if role, ok := r.roles[roleID]; ok {
				result[userID] = append(result[userID], role)
			}
		}
	}
	return result, nil
}

func (r *fakeRepo) ReplaceUserRoles(_ context.Context, userID int64, rels []*entity.UserRoleRel, _ int64) error {
	ids := make([]int64, 0, len(rels))
	for _, rel := range rels {
		ids = append(ids, rel.RoleID)
	}
	r.userRoles[userID] = ids
	return nil
}

func (r *fakeRepo) UserHasEnabledSuperAdmin(_ context.Context, userID int64) (bool, error) {
	for _, roleID := range r.userRoles[userID] {
		role := r.roles[roleID]
		if role != nil && role.Code == enums.RoleCodeSuperAdmin && role.Status == enums.StatusEnabled {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeRepo) ListUserIDsWithEnabledSuperAdmin(_ context.Context) ([]int64, error) {
	ids := make([]int64, 0)
	for userID := range r.userRoles {
		if ok, _ := r.UserHasEnabledSuperAdmin(context.Background(), userID); ok {
			ids = append(ids, userID)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}

func (r *fakeRepo) ExistsRouteAccess(_ context.Context, userID int64, method, route string) (bool, error) {
	for _, permission := range r.enabledPermissionsOf(userID) {
		if permission.Type != enums.PermissionTypeAPI {
			continue
		}
		if strings.EqualFold(permission.HTTPMethod, method) && permission.HTTPPath == route {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeRepo) ListMenusByUserID(_ context.Context, userID int64) ([]*entity.Permission, error) {
	menus := make([]*entity.Permission, 0)
	for _, permission := range r.enabledPermissionsOf(userID) {
		if permission.Type == enums.PermissionTypeMenu {
			menus = append(menus, permission)
		}
	}
	sort.Slice(menus, func(i, j int) bool { return menus[i].ID < menus[j].ID })
	return menus, nil
}

func (r *fakeRepo) ListEnabledMenus(_ context.Context) ([]*entity.Permission, error) {
	menus := make([]*entity.Permission, 0)
	for _, permission := range r.permissions {
		if permission.Type == enums.PermissionTypeMenu && permission.Status == enums.StatusEnabled {
			menus = append(menus, permission)
		}
	}
	sort.Slice(menus, func(i, j int) bool { return menus[i].ID < menus[j].ID })
	return menus, nil
}

func (r *fakeRepo) ListPermissionCodesByUserID(_ context.Context, userID int64) ([]string, error) {
	codes := make([]string, 0)
	for _, permission := range r.enabledPermissionsOf(userID) {
		codes = append(codes, permission.Code)
	}
	sort.Strings(codes)
	return codes, nil
}

func (r *fakeRepo) ListEnabledPermissionCodes(_ context.Context) ([]string, error) {
	codes := make([]string, 0)
	for _, permission := range r.permissions {
		if permission.Status == enums.StatusEnabled {
			codes = append(codes, permission.Code)
		}
	}
	sort.Strings(codes)
	return codes, nil
}

// enabledPermissionsOf 按用户 → 启用角色 → 启用权限组 → 启用权限聚合，按权限 ID 去重
func (r *fakeRepo) enabledPermissionsOf(userID int64) map[int64]*entity.Permission {
	result := make(map[int64]*entity.Permission)
	for _, roleID := range r.userRoles[userID] {
		role := r.roles[roleID]
		if role == nil || role.Status != enums.StatusEnabled {
			continue
		}
		for _, groupID := range r.roleGroups[roleID] {
			group := r.permissionGroups[groupID]
			if group == nil || group.Status != enums.StatusEnabled {
				continue
			}
			for _, permissionID := range r.groupPermissions[groupID] {
				permission := r.permissions[permissionID]
				if permission == nil || permission.Status != enums.StatusEnabled {
					continue
				}
				result[permission.ID] = permission
			}
		}
	}
	return result
}

func paginateRoles(roles []*entity.Role, page, pageSize int) []*entity.Role {
	start := (page - 1) * pageSize
	if start >= len(roles) {
		return []*entity.Role{}
	}
	end := start + pageSize
	if end > len(roles) {
		end = len(roles)
	}
	return roles[start:end]
}

func paginateGroups(groups []*entity.PermissionGroup, page, pageSize int) []*entity.PermissionGroup {
	start := (page - 1) * pageSize
	if start >= len(groups) {
		return []*entity.PermissionGroup{}
	}
	end := start + pageSize
	if end > len(groups) {
		end = len(groups)
	}
	return groups[start:end]
}

func paginatePermissions(permissions []*entity.Permission, page, pageSize int) []*entity.Permission {
	start := (page - 1) * pageSize
	if start >= len(permissions) {
		return []*entity.Permission{}
	}
	end := start + pageSize
	if end > len(permissions) {
		end = len(permissions)
	}
	return permissions[start:end]
}
