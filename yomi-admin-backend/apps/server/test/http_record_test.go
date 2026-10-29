// Package test HTTP 访问记录接口集成测试（须先启动 yomi-admin-backend 服务并导入 030 表）
package test

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"

	"gitee.com/hq_project/yomi-boot/component/common/testclient"
)

// httpRecordItem 是访问记录列表项
type httpRecordItem struct {
	ID         string `json:"id"`
	TraceID    string `json:"traceId"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"statusCode"`
	RequestAt  int64  `json:"requestAt"`
}

// TestHttpRecordPageAndDetail 覆盖访问记录落库后的分页查询与详情查询
func TestHttpRecordPageAndDetail(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)

	keyword := url.QueryEscape("/api/system/http-record")
	deadline := time.Now().Add(5 * time.Second)
	var record httpRecordItem

	for {
		body, ok := testclient.Get(API("/system/http-record?page=1&pageSize=10&keyword="+keyword), adminToken)
		if !ok {
			t.Fatalf("访问记录列表接口不可达: %s", body)
		}
		page := parsePage[httpRecordItem](t, body)

		if len(page.Items) > 0 {
			record = page.Items[0]
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("等待访问记录落库超时，请确认 http_record.enabled=true 且消费者已启动")
		}
		time.Sleep(100 * time.Millisecond)
	}

	if record.ID == "" || record.RequestAt == 0 {
		t.Fatalf("访问记录字段不完整: %+v", record)
	}
	if record.Method != "GET" || !strings.HasPrefix(record.Path, "/api/system/http-record") {
		t.Fatalf("访问记录内容不符合预期: %+v", record)
	}

	body, ok := testclient.Get(API("/system/http-record/detail?id="+record.ID), adminToken)
	if !ok {
		t.Fatalf("访问记录详情接口不可达: %s", body)
	}
	detail := parseSuccess(t, body)
	var data struct {
		ID          string `json:"id"`
		Path        string `json:"path"`
		StatusCode  int    `json:"statusCode"`
		RequestAt   int64  `json:"requestAt"`
		ServiceType string `json:"serviceType"`
	}
	if err := json.Unmarshal(detail.Data, &data); err != nil {
		t.Fatalf("解析访问记录详情失败: %v body=%s", err, body)
	}
	if data.ID != record.ID || data.RequestAt != record.RequestAt {
		t.Fatalf("详情与列表不一致: %+v", data)
	}
}

// TestHttpRecordPageRejectsInvalidStatusClass 验证非法的状态码分段被拒绝
func TestHttpRecordPageRejectsInvalidStatusClass(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)

	body, ok := testclient.Get(API("/system/http-record?statusClass=6xx"), adminToken)
	if !ok {
		t.Fatalf("访问记录列表接口不可达: %s", body)
	}
	assertFailure(t, body)
}
