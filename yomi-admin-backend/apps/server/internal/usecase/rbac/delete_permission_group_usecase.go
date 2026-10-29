package rbac

import (
	"context"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// DeletePermissionGroupInput HTTP 删除权限组入参
type DeletePermissionGroupInput struct {
	ID         dto.JSONInt64 `json:"id"`
	OperatorID int64         `json:"-"`
}

// DeletePermissionGroupUsecase 删除权限组编排
type DeletePermissionGroupUsecase struct {
	rbac rbaccontract.RbacAdminService
	tx   dbtx.Executor
}

// NewDeletePermissionGroupUsecase 创建删除权限组用例
func NewDeletePermissionGroupUsecase(rbac rbaccontract.RbacAdminService, tx dbtx.Executor) *DeletePermissionGroupUsecase {
	return &DeletePermissionGroupUsecase{rbac: rbac, tx: tx}
}

// Execute 在事务中硬删除权限组
func (uc *DeletePermissionGroupUsecase) Execute(ctx context.Context, in *DeletePermissionGroupInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return uc.tx.Run(ctx, func(txCtx context.Context) error {
		return uc.rbac.DeletePermissionGroup(txCtx, &rbacinput.DeletePermissionGroupInput{
			ID:         id,
			OperatorID: in.OperatorID,
		})
	})
}
