package sysparam

import (
	"context"

	sysparaminput "yomi-admin-backend/domains/sysparam/pkg/shared/input"
	sysparamcontract "yomi-admin-backend/domains/sysparam/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// DeleteParamInput HTTP 删除参数入参
type DeleteParamInput struct {
	ID         dto.JSONInt64 `json:"id"`
	OperatorID int64         `json:"-"`
}

// DeleteParamUsecase 删除参数编排
type DeleteParamUsecase struct {
	sysparam sysparamcontract.SysParamAdminService
}

// NewDeleteParamUsecase 创建删除参数用例
func NewDeleteParamUsecase(sysparam sysparamcontract.SysParamAdminService) *DeleteParamUsecase {
	return &DeleteParamUsecase{sysparam: sysparam}
}

// Execute 物理删除参数
func (uc *DeleteParamUsecase) Execute(ctx context.Context, in *DeleteParamInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(sysparamcontract.ErrInvalidInput)
	}
	return uc.sysparam.DeleteParam(ctx, &sysparaminput.DeleteParamInput{
		ID:         id,
		OperatorID: in.OperatorID,
	})
}
