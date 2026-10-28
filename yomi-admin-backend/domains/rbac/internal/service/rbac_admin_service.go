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
	maxRoleCodeLength = 32
	maxRoleNameRunes  = 32
	maxRemarkRunes    = 255
)

// roleCodePattern 约束角色编码为小写字母开头的下划线组合
var roleCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,31}$`)

// RbacAdminService 角色与权限管理实现
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

// GetRole 查询角色详情
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

	permissionIDs, err := s.repo.ListPermissionIDsByRoleID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &entity.RoleDetail{Role: role, PermissionIDs: permissionIDs}, nil
}

// CreateRole 创建角色
func (s *RbacAdminService) CreateRole(ctx context.Context, in *input.CreateRoleInput) (*entity.Role, error) {
	if in == nil {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	roleCode := strings.TrimSpace(in.Code)
	roleName := strings.TrimSpace(in.Name)
	remark := strings.TrimSpace(in.Remark)
	if !validRoleCode(roleCode) || !validRoleName(roleName) || !validRemark(remark) {
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
	if !validRoleName(roleName) || !validRemark(remark) || !enums.IsValidStatus(in.Status) {
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

// BindRolePermissions 覆盖角色权限
func (s *RbacAdminService) BindRolePermissions(ctx context.Context, in *input.BindRolePermissionsInput) error {
	if in == nil || in.RoleID <= 0 {
		return errors.Sentinel(service_contract.ErrRoleNotFound)
	}

	permissionIDs, err := normalizeIDs(in.PermissionIDs)
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

	rels := make([]*entity.RolePermissionRel, 0, len(permissionIDs))
	for _, permissionID := range permissionIDs {
		id, err := s.ids.NextID()
		if err != nil {
			return errors.Wrap(err, "生成角色权限关系 ID 失败")
		}
		rels = append(rels, &entity.RolePermissionRel{
			ID:           id,
			RoleID:       role.ID,
			PermissionID: permissionID,
		})
	}
	return s.repo.ReplaceRolePermissions(ctx, role.ID, rels, in.OperatorID)
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

// BindUserRoles 覆盖用户角色；超管相关授权规则由编排层策略负责
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

// validRoleCode 判断角色编码是否合法
func validRoleCode(code string) bool {
	return len(code) <= maxRoleCodeLength && roleCodePattern.MatchString(code)
}

// validRoleName 判断角色名称是否合法
func validRoleName(name string) bool {
	return name != "" && utf8.RuneCountInString(name) <= maxRoleNameRunes
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
