package sysparam

import (
	"context"

	sysparamcontract "yomi-admin-backend/domains/sysparam/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// GetParamUsecase 参数详情编排
type GetParamUsecase struct {
	sysparam sysparamcontract.SysParamAdminService
}

// NewGetParamUsecase 创建参数详情用例
func NewGetParamUsecase(sysparam sysparamcontract.SysParamAdminService) *GetParamUsecase {
	return &GetParamUsecase{sysparam: sysparam}
}

// Execute 查询参数详情
func (uc *GetParamUsecase) Execute(ctx context.Context, id int64) (*ParamItem, error) {
	if id <= 0 {
		return nil, errors.Sentinel(sysparamcontract.ErrInvalidInput)
	}

	param, err := uc.sysparam.GetParam(ctx, id)
	if err != nil {
		return nil, err
	}
	item := toParamItem(param)
	return &item, nil
}
