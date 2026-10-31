package rbac

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"

	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	frameworkidem "gitee.com/hq_project/yomi-boot/component/common/idempotency"
)

// operationCreatePermission 是创建权限的幂等操作类型
const operationCreatePermission = "CREATE_PERMISSION"

// CreatePermissionInput HTTP 创建权限入参
type CreatePermissionInput struct {
	Code       string                   `json:"code"`
	Name       string                   `json:"name"`
	Type       rbacenums.PermissionType `json:"type"`
	UIPath     string                   `json:"uiPath"`
	HTTPMethod string                   `json:"httpMethod"`
	HTTPPath   string                   `json:"httpPath"`
	Remark     string                   `json:"remark"`
	OperatorID int64                    `json:"-"`

	// IdempotencyKey 来自请求头 X-Idempotency-Key；为空时不做幂等保护
	IdempotencyKey string `json:"-"`
}

// CreatePermissionUsecase 创建权限编排
type CreatePermissionUsecase struct {
	rbac rbaccontract.RbacAdminService
	idem *frameworkidem.Guard
	tx   dbtx.Runner
}

// NewCreatePermissionUsecase 创建权限用例
func NewCreatePermissionUsecase(rbac rbaccontract.RbacAdminService, idem *frameworkidem.Guard, tx dbtx.Runner) *CreatePermissionUsecase {
	return &CreatePermissionUsecase{rbac: rbac, idem: idem, tx: tx}
}

// Execute 在幂等保护与事务中创建权限
func (uc *CreatePermissionUsecase) Execute(ctx context.Context, in *CreatePermissionInput) (*PermissionItem, error) {
	key, err := support.IdempotencyKey(in.IdempotencyKey, operationCreatePermission)
	if err != nil {
		return nil, err
	}

	return support.RunIdempotent(uc.idem, uc.tx, ctx, key, func(txCtx context.Context) (*PermissionItem, error) {
		permission, createErr := uc.rbac.CreatePermission(txCtx, &rbacinput.CreatePermissionInput{
			Code:       in.Code,
			Name:       in.Name,
			Type:       in.Type,
			UIPath:     in.UIPath,
			HTTPMethod: in.HTTPMethod,
			HTTPPath:   in.HTTPPath,
			Remark:     in.Remark,
			OperatorID: in.OperatorID,
		})
		if createErr != nil {
			return nil, createErr
		}
		item := toPermissionItem(permission)
		return &item, nil
	})
}
