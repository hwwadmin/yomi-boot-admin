package rbac

import (
	"context"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// BindGroupPermissionsInput HTTP 权限组权限绑定入参
type BindGroupPermissionsInput struct {
	GroupID       dto.JSONInt64      `json:"groupId"`
	PermissionIDs dto.JSONInt64Slice `json:"permissionIds"`
	OperatorID    int64              `json:"-"`
}

// BindGroupPermissionsUsecase 权限组权限绑定编排
type BindGroupPermissionsUsecase struct {
	rbac rbaccontract.RbacAdminService
	tx   dbtx.Executor
}

// NewBindGroupPermissionsUsecase 创建权限组权限绑定用例
func NewBindGroupPermissionsUsecase(rbac rbaccontract.RbacAdminService, tx dbtx.Executor) *BindGroupPermissionsUsecase {
	return &BindGroupPermissionsUsecase{rbac: rbac, tx: tx}
}

// Execute 在事务中覆盖权限组权限
func (uc *BindGroupPermissionsUsecase) Execute(ctx context.Context, in *BindGroupPermissionsInput) error {
	groupID := in.GroupID.Int64()
	if groupID <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return uc.tx.Run(ctx, func(txCtx context.Context) error {
		return uc.rbac.BindGroupPermissions(txCtx, &rbacinput.BindGroupPermissionsInput{
			GroupID:       groupID,
			PermissionIDs: in.PermissionIDs.Int64Slice(),
			OperatorID:    in.OperatorID,
		})
	})
}
