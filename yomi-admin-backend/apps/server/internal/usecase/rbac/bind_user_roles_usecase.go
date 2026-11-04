package rbac

import (
	"context"

	rbacentity "yomi-admin-backend/domains/rbac/pkg/shared/entity"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	dbtx "github.com/hwwadmin/yomi-boot/component/common/db/tx"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// BindUserRolesInput HTTP 用户角色绑定入参
type BindUserRolesInput struct {
	UserID     dto.JSONInt64      `json:"userId"`
	RoleIDs    dto.JSONInt64Slice `json:"roleIds"`
	OperatorID int64              `json:"-"`
}

// superAdminPolicy 超管相关编排策略
type superAdminPolicy interface {
	WithLock(ctx context.Context, fn func(context.Context) error) error
	EnsureCanBindRoles(ctx context.Context, operatorID, targetUserID int64, roles []*rbacentity.Role) error
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

// Execute 在超管锁内覆盖用户角色
func (uc *BindUserRolesUsecase) Execute(ctx context.Context, in *BindUserRolesInput) error {
	userID := in.UserID.Int64()
	if userID <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}

	roleIDs := in.RoleIDs.Int64Slice()
	return uc.guard.WithLock(ctx, func(lockCtx context.Context) error {
		return uc.tx.Run(lockCtx, func(txCtx context.Context) error {
			// 角色校验只在这里做一次：结果同时喂给超管策略与实际绑定，策略层不再重复实现规则
			roles, err := uc.rbac.ResolveBindingRoles(txCtx, roleIDs)
			if err != nil {
				return err
			}
			if err := uc.guard.EnsureCanBindRoles(txCtx, in.OperatorID, userID, roles); err != nil {
				return err
			}
			return uc.rbac.BindUserRoles(txCtx, &rbacinput.BindUserRolesInput{
				UserID:     userID,
				RoleIDs:    roleIDs,
				OperatorID: in.OperatorID,
			})
		})
	})
}
