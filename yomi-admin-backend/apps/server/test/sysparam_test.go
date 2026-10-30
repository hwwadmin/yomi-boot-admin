// Package test 系统参数接口集成测试（须先启动 yomi-admin-backend 服务）
package test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"gitee.com/hq_project/yomi-boot/component/common/testclient"
)

// paramItem 是参数列表项
type paramItem struct {
	ID      string `json:"id"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Type    string `json:"valueType"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Updated int64  `json:"updateTime"`
}

// findParamByKey 在分页结果中按 key 查找参数
func findParamByKey(t *testing.T, token, key string) *paramItem {
	t.Helper()
	body, ok := testclient.Get(API("/system/param?page=1&pageSize=100&keyword="+key), token)
	if !ok {
		t.Fatalf("参数列表接口不可达: %s", body)
	}
	for _, item := range parsePage[paramItem](t, body).Items {
		if item.Key == key {
			return &item
		}
	}
	return nil
}

// TestSysParamFlow 覆盖参数 CRUD、启停、唯一键与取值校验；读取直查数据库，保存即生效
func TestSysParamFlow(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)
	suffix := time.Now().UnixNano()
	paramKey := fmt.Sprintf("it.test.auto_cancel_days_%d", suffix)

	t.Cleanup(func() {
		if item := findParamByKey(t, adminToken, paramKey); item != nil {
			_, _ = testclient.Post(API("/system/param/delete"), adminToken, map[string]any{"id": item.ID})
		}
	})

	// 创建：启用态落库
	body, ok := testclient.Post(API("/system/param/create"), adminToken, map[string]any{
		"key":       paramKey,
		"value":     "7",
		"valueType": "INT",
		"name":      "集成测试参数",
		"remark":    "集成测试临时参数",
	})
	if !ok {
		t.Fatalf("创建参数接口不可达: %s", body)
	}
	created := parseSuccess(t, body)
	var param struct {
		ID         string `json:"id"`
		Key        string `json:"key"`
		Status     string `json:"status"`
		CreateTime int64  `json:"createTime"`
		UpdateTime int64  `json:"updateTime"`
	}
	if err := json.Unmarshal(created.Data, &param); err != nil || param.ID == "" {
		t.Fatalf("解析创建结果失败: %v body=%s", err, body)
	}
	if param.Status != "ENABLED" {
		t.Fatalf("新建参数状态 = %s, want ENABLED", param.Status)
	}
	// 创建响应的时间戳必须按契约给出真实值，不能是 0
	if param.CreateTime <= 0 || param.UpdateTime <= 0 {
		t.Fatalf("创建响应时间戳缺失: createTime=%d updateTime=%d", param.CreateTime, param.UpdateTime)
	}

	// 重复键被拒绝
	body, _ = testclient.Post(API("/system/param/create"), adminToken, map[string]any{
		"key": paramKey, "value": "8", "valueType": "INT", "name": "重复键",
	})
	assertFailure(t, body)

	// 非法取值被拒绝（INT 参数不接受布尔值）
	body, _ = testclient.Post(API("/system/param/create"), adminToken, map[string]any{
		"key": fmt.Sprintf("it.test.bad_%d", suffix), "value": "true", "valueType": "INT", "name": "非法值",
	})
	assertFailure(t, body)

	// 详情
	body, ok = testclient.Get(API("/system/param/detail?id="+param.ID), adminToken)
	if !ok {
		t.Fatalf("参数详情接口不可达: %s", body)
	}
	detail := parseSuccess(t, body)
	var got paramItem
	if err := json.Unmarshal(detail.Data, &got); err != nil || got.Key != paramKey || got.Value != "7" {
		t.Fatalf("详情不符: %v body=%s", err, body)
	}

	// 更新：只改取值与资料，键与类型不可变
	body, ok = testclient.Post(API("/system/param/update"), adminToken, map[string]any{
		"id": param.ID, "value": "14", "name": "集成测试参数（更新）", "remark": "更新后的备注",
	})
	if !ok {
		t.Fatalf("更新参数接口不可达: %s", body)
	}
	parseSuccess(t, body)
	updated := findParamByKey(t, adminToken, paramKey)
	if updated == nil {
		t.Fatal("更新后查不到参数")
	}
	if updated.Value != "14" || updated.Type != "INT" || updated.Key != paramKey {
		t.Fatalf("更新后不符: %+v", updated)
	}

	// 禁用后读取过滤生效
	body, ok = testclient.Post(API("/system/param/set-status"), adminToken, map[string]any{
		"id": param.ID, "status": "DISABLED",
	})
	if !ok {
		t.Fatalf("参数启停接口不可达: %s", body)
	}
	parseSuccess(t, body)
	disabled := findParamByKey(t, adminToken, paramKey)
	if disabled == nil || disabled.Status != "DISABLED" {
		t.Fatalf("禁用未生效: %+v", disabled)
	}
	body, ok = testclient.Get(API("/system/param?status=ENABLED&page=1&pageSize=100&keyword="+paramKey), adminToken)
	if !ok {
		t.Fatalf("参数列表接口不可达: %s", body)
	}
	for _, item := range parsePage[paramItem](t, body).Items {
		if item.Key == paramKey {
			t.Fatal("禁用参数仍出现在启用过滤结果中")
		}
	}

	// 小写状态同样可用：query 与请求体的枚举口径一致，不应报参数不合法
	body, ok = testclient.Get(API("/system/param?status=enabled&valueType=int&page=1&pageSize=100&keyword="+paramKey), adminToken)
	if !ok {
		t.Fatalf("小写过滤参数接口不可达: %s", body)
	}
	if len(parsePage[paramItem](t, body).Items) != 0 {
		t.Fatal("小写 status/valueType 过滤未生效")
	}

	// 删除（物理删除）→ 查不到
	body, ok = testclient.Post(API("/system/param/delete"), adminToken, map[string]any{"id": param.ID})
	if !ok {
		t.Fatalf("删除参数接口不可达: %s", body)
	}
	parseSuccess(t, body)
	if findParamByKey(t, adminToken, paramKey) != nil {
		t.Fatal("删除后仍能查到参数")
	}
}

// TestSysParamValidation 覆盖键格式、名称与值类型校验
func TestSysParamValidation(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)
	suffix := time.Now().UnixNano()

	cases := []struct {
		name string
		body map[string]any
	}{
		{"非法键格式", map[string]any{"key": "UPPER.Case", "value": "1", "valueType": "INT", "name": "非法键"}},
		{"布尔取值非法", map[string]any{"key": fmt.Sprintf("it.test.bool_%d", suffix), "value": "yes", "valueType": "BOOL", "name": "非法值"}},
		{"JSON 非法", map[string]any{"key": fmt.Sprintf("it.test.json_%d", suffix), "value": "{not-json", "valueType": "JSON", "name": "非法 JSON"}},
		{"名称缺失", map[string]any{"key": fmt.Sprintf("it.test.name_%d", suffix), "value": "1", "valueType": "INT", "name": ""}},
	}
	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			body, _ := testclient.Post(API("/system/param/create"), adminToken, tt.body)
			assertFailure(t, body)
		})
	}
}
