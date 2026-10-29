package rbac

import (
	"context"

	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// UpdatePermissionGroupInput HTTP 更新权限组入参；编码不可变
type UpdatePermissionGroupInput struct {
	ID         dto.JSONInt64    `json:"id"`
	Name       string           `json:"name"`
	Remark     string           `json:"remark"`
	Status     rbacenums.Status `json:"status"`
	OperatorID int64            `json:"-"`
}

// UpdatePermissionGroupUsecase 更新权限组编排
type UpdatePermissionGroupUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewUpdatePermissionGroupUsecase 创建更新权限组用例
func NewUpdatePermissionGroupUsecase(rbac rbaccontract.RbacAdminService) *UpdatePermissionGroupUsecase {
	return &UpdatePermissionGroupUsecase{rbac: rbac}
}

// Execute 更新权限组
func (uc *UpdatePermissionGroupUsecase) Execute(ctx context.Context, in *UpdatePermissionGroupInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return uc.rbac.UpdatePermissionGroup(ctx, &rbacinput.UpdatePermissionGroupInput{
		ID:         id,
		Name:       in.Name,
		Remark:     in.Remark,
		Status:     in.Status,
		OperatorID: in.OperatorID,
	})
}
