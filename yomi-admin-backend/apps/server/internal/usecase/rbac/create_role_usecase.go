package rbac

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	frameworkidem "gitee.com/hq_project/yomi-boot/component/common/idempotency"
)

// operationCreateRole 是创建角色的幂等操作类型
const operationCreateRole = "CREATE_ROLE"

// CreateRoleInput HTTP 创建角色入参
type CreateRoleInput struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Remark     string `json:"remark"`
	OperatorID int64  `json:"-"`

	// IdempotencyKey 来自请求头 X-Idempotency-Key；为空时不做幂等保护
	IdempotencyKey string `json:"-"`
}

// CreateRoleUsecase 创建角色编排
type CreateRoleUsecase struct {
	rbac rbaccontract.RbacAdminService
	idem *frameworkidem.Guard
	tx   dbtx.Runner
}

// NewCreateRoleUsecase 创建角色用例
func NewCreateRoleUsecase(rbac rbaccontract.RbacAdminService, idem *frameworkidem.Guard, tx dbtx.Runner) *CreateRoleUsecase {
	return &CreateRoleUsecase{rbac: rbac, idem: idem, tx: tx}
}

// Execute 在幂等保护与事务中创建角色
func (uc *CreateRoleUsecase) Execute(ctx context.Context, in *CreateRoleInput) (*RoleItem, error) {
	key, err := support.IdempotencyKey(in.IdempotencyKey, operationCreateRole)
	if err != nil {
		return nil, err
	}

	return support.RunIdempotent(uc.idem, uc.tx, ctx, key, func(txCtx context.Context) (*RoleItem, error) {
		role, createErr := uc.rbac.CreateRole(txCtx, &rbacinput.CreateRoleInput{
			Code:       in.Code,
			Name:       in.Name,
			Remark:     in.Remark,
			OperatorID: in.OperatorID,
		})
		if createErr != nil {
			return nil, createErr
		}
		item := toRoleItem(role)
		return &item, nil
	})
}
