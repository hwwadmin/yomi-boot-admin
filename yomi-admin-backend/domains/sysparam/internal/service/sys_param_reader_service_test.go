package service

import (
	"context"
	"strconv"
	"testing"
	"time"
	"yomi-admin-backend/domains/sysparam/pkg/shared/entity"
	"yomi-admin-backend/domains/sysparam/pkg/shared/enums"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
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

func TestReaderWarningThrottledByKey(t *testing.T) {
	reader := newTestReader(t, readerTestParams())
	base := time.Unix(1000, 0)
	reader.now = func() time.Time { return base }

	calls := 0
	record := func() { calls++ }

	reader.warnByKey("a.key", record)
	reader.warnByKey("a.key", record) // 同一窗口内同键应被抑制
	if calls != 1 {
		t.Fatalf("同键窗口内应只告警 1 次，实际 %d", calls)
	}

	reader.warnByKey("b.key", record) // 不同键互不影响
	if calls != 2 {
		t.Fatalf("不同键应各自告警，实际 %d", calls)
	}

	reader.now = func() time.Time { return base.Add(fallbackWarnInterval + time.Second) }
	reader.warnByKey("a.key", record) // 窗口外同键应重新告警
	if calls != 3 {
		t.Fatalf("窗口外应重新告警，实际 %d", calls)
	}
}

func TestReaderGetJSONRequiresPointer(t *testing.T) {
	reader := newTestReader(t, readerTestParams())
	ctx := context.Background()

	var cfg struct {
		Max int `json:"max"`
	}
	// 传值而不是指针：属于调用方写错，应直接返回 false，不 panic、不动目标
	if reader.GetJSON(ctx, keyJSON, cfg) {
		t.Fatal("GetJSON(非指针) 应返回 false")
	}
	if reader.GetJSON(ctx, keyJSON, nil) {
		t.Fatal("GetJSON(nil) 应返回 false")
	}
	if !reader.GetJSON(ctx, keyJSON, &cfg) || cfg.Max != 10 {
		t.Fatalf("GetJSON(指针) 应成功，max=%d", cfg.Max)
	}
}

func TestReaderGetMany(t *testing.T) {
	reader := newTestReader(t, readerTestParams())
	ctx := context.Background()

	// 重复键、空键、未配置键混合：只返回已配置且启用的项，且不重复
	got := reader.GetMany(ctx, []string{keyBool, keyBool, "", keyInt, "missing.key"})
	if len(got) != 2 || got[keyBool] != "true" || got[keyInt] != "3" {
		t.Fatalf("GetMany = %+v", got)
	}
	if _, ok := got["missing.key"]; ok {
		t.Fatal("未配置的键不应出现在结果中")
	}
	if _, ok := got[""]; ok {
		t.Fatal("空键不应出现在结果中")
	}

	// 空列表：返回空映射，不查库
	if got := reader.GetMany(ctx, nil); len(got) != 0 {
		t.Fatalf("空键列表应返回空映射，实际 %+v", got)
	}
}

func TestReaderGetManyDegradesOnDBError(t *testing.T) {
	reader := NewSysParamReaderService(&fakeRepo{rows: readerTestParams(), getErr: errors.New("db down")})

	if got := reader.GetMany(context.Background(), []string{keyBool, keyInt}); len(got) != 0 {
		t.Fatalf("数据库故障应返回空映射，实际 %+v", got)
	}
}

func TestReaderGetManyChunksLargeKeySet(t *testing.T) {
	reader := newTestReader(t, readerTestParams())
	ctx := context.Background()

	// 超过单次上限的键集：分批后仍能取到已配置的键
	keys := make([]string, 0, maxKeysPerQuery+2)
	for i := 0; i < maxKeysPerQuery+1; i++ {
		keys = append(keys, "filler.key."+strconv.Itoa(i))
	}
	keys = append(keys, keyString)

	got := reader.GetMany(ctx, keys)
	if len(got) != 1 || got[keyString] != "你好" {
		t.Fatalf("分批读取结果 = %+v", got)
	}
}
