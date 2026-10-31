package rbac

import (
	"context"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// DeleteRoleInput HTTP 删除角色入参
type DeleteRoleInput struct {
	ID         dto.JSONInt64 `json:"id"`
	OperatorID int64         `json:"-"`
}

// DeleteRoleUsecase 删除角色编排
type DeleteRoleUsecase struct {
	rbac rbaccontract.RbacAdminService
	tx   dbtx.Executor
}

// NewDeleteRoleUsecase 创建删除角色用例
func NewDeleteRoleUsecase(rbac rbaccontract.RbacAdminService, tx dbtx.Executor) *DeleteRoleUsecase {
	return &DeleteRoleUsecase{rbac: rbac, tx: tx}
}

// Execute 在事务中硬删除角色；重复删除按成功处理
func (uc *DeleteRoleUsecase) Execute(ctx context.Context, in *DeleteRoleInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	err := uc.tx.Run(ctx, func(txCtx context.Context) error {
		return uc.rbac.DeleteRole(txCtx, &rbacinput.DeleteRoleInput{
			ID:         id,
			OperatorID: in.OperatorID,
		})
	})
	// 首次删除已提交后重试（超时重放、双击）不应再报 NOT_FOUND
	if errors.Is(err, rbaccontract.ErrRoleNotFound) {
		return nil
	}
	return err
}
