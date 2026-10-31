package rbac

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// RoleDetailOutput 角色详情：已绑权限组、组内权限的去重聚合，以及实际生效的权限子集
type RoleDetailOutput struct {
	RoleItem
	GroupIDs               dto.JSONInt64Slice `json:"groupIds"`
	PermissionIDs          dto.JSONInt64Slice `json:"permissionIds"`
	EffectivePermissionIDs dto.JSONInt64Slice `json:"effectivePermissionIds"`
}

// GetRoleUsecase 角色详情编排
type GetRoleUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewGetRoleUsecase 创建角色详情用例
func NewGetRoleUsecase(rbac rbaccontract.RbacAdminService) *GetRoleUsecase {
	return &GetRoleUsecase{rbac: rbac}
}

// Execute 查询角色详情
func (uc *GetRoleUsecase) Execute(ctx context.Context, id int64) (*RoleDetailOutput, error) {
	if id <= 0 {
		return nil, errors.Sentinel(rbaccontract.ErrInvalidInput)
	}

	detail, err := uc.rbac.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	return &RoleDetailOutput{
		RoleItem:               toRoleItem(detail.Role),
		GroupIDs:               support.NonNilIDs(dto.NewJSONInt64Slice(detail.GroupIDs)),
		PermissionIDs:          support.NonNilIDs(dto.NewJSONInt64Slice(detail.PermissionIDs)),
		EffectivePermissionIDs: support.NonNilIDs(dto.NewJSONInt64Slice(detail.EffectivePermissionIDs)),
	}, nil
}

// toRoleItem 将角色实体转换为 HTTP 视图
func toRoleItem(role *entity.Role) RoleItem {
	return RoleItem{
		ID:         dto.JSONInt64(role.ID),
		Code:       role.Code,
		Name:       role.Name,
		Builtin:    role.Builtin,
		Status:     role.Status,
		Remark:     role.Remark,
		CreateTime: role.CreateTime,
		UpdateTime: role.UpdateTime,
	}
}
