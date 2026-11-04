// Package service RBAC 管理服务实现
package service

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"
	"yomi-admin-backend/domains/rbac/pkg/shared/input"
	"yomi-admin-backend/domains/rbac/pkg/shared/output"
	"yomi-admin-backend/domains/rbac/service_contract"

	"github.com/hwwadmin/yomi-boot/component/common/snowflake"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

const (
	maxCodeLength     = 32
	maxPermissionCode = 96
	maxNameRunes      = 32
	maxRemarkRunes    = 255
	maxResourceLength = 191
)

// codePattern 约束角色与权限组编码为小写字母开头的下划线组合
var codePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,31}$`)

// permissionCodePattern 约束权限编码为小写字母开头、冒号分段的小写标识
var permissionCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?::[a-z][a-z0-9_]*)*$`)

// RbacAdminService 角色、权限组与权限管理实现
type RbacAdminService struct {
	repo repo_contract.RbacRepo
	ids  snowflake.Generator
}

var _ service_contract.RbacAdminService = (*RbacAdminService)(nil)

// NewRbacAdminService 创建 RBAC 管理服务
func NewRbacAdminService(repo repo_contract.RbacRepo, ids snowflake.Generator) *RbacAdminService {
	return &RbacAdminService{repo: repo, ids: ids}
}

// PageRoles 分页查询角色
func (s *RbacAdminService) PageRoles(ctx context.Context, in *input.PageRolesInput) (*output.PageRolesOutput, error) {
	if in == nil {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.Status != "" && !enums.IsValidStatus(in.Status) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	items, total, err := s.repo.PageRoles(ctx, in.Keyword, in.Status, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}
	return &output.PageRolesOutput{Items: items, Total: total}, nil
}

// GetRole 查询角色详情：已绑权限组、组内权限的去重聚合，以及实际生效的权限子集
func (s *RbacAdminService) GetRole(ctx context.Context, id int64) (*entity.RoleDetail, error) {
	if id <= 0 {
		return nil, errors.Sentinel(service_contract.ErrRoleNotFound)
	}

	role, found, err := s.repo.FindRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.Sentinel(service_contract.ErrRoleNotFound)
	}

	groupIDs, err := s.repo.ListGroupIDsByRoleID(ctx, id)
	if err != nil {
		return nil, err
	}
	permissionIDs, err := s.repo.ListPermissionIDsByRoleID(ctx, id)
	if err != nil {
		return nil, err
	}
	effectiveIDs, err := s.repo.ListEffectivePermissionIDsByRoleID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.RoleDetail{
		Role:                   role,
		GroupIDs:               groupIDs,
		PermissionIDs:          permissionIDs,
		EffectivePermissionIDs: effectiveIDs,
	}, nil
}

// CreateRole 创建角色
func (s *RbacAdminService) CreateRole(ctx context.Context, in *input.CreateRoleInput) (*entity.Role, error) {
	if in == nil {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	roleCode := strings.TrimSpace(in.Code)
	roleName := strings.TrimSpace(in.Name)
	remark := strings.TrimSpace(in.Remark)
	if !validCode(roleCode) || !validName(roleName) || !validRemark(remark) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if roleCode == enums.RoleCodeSuperAdmin {
		return nil, errors.Sentinel(service_contract.ErrBuiltinRoleImmutable)
	}

	_, found, err := s.repo.FindRoleByCode(ctx, roleCode)
	if err != nil {
		return nil, err
	}
	if found {
		return nil, errors.Sentinel(service_contract.ErrRoleCodeExists)
	}

	id, err := s.ids.NextID()
	if err != nil {
		return nil, errors.Wrap(err, "生成角色 ID 失败")
	}
	role := &entity.Role{
		ID:     id,
		Code:   roleCode,
		Name:   roleName,
		Status: enums.StatusEnabled,
		Remark: remark,
	}
	if err := s.repo.CreateRole(ctx, role, in.OperatorID); err != nil {
		if errors.Is(err, repo_contract.ErrDuplicatedKey) {
			return nil, errors.Sentinel(service_contract.ErrRoleCodeExists)
		}
		return nil, err
	}
	return role, nil
}

// UpdateRole 更新角色资料；启停走独立入口
func (s *RbacAdminService) UpdateRole(ctx context.Context, in *input.UpdateRoleInput) error {
	if in == nil || in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrRoleNotFound)
	}

	roleName := strings.TrimSpace(in.Name)
	remark := strings.TrimSpace(in.Remark)
	if !validName(roleName) || !validRemark(remark) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	role, found, err := s.repo.FindRoleByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrRoleNotFound)
	}

	role.Name = roleName
	role.Remark = remark
	return s.repo.UpdateRole(ctx, role, in.OperatorID)
}

