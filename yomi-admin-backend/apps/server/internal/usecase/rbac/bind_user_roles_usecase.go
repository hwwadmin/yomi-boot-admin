package rbac

import (
	"context"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// BindUserRolesInput HTTP 用户角色绑定入参
type BindUserRolesInput struct {
	RoleIDs    dto.JSONInt64Slice `json:"roleIds"`
	OperatorID int64              `json:"-"`
}

// superAdminPolicy 超管相关编排策略
type superAdminPolicy interface {
	EnsureCanBindRoles(ctx context.Context, operatorID, targetUserID int64, roleIDs []int64) error
}

// BindUserRolesUsecase 用户角色绑定编排
type BindUserRolesUsecase struct {
	rbac  rbaccontract.RbacAdminService
	guard superAdminPolicy
	tx    dbtx.Executor
}

// NewBindUserRolesUsecase 创建用户角色绑定用例
func NewBindUserRolesUsecase(rbac rbaccontract.RbacAdminService, guard superAdminPolicy, tx dbtx.Executor) *BindUserRolesUsecase {
	return &BindUserRolesUsecase{rbac: rbac, guard: guard, tx: tx}
}

// Execute 覆盖用户角色
func (uc *BindUserRolesUsecase) Execute(ctx context.Context, userID int64, in *BindUserRolesInput) error {
	if userID <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}

	roleIDs := in.RoleIDs.Int64Slice()
	return uc.tx.Run(ctx, func(txCtx context.Context) error {
		if err := uc.guard.EnsureCanBindRoles(txCtx, in.OperatorID, userID, roleIDs); err != nil {
			return err
		}
		return uc.rbac.BindUserRoles(txCtx, &rbacinput.BindUserRolesInput{
			UserID:     userID,
			RoleIDs:    roleIDs,
			OperatorID: in.OperatorID,
		})
	})
}
