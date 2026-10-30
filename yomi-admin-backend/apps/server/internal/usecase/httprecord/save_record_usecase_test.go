package httprecord

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	domaininput "yomi-admin-backend/domains/httprecord/pkg/shared/input"
	domaincontract "yomi-admin-backend/domains/httprecord/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// fakeRecordService 只实现 Save，其余契约方法由嵌入接口兜底
type fakeRecordService struct {
	domaincontract.RecordService
	saved *domaininput.SaveRecordInput
	err   error
}

func (f *fakeRecordService) Save(_ context.Context, in *domaininput.SaveRecordInput) error {
	f.saved = in
	return f.err
}

// testSaveRecordInput 返回字段取值为可区分的消费入参，便于逐字段核对映射
func testSaveRecordInput() *SaveRecordInput {
	return &SaveRecordInput{
		ServiceType:           "admin-bff",
		ServiceVersion:        "v1",
		TraceID:               "trace-1",
		RequestAt:             1790000000123,
		DurationMS:            150,
		Method:                "POST",
		Path:                  "/api/system/user/create",
		Route:                 "/api/system/user/create",
		Query:                 "a=1&b=2",
		ClientIP:              "10.0.0.1",
		UserAgent:             "test-agent",
		RequestHeaders:        map[string]string{"X-Trace": "abc"},
		RequestBody:           `{"username":"admin"}`,
		RequestBodyTruncated:  true,
		UserID:                42,
		ClientType:            "admin",
		StatusCode:            200,
		ResponseBody:          `{"code":"OK"}`,
		ResponseBodyTruncated: true,
		BizCode:               "OK",
		BizMsg:                "成功",
	}
}

func TestSaveRecordUsecaseMapsAllFields(t *testing.T) {
	records := &fakeRecordService{}
	if err := NewSaveRecordUsecase(records).Execute(context.Background(), testSaveRecordInput()); err != nil {
		t.Fatalf("Execute 失败: %v", err)
	}

	want := &domaininput.SaveRecordInput{
		ServiceType:           "admin-bff",
		ServiceVersion:        "v1",
		TraceID:               "trace-1",
		RequestAt:             1790000000123,
		DurationMS:            150,
		Method:                "POST",
		Path:                  "/api/system/user/create",
		Route:                 "/api/system/user/create",
		Query:                 "a=1&b=2",
		ClientIP:              "10.0.0.1",
		UserAgent:             "test-agent",
		RequestHeaders:        map[string]string{"X-Trace": "abc"},
		RequestBody:           `{"username":"admin"}`,
		RequestBodyTruncated:  true,
		UserID:                42,
		ClientType:            "admin",
		StatusCode:            200,
		ResponseBody:          `{"code":"OK"}`,
		ResponseBodyTruncated: true,
		BizCode:               "OK",
		BizMsg:                "成功",
	}
	if !reflect.DeepEqual(records.saved, want) {
		t.Fatalf("领域入参映射不完整:\n got=%+v\nwant=%+v", records.saved, want)
	}
}

func TestSaveRecordUsecaseClassifiesInvalidRecord(t *testing.T) {
	records := &fakeRecordService{err: errors.Sentinel(domaincontract.ErrInvalidInput)}

	err := NewSaveRecordUsecase(records).Execute(context.Background(), testSaveRecordInput())
	if !errors.Is(err, ErrInvalidRecord) {
		t.Fatalf("期望 ErrInvalidRecord，实际 %v", err)
	}
	if !errors.Is(err, domaincontract.ErrInvalidInput) {
		t.Fatalf("期望保留领域错误链，实际 %v", err)
	}
}

func TestSaveRecordUsecasePassesThroughOtherErrors(t *testing.T) {
	cause := fmt.Errorf("db unavailable")
	records := &fakeRecordService{err: cause}

	err := NewSaveRecordUsecase(records).Execute(context.Background(), testSaveRecordInput())
	if errors.Is(err, ErrInvalidRecord) {
		t.Fatalf("非不合法错误不应归类为 ErrInvalidRecord: %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("期望透传原始错误，实际 %v", err)
	}
}
