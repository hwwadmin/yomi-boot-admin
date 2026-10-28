// Package user 用户管理编排
package user

import (
	"context"
	userinput "yomi-admin-backend/domains/user/pkg/shared/input"
	"yomi-admin-backend/domains/user/service_contract"

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
	ID        int64  `json:"id,string"`
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
