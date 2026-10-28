// Package usecase 当前用户编排用例
package usecase

import (
	"context"
	"yomi-admin-backend/domains/user/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/token"
)

// ProfileUsecase 当前登录用户信息编排
type ProfileUsecase struct {
	users service_contract.UserService
}

// NewProfileUsecase 创建当前用户用例
func NewProfileUsecase(users service_contract.UserService) *ProfileUsecase {
	return &ProfileUsecase{users: users}
}

// ProfileOutput HTTP 出参
type ProfileOutput struct {
	Username    string   `json:"username"`
	Nickname    string   `json:"nickname"`
	Permissions []string `json:"permissions"`
}

// Execute 查询当前登录用户
func (uc *ProfileUsecase) Execute(ctx context.Context) (*ProfileOutput, error) {
	userID, err := token.OperatorID(ctx)
	if err != nil {
		return nil, err
	}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &ProfileOutput{
		Username:    user.Username,
		Nickname:    user.Nickname,
		Permissions: permissionsOf(user.Username),
	}, nil
}
