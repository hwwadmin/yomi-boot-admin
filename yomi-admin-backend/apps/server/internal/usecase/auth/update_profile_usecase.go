// Package auth 认证相关编排
package auth

import (
	"context"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	sharederr "gitee.com/hq_project/yomi-boot/component/core/shared/error"
)

// UpdateProfileInput HTTP 更新自身资料入参
type UpdateProfileInput struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// UpdateProfileUsecase 更新当前登录用户资料编排
type UpdateProfileUsecase struct {
	users service_contract.UserService
	auth  rbaccontract.RbacAuthService
}

// NewUpdateProfileUsecase 创建更新自身资料用例
func NewUpdateProfileUsecase(users service_contract.UserService, auth rbaccontract.RbacAuthService) *UpdateProfileUsecase {
	return &UpdateProfileUsecase{users: users, auth: auth}
}

// Execute 更新当前登录用户的昵称、邮箱与手机号，返回更新后的身份视图
func (uc *UpdateProfileUsecase) Execute(ctx context.Context, in *UpdateProfileInput, operatorID int64) (*IdentityView, error) {
	if operatorID <= 0 {
		return nil, errors.Sentinel(sharederr.ErrInvalidInput)
	}

	user, err := uc.users.UpdateProfile(ctx, &userinput.UpdateProfileInput{
		ID:         operatorID,
		Nickname:   in.Nickname,
		Email:      in.Email,
		Phone:      in.Phone,
		OperatorID: operatorID,
	})
	if err != nil {
		return nil, err
	}
	return buildIdentityView(ctx, user, uc.auth)
}
