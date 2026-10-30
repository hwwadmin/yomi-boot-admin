package repo

import (
	"context"
	"testing"
	"yomi-admin-backend/domains/sysparam/internal/repo_contract"
	"yomi-admin-backend/domains/sysparam/pkg/shared/entity"
	"yomi-admin-backend/domains/sysparam/pkg/shared/enums"

	"gitee.com/hq_project/yomi-boot/component/common/db/dbtest"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	sharedenums "gitee.com/hq_project/yomi-boot/component/core/shared/enums"
)

func newTestRepo(t *testing.T) (*SysParamRepo, context.Context) {
	t.Helper()
	gormDB, cleanup := dbtest.New(t, dbtest.WithTables(&sysParamModel{}))
	t.Cleanup(cleanup)
	return NewSysParamRepo(gormDB, sharedenums.MySQL), context.Background()
}

// mustCreate 写入一条参数供后续断言使用
func mustCreate(t *testing.T, r *SysParamRepo, ctx context.Context, id int64, key, value string, valueType enums.ParamValueType, status enums.Status) *entity.Param {
	t.Helper()
	param := &entity.Param{
		ID:        id,
		Key:       key,
		Value:     value,
		ValueType: valueType,
		Name:      "参数" + key,
		Status:    status,
	}
	if err := r.CreateParam(ctx, param, 100); err != nil {
		t.Fatalf("CreateParam(%s) error = %v", key, err)
	}
	if param.CreateTime == 0 || param.UpdateTime == 0 {
		t.Fatalf("CreateParam(%s) 未回填时间戳: createTime=%d updateTime=%d", key, param.CreateTime, param.UpdateTime)
	}
	return param
}

func TestSysParamRepoCreateAndFind(t *testing.T) {
	r, ctx := newTestRepo(t)

	created := mustCreate(t, r, ctx, 1, "order.auto_cancel_days", "7", enums.ParamTypeInt, enums.StatusEnabled)

	found, ok, err := r.FindByID(ctx, created.ID)
	if err != nil || !ok {
		t.Fatalf("FindByID() ok=%v err=%v", ok, err)
	}
	if found.Key != "order.auto_cancel_days" || found.Value != "7" || found.ValueType != enums.ParamTypeInt {
		t.Fatalf("found = %+v", found)
	}
	if found.CreateTime == 0 {
		t.Fatal("CreateTime 未回填")
	}

	byKey, ok, err := r.FindByKey(ctx, "order.auto_cancel_days")
	if err != nil || !ok || byKey.ID != created.ID {
		t.Fatalf("FindByKey() ok=%v err=%v byKey=%+v", ok, err, byKey)
	}

	// 同键重复创建应返回唯一键冲突
	duplicate := *created
	duplicate.ID = 2
	if err := r.CreateParam(ctx, &duplicate, 100); !errors.Is(err, repo_contract.ErrDuplicatedKey) {
		t.Fatalf("重复 CreateParam() error = %v, want ErrDuplicatedKey", err)
	}
}

func TestSysParamRepoPage(t *testing.T) {
	r, ctx := newTestRepo(t)

	// 故意让键序与 id 序相反：只有按 id 倒序，最后创建的 a.enabled 才会排最前
	mustCreate(t, r, ctx, 1, "c.greeting", "你好", enums.ParamTypeString, enums.StatusEnabled)
	mustCreate(t, r, ctx, 2, "b.retries", "3", enums.ParamTypeInt, enums.StatusDisabled)
	mustCreate(t, r, ctx, 3, "a.enabled", "true", enums.ParamTypeBool, enums.StatusEnabled)

	tests := []struct {
		name   string
		filter repo_contract.ParamFilter
		want   int64
	}{
		{name: "全部", filter: repo_contract.ParamFilter{Page: 1, PageSize: 10}, want: 3},
		{name: "关键字命中键", filter: repo_contract.ParamFilter{Keyword: "a.", Page: 1, PageSize: 10}, want: 1},
		{name: "关键字命中名称", filter: repo_contract.ParamFilter{Keyword: "参数b", Page: 1, PageSize: 10}, want: 1},
		{name: "状态过滤", filter: repo_contract.ParamFilter{Status: enums.StatusDisabled, Page: 1, PageSize: 10}, want: 1},
		{name: "类型过滤", filter: repo_contract.ParamFilter{ValueType: enums.ParamTypeBool, Page: 1, PageSize: 10}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := r.Page(ctx, &tt.filter)
			if err != nil {
				t.Fatalf("Page() error = %v", err)
			}
			if total != tt.want || int64(len(items)) != tt.want {
				t.Fatalf("total=%d items=%d, want %d", total, len(items), tt.want)
			}
		})
	}

	// 按主键倒序：id 3 → id 1，即最后创建的排最前
	items, _, err := r.Page(ctx, &repo_contract.ParamFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("Page() error = %v", err)
	}
	if items[0].Key != "a.enabled" || items[2].Key != "c.greeting" {
		t.Fatalf("按主键倒序失败: %s, %s, %s", items[0].Key, items[1].Key, items[2].Key)
	}
}

