// Package auth 认证相关编排
package auth

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	sharederr "gitee.com/hq_project/yomi-boot/component/core/shared/error"
)

// OperationChangePassword 是改密的幂等操作类型；幂等边界在 Handler 层（需同时包住审计写入）
const OperationChangePassword = "CHANGE_PASSWORD"

// ChangePasswordInput HTTP 修改自身密码入参
type ChangePasswordInput struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// ChangePasswordUsecase 修改当前登录用户密码：校验原密码 + 更新密码 + 撤销该用户全部令牌
type ChangePasswordUsecase struct {
	users service_contract.UserService
	token token.TokenService
}

// NewChangePasswordUsecase 创建修改密码用例
func NewChangePasswordUsecase(users service_contract.UserService, tokenService token.TokenService) *ChangePasswordUsecase {
	return &ChangePasswordUsecase{users: users, token: tokenService}
}

// Execute 修改当前登录用户的密码；成功后该账号所有会话失效，需要用新密码重新登录
func (uc *ChangePasswordUsecase) Execute(ctx context.Context, in *ChangePasswordInput, operatorID int64) error {
	if operatorID <= 0 {
		return errors.Sentinel(sharederr.ErrInvalidInput)
	}

	if err := uc.users.ChangePassword(ctx, &userinput.ChangePasswordInput{
		ID:          operatorID,
		OldPassword: in.OldPassword,
		NewPassword: in.NewPassword,
	}); err != nil {
		return err
	}

	if err := support.RevokeUserTokens(ctx, uc.token, operatorID); err != nil {
		// 调用方在幂等事务内执行本用例：撤销失败会让密码更新一并回滚，提示按"改密未完成"处理
		return errors.WrapBiz(err, support.ErrPasswordChangeTokenRevokeFailed.Code,
			support.ErrPasswordChangeTokenRevokeFailed.Message)
	}
	return nil
}
