package service

import (
	"context"
	"testing"
	"yomi-admin-backend/domains/sysparam/pkg/shared/enums"
	"yomi-admin-backend/domains/sysparam/pkg/shared/input"
	"yomi-admin-backend/domains/sysparam/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

func newAdminService(t *testing.T) (*SysParamAdminService, *fakeRepo) {
	t.Helper()
	repo := &fakeRepo{}
	return NewSysParamAdminService(repo, &newTestIDs{}), repo
}

func TestSysParamAdminServiceCreateParam(t *testing.T) {
	svc, repo := newAdminService(t)
	ctx := context.Background()

	created, err := svc.CreateParam(ctx, &input.CreateParamInput{
		Key:       "order.auto_cancel_days",
		Value:     "7",
		ValueType: enums.ParamTypeInt,
		Name:      "自动取消天数",
		Remark:    "超时未支付订单自动取消",
	})
	if err != nil {
		t.Fatalf("CreateParam() error = %v", err)
	}
	if created.ID == 0 || created.Status != enums.StatusEnabled || created.ValueType != enums.ParamTypeInt {
		t.Fatalf("created = %+v", created)
	}
	if len(repo.rows) != 1 {
		t.Fatalf("仓储未写入: %+v", repo.rows)
	}

	// 重复键
	if _, err := svc.CreateParam(ctx, &input.CreateParamInput{
		Key: "order.auto_cancel_days", Value: "8", ValueType: enums.ParamTypeInt, Name: "重复",
	}); !errors.Is(err, service_contract.ErrParamKeyExists) {
		t.Fatalf("重复键 error = %v, want ErrParamKeyExists", err)
	}

	// 仓储层唯一键冲突同样归一
	repo.dupOnNew = true
	if _, err := svc.CreateParam(ctx, &input.CreateParamInput{
		Key: "other.key", Value: "1", ValueType: enums.ParamTypeInt, Name: "冲突",
	}); !errors.Is(err, service_contract.ErrParamKeyExists) {
		t.Fatalf("仓储冲突 error = %v, want ErrParamKeyExists", err)
	}

	// 非法键格式
	invalidKeys := []string{"", "UPPER.CASE", "1start.with.digit", "中文.key", "trailing.dot.", "a..b", longString(129)}
	for _, key := range invalidKeys {
		if _, err := svc.CreateParam(ctx, &input.CreateParamInput{
			Key: key, Value: "1", ValueType: enums.ParamTypeInt, Name: "非法键",
		}); !errors.Is(err, service_contract.ErrInvalidInput) {
			t.Fatalf("非法键 %q error = %v, want ErrInvalidInput", key, err)
		}
	}

	// 非法取值
	invalidValues := []struct {
		value     string
		valueType enums.ParamValueType
	}{
		{"yes", enums.ParamTypeBool},
		{"3.14", enums.ParamTypeInt},
		{"{not-json", enums.ParamTypeJSON},
		{longString(2049), enums.ParamTypeString},
		{"1", "UNKNOWN"},
	}
	for _, tt := range invalidValues {
		if _, err := svc.CreateParam(ctx, &input.CreateParamInput{
			Key: "bad.value", Value: tt.value, ValueType: tt.valueType, Name: "非法值",
		}); !errors.Is(err, service_contract.ErrInvalidInput) {
			t.Fatalf("非法值 %q/%s error = %v, want ErrInvalidInput", tt.value, tt.valueType, err)
		}
	}

	// 名称缺失
	if _, err := svc.CreateParam(ctx, &input.CreateParamInput{
		Key: "no.name", Value: "1", ValueType: enums.ParamTypeInt,
	}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("缺名称 error = %v, want ErrInvalidInput", err)
	}
}

func TestSysParamAdminServiceUpdateParam(t *testing.T) {
	svc, repo := newAdminService(t)
	ctx := context.Background()

	created, err := svc.CreateParam(ctx, &input.CreateParamInput{
		Key: "ui.banner", Value: "hello", ValueType: enums.ParamTypeString, Name: "横幅",
	})
	if err != nil {
		t.Fatalf("CreateParam() error = %v", err)
	}

	// 正常更新
	if err := svc.UpdateParam(ctx, &input.UpdateParamInput{
		ID: created.ID, Value: "welcome", Name: "横幅文案", Remark: "首页",
	}); err != nil {
		t.Fatalf("UpdateParam() error = %v", err)
	}
	if repo.rows[0].Value != "welcome" || repo.rows[0].Name != "横幅文案" {
		t.Fatalf("更新未生效: %+v", repo.rows[0])
	}
	// 键与类型不可变
	if repo.rows[0].Key != "ui.banner" || repo.rows[0].ValueType != enums.ParamTypeString {
		t.Fatalf("键或类型被改动: %+v", repo.rows[0])
	}

	// 值按参数既有类型校验：INT 参数不接受布尔取值
	intParam, err := svc.CreateParam(ctx, &input.CreateParamInput{
		Key: "a.count", Value: "1", ValueType: enums.ParamTypeInt, Name: "计数",
	})
	if err != nil {
		t.Fatalf("CreateParam() error = %v", err)
	}
	if err := svc.UpdateParam(ctx, &input.UpdateParamInput{
		ID: intParam.ID, Value: "true", Name: "计数",
	}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("类型不符更新 error = %v, want ErrInvalidInput", err)
	}

	// 参数不存在
	if err := svc.UpdateParam(ctx, &input.UpdateParamInput{ID: 999, Value: "v", Name: "n"}); !errors.Is(err, service_contract.ErrParamNotFound) {
		t.Fatalf("不存在更新 error = %v, want ErrParamNotFound", err)
	}
}

func TestSysParamAdminServiceStatusAndDelete(t *testing.T) {
	svc, repo := newAdminService(t)
	ctx := context.Background()

	created, err := svc.CreateParam(ctx, &input.CreateParamInput{
		Key: "a.enabled", Value: "true", ValueType: enums.ParamTypeBool, Name: "开关",
	})
	if err != nil {
		t.Fatalf("CreateParam() error = %v", err)
	}

	// 禁用
	if err := svc.SetParamStatus(ctx, &input.SetParamStatusInput{ID: created.ID, Status: enums.StatusDisabled}); err != nil {
		t.Fatalf("SetParamStatus() error = %v", err)
	}
	if repo.rows[0].Status != enums.StatusDisabled {
		t.Fatalf("禁用未生效: %+v", repo.rows[0])
	}
	// 非法状态
	if err := svc.SetParamStatus(ctx, &input.SetParamStatusInput{ID: created.ID, Status: "ON"}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法状态 error = %v, want ErrInvalidInput", err)
	}
	// 不存在
	if err := svc.SetParamStatus(ctx, &input.SetParamStatusInput{ID: 999, Status: enums.StatusEnabled}); !errors.Is(err, service_contract.ErrParamNotFound) {
		t.Fatalf("不存在启停 error = %v, want ErrParamNotFound", err)
	}

	// 删除
	if err := svc.DeleteParam(ctx, &input.DeleteParamInput{ID: created.ID}); err != nil {
		t.Fatalf("DeleteParam() error = %v", err)
	}
	if len(repo.rows) != 0 {
		t.Fatalf("删除未生效: %+v", repo.rows)
	}
	if err := svc.DeleteParam(ctx, &input.DeleteParamInput{ID: created.ID}); !errors.Is(err, service_contract.ErrParamNotFound) {
		t.Fatalf("重复删除 error = %v, want ErrParamNotFound", err)
	}
}

func TestSysParamAdminServicePageAndGet(t *testing.T) {
	svc, _ := newAdminService(t)
	ctx := context.Background()

	for _, in := range []*input.CreateParamInput{
		{Key: "a.enabled", Value: "true", ValueType: enums.ParamTypeBool, Name: "开关"},
		{Key: "b.retries", Value: "3", ValueType: enums.ParamTypeInt, Name: "重试"},
	} {
		if _, err := svc.CreateParam(ctx, in); err != nil {
			t.Fatalf("CreateParam(%s) error = %v", in.Key, err)
		}
	}

	page, err := svc.PageParams(ctx, &input.PageParamsInput{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("PageParams() error = %v", err)
	}
	if page.Total != 2 || len(page.Items) != 2 {
		t.Fatalf("page = %d/%d, want 2/2", page.Total, len(page.Items))
	}

	filtered, err := svc.PageParams(ctx, &input.PageParamsInput{ValueType: enums.ParamTypeBool, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("PageParams(过滤) error = %v", err)
	}
	if filtered.Total != 1 || filtered.Items[0].Key != "a.enabled" {
		t.Fatalf("类型过滤失败: %+v", filtered)
	}

	if _, err := svc.PageParams(ctx, &input.PageParamsInput{Status: "ON", Page: 1, PageSize: 10}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法状态过滤 error = %v, want ErrInvalidInput", err)
	}

	detail, err := svc.GetParam(ctx, page.Items[0].ID)
	if err != nil {
		t.Fatalf("GetParam() error = %v", err)
	}
	if detail.Key != page.Items[0].Key {
		t.Fatalf("详情键不符: %+v", detail)
	}
	if _, err := svc.GetParam(ctx, 999); !errors.Is(err, service_contract.ErrParamNotFound) {
		t.Fatalf("不存在详情 error = %v, want ErrParamNotFound", err)
	}
}

func TestValidateKeyPattern(t *testing.T) {
	valid := []string{"a", "a.b", "order.auto_cancel_days", "ui.banner_text", "x1.y2"}
	for _, key := range valid {
		if err := validateKey(key); err != nil {
			t.Fatalf("合法键 %q 被拒绝: %v", key, err)
		}
	}
	invalid := []string{"", "A.b", ".lead", "trail.", "a..b", "a b", longString(129)}
	for _, key := range invalid {
		if err := validateKey(key); err == nil {
			t.Fatalf("非法键 %q 被接受", key)
		}
	}
}
