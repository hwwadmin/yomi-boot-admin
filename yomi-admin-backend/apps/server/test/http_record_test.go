// Package test HTTP 访问记录接口集成测试（须先启动 yomi-admin-backend 服务并导入 030 表）
package test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"
)

// httpRecordItem 是访问记录列表项
type httpRecordItem struct {
	ID          string `json:"id"`
	TraceParent string `json:"traceParent"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	StatusCode  int    `json:"statusCode"`
	RequestAt   int64  `json:"requestAt"`
}

// TestHttpRecordPageAndDetail 覆盖访问记录落库后的分页查询与详情查询
func TestHttpRecordPageAndDetail(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)

	// 登录请求本身会落一条访问记录，用它作为样本；访问记录页面自身的请求
	// 已由 http_record.skip_paths 忽略，不能再作为样本来源
	keyword := url.QueryEscape("/api/auth/login")
	// 列表查询必须带完整时间范围；刚落库的记录就在当前时刻
	rangeQuery := "&startTime=" + fmt.Sprint(time.Now().Add(-time.Hour).UnixMilli()) +
		"&endTime=" + fmt.Sprint(time.Now().Add(time.Minute).UnixMilli())
	deadline := time.Now().Add(5 * time.Second)
	var record httpRecordItem

	for {
		body, ok := Get(API("/system/http-record?page=1&pageSize=10&keyword="+keyword+rangeQuery), adminToken)
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
	if record.Method != "POST" || !strings.HasPrefix(record.Path, "/api/auth/login") {
		t.Fatalf("访问记录内容不符合预期: %+v", record)
	}

	body, ok := Get(API("/system/http-record/detail?id="+record.ID), adminToken)
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

// TestHttpRecordPageSkipsSelf 验证访问记录页面自身的请求不再被采集
func TestHttpRecordPageSkipsSelf(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)

	// 只观察本次测试开始后的记录，排除历史数据
	startedAt := time.Now()
	rangeQuery := fmt.Sprintf("&startTime=%d&endTime=%d",
		startedAt.UnixMilli(), time.Now().Add(time.Minute).UnixMilli())

	// 访问记录页面的查询应被采集器跳过
	for i := 0; i < 3; i++ {
		body, ok := Get(API("/system/http-record?page=1&pageSize=1"+rangeQuery), adminToken)
		if !ok {
			t.Fatalf("访问记录列表接口不可达: %s", body)
		}
	}

	// 观察窗口内这些查询都不应落库；若 skip_paths 失效会很快出现记录
	keyword := url.QueryEscape("/api/system/http-record")
	deadline := time.Now().Add(2 * time.Second)
	for {
		body, ok := Get(API("/system/http-record?page=1&pageSize=10&keyword="+keyword+rangeQuery), adminToken)
		if !ok {
			t.Fatalf("访问记录列表接口不可达: %s", body)
		}
		page := parsePage[httpRecordItem](t, body)
		if len(page.Items) > 0 {
			t.Fatalf("访问记录页面的请求不应被采集: %+v", page.Items[0])
		}
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// TestHttpRecordPageRejectsInvalidStatusClass 验证非法的状态码分段被拒绝
func TestHttpRecordPageRejectsInvalidStatusClass(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)

	rangeQuery := fmt.Sprintf("&startTime=%d&endTime=%d",
		time.Now().Add(-time.Hour).UnixMilli(), time.Now().Add(time.Minute).UnixMilli())
	body, ok := Get(API("/system/http-record?statusClass=6xx"+rangeQuery), adminToken)
	if !ok {
		t.Fatalf("访问记录列表接口不可达: %s", body)
	}
	assertFailure(t, body)
}
