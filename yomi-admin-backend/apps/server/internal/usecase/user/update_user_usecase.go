package user

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"
	"yomi-admin-backend/domains/user/pkg/shared/enums"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// UpdateUserInput HTTP 更新用户入参；用户名不可变
type UpdateUserInput struct {
	ID         dto.JSONInt64    `json:"id"`
	Nickname   string           `json:"nickname"`
	Email      string           `json:"email"`
	Phone      string           `json:"phone"`
	Status     enums.UserStatus `json:"status"`
	OperatorID int64            `json:"-"`
}

// UpdateUserUsecase 更新用户编排
type UpdateUserUsecase struct {
	users service_contract.UserService
	guard superAdminGuard
	tx    dbtx.Executor
	token token.TokenService
}

// NewUpdateUserUsecase 创建更新用户用例
func NewUpdateUserUsecase(users service_contract.UserService, guard superAdminGuard, tx dbtx.Executor, tokenService token.TokenService) *UpdateUserUsecase {
	return &UpdateUserUsecase{users: users, guard: guard, tx: tx, token: tokenService}
}

// Execute 更新用户；状态改为非启用后在超管锁内校验并撤销该用户全部 Token
func (uc *UpdateUserUsecase) Execute(ctx context.Context, in *UpdateUserInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	// 状态合法性由领域校验；这里只判断本次是否移除启用态，决定要不要进超管锁。
	// 非法状态会走不加锁路径，最终由领域以 ErrInvalidInput 拒绝。
	disabling := in.Status == enums.UserStatusDisabled
	if id == in.OperatorID && disabling {
		return support.ErrCannotOperateSelf
	}

	update := func(txCtx context.Context) error {
		if disabling {
			if err := uc.guard.EnsureCanManageUser(txCtx, in.OperatorID, id); err != nil {
				return err
			}
		}
		return uc.users.Update(txCtx, &userinput.UpdateUserInput{
			ID:         id,
			Nickname:   in.Nickname,
			Email:      in.Email,
			Phone:      in.Phone,
			Status:     in.Status,
			OperatorID: in.OperatorID,
		})
	}

	var err error
	if disabling {
		// 停用可能移除超管账号或减少启用态超管集合，必须串行化。
		err = uc.guard.WithLock(ctx, func(lockCtx context.Context) error {
			return uc.tx.Run(lockCtx, update)
		})
	} else {
		// 启用不会减少集合，无需加锁；并发禁用可能因此保守拒绝。
		err = uc.tx.Run(ctx, update)
	}
	if err != nil {
		return err
	}

	if disabling {
		return support.RevokeUserTokens(ctx, uc.token, id)
	}
	return nil
}
