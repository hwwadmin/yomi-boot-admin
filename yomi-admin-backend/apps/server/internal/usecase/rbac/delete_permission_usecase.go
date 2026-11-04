package rbac

import (
	"context"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	dbtx "github.com/hwwadmin/yomi-boot/component/common/db/tx"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// DeletePermissionInput HTTP 删除权限入参
type DeletePermissionInput struct {
	ID         dto.JSONInt64 `json:"id"`
	OperatorID int64         `json:"-"`
}

// DeletePermissionUsecase 删除权限编排
type DeletePermissionUsecase struct {
	rbac rbaccontract.RbacAdminService
	tx   dbtx.Executor
}

// NewDeletePermissionUsecase 创建删除权限用例
func NewDeletePermissionUsecase(rbac rbaccontract.RbacAdminService, tx dbtx.Executor) *DeletePermissionUsecase {
	return &DeletePermissionUsecase{rbac: rbac, tx: tx}
}

// Execute 在事务中硬删除权限；重复删除按成功处理
func (uc *DeletePermissionUsecase) Execute(ctx context.Context, in *DeletePermissionInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	err := uc.tx.Run(ctx, func(txCtx context.Context) error {
		return uc.rbac.DeletePermission(txCtx, &rbacinput.DeletePermissionInput{
			ID:         id,
			OperatorID: in.OperatorID,
		})
	})
	// 首次删除已提交后重试（超时重放、双击）不应再报 NOT_FOUND
	if errors.Is(err, rbaccontract.ErrPermissionNotFound) {
		return nil
	}
	return err
}
