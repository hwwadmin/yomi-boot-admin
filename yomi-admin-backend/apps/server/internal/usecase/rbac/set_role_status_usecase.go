package rbac

import (
	"context"

	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// SetRoleStatusInput HTTP 角色启停入参
type SetRoleStatusInput struct {
	ID         dto.JSONInt64    `json:"id"`
	Status     rbacenums.Status `json:"status"`
	OperatorID int64            `json:"-"`
}

// SetRoleStatusUsecase 角色启停编排
type SetRoleStatusUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewSetRoleStatusUsecase 创建角色启停用例
func NewSetRoleStatusUsecase(rbac rbaccontract.RbacAdminService) *SetRoleStatusUsecase {
	return &SetRoleStatusUsecase{rbac: rbac}
}

// Execute 设置角色启停
func (uc *SetRoleStatusUsecase) Execute(ctx context.Context, in *SetRoleStatusInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return uc.rbac.SetRoleStatus(ctx, &rbacinput.SetRoleStatusInput{
		ID:         id,
		Status:     in.Status,
		OperatorID: in.OperatorID,
	})
}
