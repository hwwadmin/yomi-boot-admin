package rbac

import (
	"context"

	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// SetPermissionStatusInput HTTP 权限启停入参
type SetPermissionStatusInput struct {
	Status     rbacenums.Status `json:"status"`
	OperatorID int64            `json:"-"`
}

// SetPermissionStatusUsecase 权限启停编排
type SetPermissionStatusUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewSetPermissionStatusUsecase 创建权限启停用例
func NewSetPermissionStatusUsecase(rbac rbaccontract.RbacAdminService) *SetPermissionStatusUsecase {
	return &SetPermissionStatusUsecase{rbac: rbac}
}

// Execute 设置权限启停
func (uc *SetPermissionStatusUsecase) Execute(ctx context.Context, id int64, in *SetPermissionStatusInput) error {
	if id <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return uc.rbac.SetPermissionStatus(ctx, &rbacinput.SetPermissionStatusInput{
		ID:         id,
		Status:     in.Status,
		OperatorID: in.OperatorID,
	})
}