// DeleteRole 硬删除角色：仍有用户绑定时拒绝，并同步清理角色权限组关系
func (s *RbacAdminService) DeleteRole(ctx context.Context, in *input.DeleteRoleInput) error {
	if in == nil || in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrRoleNotFound)
	}

	role, found, err := s.repo.FindRoleByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrRoleNotFound)
	}
	if role.Builtin {
		return errors.Sentinel(service_contract.ErrBuiltinRoleImmutable)
	}

	boundUsers, err := s.repo.CountUserRolesByRoleID(ctx, role.ID)
	if err != nil {
		return err
	}
	if boundUsers > 0 {
		return errors.Sentinel(service_contract.ErrRoleBoundToUsers)
	}

	if err := s.repo.DeleteRoleGroupsByRoleID(ctx, role.ID); err != nil {
		return err
	}
	return s.repo.DeleteRole(ctx, role.ID)
}

// SetRoleStatus 设置角色启停；内置超管不可禁用
func (s *RbacAdminService) SetRoleStatus(ctx context.Context, in *input.SetRoleStatusInput) error {
	if in == nil || in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrRoleNotFound)
	}
	if !enums.IsValidStatus(in.Status) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	role, found, err := s.repo.FindRoleByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrRoleNotFound)
	}
	if role.Builtin && in.Status == enums.StatusDisabled {
		return errors.Sentinel(service_contract.ErrBuiltinRoleImmutable)
	}
	return s.repo.UpdateRoleStatus(ctx, role.ID, in.Status, in.OperatorID)
}

// BindRoleGroups 覆盖角色权限组；角色权限始终来自权限组的去重聚合。
//
// 停用的权限组允许绑定：它不会授予任何权限，由界面标注状态；
// 实际生效范围取决于权限组与权限本身的启停。
func (s *RbacAdminService) BindRoleGroups(ctx context.Context, in *input.BindRoleGroupsInput) error {
	if in == nil || in.RoleID <= 0 {
		return errors.Sentinel(service_contract.ErrRoleNotFound)
	}

	groupIDs, err := normalizeIDs(in.GroupIDs)
	if err != nil {
		return err
	}

	role, found, err := s.repo.FindRoleByID(ctx, in.RoleID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrRoleNotFound)
	}
	if role.Builtin {
		return errors.Sentinel(service_contract.ErrBuiltinRoleImmutable)
	}

	if len(groupIDs) > 0 {
		groups, err := s.repo.FindPermissionGroupsByIDs(ctx, groupIDs)
		if err != nil {
			return err
		}
		if len(groups) != len(groupIDs) {
			return errors.Sentinel(service_contract.ErrPermissionGroupNotFound)
		}
	}

	rels := make([]*entity.RoleGroupRel, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		id, err := s.ids.NextID()
		if err != nil {
			return errors.Wrap(err, "生成角色权限组关系 ID 失败")
		}
		rels = append(rels, &entity.RoleGroupRel{
			ID:      id,
			RoleID:  role.ID,
			GroupID: groupID,
		})
	}
	return s.repo.ReplaceRoleGroups(ctx, role.ID, rels, in.OperatorID)
}

// PagePermissionGroups 分页查询权限组
func (s *RbacAdminService) PagePermissionGroups(ctx context.Context, in *input.PagePermissionGroupsInput) (*output.PagePermissionGroupsOutput, error) {
	if in == nil {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.Status != "" && !enums.IsValidStatus(in.Status) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	items, total, err := s.repo.PagePermissionGroups(ctx, in.Keyword, in.Status, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}
	return &output.PagePermissionGroupsOutput{Items: items, Total: total}, nil
}

// GetPermissionGroup 查询权限组详情
func (s *RbacAdminService) GetPermissionGroup(ctx context.Context, id int64) (*entity.PermissionGroupDetail, error) {
	if id <= 0 {
		return nil, errors.Sentinel(service_contract.ErrPermissionGroupNotFound)
	}

	group, found, err := s.repo.FindPermissionGroupByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.Sentinel(service_contract.ErrPermissionGroupNotFound)
	}

	permissionIDs, err := s.repo.ListPermissionIDsByGroupID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.PermissionGroupDetail{Group: group, PermissionIDs: permissionIDs}, nil
}

