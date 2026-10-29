package user

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// DeleteUserInput HTTP 删除用户入参
type DeleteUserInput struct {
	ID         dto.JSONInt64 `json:"id"`
	OperatorID int64         `json:"-"`
}

// DeleteUserUsecase 删除用户编排
type DeleteUserUsecase struct {
	users service_contract.UserService
	rbac  rbaccontract.RbacAdminService
	guard superAdminGuard
	tx    dbtx.Executor
	token token.TokenService
}

// NewDeleteUserUsecase 创建删除用户用例
func NewDeleteUserUsecase(users service_contract.UserService, rbac rbaccontract.RbacAdminService, guard superAdminGuard, tx dbtx.Executor, tokenService token.TokenService) *DeleteUserUsecase {
	return &DeleteUserUsecase{users: users, rbac: rbac, guard: guard, tx: tx, token: tokenService}
}

// Execute 在超管锁内软删除用户、清理角色绑定，并撤销其全部 Token
func (uc *DeleteUserUsecase) Execute(ctx context.Context, in *DeleteUserInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if id == in.OperatorID {
		return support.ErrCannotOperateSelf
	}

	err := uc.guard.WithLock(ctx, func(lockCtx context.Context) error {
		return uc.tx.Run(lockCtx, func(txCtx context.Context) error {
			if err := uc.guard.EnsureCanManageUser(txCtx, in.OperatorID, id); err != nil {
				return err
			}
			if err := uc.users.Delete(txCtx, &userinput.DeleteUserInput{
				ID:         id,
				OperatorID: in.OperatorID,
			}); err != nil {
				return err
			}
			return uc.rbac.DeleteUserRoles(txCtx, id)
		})
	})
	if err != nil && !errors.Is(err, service_contract.ErrUserNotFound) {
		return err
	}

	// 用户可能已被软删除；仍然补撤 Token，保证失败重试能生效。
	if revokeErr := revokeUserTokens(ctx, uc.token, id); revokeErr != nil {
		return revokeErr
	}
	return err
}
