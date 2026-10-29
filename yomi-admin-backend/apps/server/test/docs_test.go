// API 文档集成测试（须先启动 yomi-admin-backend 服务，且 biz.docs.enabled 开启）
package test

import (
	"strings"
	"testing"

	"gitee.com/hq_project/yomi-boot/component/common/testclient"
)

func TestApiDocSchema(t *testing.T) {
	requireServer(t)

	body, ok := testclient.Get(API("/system/api-doc/schema"), "")
	if !ok {
		t.Fatalf("API 文档 schema 接口不可达: %s", body)
	}
	if !strings.Contains(body, "openapi: 3.1.0") {
		t.Fatalf("schema 不是 OpenAPI 3.1: %.200s", body)
	}
	if !strings.Contains(body, "/api/system/user") {
		t.Fatalf("schema 缺少业务路由: %.200s", body)
	}
}

func TestApiDocPage(t *testing.T) {
	requireServer(t)

	body, ok := testclient.Get(API("/system/api-doc"), "")
	if !ok {
		t.Fatalf("API 文档页面不可达: %s", body)
	}
	if !strings.Contains(body, "swagger-ui-bundle.js") {
		t.Fatalf("文档页面缺少 Swagger UI 资源引用: %.200s", body)
	}
}
