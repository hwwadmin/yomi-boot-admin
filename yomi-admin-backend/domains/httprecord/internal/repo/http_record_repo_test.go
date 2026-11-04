package repo

import (
	"context"
	"testing"
	"time"
	"yomi-admin-backend/domains/httprecord/internal/repo_contract"
	"yomi-admin-backend/domains/httprecord/pkg/shared/entity"

	"github.com/hwwadmin/yomi-boot/component/common/db/dbtest"
)

func TestRecordRepoCreateAndPage(t *testing.T) {
	gormDB, cleanup := dbtest.New(t, dbtest.WithTables(&httpRecordModel{}))
	defer cleanup()

	repo := NewRecordRepo(gormDB)
	ctx := context.Background()
	base := time.Now().UTC().Truncate(time.Millisecond)

	first := &entity.Record{
		ID:          1,
		ServiceType: "yomi-admin-backend",
		TraceID:     "trace-1",
		RequestAt:   base.UnixMilli(),
		DurationMS:  12,
		Method:      "GET",
		Path:        "/api/system/user",
		Route:       "/api/system/user",
		Query:       "page=1",
		ClientIP:    "127.0.0.1",
		UserID:      100,
		StatusCode:  200,
		BizCode:     "OK",
	}
	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, ok, err := repo.FindByID(ctx, first.ID)
	if err != nil || !ok {
		t.Fatalf("FindByID() ok=%v err=%v", ok, err)
	}
	if found.TraceID != "trace-1" || found.Method != "GET" || found.UserID != 100 {
		t.Fatalf("found = %+v", found)
	}
	if found.CreateTime == 0 {
		t.Fatal("CreateTime 未回填")
	}

	// 同一 trace_id + request_at 重复投递应被忽略
	duplicate := *first
	duplicate.ID = 2
	if err := repo.Create(ctx, &duplicate); err != nil {
		t.Fatalf("重复 Create() error = %v", err)
	}

	second := &entity.Record{
		ID:         3,
		TraceID:    "trace-2",
		RequestAt:  base.Add(time.Second).UnixMilli(),
		Method:     "POST",
		Path:       "/api/auth/login",
		ClientIP:   "10.0.0.1",
		UserID:     200,
		StatusCode: 500,
	}
	if err := repo.Create(ctx, second); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	tests := []struct {
		name   string
		filter repo_contract.RecordFilter
		want   int64
	}{
		{name: "全部", filter: repo_contract.RecordFilter{Page: 1, PageSize: 10}, want: 2},
		{name: "关键字命中路径", filter: repo_contract.RecordFilter{Keyword: "user", Page: 1, PageSize: 10}, want: 1},
		{name: "方法", filter: repo_contract.RecordFilter{Method: "POST", Page: 1, PageSize: 10}, want: 1},
		{name: "状态分段", filter: repo_contract.RecordFilter{StatusClass: "5xx", Page: 1, PageSize: 10}, want: 1},
		{name: "用户", filter: repo_contract.RecordFilter{UserID: 100, Page: 1, PageSize: 10}, want: 1},
		{name: "时间区间", filter: repo_contract.RecordFilter{StartTime: base.Add(500 * time.Millisecond).UnixMilli(), Page: 1, PageSize: 10}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := repo.Page(ctx, &tt.filter)
			if err != nil {
				t.Fatalf("Page() error = %v", err)
			}
			if total != tt.want || int64(len(items)) != tt.want {
				t.Fatalf("total=%d items=%d, want %d", total, len(items), tt.want)
			}
		})
	}

	items, _, err := repo.Page(ctx, &repo_contract.RecordFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Page() error = %v", err)
	}
	if len(items) != 2 || items[0].TraceID != "trace-2" {
		t.Fatalf("按请求时间倒序失败: %+v", items)
	}
}

func TestRecordRepoPageEmpty(t *testing.T) {
	gormDB, cleanup := dbtest.New(t, dbtest.WithTables(&httpRecordModel{}))
	defer cleanup()

	items, total, err := NewRecordRepo(gormDB).Page(context.Background(), &repo_contract.RecordFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Page() error = %v", err)
	}
	if total != 0 || len(items) != 0 {
		t.Fatalf("total=%d items=%d, want 0", total, len(items))
	}
}
