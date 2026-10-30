package sysparam

import (
	"context"

	sysparamenums "yomi-admin-backend/domains/sysparam/pkg/shared/enums"
	sysparaminput "yomi-admin-backend/domains/sysparam/pkg/shared/input"
	sysparamcontract "yomi-admin-backend/domains/sysparam/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// SetParamStatusInput HTTP 参数启停入参
type SetParamStatusInput struct {
	ID         dto.JSONInt64        `json:"id"`
	Status     sysparamenums.Status `json:"status"`
	OperatorID int64                `json:"-"`
}

// SetParamStatusUsecase 参数启停编排
type SetParamStatusUsecase struct {
	sysparam sysparamcontract.SysParamAdminService
}

// NewSetParamStatusUsecase 创建参数启停用例
func NewSetParamStatusUsecase(sysparam sysparamcontract.SysParamAdminService) *SetParamStatusUsecase {
	return &SetParamStatusUsecase{sysparam: sysparam}
}

// Execute 设置参数启停
func (uc *SetParamStatusUsecase) Execute(ctx context.Context, in *SetParamStatusInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(sysparamcontract.ErrInvalidInput)
	}
	return uc.sysparam.SetParamStatus(ctx, &sysparaminput.SetParamStatusInput{
		ID:         id,
		Status:     in.Status,
		OperatorID: in.OperatorID,
	})
}
