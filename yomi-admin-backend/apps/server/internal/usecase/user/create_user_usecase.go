package user

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"
	"yomi-admin-backend/domains/user/pkg/shared/enums"
	userinput "yomi-admin-backend/domains/user/pkg/shared/input"
	"yomi-admin-backend/domains/user/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	frameworkidem "gitee.com/hq_project/yomi-boot/component/common/idempotency"
)

// operationCreateUser 是创建用户的幂等操作类型
const operationCreateUser = "CREATE_USER"

// CreateUserInput HTTP 创建用户入参
type CreateUserInput struct {
	Username   string           `json:"username"`
	Password   string           `json:"password"`
	Nickname   string           `json:"nickname"`
	Email      string           `json:"email"`
	Phone      string           `json:"phone"`
	Status     enums.UserStatus `json:"status"`
	OperatorID int64            `json:"-"`

	// IdempotencyKey 来自请求头 X-Idempotency-Key；为空时不做幂等保护
	IdempotencyKey string `json:"-"`
}

// CreateUserUsecase 创建用户编排
type CreateUserUsecase struct {
	users service_contract.UserService
	idem  *frameworkidem.Guard
	tx    dbtx.Runner
}

// NewCreateUserUsecase 创建用户用例
func NewCreateUserUsecase(users service_contract.UserService, idem *frameworkidem.Guard, tx dbtx.Runner) *CreateUserUsecase {
	return &CreateUserUsecase{users: users, idem: idem, tx: tx}
}

// Execute 在幂等保护与事务中创建用户
func (uc *CreateUserUsecase) Execute(ctx context.Context, in *CreateUserInput) error {
	key, err := support.IdempotencyKey(in.IdempotencyKey, operationCreateUser)
	if err != nil {
		return err
	}

	_, err = support.RunIdempotent(uc.idem, uc.tx, ctx, key, func(txCtx context.Context) (struct{}, error) {
		_, createErr := uc.users.Create(txCtx, &userinput.CreateUserInput{
			Username:   in.Username,
			Password:   in.Password,
			Nickname:   in.Nickname,
			Email:      in.Email,
			Phone:      in.Phone,
			Status:     in.Status,
			OperatorID: in.OperatorID,
		})
		return struct{}{}, createErr
	})
	return err
}
