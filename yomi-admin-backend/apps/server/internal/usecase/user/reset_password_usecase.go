package user

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	"github.com/hwwadmin/yomi-boot/component/common/token"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// ResetUserPasswordInput HTTP 重置他人密码入参
type ResetUserPasswordInput struct {
	ID          dto.JSONInt64 `json:"id"`
	NewPassword string        `json:"newPassword"`
	OperatorID  int64         `json:"-"`
}

// resetPasswordGuard 判断操作者能否重置他人密码
type resetPasswordGuard interface {
	// EnsureCanResetPassword 只有启用态超管可以重置他人密码
	EnsureCanResetPassword(ctx context.Context, operatorID int64) error
}

// ResetUserPasswordUsecase 超级管理员重置他人密码编排
type ResetUserPasswordUsecase struct {
	users service_contract.UserService
	guard resetPasswordGuard
	token token.TokenService
}

// NewResetUserPasswordUsecase 创建重置密码用例
func NewResetUserPasswordUsecase(users service_contract.UserService, guard resetPasswordGuard, tokenService token.TokenService) *ResetUserPasswordUsecase {
	return &ResetUserPasswordUsecase{users: users, guard: guard, token: tokenService}
}

// Execute 重置目标用户密码并撤销其全部令牌；重置自己必须走"修改密码"校验原密码
func (uc *ResetUserPasswordUsecase) Execute(ctx context.Context, in *ResetUserPasswordInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if id == in.OperatorID {
		return support.ErrCannotOperateSelf
	}
	if err := uc.guard.EnsureCanResetPassword(ctx, in.OperatorID); err != nil {
		return err
	}

	if err := uc.users.ResetPassword(ctx, &userinput.ResetPasswordInput{
		ID:          id,
		NewPassword: in.NewPassword,
		OperatorID:  in.OperatorID,
	}); err != nil {
		return err
	}
	return support.RevokeUserTokens(ctx, uc.token, id)
}
