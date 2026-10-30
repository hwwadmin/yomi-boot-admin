package repo

import (
	"context"
	"testing"
	"yomi-admin-backend/domains/auditlog/internal/repo_contract"
	"yomi-admin-backend/domains/auditlog/pkg/shared/entity"
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"

	"gitee.com/hq_project/yomi-boot/component/common/db/dbtest"
)

func newTestRepo(t *testing.T) (*AuditLogRepo, context.Context) {
	t.Helper()
	gormDB, cleanup := dbtest.New(t, dbtest.WithTables(&auditLogModel{}))
	t.Cleanup(cleanup)
	return NewAuditLogRepo(gormDB), context.Background()
}

// mustCreate 写入一条审计日志
func mustCreate(t *testing.T, r *AuditLogRepo, ctx context.Context, log *entity.Log) {
	t.Helper()
	if err := r.Create(ctx, log); err != nil {
		t.Fatalf("Create(%d) error = %v", log.ID, err)
	}
}

// seedLogs 写入三条不同时间、事件与结果的记录
func seedLogs(t *testing.T, r *AuditLogRepo, ctx context.Context) {
	t.Helper()
	mustCreate(t, r, ctx, &entity.Log{
		ID: 1, AuditID: 1001, Phase: enums.PhaseFinished, EventCode: "auth.login", OperatorID: 7, OperatorName: "admin",
		ClientIP: "10.0.0.1", Result: enums.ResultSuccess, BizCode: "OK", CreateTime: 3000,
	})
	mustCreate(t, r, ctx, &entity.Log{
		ID: 2, AuditID: 1002, Phase: enums.PhaseFinished, EventCode: "auth.change_password", OperatorID: 8, OperatorName: "alice",
		ClientIP: "10.0.0.2", Result: enums.ResultFailure, BizCode: "ADMIN_OLD_PASSWORD_MISMATCH", CreateTime: 1000,
	})
	mustCreate(t, r, ctx, &entity.Log{
		ID: 3, AuditID: 1003, Phase: enums.PhaseFinished, EventCode: "auth.reset_password", OperatorID: 7, OperatorName: "admin",
		ClientIP: "10.0.0.3", TargetType: "user", TargetID: 9,
		Result: enums.ResultSuccess, BizCode: "OK", CreateTime: 2000,
	})
}

func TestAuditLogRepoCreateAndPage(t *testing.T) {
	r, ctx := newTestRepo(t)
	seedLogs(t, r, ctx)

	items, total, err := r.Page(ctx, &repo_contract.Filter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Page() error = %v", err)
	}
	if total != 3 || len(items) != 3 {
		t.Fatalf("total=%d len=%d, want 3/3", total, len(items))
	}
	// 按事件时间倒序
	if items[0].ID != 1 || items[1].ID != 3 || items[2].ID != 2 {
		t.Fatalf("排序错误: %d,%d,%d", items[0].ID, items[1].ID, items[2].ID)
	}
	// 实体映射与快照字段
	if items[1].AuditID != 1003 || items[1].Phase != enums.PhaseFinished ||
		items[1].TargetType != "user" || items[1].TargetID != 9 {
		t.Fatalf("目标字段映射错误: %+v", items[1])
	}
	if items[0].CreateTime != 3000 {
		t.Fatalf("CreateTime = %d, want 3000", items[0].CreateTime)
	}
}

func TestAuditLogRepoFilters(t *testing.T) {
	r, ctx := newTestRepo(t)
	seedLogs(t, r, ctx)

	// keyword 命中操作者/来源 IP（目标用户名不落库，不参与 SQL 过滤）
	for _, keyword := range []string{"ali", "10.0.0.2"} {
		items, total, err := r.Page(ctx, &repo_contract.Filter{Keyword: keyword, Page: 1, PageSize: 10})
		if err != nil || total != 1 || len(items) != 1 {
			t.Fatalf("keyword=%s total=%d err=%v", keyword, total, err)
		}
	}

	// 事件编码
	items, total, err := r.Page(ctx, &repo_contract.Filter{EventCode: "auth.login", Page: 1, PageSize: 10})
	if err != nil || total != 1 || items[0].ID != 1 {
		t.Fatalf("eventCode 过滤错误: total=%d err=%v", total, err)
	}

	// 结果
	_, total, err = r.Page(ctx, &repo_contract.Filter{Result: enums.ResultFailure, Page: 1, PageSize: 10})
	if err != nil || total != 1 {
		t.Fatalf("result 过滤错误: total=%d err=%v", total, err)
	}

	// 阶段
	_, total, err = r.Page(ctx, &repo_contract.Filter{Phase: enums.PhaseFinished, Page: 1, PageSize: 10})
	if err != nil || total != 3 {
		t.Fatalf("phase 过滤错误: total=%d err=%v", total, err)
	}

	// 审计事件实例
	items, total, err = r.Page(ctx, &repo_contract.Filter{AuditID: 1002, Page: 1, PageSize: 10})
	if err != nil || total != 1 || items[0].ID != 2 {
		t.Fatalf("auditId 过滤错误: total=%d err=%v", total, err)
	}

	// 时间范围（含边界）
	items, total, err = r.Page(ctx, &repo_contract.Filter{StartTime: 2000, EndTime: 3000, Page: 1, PageSize: 10})
	if err != nil || total != 2 || items[0].ID != 1 || items[1].ID != 3 {
		t.Fatalf("时间过滤错误: total=%d err=%v", total, err)
	}

	// 分页切片
	items, total, err = r.Page(ctx, &repo_contract.Filter{Page: 2, PageSize: 1})
	if err != nil || total != 3 || len(items) != 1 || items[0].ID != 3 {
		t.Fatalf("分页错误: total=%d items=%+v err=%v", total, items, err)
	}
}
