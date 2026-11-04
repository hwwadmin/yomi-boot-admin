package service

import (
	"context"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
	"yomi-admin-backend/domains/httprecord/internal/repo_contract"
	"yomi-admin-backend/domains/httprecord/pkg/shared/entity"
	"yomi-admin-backend/domains/httprecord/pkg/shared/input"
	"yomi-admin-backend/domains/httprecord/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

// fakeRecordRepo 是内存访问记录仓储
type fakeRecordRepo struct {
	records    []*entity.Record
	lastFilter *repo_contract.RecordFilter
}

func (r *fakeRecordRepo) Create(_ context.Context, record *entity.Record) error {
	r.records = append(r.records, record)
	return nil
}

func (r *fakeRecordRepo) FindByID(_ context.Context, id int64) (*entity.Record, bool, error) {
	for _, record := range r.records {
		if record.ID == id {
			return record, true, nil
		}
	}
	return nil, false, nil
}

func (r *fakeRecordRepo) Page(_ context.Context, filter *repo_contract.RecordFilter) ([]*entity.Record, int64, error) {
	r.lastFilter = filter
	start := (filter.Page - 1) * filter.PageSize
	if start >= len(r.records) {
		return []*entity.Record{}, int64(len(r.records)), nil
	}
	end := start + filter.PageSize
	if end > len(r.records) {
		end = len(r.records)
	}
	return r.records[start:end], int64(len(r.records)), nil
}

// fixedIDGenerator 返回固定雪花 ID
type fixedIDGenerator struct {
	id int64
}

func (g fixedIDGenerator) NextID() (int64, error) {
	return g.id, nil
}

func TestRecordServiceSave(t *testing.T) {
	t.Parallel()

	records := &fakeRecordRepo{}
	svc := NewRecordService(records, fixedIDGenerator{id: 100})

	longPath := strings.Repeat("路径", 400)
	err := svc.Save(context.Background(), &input.SaveRecordInput{
		ServiceType:    "yomi-admin-backend",
		RequestAt:      1761619200000,
		Method:         "post",
		Path:           longPath,
		UserID:         100000000000000001,
		StatusCode:     200,
		RequestHeaders: map[string]string{"Authorization": "***"},
		RequestBody:    `{"username":"admin"}`,
	})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if len(records.records) != 1 {
		t.Fatalf("records = %d, want 1", len(records.records))
	}
	saved := records.records[0]
	if saved.ID != 100 {
		t.Fatalf("ID = %d, want 100", saved.ID)
	}
	if saved.Method != "POST" {
		t.Fatalf("Method = %s, want POST", saved.Method)
	}
	if len(saved.Path) > maxPathBytes {
		t.Fatalf("Path 未截断: %d bytes", len(saved.Path))
	}
	if !utf8.ValidString(saved.Path) {
		t.Fatal("Path 截断后不是合法 UTF-8")
	}
	if saved.RequestAt != 1761619200000 {
		t.Fatalf("RequestAt = %d", saved.RequestAt)
	}
	if saved.RequestHeaders == "" || !strings.Contains(saved.RequestHeaders, "Authorization") {
		t.Fatalf("RequestHeaders = %q", saved.RequestHeaders)
	}
}

func TestRecordServiceSaveDefaultsRequestAt(t *testing.T) {
	t.Parallel()

	records := &fakeRecordRepo{}
	svc := NewRecordService(records, fixedIDGenerator{id: 101})
	before := time.Now().UnixMilli()

	if err := svc.Save(context.Background(), &input.SaveRecordInput{Method: "GET", Path: "/api/x"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if records.records[0].RequestAt < before {
		t.Fatalf("RequestAt = %d, want >= %d", records.records[0].RequestAt, before)
	}
}

func TestRecordServiceSaveRejectsEmptyPath(t *testing.T) {
	t.Parallel()

	svc := NewRecordService(&fakeRecordRepo{}, fixedIDGenerator{id: 102})
	err := svc.Save(context.Background(), &input.SaveRecordInput{Method: "GET"})
	if !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestRecordServicePageValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   *input.PageRecordsInput
	}{
		{name: "非法方法", in: &input.PageRecordsInput{Method: "TRACE", StartTime: 1, EndTime: 2}},
		{name: "非法状态分段", in: &input.PageRecordsInput{StatusClass: "6xx", StartTime: 1, EndTime: 2}},
		// 列表查询必须带完整时间范围
		{name: "缺少时间范围", in: &input.PageRecordsInput{}},
		{name: "只有起始时间", in: &input.PageRecordsInput{StartTime: 1}},
		{name: "只有结束时间", in: &input.PageRecordsInput{EndTime: 2}},
		{name: "时间区间倒置", in: &input.PageRecordsInput{StartTime: 2, EndTime: 1}},
	}

	svc := NewRecordService(&fakeRecordRepo{}, fixedIDGenerator{id: 103})
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := svc.Page(context.Background(), tt.in); !errors.Is(err, service_contract.ErrInvalidInput) {
				t.Fatalf("err = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestRecordServicePageNormalizesFilter(t *testing.T) {
	t.Parallel()

	records := &fakeRecordRepo{}
	svc := NewRecordService(records, fixedIDGenerator{id: 104})
	if _, err := svc.Page(context.Background(), &input.PageRecordsInput{
		Method:      "get",
		StatusClass: "4XX",
		StartTime:   1,
		EndTime:     2,
	}); err != nil {
		t.Fatalf("Page() error = %v", err)
	}

	filter := records.lastFilter
	if filter.Method != "GET" || filter.StatusClass != "4xx" {
		t.Fatalf("filter = %+v", filter)
	}
}

func TestRecordServiceGetNotFound(t *testing.T) {
	t.Parallel()

	svc := NewRecordService(&fakeRecordRepo{}, fixedIDGenerator{id: 105})
	if _, err := svc.Get(context.Background(), 1); !errors.Is(err, service_contract.ErrRecordNotFound) {
		t.Fatalf("err = %v, want ErrRecordNotFound", err)
	}
	if _, err := svc.Get(context.Background(), 0); !errors.Is(err, service_contract.ErrRecordNotFound) {
		t.Fatalf("err = %v, want ErrRecordNotFound", err)
	}
}
