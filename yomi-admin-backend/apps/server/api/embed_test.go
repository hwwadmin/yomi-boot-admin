// Package api 内嵌契约文档测试
package api

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestSpecIsValidOpenAPI31 校验内嵌 openapi.yaml 是合法的 OpenAPI 3.1 文档
func TestSpecIsValidOpenAPI31(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(Spec)
	if err != nil {
		t.Fatalf("openapi.yaml 解析失败: %v", err)
	}
	if doc.OpenAPI != "3.1.0" {
		t.Fatalf("openapi 版本应为 3.1.0，实际 %q", doc.OpenAPI)
	}
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("openapi.yaml 校验失败: %v", err)
	}
	if doc.Paths == nil || doc.Paths.Len() == 0 {
		t.Fatal("openapi.yaml 未包含任何路径")
	}
}
