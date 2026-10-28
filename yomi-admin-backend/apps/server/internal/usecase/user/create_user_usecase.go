package user

import (
	"context"
	"yomi-admin-backend/domains/user/pkg/shared/enums"
	userinput "yomi-admin-backend/domains/user/pkg/shared/input"
	"yomi-admin-backend/domains/user/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
)

// CreateUserInput HTTP 创建用户入参
type CreateUserInput struct {
	Username   string           `json:"username"`
	Password   string           `json:"password"`
	Nickname   string           `json:"nickname"`
	Email      string           `json:"email"`
	Phone      string           `json:"phone"`
	Status     enums.UserStatus `json:"status"`
	OperatorID int64            `json:"-"`
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
	return uc.tx.Run(ctx, func(txCtx context.Context) error {
		_, err := uc.users.Create(txCtx, &userinput.CreateUserInput{
			Username:   in.Username,
			Password:   in.Password,
			Nickname:   in.Nickname,
			Email:      in.Email,
			Phone:      in.Phone,
			Status:     in.Status,
			OperatorID: in.OperatorID,
		})
		return err
	})
}
