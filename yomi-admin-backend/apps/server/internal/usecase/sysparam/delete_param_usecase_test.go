package sysparam

import (
	"context"
	"testing"

	sysparamentity "yomi-admin-backend/domains/sysparam/pkg/shared/entity"
	sysparaminput "yomi-admin-backend/domains/sysparam/pkg/shared/input"
	sysparamcontract "yomi-admin-backend/domains/sysparam/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// fakeParamAdmin 只实现创建与删除，其余方法由嵌入接口兜底
type fakeParamAdmin struct {
	sysparamcontract.SysParamAdminService
	deleteErr error
}

func (f *fakeParamAdmin) DeleteParam(context.Context, *sysparaminput.DeleteParamInput) error {
	return f.deleteErr
}

func (f *fakeParamAdmin) CreateParam(context.Context, *sysparaminput.CreateParamInput) (*sysparamentity.Param, error) {
	return &sysparamentity.Param{ID: 5, Key: "site.name", Name: "站点名称"}, nil
}

func TestDeleteParamTreatsNotFoundAsSuccess(t *testing.T) {
	uc := NewDeleteParamUsecase(&fakeParamAdmin{deleteErr: sysparamcontract.ErrParamNotFound})
	err := uc.Execute(context.Background(), &DeleteParamInput{ID: dto.JSONInt64(5), OperatorID: 1})
	if err != nil {
		t.Fatalf("重复删除参数应返回 nil, got %v", err)
	}
}

func TestDeleteParamKeepsOtherErrors(t *testing.T) {
	other := errors.NewBiz("ADMIN_PARAM_DELETE_FAILED", "删除失败")
	uc := NewDeleteParamUsecase(&fakeParamAdmin{deleteErr: other})
	err := uc.Execute(context.Background(), &DeleteParamInput{ID: dto.JSONInt64(5), OperatorID: 1})
	if !errors.Is(err, other) {
		t.Fatalf("非 NOT_FOUND 错误应原样返回, got %v", err)
	}
}
