// Package service RBAC 管理服务实现
package service

import (
	"context"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"
	"yomi-admin-backend/domains/rbac/pkg/shared/input"
	"yomi-admin-backend/domains/rbac/pkg/shared/output"
	"yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/snowflake"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

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
	items, total, err := s.repo.PageRoles(ctx, in.Keyword, in.Status, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}
	return &output.PageRolesOutput{Items: items, Total: total}, nil
}

// GetRole 查询角色详情
func (s *RbacAdminService) GetRole(ctx context.Context, id int64) (*entity.RoleDetail, error) {
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
	if in.Code == enums.RoleCodeSuperAdmin {
		return nil, errors.Sentinel(service_contract.ErrBuiltinRoleImmutable)
	}

	_, found, err := s.repo.FindRoleByCode(ctx, in.Code)
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
		Code:   in.Code,
		Name:   in.Name,
		Status: enums.StatusEnabled,
		Remark: in.Remark,
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

	role.Name = in.Name
	role.Remark = in.Remark
	role.Status = in.Status
	return s.repo.UpdateRole(ctx, role, in.OperatorID)
}

// BindRolePermissions 覆盖角色权限
func (s *RbacAdminService) BindRolePermissions(ctx context.Context, in *input.BindRolePermissionsInput) error {
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

	permissionIDs := uniqueIDs(in.PermissionIDs)
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
	items, total, err := s.repo.PagePermissions(ctx, in.Keyword, in.Type, in.Status, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}
	return &output.PagePermissionsOutput{Items: items, Total: total}, nil
}

// SetPermissionStatus 设置权限启停
func (s *RbacAdminService) SetPermissionStatus(ctx context.Context, in *input.SetPermissionStatusInput) error {
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

// BindUserRoles 覆盖用户角色
func (s *RbacAdminService) BindUserRoles(ctx context.Context, in *input.BindUserRolesInput) error {
	roleIDs := uniqueIDs(in.RoleIDs)

	roles, err := s.repo.FindRolesByIDs(ctx, roleIDs)
	if err != nil {
		return err
	}
	if len(roles) != len(roleIDs) {
		return errors.Sentinel(service_contract.ErrRoleNotFound)
	}

	willHaveSuper := false
	for _, role := range roles {
		if role.Status != enums.StatusEnabled {
			return errors.Sentinel(service_contract.ErrRelatedDisabled)
		}
		if role.Code == enums.RoleCodeSuperAdmin {
			willHaveSuper = true
		}
	}

	actorSuper, err := s.repo.UserHasEnabledSuperAdmin(ctx, in.OperatorID)
	if err != nil {
		return err
	}
	targetSuper, err := s.repo.UserHasEnabledSuperAdmin(ctx, in.UserID)
	if err != nil {
		return err
	}

	// 授予或回收超管角色都要求操作者本身是超管
	if (willHaveSuper || targetSuper) && !actorSuper {
		return errors.Sentinel(service_contract.ErrGrantSuperDenied)
	}
	// 超管不能解绑自己的超管角色
	if in.OperatorID == in.UserID && targetSuper && !willHaveSuper {
		return errors.Sentinel(service_contract.ErrUnbindSelfSuper)
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

// ListUserIDsWithEnabledSuperAdmin 列出拥有启用态超管角色的用户 ID
func (s *RbacAdminService) ListUserIDsWithEnabledSuperAdmin(ctx context.Context) ([]int64, error) {
	return s.repo.ListUserIDsWithEnabledSuperAdmin(ctx)
}

// uniqueIDs 按首次出现顺序去重
func uniqueIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