// CreatePermissionGroup 创建权限组
func (s *RbacAdminService) CreatePermissionGroup(ctx context.Context, in *input.CreatePermissionGroupInput) (*entity.PermissionGroup, error) {
	if in == nil {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	code := strings.TrimSpace(in.Code)
	groupName := strings.TrimSpace(in.Name)
	remark := strings.TrimSpace(in.Remark)
	if !validCode(code) || !validName(groupName) || !validRemark(remark) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	_, found, err := s.repo.FindPermissionGroupByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if found {
		return nil, errors.Sentinel(service_contract.ErrPermissionGroupCodeExists)
	}

	id, err := s.ids.NextID()
	if err != nil {
		return nil, errors.Wrap(err, "生成权限组 ID 失败")
	}
	group := &entity.PermissionGroup{
		ID:     id,
		Code:   code,
		Name:   groupName,
		Status: enums.StatusEnabled,
		Remark: remark,
	}
	if err := s.repo.CreatePermissionGroup(ctx, group, in.OperatorID); err != nil {
		if errors.Is(err, repo_contract.ErrDuplicatedKey) {
			return nil, errors.Sentinel(service_contract.ErrPermissionGroupCodeExists)
		}
		return nil, err
	}
	return group, nil
}

// UpdatePermissionGroup 更新权限组资料；启停走独立入口
func (s *RbacAdminService) UpdatePermissionGroup(ctx context.Context, in *input.UpdatePermissionGroupInput) error {
	if in == nil || in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrPermissionGroupNotFound)
	}

	groupName := strings.TrimSpace(in.Name)
	remark := strings.TrimSpace(in.Remark)
	if !validName(groupName) || !validRemark(remark) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	group, found, err := s.repo.FindPermissionGroupByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrPermissionGroupNotFound)
	}

	group.Name = groupName
	group.Remark = remark
	return s.repo.UpdatePermissionGroup(ctx, group, in.OperatorID)
}

// DeletePermissionGroup 硬删除权限组：仍有角色绑定时拒绝，并同步清理权限组权限关系
func (s *RbacAdminService) DeletePermissionGroup(ctx context.Context, in *input.DeletePermissionGroupInput) error {
	if in == nil || in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrPermissionGroupNotFound)
	}

	group, found, err := s.repo.FindPermissionGroupByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrPermissionGroupNotFound)
	}

	boundRoles, err := s.repo.CountRoleGroupsByGroupID(ctx, group.ID)
	if err != nil {
		return err
	}
	if boundRoles > 0 {
		return errors.Sentinel(service_contract.ErrPermissionGroupBoundToRoles)
	}

	if err := s.repo.DeleteGroupPermissionsByGroupID(ctx, group.ID); err != nil {
		return err
	}
	return s.repo.DeletePermissionGroup(ctx, group.ID)
}

// SetPermissionGroupStatus 设置权限组启停
func (s *RbacAdminService) SetPermissionGroupStatus(ctx context.Context, in *input.SetPermissionGroupStatusInput) error {
	if in == nil || in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrPermissionGroupNotFound)
	}
	if !enums.IsValidStatus(in.Status) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	group, found, err := s.repo.FindPermissionGroupByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrPermissionGroupNotFound)
	}
	return s.repo.UpdatePermissionGroupStatus(ctx, group.ID, in.Status, in.OperatorID)
}

