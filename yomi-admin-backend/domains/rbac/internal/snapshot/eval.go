package snapshot

import (
	"sort"
	"strings"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"
)

// index 是按角色聚合后的求值索引；由快照一次性构建，之后只读
type index struct {
	roles        map[int64]Role
	permissions  map[int64]Permission
	groups       map[int64]PermissionGroup
	superUsers   map[int64]struct{}
	apiByRole    map[int64]map[string]struct{}
	menusByRole  map[int64][]Permission
	codesByRole  map[int64][]string
	enabledMenus []Permission
}

// CanAccessAPI 判断用户能否访问指定的 HTTP 方法与路由模板。
// 超管直接放行；其余用户沿 启用角色 → 启用权限组 → 启用接口权限 链路精确匹配。
func (s *Snapshot) CanAccessAPI(userID int64, method, path string) bool {
	if s == nil || userID <= 0 {
		return false
	}

	idx := s.ensureIndex()
	if _, ok := idx.superUsers[userID]; ok {
		return true
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)
	if method == "" || path == "" {
		return false
	}

	key := method + " " + path
	for _, roleID := range s.UserRoles[userID] {
		if _, hit := idx.apiByRole[roleID][key]; hit {
			return true
		}
	}
	return false
}

// ListMenus 列出用户可见的启用菜单，按权限 ID 升序；超管返回全部启用菜单
func (s *Snapshot) ListMenus(userID int64) []*entity.Permission {
	if s == nil || userID <= 0 {
		return []*entity.Permission{}
	}

	idx := s.ensureIndex()
	if _, ok := idx.superUsers[userID]; ok {
		return toEntities(idx.enabledMenus)
	}

	merged := make(map[int64]Permission)
	for _, roleID := range s.UserRoles[userID] {
		for _, menu := range idx.menusByRole[roleID] {
			merged[menu.ID] = menu
		}
	}
	if len(merged) == 0 {
		return []*entity.Permission{}
	}

	menus := make([]Permission, 0, len(merged))
	for _, menu := range merged {
		menus = append(menus, menu)
	}
	sort.Slice(menus, func(i, j int) bool { return menus[i].ID < menus[j].ID })
	return toEntities(menus)
}

// ListPermissionCodes 列出用户拥有的启用权限编码，按编码升序；超管返回 ["*"]
func (s *Snapshot) ListPermissionCodes(userID int64) []string {
	if s == nil || userID <= 0 {
		return []string{}
	}

	idx := s.ensureIndex()
	if _, ok := idx.superUsers[userID]; ok {
		return []string{"*"}
	}

	merged := make(map[string]struct{})
	for _, roleID := range s.UserRoles[userID] {
		for _, code := range idx.codesByRole[roleID] {
			merged[code] = struct{}{}
		}
	}
	if len(merged) == 0 {
		return []string{}
	}

	codes := make([]string, 0, len(merged))
	for code := range merged {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

// buildIndex 按 启用角色 → 启用权限组 → 启用权限 聚合出角色级索引
func buildIndex(s *Snapshot) *index {
	idx := &index{
		roles:        make(map[int64]Role, len(s.Roles)),
		permissions:  make(map[int64]Permission, len(s.Permissions)),
		groups:       make(map[int64]PermissionGroup, len(s.PermissionGroups)),
		superUsers:   make(map[int64]struct{}),
		apiByRole:    make(map[int64]map[string]struct{}),
		menusByRole:  make(map[int64][]Permission),
		codesByRole:  make(map[int64][]string),
		enabledMenus: make([]Permission, 0, len(s.Permissions)),
	}

	for _, role := range s.Roles {
		idx.roles[role.ID] = role
	}
	for _, permission := range s.Permissions {
		idx.permissions[permission.ID] = permission
		if permission.Status == enums.StatusEnabled && permission.Type == enums.PermissionTypeMenu {
			idx.enabledMenus = append(idx.enabledMenus, permission)
		}
	}
	for _, group := range s.PermissionGroups {
		idx.groups[group.ID] = group
	}
	sort.Slice(idx.enabledMenus, func(i, j int) bool { return idx.enabledMenus[i].ID < idx.enabledMenus[j].ID })

	for userID, roleIDs := range s.UserRoles {
		for _, roleID := range roleIDs {
			if role, ok := idx.roles[roleID]; ok && isSuperRole(role) {
				idx.superUsers[userID] = struct{}{}
				break
			}
		}
	}

	for _, role := range s.Roles {
		if role.Status != enums.StatusEnabled {
			continue
		}
		api := make(map[string]struct{})
		menus := make(map[int64]Permission)
		codes := make(map[string]struct{})

		for _, groupID := range s.RoleGroups[role.ID] {
			group, ok := idx.groups[groupID]
			if !ok || group.Status != enums.StatusEnabled {
				continue
			}
			for _, permissionID := range s.GroupPermissions[groupID] {
				permission, ok := idx.permissions[permissionID]
				if !ok || permission.Status != enums.StatusEnabled {
					continue
				}
				switch permission.Type {
				case enums.PermissionTypeAPI:
					method := strings.ToUpper(strings.TrimSpace(permission.HTTPMethod))
					path := strings.TrimSpace(permission.HTTPPath)
					if method != "" && path != "" {
						api[method+" "+path] = struct{}{}
					}
				case enums.PermissionTypeMenu:
					menus[permission.ID] = permission
				}
				if permission.Code != "" {
					codes[permission.Code] = struct{}{}
				}
			}
		}

		idx.apiByRole[role.ID] = api
		if len(menus) > 0 {
			groupMenus := make([]Permission, 0, len(menus))
			for _, menu := range menus {
				groupMenus = append(groupMenus, menu)
			}
			sort.Slice(groupMenus, func(i, j int) bool { return groupMenus[i].ID < groupMenus[j].ID })
			idx.menusByRole[role.ID] = groupMenus
		}
		if len(codes) > 0 {
			roleCodes := make([]string, 0, len(codes))
			for code := range codes {
				roleCodes = append(roleCodes, code)
			}
			sort.Strings(roleCodes)
			idx.codesByRole[role.ID] = roleCodes
		}
	}

	return idx
}

// isSuperRole 判断是否为启用态超管角色
func isSuperRole(role Role) bool {
	return role.Code == enums.RoleCodeSuperAdmin && role.Status == enums.StatusEnabled
}

// toEntities 将快照权限转换为领域实体
func toEntities(permissions []Permission) []*entity.Permission {
	items := make([]*entity.Permission, 0, len(permissions))
	for _, permission := range permissions {
		items = append(items, &entity.Permission{
			ID:         permission.ID,
			Code:       permission.Code,
			Name:       permission.Name,
			Type:       permission.Type,
			UIPath:     permission.UIPath,
			HTTPMethod: permission.HTTPMethod,
			HTTPPath:   permission.HTTPPath,
			Status:     permission.Status,
		})
	}
	return items
}
