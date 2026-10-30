package sysparam

import (
	"context"

	sysparaminput "yomi-admin-backend/domains/sysparam/pkg/shared/input"
	sysparamcontract "yomi-admin-backend/domains/sysparam/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// UpdateParamInput HTTP 更新参数入参；键与值类型不可变，启停走独立接口
type UpdateParamInput struct {
	ID         dto.JSONInt64 `json:"id"`
	Value      string        `json:"value"`
	Name       string        `json:"name"`
	Remark     string        `json:"remark"`
	OperatorID int64         `json:"-"`
}

// UpdateParamUsecase 更新参数编排
type UpdateParamUsecase struct {
	sysparam sysparamcontract.SysParamAdminService
}

// NewUpdateParamUsecase 创建更新参数用例
func NewUpdateParamUsecase(sysparam sysparamcontract.SysParamAdminService) *UpdateParamUsecase {
	return &UpdateParamUsecase{sysparam: sysparam}
}

// Execute 更新参数
func (uc *UpdateParamUsecase) Execute(ctx context.Context, in *UpdateParamInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(sysparamcontract.ErrInvalidInput)
	}
	return uc.sysparam.UpdateParam(ctx, &sysparaminput.UpdateParamInput{
		ID:         id,
		Value:      in.Value,
		Name:       in.Name,
		Remark:     in.Remark,
		OperatorID: in.OperatorID,
	})
}