// BindGroupPermissions 覆盖权限组权限。
//
// 停用的权限允许绑定：它不会授予任何权限，由界面标注状态。
func (s *RbacAdminService) BindGroupPermissions(ctx context.Context, in *input.BindGroupPermissionsInput) error {
	if in == nil || in.GroupID <= 0 {
		return errors.Sentinel(service_contract.ErrPermissionGroupNotFound)
	}

	permissionIDs, err := normalizeIDs(in.PermissionIDs)
	if err != nil {
		return err
	}

	group, found, err := s.repo.FindPermissionGroupByID(ctx, in.GroupID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrPermissionGroupNotFound)
	}

	if len(permissionIDs) > 0 {
		permissions, err := s.repo.FindPermissionsByIDs(ctx, permissionIDs)
		if err != nil {
			return err
		}
		if len(permissions) != len(permissionIDs) {
			return errors.Sentinel(service_contract.ErrPermissionNotFound)
		}
	}

	rels := make([]*entity.GroupPermissionRel, 0, len(permissionIDs))
	for _, permissionID := range permissionIDs {
		id, err := s.ids.NextID()
		if err != nil {
			return errors.Wrap(err, "生成权限组权限关系 ID 失败")
		}
		rels = append(rels, &entity.GroupPermissionRel{
			ID:           id,
			GroupID:      group.ID,
			PermissionID: permissionID,
		})
	}
	return s.repo.ReplaceGroupPermissions(ctx, group.ID, rels, in.OperatorID)
}

// PagePermissions 分页查询权限目录
func (s *RbacAdminService) PagePermissions(ctx context.Context, in *input.PagePermissionsInput) (*output.PagePermissionsOutput, error) {
	if in == nil {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.Type != "" && !enums.IsValidPermissionType(in.Type) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.Status != "" && !enums.IsValidStatus(in.Status) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	items, total, err := s.repo.PagePermissions(ctx, in.Keyword, in.Type, in.Status, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}
	return &output.PagePermissionsOutput{Items: items, Total: total}, nil
}

// CreatePermission 创建权限
func (s *RbacAdminService) CreatePermission(ctx context.Context, in *input.CreatePermissionInput) (*entity.Permission, error) {
	if in == nil {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	code := strings.TrimSpace(in.Code)
	permissionName := strings.TrimSpace(in.Name)
	remark := strings.TrimSpace(in.Remark)
	if !validPermissionCode(code) || !validName(permissionName) || !validRemark(remark) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	permission, ok := normalizePermissionResource(in.Type, in.UIPath, in.HTTPMethod, in.HTTPPath)
	if !ok {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	_, found, err := s.repo.FindPermissionByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if found {
		return nil, errors.Sentinel(service_contract.ErrPermissionCodeExists)
	}
	if err := s.ensurePermissionResourceFree(ctx, permission, 0); err != nil {
		return nil, err
	}

	id, err := s.ids.NextID()
	if err != nil {
		return nil, errors.Wrap(err, "生成权限 ID 失败")
	}
	permission.ID = id
	permission.Code = code
	permission.Name = permissionName
	permission.Status = enums.StatusEnabled
	permission.Remark = remark
	if err := s.repo.CreatePermission(ctx, permission, in.OperatorID); err != nil {
		if errors.Is(err, repo_contract.ErrDuplicatedKey) {
			return nil, s.resolveCreatePermissionConflict(ctx, code, permission)
		}
		return nil, err
	}
	return permission, nil
}

// UpdatePermission 更新权限资料；编码与类型不可变
func (s *RbacAdminService) UpdatePermission(ctx context.Context, in *input.UpdatePermissionInput) error {
	if in == nil || in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrPermissionNotFound)
	}

	permissionName := strings.TrimSpace(in.Name)
	remark := strings.TrimSpace(in.Remark)
	if !validName(permissionName) || !validRemark(remark) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	permission, found, err := s.repo.FindPermissionByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrPermissionNotFound)
	}

	resource, ok := normalizePermissionResource(permission.Type, in.UIPath, in.HTTPMethod, in.HTTPPath)
	if !ok {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if err := s.ensurePermissionResourceFree(ctx, resource, permission.ID); err != nil {
		return err
	}

	permission.Name = permissionName
	permission.UIPath = resource.UIPath
	permission.HTTPMethod = resource.HTTPMethod
	permission.HTTPPath = resource.HTTPPath
	permission.Remark = remark
	if err := s.repo.UpdatePermission(ctx, permission, in.OperatorID); err != nil {
		if errors.Is(err, repo_contract.ErrDuplicatedKey) {
			return errors.Sentinel(service_contract.ErrPermissionResourceExists)
		}
		return err
	}
	return nil
}

