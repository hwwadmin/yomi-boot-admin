package user

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support/superadmin"
	"yomi-admin-backend/domains/user/pkg/shared/enums"
	userinput "yomi-admin-backend/domains/user/pkg/shared/input"
	"yomi-admin-backend/domains/user/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

// UpdateUserInput HTTP 更新用户入参；用户名不可变
type UpdateUserInput struct {
	Nickname   string           `json:"nickname"`
	Email      string           `json:"email"`
	Phone      string           `json:"phone"`
	Status     enums.UserStatus `json:"status"`
	OperatorID int64            `json:"-"`
}

// UpdateUserUsecase 更新用户编排
type UpdateUserUsecase struct {
	users service_contract.UserService
	guard *superadmin.Guard
	tx    dbtx.Executor
	token token.TokenService
}

// NewUpdateUserUsecase 创建更新用户用例
func NewUpdateUserUsecase(users service_contract.UserService, guard *superadmin.Guard, tx dbtx.Executor, tokenService token.TokenService) *UpdateUserUsecase {
	return &UpdateUserUsecase{users: users, guard: guard, tx: tx, token: tokenService}
}

// Execute 更新用户；状态改为禁用后撤销该用户全部 Token
func (uc *UpdateUserUsecase) Execute(ctx context.Context, id int64, in *UpdateUserInput) error {
	if err := uc.tx.Run(ctx, func(txCtx context.Context) error {
		if in.Status == enums.UserStatusDisabled {
			if err := uc.guard.EnsureNotLastSuper(txCtx, id); err != nil {
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
	}); err != nil {
		return err
	}

	if in.Status == enums.UserStatusDisabled {
		if err := uc.token.RevokeByUserID(ctx, id); err != nil {
			logger.Error(ctx, "撤销用户令牌失败", "userID", id, "error", err)
		}
	}
	return nil
}
