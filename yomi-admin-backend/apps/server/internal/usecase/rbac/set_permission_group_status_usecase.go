package rbac

import (
	"context"

	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// SetPermissionGroupStatusInput HTTP 权限组启停入参
type SetPermissionGroupStatusInput struct {
	ID         dto.JSONInt64    `json:"id"`
	Status     rbacenums.Status `json:"status"`
	OperatorID int64            `json:"-"`
}

// SetPermissionGroupStatusUsecase 权限组启停编排
type SetPermissionGroupStatusUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewSetPermissionGroupStatusUsecase 创建权限组启停用例
func NewSetPermissionGroupStatusUsecase(rbac rbaccontract.RbacAdminService) *SetPermissionGroupStatusUsecase {
	return &SetPermissionGroupStatusUsecase{rbac: rbac}
}

// Execute 设置权限组启停
func (uc *SetPermissionGroupStatusUsecase) Execute(ctx context.Context, in *SetPermissionGroupStatusInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return uc.rbac.SetPermissionGroupStatus(ctx, &rbacinput.SetPermissionGroupStatusInput{
		ID:         id,
		Status:     in.Status,
		OperatorID: in.OperatorID,
	})
}
