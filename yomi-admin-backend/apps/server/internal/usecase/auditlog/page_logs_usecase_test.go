package auditlog

import (
	"context"
	"errors"
	"testing"
	"yomi-admin-backend/apps/server/internal/support/audit"
	"yomi-admin-backend/domains/auditlog/pkg/shared/entity"
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"
	domaininput "yomi-admin-backend/domains/auditlog/pkg/shared/input"
	domainoutput "yomi-admin-backend/domains/auditlog/pkg/shared/output"
	usercontract "yomi-admin-backend/domains/user/service_contract"
)

// fakeLogReader 是返回固定结果的分页读取器
type fakeLogReader struct {
	items []*entity.Log
}

// Page 返回预置记录
func (f *fakeLogReader) Page(context.Context, *domaininput.PageInput) (*domainoutput.PageLogsOutput, error) {
	return &domainoutput.PageLogsOutput{Items: f.items, Total: int64(len(f.items))}, nil
}

// fakeUserService 只覆盖取名能力，其余方法用内嵌接口占位
type fakeUserService struct {
	usercontract.UserService
	askedIDs []int64
	names    map[int64]string
	err      error
}

// GetNamesByIDs 返回预置用户名
func (f *fakeUserService) GetNamesByIDs(_ context.Context, ids []int64) (map[int64]string, error) {
	f.askedIDs = ids
	return f.names, f.err
}

// TestPageLogsFillsTargetName 验证只为用户目标回填用户名，无目标的记录保持为空
func TestPageLogsFillsTargetName(t *testing.T) {
	users := &fakeUserService{names: map[int64]string{9: "bob"}}
	uc := NewPageLogsUsecase(&fakeLogReader{items: []*entity.Log{
		{
			ID: 1, AuditID: 1001, Phase: enums.PhaseFinished, EventCode: "auth.reset_password",
			TargetType: audit.TargetTypeUser, TargetID: 9,
		},
		{ID: 2, AuditID: 1002, Phase: enums.PhaseFinished, EventCode: "auth.login"},
	}}, users)

	out, err := uc.Execute(context.Background(), &PageLogsInput{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(out.Items) != 2 || out.Items[0].TargetName != "bob" || out.Items[1].TargetName != "" {
		t.Fatalf("目标用户名回填错误: %+v", out.Items)
	}
	if len(users.askedIDs) != 1 || users.askedIDs[0] != 9 {
		t.Fatalf("应只按目标用户 ID 查询用户名: %+v", users.askedIDs)
	}
}

// TestPageLogsKeepsItemsWhenNameLookupFails 验证取用户名失败时不报错、列表原样返回
func TestPageLogsKeepsItemsWhenNameLookupFails(t *testing.T) {
	users := &fakeUserService{err: errors.New("boom")}
	uc := NewPageLogsUsecase(&fakeLogReader{items: []*entity.Log{
		{
			ID: 1, AuditID: 1001, Phase: enums.PhaseFinished, EventCode: "auth.reset_password",
			TargetType: audit.TargetTypeUser, TargetID: 9,
		},
	}}, users)

	out, err := uc.Execute(context.Background(), &PageLogsInput{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].TargetName != "" || out.Items[0].TargetID != 9 {
		t.Fatalf("回填失败时列表应原样返回: %+v", out.Items)
	}
}
