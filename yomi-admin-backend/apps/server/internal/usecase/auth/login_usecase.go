// Package auth 认证相关编排
package auth

import (
	"context"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// LoginUsecase 登录编排：领域鉴权 + 签发 Token
type LoginUsecase struct {
	users service_contract.UserService
	token token.TokenService
}

// NewLoginUsecase 创建登录用例
func NewLoginUsecase(users service_contract.UserService, tokenService token.TokenService) *LoginUsecase {
	return &LoginUsecase{users: users, token: tokenService}
}

// LoginInput HTTP 登录入参
type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginUser 登录返回的用户信息
type LoginUser struct {
	Username    string   `json:"username"`
	Nickname    string   `json:"nickname"`
	Permissions []string `json:"permissions"`
}

// LoginOutput HTTP 登录出参
type LoginOutput struct {
	Token string    `json:"token"`
	User  LoginUser `json:"user"`
}

// Execute 编排登录
func (uc *LoginUsecase) Execute(ctx context.Context, in *LoginInput) (*LoginOutput, error) {
	out, err := uc.users.Authenticate(ctx, &userinput.AuthenticateInput{
		Username: in.Username,
		Password: in.Password,
	})
	if err != nil {
		return nil, err
	}

	session := &token.TokenSession{
		UserID:     out.User.ID,
		ClientType: "admin",
	}
	session.SetContextValue("username", out.User.Username)
	session.SetContextValue("nickname", out.User.Nickname)

	tokenValue, err := uc.token.Issue(ctx, session)
	if err != nil {
		return nil, errors.Wrap(err, "签发令牌失败")
	}

	return &LoginOutput{
		Token: tokenValue,
		User: LoginUser{
			Username:    out.User.Username,
			Nickname:    out.User.Nickname,
			Permissions: permissionsOf(out.User.Username),
		},
	}, nil
}
