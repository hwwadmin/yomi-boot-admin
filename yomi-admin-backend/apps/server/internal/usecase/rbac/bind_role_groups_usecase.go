package rbac

import (
	"context"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	dbtx "github.com/hwwadmin/yomi-boot/component/common/db/tx"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// BindRoleGroupsInput HTTP 角色权限组绑定入参
type BindRoleGroupsInput struct {
	RoleID     dto.JSONInt64      `json:"roleId"`
	GroupIDs   dto.JSONInt64Slice `json:"groupIds"`
	OperatorID int64              `json:"-"`
}

// BindRoleGroupsUsecase 角色权限组绑定编排
type BindRoleGroupsUsecase struct {
	rbac rbaccontract.RbacAdminService
	tx   dbtx.Executor
}

// NewBindRoleGroupsUsecase 创建角色权限组绑定用例
func NewBindRoleGroupsUsecase(rbac rbaccontract.RbacAdminService, tx dbtx.Executor) *BindRoleGroupsUsecase {
	return &BindRoleGroupsUsecase{rbac: rbac, tx: tx}
}

// Execute 在事务中覆盖角色权限组
func (uc *BindRoleGroupsUsecase) Execute(ctx context.Context, in *BindRoleGroupsInput) error {
	roleID := in.RoleID.Int64()
	if roleID <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return uc.tx.Run(ctx, func(txCtx context.Context) error {
		return uc.rbac.BindRoleGroups(txCtx, &rbacinput.BindRoleGroupsInput{
			RoleID:     roleID,
			GroupIDs:   in.GroupIDs.Int64Slice(),
			OperatorID: in.OperatorID,
		})
	})
}
