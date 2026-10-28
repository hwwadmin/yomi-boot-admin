// Package auth 认证相关编排
package auth

import (
	"context"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
	"yomi-admin-backend/domains/user/service_contract"
)

// ProfileUsecase 当前登录用户信息编排
type ProfileUsecase struct {
	users service_contract.UserService
	auth  rbaccontract.RbacAuthService
}

// NewProfileUsecase 创建当前用户用例
func NewProfileUsecase(users service_contract.UserService, auth rbaccontract.RbacAuthService) *ProfileUsecase {
	return &ProfileUsecase{users: users, auth: auth}
}

// Execute 查询当前登录用户
func (uc *ProfileUsecase) Execute(ctx context.Context, userID int64) (*IdentityView, error) {
	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return buildIdentityView(ctx, user, uc.auth)
}
