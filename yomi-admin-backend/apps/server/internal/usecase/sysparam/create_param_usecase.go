package sysparam

import (
	"context"

	sysparamenums "yomi-admin-backend/domains/sysparam/pkg/shared/enums"
	sysparaminput "yomi-admin-backend/domains/sysparam/pkg/shared/input"
	sysparamcontract "yomi-admin-backend/domains/sysparam/service_contract"
)

// CreateParamInput HTTP 创建参数入参
type CreateParamInput struct {
	Key        string                       `json:"key"`
	Value      string                       `json:"value"`
	ValueType  sysparamenums.ParamValueType `json:"valueType"`
	Name       string                       `json:"name"`
	Remark     string                       `json:"remark"`
	OperatorID int64                        `json:"-"`
}

// CreateParamUsecase 创建参数编排
type CreateParamUsecase struct {
	sysparam sysparamcontract.SysParamAdminService
}

// NewCreateParamUsecase 创建参数用例
func NewCreateParamUsecase(sysparam sysparamcontract.SysParamAdminService) *CreateParamUsecase {
	return &CreateParamUsecase{sysparam: sysparam}
}

// Execute 创建参数
func (uc *CreateParamUsecase) Execute(ctx context.Context, in *CreateParamInput) (*ParamItem, error) {
	param, err := uc.sysparam.CreateParam(ctx, &sysparaminput.CreateParamInput{
		Key:        in.Key,
		Value:      in.Value,
		ValueType:  in.ValueType,
		Name:       in.Name,
		Remark:     in.Remark,
		OperatorID: in.OperatorID,
	})
	if err != nil {
		return nil, err
	}
	item := toParamItem(param)
	return &item, nil
}