func TestSysParamRepoEnabledQueries(t *testing.T) {
	r, ctx := newTestRepo(t)

	mustCreate(t, r, ctx, 1, "on.enabled", "true", enums.ParamTypeBool, enums.StatusEnabled)
	mustCreate(t, r, ctx, 2, "off.enabled", "false", enums.ParamTypeBool, enums.StatusDisabled)

	// GetEnabledByKey 只返回启用参数
	if _, ok, err := r.GetEnabledByKey(ctx, "on.enabled"); err != nil || !ok {
		t.Fatalf("GetEnabledByKey(启用) ok=%v err=%v", ok, err)
	}
	if _, ok, err := r.GetEnabledByKey(ctx, "off.enabled"); err != nil || ok {
		t.Fatalf("GetEnabledByKey(禁用) ok=%v err=%v, want 不返回", ok, err)
	}
}

func TestSysParamRepoFindEnabledByKeys(t *testing.T) {
	r, ctx := newTestRepo(t)

	mustCreate(t, r, ctx, 1, "on.enabled", "true", enums.ParamTypeBool, enums.StatusEnabled)
	mustCreate(t, r, ctx, 2, "off.enabled", "false", enums.ParamTypeBool, enums.StatusDisabled)
	mustCreate(t, r, ctx, 3, "other.key", "1", enums.ParamTypeInt, enums.StatusEnabled)

	// 只返回启用项；禁用项与不存在的键都不出现在结果里
	items, err := r.FindEnabledByKeys(ctx, []string{"on.enabled", "off.enabled", "other.key", "missing.key"})
	if err != nil {
		t.Fatalf("FindEnabledByKeys() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("FindEnabledByKeys() = %+v, want 2 条启用项", items)
	}
	for _, item := range items {
		if item.Key == "off.enabled" {
			t.Fatal("禁用参数不应出现在批量结果中")
		}
	}

	// 空键列表不查库，返回空结果
	empty, err := r.FindEnabledByKeys(ctx, nil)
	if err != nil || len(empty) != 0 {
		t.Fatalf("FindEnabledByKeys(nil) = %+v, err = %v", empty, err)
	}
}

func TestSysParamRepoWriteOperations(t *testing.T) {
	r, ctx := newTestRepo(t)

	param := mustCreate(t, r, ctx, 1, "ui.banner", "hello", enums.ParamTypeString, enums.StatusEnabled)

	// 更新取值与资料
	param.Value = "welcome"
	param.Name = "横幅文案"
	param.Remark = "首页横幅"
	if err := r.UpdateParam(ctx, param, 200); err != nil {
		t.Fatalf("UpdateParam() error = %v", err)
	}
	found, _, err := r.FindByID(ctx, param.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.Value != "welcome" || found.Name != "横幅文案" || found.Remark != "首页横幅" {
		t.Fatalf("更新未生效: %+v", found)
	}

	// 启停
	if err := r.UpdateParamStatus(ctx, param.ID, enums.StatusDisabled, 200); err != nil {
		t.Fatalf("UpdateParamStatus() error = %v", err)
	}
	if _, ok, _ := r.GetEnabledByKey(ctx, param.Key); ok {
		t.Fatal("禁用后 GetEnabledByKey 仍返回参数")
	}

	// 物理删除
	if err := r.DeleteParam(ctx, param.ID); err != nil {
		t.Fatalf("DeleteParam() error = %v", err)
	}
	if _, ok, err := r.FindByID(ctx, param.ID); err != nil || ok {
		t.Fatalf("删除后 FindByID ok=%v err=%v, want 不返回", ok, err)
	}
}
