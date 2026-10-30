package service

import (
	"context"
	"testing"
	"yomi-admin-backend/domains/sysparam/pkg/shared/entity"
	"yomi-admin-backend/domains/sysparam/pkg/shared/enums"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

const (
	keyBool   = "a.enabled"
	keyInt    = "b.retries"
	keyString = "c.greeting"
	keyJSON   = "d.config"
)

// readerTestParams 构造覆盖四种类型的启用参数
func readerTestParams() []*entity.Param {
	return []*entity.Param{
		{ID: 1, Key: keyBool, Value: "true", ValueType: enums.ParamTypeBool, Status: enums.StatusEnabled},
		{ID: 2, Key: keyInt, Value: "3", ValueType: enums.ParamTypeInt, Status: enums.StatusEnabled},
		{ID: 3, Key: keyString, Value: "你好", ValueType: enums.ParamTypeString, Status: enums.StatusEnabled},
		{ID: 4, Key: keyJSON, Value: `{"max":10}`, ValueType: enums.ParamTypeJSON, Status: enums.StatusEnabled},
	}
}

func newTestReader(t *testing.T, params []*entity.Param) *SysParamReaderService {
	t.Helper()
	return NewSysParamReaderService(&fakeRepo{rows: params})
}

func TestReaderTypedGetters(t *testing.T) {
	reader := newTestReader(t, readerTestParams())
	ctx := context.Background()

	if v := reader.GetBool(ctx, keyBool, false); v != true {
		t.Fatalf("GetBool = %v, want true", v)
	}
	if v := reader.GetInt64(ctx, keyInt, 0); v != 3 {
		t.Fatalf("GetInt64 = %d, want 3", v)
	}
	if v := reader.GetString(ctx, keyString, "x"); v != "你好" {
		t.Fatalf("GetString = %q, want 你好", v)
	}
	var cfg struct {
		Max int `json:"max"`
	}
	if !reader.GetJSON(ctx, keyJSON, &cfg) || cfg.Max != 10 {
		t.Fatalf("GetJSON ok=true max=%d", cfg.Max)
	}
	if !reader.Exists(ctx, keyBool) || reader.Exists(ctx, "missing.key") {
		t.Fatal("Exists 语义不符")
	}
}

func TestReaderFallbackSemantics(t *testing.T) {
	params := readerTestParams()
	params = append(params,
		&entity.Param{ID: 5, Key: "off.enabled", Value: "true", ValueType: enums.ParamTypeBool, Status: enums.StatusDisabled},
	)
	reader := newTestReader(t, params)
	ctx := context.Background()

	// 不存在的键 → fallback
	if v := reader.GetBool(ctx, "missing.key", true); v != true {
		t.Fatalf("GetBool(缺失) = %v, want fallback", v)
	}
	if v := reader.GetInt64(ctx, "missing.key", 42); v != 42 {
		t.Fatalf("GetInt64(缺失) = %v, want fallback", v)
	}
	if v := reader.GetString(ctx, "missing.key", "def"); v != "def" {
		t.Fatalf("GetString(缺失) = %v, want fallback", v)
	}
	var out map[string]any
	if reader.GetJSON(ctx, "missing.key", &out) {
		t.Fatal("GetJSON(缺失) 应返回 false")
	}

	// 类型不匹配 → fallback
	if v := reader.GetBool(ctx, keyInt, true); v != true {
		t.Fatalf("GetBool(类型不符) = %v, want fallback", v)
	}
	if v := reader.GetInt64(ctx, keyString, 7); v != 7 {
		t.Fatalf("GetInt64(类型不符) = %v, want fallback", v)
	}

	// 禁用参数 → fallback
	if v := reader.GetBool(ctx, "off.enabled", false); v != false {
		t.Fatalf("GetBool(禁用) = %v, want fallback", v)
	}
	if reader.Exists(ctx, "off.enabled") {
		t.Fatal("禁用参数 Exists 应为 false")
	}
}

func TestReaderDegradesOnDBError(t *testing.T) {
	// 数据库查询报错：全部回退 fallback，不返回错误
	reader := NewSysParamReaderService(&fakeRepo{rows: readerTestParams(), getErr: errors.New("db down")})
	ctx := context.Background()

	if v := reader.GetBool(ctx, keyBool, true); v != true {
		t.Fatalf("GetBool = %v, want fallback", v)
	}
	if v := reader.GetInt64(ctx, keyInt, 9); v != 9 {
		t.Fatalf("GetInt64 = %v, want fallback", v)
	}
	if reader.Exists(ctx, keyBool) {
		t.Fatal("Exists 应为 false")
	}
}
