package rbac

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	frameworkidem "gitee.com/hq_project/yomi-boot/component/common/idempotency"
)

// operationCreatePermissionGroup 是创建权限组的幂等操作类型
const operationCreatePermissionGroup = "CREATE_PERMISSION_GROUP"

// CreatePermissionGroupInput HTTP 创建权限组入参
type CreatePermissionGroupInput struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Remark     string `json:"remark"`
	OperatorID int64  `json:"-"`

	// IdempotencyKey 来自请求头 X-Idempotency-Key；为空时不做幂等保护
	IdempotencyKey string `json:"-"`
}

// CreatePermissionGroupUsecase 创建权限组编排
type CreatePermissionGroupUsecase struct {
	rbac rbaccontract.RbacAdminService
	idem *frameworkidem.Guard
	tx   dbtx.Runner
}

// NewCreatePermissionGroupUsecase 创建权限组用例
func NewCreatePermissionGroupUsecase(rbac rbaccontract.RbacAdminService, idem *frameworkidem.Guard, tx dbtx.Runner) *CreatePermissionGroupUsecase {
	return &CreatePermissionGroupUsecase{rbac: rbac, idem: idem, tx: tx}
}

// Execute 在幂等保护与事务中创建权限组
func (uc *CreatePermissionGroupUsecase) Execute(ctx context.Context, in *CreatePermissionGroupInput) (*PermissionGroupItem, error) {
	key, err := support.IdempotencyKey(in.IdempotencyKey, operationCreatePermissionGroup)
	if err != nil {
		return nil, err
	}

	return support.RunIdempotent(uc.idem, uc.tx, ctx, key, func(txCtx context.Context) (*PermissionGroupItem, error) {
		group, createErr := uc.rbac.CreatePermissionGroup(txCtx, &rbacinput.CreatePermissionGroupInput{
			Code:       in.Code,
			Name:       in.Name,
			Remark:     in.Remark,
			OperatorID: in.OperatorID,
		})
		if createErr != nil {
			return nil, createErr
		}
		item := toPermissionGroupItem(group)
		return &item, nil
	})
}
