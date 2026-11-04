package service

import (
	"context"
	"strings"
	"testing"
	"yomi-admin-backend/domains/auditlog/internal/repo_contract"
	"yomi-admin-backend/domains/auditlog/pkg/shared/entity"
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"
	"yomi-admin-backend/domains/auditlog/pkg/shared/input"
	"yomi-admin-backend/domains/auditlog/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

// fakeRepo 记录仓储调用，供服务单测断言
type fakeRepo struct {
	created   []*entity.Log
	createErr error
	pageItems []*entity.Log
	pageTotal int64
	pageErr   error
	gotFilter *repo_contract.Filter
}

func (f *fakeRepo) Create(_ context.Context, log *entity.Log) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, log)
	return nil
}

func (f *fakeRepo) Page(_ context.Context, filter *repo_contract.Filter) ([]*entity.Log, int64, error) {
	f.gotFilter = filter
	return f.pageItems, f.pageTotal, f.pageErr
}

// fakeIDs 生成自增 ID
type fakeIDs struct{ next int64 }

func (f *fakeIDs) NextID() (int64, error) {
	f.next++
	return f.next, nil
}

// errIDs 模拟雪花 ID 生成失败
type errIDs struct{}

func (errIDs) NextID() (int64, error) { return 0, errors.New("id down") }

func newService() (*AuditLogService, *fakeRepo) {
	repo := &fakeRepo{}
	return NewAuditLogService(repo, &fakeIDs{}), repo
}

func TestRecordPersistsSanitizedLog(t *testing.T) {
	svc, repo := newService()

	err := svc.Record(context.Background(), &input.RecordInput{
		AuditID:      99,
		Phase:        enums.PhaseFinished,
		EventCode:    "  auth.login  ",
		OperatorID:   7,
		OperatorName: "  admin  ",
		ClientIP:     " 1.2.3.4 ",
		TargetType:   " user ",
		TargetID:     8,
		Result:       enums.ResultSuccess,
		BizCode:      " OK ",
	})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("仓储未写入: %+v", repo.created)
	}
	log := repo.created[0]
	if log.ID == 0 || log.AuditID != 99 || log.Phase != enums.PhaseFinished ||
		log.EventCode != "auth.login" || log.OperatorName != "admin" ||
		log.ClientIP != "1.2.3.4" || log.TargetType != "user" ||
		log.BizCode != "OK" || log.Result != enums.ResultSuccess || log.CreateTime <= 0 {
		t.Fatalf("落库字段不符合预期: %+v", log)
	}
}

func TestRecordTruncatesLongFields(t *testing.T) {
	svc, repo := newService()
	long := strings.Repeat("x", 200)

	err := svc.Record(context.Background(), &input.RecordInput{
		AuditID:      99,
		Phase:        enums.PhaseFinished,
		EventCode:    "auth.login",
		OperatorName: long,
		ClientIP:     long,
		TargetType:   long,
		Result:       enums.ResultFailure,
		BizCode:      long,
	})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	log := repo.created[0]
	if len(log.OperatorName) > 64 || len(log.ClientIP) > 64 || len(log.TargetType) > 32 ||
		len(log.BizCode) > 64 {
		t.Fatalf("超长字段未截断: %+v", log)
	}
}

func TestRecordRejectsInvalidInput(t *testing.T) {
	svc, _ := newService()
	cases := []*input.RecordInput{
		nil,
		{AuditID: 1, Phase: enums.PhaseFinished, EventCode: "", Result: enums.ResultSuccess},
		{AuditID: 1, Phase: enums.PhaseFinished, EventCode: strings.Repeat("x", 65), Result: enums.ResultSuccess},
		{AuditID: 0, Phase: enums.PhaseFinished, EventCode: "auth.login", Result: enums.ResultSuccess},
		{AuditID: 1, Phase: "UNKNOWN", EventCode: "auth.login", Result: enums.ResultSuccess},
		{AuditID: 1, Phase: enums.PhaseStarted, EventCode: "auth.login", Result: enums.ResultSuccess},
		{AuditID: 1, Phase: enums.PhaseFinished, EventCode: "auth.login", Result: enums.ResultPending},
		{AuditID: 1, Phase: enums.PhaseFinished, EventCode: "auth.login", Result: "UNKNOWN"},
		{AuditID: 1, Phase: enums.PhaseFinished, EventCode: "auth.login", Result: enums.ResultSuccess, OperatorID: -1},
		{AuditID: 1, Phase: enums.PhaseFinished, EventCode: "auth.login", Result: enums.ResultSuccess, TargetID: -1},
	}
	for _, in := range cases {
		if err := svc.Record(context.Background(), in); !errors.Is(err, service_contract.ErrInvalidInput) {
			t.Fatalf("Record(%+v) error = %v, want ErrInvalidInput", in, err)
		}
	}
}

