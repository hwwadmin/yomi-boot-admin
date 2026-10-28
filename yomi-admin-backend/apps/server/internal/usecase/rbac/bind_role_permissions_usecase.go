package rbac

import (
	"context"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// BindRolePermissionsInput HTTP 角色权限绑定入参
type BindRolePermissionsInput struct {
	PermissionIDs dto.JSONInt64Slice `json:"permissionIds"`
	OperatorID    int64              `json:"-"`
}

// BindRolePermissionsUsecase 角色权限绑定编排
type BindRolePermissionsUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewBindRolePermissionsUsecase 创建角色权限绑定用例
func NewBindRolePermissionsUsecase(rbac rbaccontract.RbacAdminService) *BindRolePermissionsUsecase {
	return &BindRolePermissionsUsecase{rbac: rbac}
}

// Execute 覆盖角色权限
func (uc *BindRolePermissionsUsecase) Execute(ctx context.Context, roleID int64, in *BindRolePermissionsInput) error {
	if roleID <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return uc.rbac.BindRolePermissions(ctx, &rbacinput.BindRolePermissionsInput{
		RoleID:        roleID,
		PermissionIDs: in.PermissionIDs.Int64Slice(),
		OperatorID:    in.OperatorID,
	})
}