// DeletePermission 硬删除权限，并同步清理权限组对它的引用
func (s *RbacAdminService) DeletePermission(ctx context.Context, in *input.DeletePermissionInput) error {
	if in == nil || in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrPermissionNotFound)
	}

	_, found, err := s.repo.FindPermissionByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrPermissionNotFound)
	}

	if err := s.repo.DeleteGroupPermissionsByPermissionID(ctx, in.ID); err != nil {
		return err
	}
	return s.repo.DeletePermission(ctx, in.ID)
}

// SetPermissionStatus 设置权限启停
func (s *RbacAdminService) SetPermissionStatus(ctx context.Context, in *input.SetPermissionStatusInput) error {
	if in == nil || in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrPermissionNotFound)
	}
	if !enums.IsValidStatus(in.Status) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	_, found, err := s.repo.FindPermissionByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrPermissionNotFound)
	}
	return s.repo.UpdatePermissionStatus(ctx, in.ID, in.Status, in.OperatorID)
}

// ListUserRoles 查询用户角色绑定视图
func (s *RbacAdminService) ListUserRoles(ctx context.Context, userID int64) (*output.UserRolesOutput, error) {
	if userID <= 0 {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	roles, err := s.repo.ListAllRoles(ctx)
	if err != nil {
		return nil, err
	}
	roleIDs, err := s.repo.ListRoleIDsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &output.UserRolesOutput{Roles: roles, RoleIDs: roleIDs}, nil
}

// BindUserRoles 覆盖用户角色；超管授予/回收、自解绑与最后一名超管校验由编排层策略负责
func (s *RbacAdminService) BindUserRoles(ctx context.Context, in *input.BindUserRolesInput) error {
	if in == nil || in.UserID <= 0 {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	roleIDs, _, err := s.resolveBindingRoles(ctx, in.RoleIDs)
	if err != nil {
		return err
	}

	rels := make([]*entity.UserRoleRel, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		id, err := s.ids.NextID()
		if err != nil {
			return errors.Wrap(err, "生成用户角色关系 ID 失败")
		}
		rels = append(rels, &entity.UserRoleRel{
			ID:     id,
			UserID: in.UserID,
			RoleID: roleID,
		})
	}
	return s.repo.ReplaceUserRoles(ctx, in.UserID, rels, in.OperatorID)
}

// DeleteUserRoles 硬删除用户全部角色绑定
func (s *RbacAdminService) DeleteUserRoles(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	return s.repo.DeleteUserRolesByUserID(ctx, userID)
}

// ListRolesByUserIDs 按用户批量查询已绑角色
func (s *RbacAdminService) ListRolesByUserIDs(ctx context.Context, userIDs []int64) (map[int64][]*entity.Role, error) {
	return s.repo.ListRolesByUserIDs(ctx, userIDs)
}

// ResolveBindingRoles 供编排层在超管策略检查前取得待绑定的角色。
// 规则与 BindUserRoles 完全一致：ID 去重、角色必须存在且处于启用态。
func (s *RbacAdminService) ResolveBindingRoles(ctx context.Context, ids []int64) ([]*entity.Role, error) {
	_, roles, err := s.resolveBindingRoles(ctx, ids)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// resolveBindingRoles 是“角色能否绑定”规则的唯一实现：规范化 ID、校验存在性与启用态。
func (s *RbacAdminService) resolveBindingRoles(ctx context.Context, ids []int64) ([]int64, []*entity.Role, error) {
	roleIDs, err := normalizeIDs(ids)
	if err != nil {
		return nil, nil, err
	}

	roles, err := s.repo.FindRolesByIDs(ctx, roleIDs)
	if err != nil {
		return nil, nil, err
	}
	if len(roles) != len(roleIDs) {
		return nil, nil, errors.Sentinel(service_contract.ErrRoleNotFound)
	}
	for _, role := range roles {
		if role.Status != enums.StatusEnabled {
			return nil, nil, errors.Sentinel(service_contract.ErrRelatedDisabled)
		}
	}
	return roleIDs, roles, nil
}

// ListUserIDsWithEnabledSuperAdmin 列出拥有启用态超管角色的用户 ID
func (s *RbacAdminService) ListUserIDsWithEnabledSuperAdmin(ctx context.Context) ([]int64, error) {
	return s.repo.ListUserIDsWithEnabledSuperAdmin(ctx)
}

// validCode 判断角色与权限组编码是否合法
func validCode(code string) bool {
	return len(code) <= maxCodeLength && codePattern.MatchString(code)
}

// validName 判断名称是否合法
func validName(name string) bool {
	return name != "" && utf8.RuneCountInString(name) <= maxNameRunes
}

// validRemark 判断备注长度是否合法
func validRemark(remark string) bool {
	return utf8.RuneCountInString(remark) <= maxRemarkRunes
}

// validPermissionCode 判断权限编码是否合法
func validPermissionCode(code string) bool {
	return len(code) >= 3 && len(code) <= maxPermissionCode && permissionCodePattern.MatchString(code)
}

// validUIPath 判断菜单路径是否合法
func validUIPath(uiPath string) bool {
	return strings.HasPrefix(uiPath, "/") && len(uiPath) <= maxResourceLength && !containsSpace(uiPath)
}

// validHTTPPath 判断接口路由模板是否合法：以 / 开头、长度受限、不含空白。
//
// 不强制 /api 前缀，便于对接网关改写或多前缀部署；路径是否真能匹配到路由由配置者自行保证。
func validHTTPPath(httpPath string) bool {
	return strings.HasPrefix(httpPath, "/") && len(httpPath) <= maxResourceLength && !containsSpace(httpPath)
}

// validHTTPMethod 判断 HTTP 方法是否合法
func validHTTPMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}

// containsSpace 判断字符串是否包含空白字符
func containsSpace(value string) bool {
	return strings.ContainsAny(value, " \t\r\n")
}

// normalizePermissionResource 按权限类型归一化资源字段：菜单只看 uiPath，接口只看 method + path
func normalizePermissionResource(permissionType enums.PermissionType, uiPath, httpMethod, httpPath string) (*entity.Permission, bool) {
	uiPath = strings.TrimSpace(uiPath)
	httpMethod = strings.ToUpper(strings.TrimSpace(httpMethod))
	httpPath = strings.TrimSpace(httpPath)

	switch permissionType {
	case enums.PermissionTypeMenu:
		if !validUIPath(uiPath) {
			return nil, false
		}
		return &entity.Permission{Type: permissionType, UIPath: uiPath}, true
	case enums.PermissionTypeAPI:
		if !validHTTPMethod(httpMethod) || !validHTTPPath(httpPath) {
			return nil, false
		}
		return &entity.Permission{Type: permissionType, HTTPMethod: httpMethod, HTTPPath: httpPath}, true
	default:
		return nil, false
	}
}

// resolveCreatePermissionConflict 在插入撞唯一键时回查冲突字段，返回准确的业务错误。
//
// 仓储把重复键统一收敛成一个哨兵，无法区分撞的是编码还是资源；预检查与插入之间存在
// 并发窗口，因此这里按“编码优先”的顺序回查确认，避免把编码冲突误报成资源冲突。
func (s *RbacAdminService) resolveCreatePermissionConflict(ctx context.Context, code string, permission *entity.Permission) error {
	_, found, err := s.repo.FindPermissionByCode(ctx, code)
	if err != nil {
		return err
	}
	if found {
		return errors.Sentinel(service_contract.ErrPermissionCodeExists)
	}

	if err := s.ensurePermissionResourceFree(ctx, permission, 0); err != nil {
		return err
	}

	// 冲突行在回查前已被删除等极端情况：保守返回资源冲突，不暴露数据库细节
	return errors.Sentinel(service_contract.ErrPermissionResourceExists)
}

// ensurePermissionResourceFree 校验权限资源未被其它权限占用
func (s *RbacAdminService) ensurePermissionResourceFree(ctx context.Context, permission *entity.Permission, excludeID int64) error {
	var (
		existing *entity.Permission
		found    bool
		err      error
	)
	if permission.Type == enums.PermissionTypeMenu {
		existing, found, err = s.repo.FindPermissionByUIPath(ctx, permission.UIPath)
	} else {
		existing, found, err = s.repo.FindPermissionByHTTPRoute(ctx, permission.HTTPMethod, permission.HTTPPath)
	}
	if err != nil {
		return err
	}
	if found && existing.ID != excludeID {
		return errors.Sentinel(service_contract.ErrPermissionResourceExists)
	}
	return nil
}

// normalizeIDs 校验并去重 ID 列表，保持首次出现顺序
func normalizeIDs(ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, errors.Sentinel(service_contract.ErrInvalidInput)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}
