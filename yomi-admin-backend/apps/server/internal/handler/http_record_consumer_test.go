package handler

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	httprecoreusecase "yomi-admin-backend/apps/server/internal/usecase/httprecord"

	"gitee.com/hq_project/yomi-boot/component/common/httprecord"
)

// fakeRecordSaver 记录用例入参并返回预设错误
type fakeRecordSaver struct {
	input *httprecoreusecase.SaveRecordInput
	err   error
}

func (f *fakeRecordSaver) Execute(_ context.Context, in *httprecoreusecase.SaveRecordInput) error {
	f.input = in
	return f.err
}

func TestHttpRecordConsumerDropsNilRecord(t *testing.T) {
	saver := &fakeRecordSaver{}
	got := newHttpRecordConsumer(saver)(context.Background(), nil)

	if got.Status != httprecord.RecordHandleDrop {
		t.Fatalf("nil 记录期望丢弃，实际 %v", got.Status)
	}
	if saver.input != nil {
		t.Fatalf("nil 记录不应调用保存: %+v", saver.input)
	}
}

func TestHttpRecordConsumerMapsRecord(t *testing.T) {
	saver := &fakeRecordSaver{}
	requestAt := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	record := &httprecord.HttpRecord{
		ServiceType:           "admin-bff",
		ServiceVersion:        "v1",
		TraceID:               "trace-1",
		RequestAt:             requestAt,
		Duration:              150 * time.Millisecond,
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

	got := newHttpRecordConsumer(saver)(context.Background(), record)
	if got.Status != httprecord.RecordHandleOK {
		t.Fatalf("期望处理成功，实际 %v", got.Status)
	}

	want := &httprecoreusecase.SaveRecordInput{
		ServiceType:           "admin-bff",
		ServiceVersion:        "v1",
		TraceID:               "trace-1",
		RequestAt:             requestAt.UnixMilli(),
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
	if !reflect.DeepEqual(saver.input, want) {
		t.Fatalf("用例入参映射不完整:\n got=%+v\nwant=%+v", saver.input, want)
	}
}

func TestHttpRecordConsumerDropsInvalidRecord(t *testing.T) {
	saver := &fakeRecordSaver{err: fmt.Errorf("wrapped: %w", httprecoreusecase.ErrInvalidRecord)}
	got := newHttpRecordConsumer(saver)(context.Background(), &httprecord.HttpRecord{Method: "GET", Path: "/"})

	if got.Status != httprecord.RecordHandleDrop {
		t.Fatalf("非法记录期望丢弃，实际 %v", got.Status)
	}
}

func TestHttpRecordConsumerRetriesOnSaveFailure(t *testing.T) {
	saver := &fakeRecordSaver{err: fmt.Errorf("db unavailable")}
	got := newHttpRecordConsumer(saver)(context.Background(), &httprecord.HttpRecord{Method: "GET", Path: "/"})

	if got.Status != httprecord.RecordHandleRetry {
		t.Fatalf("落库失败期望重试，实际 %v", got.Status)
	}
}