func TestRecordPropagatesIDAndRepoErrors(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewAuditLogService(repo, errIDs{})
	if err := svc.Record(context.Background(), &input.RecordInput{
		AuditID: 1, Phase: enums.PhaseFinished, EventCode: "auth.login", Result: enums.ResultSuccess,
	}); err == nil {
		t.Fatal("ID 生成失败应返回错误")
	}

	repo.createErr = errors.New("db down")
	svc = NewAuditLogService(repo, &fakeIDs{})
	if err := svc.Record(context.Background(), &input.RecordInput{
		AuditID: 1, Phase: enums.PhaseFinished, EventCode: "auth.login", Result: enums.ResultSuccess,
	}); err == nil {
		t.Fatal("仓储失败应返回错误")
	}
}

func TestPagePropagatesFilters(t *testing.T) {
	svc, repo := newService()
	repo.pageItems = []*entity.Log{{ID: 1, EventCode: "auth.login", Result: enums.ResultSuccess}}
	repo.pageTotal = 1

	out, err := svc.Page(context.Background(), &input.PageInput{
		Keyword:   "admin",
		EventCode: "  auth.login ",
		Phase:     enums.PhaseFinished,
		Result:    enums.ResultSuccess,
		AuditID:   123,
		StartTime: 100,
		EndTime:   200,
		Page:      2,
		PageSize:  20,
	})
	if err != nil {
		t.Fatalf("Page() error = %v", err)
	}
	if out.Total != 1 || len(out.Items) != 1 {
		t.Fatalf("分页出参错误: %+v", out)
	}
	if repo.gotFilter == nil || repo.gotFilter.EventCode != "auth.login" || repo.gotFilter.Phase != enums.PhaseFinished ||
		repo.gotFilter.AuditID != 123 || repo.gotFilter.Page != 2 ||
		repo.gotFilter.PageSize != 20 || repo.gotFilter.StartTime != 100 || repo.gotFilter.EndTime != 200 {
		t.Fatalf("过滤条件未透传: %+v", repo.gotFilter)
	}
}

func TestPageRejectsInvalidFilters(t *testing.T) {
	svc, _ := newService()
	cases := []*input.PageInput{
		nil,
		{Result: "UNKNOWN"},
		{Phase: "UNKNOWN"},
		{AuditID: -1},
		{StartTime: 200, EndTime: 100},
		{EventCode: strings.Repeat("x", 65)},
		// 列表查询必须带完整时间范围
		{},
		{StartTime: 100},
		{EndTime: 200},
		{StartTime: 100, EventCode: "auth.login"},
	}
	for _, in := range cases {
		if _, err := svc.Page(context.Background(), in); !errors.Is(err, service_contract.ErrInvalidInput) {
			t.Fatalf("Page(%+v) error = %v, want ErrInvalidInput", in, err)
		}
	}
}

// TestPageAllowsAuditIDWithoutRange 验证按事件 ID 点查时不受"必须带时间范围"约束
func TestPageAllowsAuditIDWithoutRange(t *testing.T) {
	svc, repo := newService()
	repo.pageItems = nil
	repo.pageTotal = 0

	if _, err := svc.Page(context.Background(), &input.PageInput{AuditID: 123, Page: 1, PageSize: 10}); err != nil {
		t.Fatalf("Page(auditId only) error = %v", err)
	}
	if repo.gotFilter == nil || repo.gotFilter.AuditID != 123 ||
		repo.gotFilter.StartTime != 0 || repo.gotFilter.EndTime != 0 {
		t.Fatalf("auditId 查询过滤条件错误: %+v", repo.gotFilter)
	}
}
