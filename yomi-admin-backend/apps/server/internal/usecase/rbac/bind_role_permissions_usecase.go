package rbac

import (
	"context"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
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
	tx   dbtx.Executor
}

// NewBindRolePermissionsUsecase 创建角色权限绑定用例
func NewBindRolePermissionsUsecase(rbac rbaccontract.RbacAdminService, tx dbtx.Executor) *BindRolePermissionsUsecase {
	return &BindRolePermissionsUsecase{rbac: rbac, tx: tx}
}

// Execute 在事务中覆盖角色权限
func (uc *BindRolePermissionsUsecase) Execute(ctx context.Context, roleID int64, in *BindRolePermissionsInput) error {
	if roleID <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return uc.tx.Run(ctx, func(txCtx context.Context) error {
		return uc.rbac.BindRolePermissions(txCtx, &rbacinput.BindRolePermissionsInput{
			RoleID:        roleID,
			PermissionIDs: in.PermissionIDs.Int64Slice(),
			OperatorID:    in.OperatorID,
		})
	})
}
