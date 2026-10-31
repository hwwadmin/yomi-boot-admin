package sysparam

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"

	sysparamenums "yomi-admin-backend/domains/sysparam/pkg/shared/enums"
	sysparaminput "yomi-admin-backend/domains/sysparam/pkg/shared/input"
	sysparamcontract "yomi-admin-backend/domains/sysparam/service_contract"

	dbtx "gitee.com/hq_project/yomi-boot/component/common/db/tx"
	frameworkidem "gitee.com/hq_project/yomi-boot/component/common/idempotency"
)

// operationCreateParam 是创建系统参数的幂等操作类型
const operationCreateParam = "CREATE_PARAM"

// CreateParamInput HTTP 创建参数入参
type CreateParamInput struct {
	Key        string                       `json:"key"`
	Value      string                       `json:"value"`
	ValueType  sysparamenums.ParamValueType `json:"valueType"`
	Name       string                       `json:"name"`
	Remark     string                       `json:"remark"`
	OperatorID int64                        `json:"-"`

	// IdempotencyKey 来自请求头 X-Idempotency-Key；为空时不做幂等保护
	IdempotencyKey string `json:"-"`
}

// CreateParamUsecase 创建参数编排
type CreateParamUsecase struct {
	sysparam sysparamcontract.SysParamAdminService
	idem     *frameworkidem.Guard
	tx       dbtx.Runner
}

// NewCreateParamUsecase 创建参数用例
func NewCreateParamUsecase(sysparam sysparamcontract.SysParamAdminService, idem *frameworkidem.Guard, tx dbtx.Runner) *CreateParamUsecase {
	return &CreateParamUsecase{sysparam: sysparam, idem: idem, tx: tx}
}

// Execute 在幂等保护与事务中创建参数
func (uc *CreateParamUsecase) Execute(ctx context.Context, in *CreateParamInput) (*ParamItem, error) {
	key, err := support.IdempotencyKey(in.IdempotencyKey, operationCreateParam)
	if err != nil {
		return nil, err
	}

	return support.RunIdempotent(uc.idem, uc.tx, ctx, key, func(txCtx context.Context) (*ParamItem, error) {
		param, createErr := uc.sysparam.CreateParam(txCtx, &sysparaminput.CreateParamInput{
			Key:        in.Key,
			Value:      in.Value,
			ValueType:  in.ValueType,
			Name:       in.Name,
			Remark:     in.Remark,
			OperatorID: in.OperatorID,
		})
		if createErr != nil {
			return nil, createErr
		}
		item := toParamItem(param)
		return &item, nil
	})
}
