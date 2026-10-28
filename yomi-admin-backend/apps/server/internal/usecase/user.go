// Package usecase 用户管理编排用例
package usecase

import (
	"context"
	"yomi-admin-backend/domains/user/pkg/shared/enums"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
	"gitee.com/hq_project/yomi-boot/component/core/utils/timeutil"
)

// PageUsersInput HTTP 分页入参
type PageUsersInput struct {
	dto.PageInput
	Keyword string `json:"keyword"`
}

// UserItem 用户列表项
type UserItem struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    int8   `json:"status"`
	CreatedAt int64  `json:"createdAt"` // Unix 毫秒时间戳，由前端按客户端时区展示
}

// PageUsersUsecase 用户分页编排
type PageUsersUsecase struct {
	users service_contract.UserService
}

// NewPageUsersUsecase 创建用户分页用例
func NewPageUsersUsecase(users service_contract.UserService) *PageUsersUsecase {
	return &PageUsersUsecase{users: users}
}

// Execute 分页查询用户
func (uc *PageUsersUsecase) Execute(ctx context.Context, in *PageUsersInput) (*dto.PageOutput[UserItem], error) {
	in.ValiPage()

	out, err := uc.users.Page(ctx, &userinput.PageUsersInput{
		Keyword:  in.Keyword,
		Page:     in.Page,
		PageSize: in.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]UserItem, 0, len(out.Items))
	for _, user := range out.Items {
		items = append(items, UserItem{
			ID:        user.ID,
			Username:  user.Username,
			Nickname:  user.Nickname,
			Email:     user.Email,
			Phone:     user.Phone,
			Status:    int8(user.Status),
			CreatedAt: timeutil.ToUnix(user.CreatedAt),
		})
	}

	return &dto.PageOutput[UserItem]{Items: items, Total: out.Total}, nil
}

// CreateUserInput HTTP 创建用户入参
type CreateUserInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Status   int8   `json:"status"`
}

// CreateUserUsecase 创建用户编排
type CreateUserUsecase struct {
	users service_contract.UserService
	tx    dbtx.Executor
}

// NewCreateUserUsecase 创建用户用例
func NewCreateUserUsecase(users service_contract.UserService, tx dbtx.Executor) *CreateUserUsecase {
	return &CreateUserUsecase{users: users, tx: tx}
}

// Execute 在事务中创建用户
func (uc *CreateUserUsecase) Execute(ctx context.Context, in *CreateUserInput) error {
	operatorID, err := token.OperatorID(ctx)
	if err != nil {
		return err
	}

	return uc.tx.Run(ctx, func(txCtx context.Context) error {
		_, err := uc.users.Create(txCtx, &userinput.CreateUserInput{
			Username:   in.Username,
			Password:   in.Password,
			Nickname:   in.Nickname,
			Email:      in.Email,
			Phone:      in.Phone,
			Status:     enums.UserStatus(in.Status),
			OperatorID: operatorID,
		})
		return err
	})
}

// UpdateUserInput HTTP 更新用户入参；用户名不可变
type UpdateUserInput struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Status   int8   `json:"status"`
}

// UpdateUserUsecase 更新用户编排
type UpdateUserUsecase struct {
	users service_contract.UserService
	tx    dbtx.Executor
	token token.TokenService
}

// NewUpdateUserUsecase 创建更新用户用例
func NewUpdateUserUsecase(users service_contract.UserService, tx dbtx.Executor, tokenService token.TokenService) *UpdateUserUsecase {
	return &UpdateUserUsecase{users: users, tx: tx, token: tokenService}
}

// Execute 更新用户；状态改为禁用后撤销该用户全部 Token
func (uc *UpdateUserUsecase) Execute(ctx context.Context, id int64, in *UpdateUserInput) error {
	operatorID, err := token.OperatorID(ctx)
	if err != nil {
		return err
	}

	if err := uc.tx.Run(ctx, func(txCtx context.Context) error {
		return uc.users.Update(txCtx, &userinput.UpdateUserInput{
			ID:         id,
			Nickname:   in.Nickname,
			Email:      in.Email,
			Phone:      in.Phone,
			Status:     enums.UserStatus(in.Status),
			OperatorID: operatorID,
		})
	}); err != nil {
		return err
	}

	if in.Status == int8(enums.UserStatusDisabled) {
		if err := uc.token.RevokeByUserID(ctx, id); err != nil {
			logger.Error(ctx, "撤销用户令牌失败", "userID", id, "error", err)
		}
	}
	return nil
}

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
