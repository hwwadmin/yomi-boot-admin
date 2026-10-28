package user

import (
	"context"
	userinput "yomi-admin-backend/domains/user/pkg/shared/input"
	"yomi-admin-backend/domains/user/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

// DeleteUserUsecase 删除用户编排
type DeleteUserUsecase struct {
	users service_contract.UserService
	tx    dbtx.Executor
	token token.TokenService
}

// NewDeleteUserUsecase 创建删除用户用例
func NewDeleteUserUsecase(users service_contract.UserService, tx dbtx.Executor, tokenService token.TokenService) *DeleteUserUsecase {
	return &DeleteUserUsecase{users: users, tx: tx, token: tokenService}
}

// Execute 软删除用户并撤销其全部 Token
func (uc *DeleteUserUsecase) Execute(ctx context.Context, id int64) error {
	operatorID, err := token.OperatorID(ctx)
	if err != nil {
		return err
	}

	if err := uc.tx.Run(ctx, func(txCtx context.Context) error {
		return uc.users.Delete(txCtx, &userinput.DeleteUserInput{
			ID:         id,
			OperatorID: operatorID,
		})
	}); err != nil {
		return err
	}

	if err := uc.token.RevokeByUserID(ctx, id); err != nil {
		logger.Error(ctx, "撤销用户令牌失败", "userID", id, "error", err)
	}
	return nil
}
