package sysparam

import (
	"context"

	sysparamentity "yomi-admin-backend/domains/sysparam/pkg/shared/entity"
	sysparamenums "yomi-admin-backend/domains/sysparam/pkg/shared/enums"
	sysparaminput "yomi-admin-backend/domains/sysparam/pkg/shared/input"
	sysparamcontract "yomi-admin-backend/domains/sysparam/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// PageParamsInput HTTP 参数分页入参
type PageParamsInput struct {
	dto.PageInput
	Keyword   string                       `json:"keyword"`
	Status    sysparamenums.Status         `json:"status"`
	ValueType sysparamenums.ParamValueType `json:"valueType"`
}

// ParamItem 参数列表项
type ParamItem struct {
	ID         dto.JSONInt64                `json:"id"`
	Key        string                       `json:"key"`
	Value      string                       `json:"value"`
	ValueType  sysparamenums.ParamValueType `json:"valueType"`
	Name       string                       `json:"name"`
	Remark     string                       `json:"remark"`
	Status     sysparamenums.Status         `json:"status"`
	CreateTime int64                        `json:"createTime"`
	UpdateTime int64                        `json:"updateTime"`
}

// PageParamsUsecase 参数分页编排
type PageParamsUsecase struct {
	sysparam sysparamcontract.SysParamAdminService
}

// NewPageParamsUsecase 创建参数分页用例
func NewPageParamsUsecase(sysparam sysparamcontract.SysParamAdminService) *PageParamsUsecase {
	return &PageParamsUsecase{sysparam: sysparam}
}

// Execute 分页查询参数
func (uc *PageParamsUsecase) Execute(ctx context.Context, in *PageParamsInput) (dto.PageOutput[ParamItem], error) {
	in.ValiPage()

	out, err := uc.sysparam.PageParams(ctx, &sysparaminput.PageParamsInput{
		Keyword:   in.Keyword,
		Status:    in.Status,
		ValueType: in.ValueType,
		Page:      in.Page,
		PageSize:  in.PageSize,
	})
	if err != nil {
		return dto.PageOutput[ParamItem]{}, err
	}

	items := make([]ParamItem, 0, len(out.Items))
	for _, param := range out.Items {
		items = append(items, toParamItem(param))
	}
	return dto.NewPageOutput(items, out.Total), nil
}

// toParamItem 将参数实体转换为 HTTP 视图
func toParamItem(param *sysparamentity.Param) ParamItem {
	return ParamItem{
		ID:         dto.JSONInt64(param.ID),
		Key:        param.Key,
		Value:      param.Value,
		ValueType:  param.ValueType,
		Name:       param.Name,
		Remark:     param.Remark,
		Status:     param.Status,
		CreateTime: param.CreateTime,
		UpdateTime: param.UpdateTime,
	}
}
