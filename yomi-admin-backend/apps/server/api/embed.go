// Package api 内嵌对外契约文档。
//
// openapi.yaml 是 API 文档的唯一来源（OpenAPI 3.1，手写维护），
// 由 internal/support/apidoc 以 Swagger UI 形式对外分发。
package api

import _ "embed"

// Spec 是内嵌的 OpenAPI 3.1 文档（YAML 原文）
//
//go:embed openapi.yaml
var Spec []byte
