// Package usecase 登出编排用例
package usecase

import (
	"context"

	"gitee.com/hq_project/yomi-boot/component/common/token"
)

// LogoutUsecase 登出编排：撤销当前 Token
type LogoutUsecase struct {
	token token.TokenService
}

// NewLogoutUsecase 创建登出用例
func NewLogoutUsecase(tokenService token.TokenService) *LogoutUsecase {
	return &LogoutUsecase{token: tokenService}
}

// Execute 撤销当前请求携带的 Token
func (uc *LogoutUsecase) Execute(ctx context.Context) error {
	tokenValue, ok := token.TokenValueFromContext(ctx)
	if !ok {
		return nil
	}
	return uc.token.Revoke(ctx, tokenValue)
}
