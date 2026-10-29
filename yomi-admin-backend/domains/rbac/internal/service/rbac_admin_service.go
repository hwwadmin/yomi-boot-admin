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

	"gitee.com/hq_project/yomi-boot/component/common/snowflake"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

const (
	maxCodeLength  = 32
	maxNameRunes   = 32
	maxRemarkRunes = 255
)

// codePattern 约束角色与权限组编码为小写字母开头的下划线组合
var codePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,31}$`)

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

// GetRole 查询角色详情：已绑权限组 + 组内权限的去重聚合
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
	return &entity.RoleDetail{Role: role, GroupIDs: groupIDs, PermissionIDs: permissionIDs}, nil
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

// UpdateRole 更新角色资料与状态
func (s *RbacAdminService) UpdateRole(ctx context.Context, in *input.UpdateRoleInput) error {
	if in == nil || in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrRoleNotFound)
	}

	roleName := strings.TrimSpace(in.Name)
	remark := strings.TrimSpace(in.Remark)
	if !validName(roleName) || !validRemark(remark) || !enums.IsValidStatus(in.Status) {
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

	role.Name = roleName
	role.Remark = remark
	role.Status = in.Status
	return s.repo.UpdateRole(ctx, role, in.OperatorID)
}

// BindRoleGroups 覆盖角色权限组；角色权限始终来自权限组的去重聚合
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
		for _, group := range groups {
			if group.Status != enums.StatusEnabled {
				return errors.Sentinel(service_contract.ErrRelatedDisabled)
			}
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

// UpdatePermissionGroup 更新权限组资料与状态
func (s *RbacAdminService) UpdatePermissionGroup(ctx context.Context, in *input.UpdatePermissionGroupInput) error {
	if in == nil || in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrPermissionGroupNotFound)
	}

	groupName := strings.TrimSpace(in.Name)
	remark := strings.TrimSpace(in.Remark)
	if !validName(groupName) || !validRemark(remark) || !enums.IsValidStatus(in.Status) {
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
	group.Status = in.Status
	return s.repo.UpdatePermissionGroup(ctx, group, in.OperatorID)
}

// BindGroupPermissions 覆盖权限组权限
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
		for _, permission := range permissions {
			if permission.Status != enums.StatusEnabled {
				return errors.Sentinel(service_contract.ErrRelatedDisabled)
			}
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

	roleIDs, err := normalizeIDs(in.RoleIDs)
	if err != nil {
		return err
	}

	roles, err := s.repo.FindRolesByIDs(ctx, roleIDs)
	if err != nil {
		return err
	}
	if len(roles) != len(roleIDs) {
		return errors.Sentinel(service_contract.ErrRoleNotFound)
	}
	for _, role := range roles {
		if role.Status != enums.StatusEnabled {
			return errors.Sentinel(service_contract.ErrRelatedDisabled)
		}
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

// ListRolesByUserIDs 按用户批量查询已绑角色
func (s *RbacAdminService) ListRolesByUserIDs(ctx context.Context, userIDs []int64) (map[int64][]*entity.Role, error) {
	return s.repo.ListRolesByUserIDs(ctx, userIDs)
}

// ListRolesByIDs 按主键批量查询角色
func (s *RbacAdminService) ListRolesByIDs(ctx context.Context, ids []int64) ([]*entity.Role, error) {
	roleIDs, err := normalizeIDs(ids)
	if err != nil {
		return nil, err
	}
	if len(roleIDs) == 0 {
		return []*entity.Role{}, nil
	}
	return s.repo.FindRolesByIDs(ctx, roleIDs)
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
