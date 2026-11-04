// Package service 系统参数管理服务实现
package service

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
	"yomi-admin-backend/domains/sysparam/internal/repo_contract"
	"yomi-admin-backend/domains/sysparam/pkg/shared/entity"
	"yomi-admin-backend/domains/sysparam/pkg/shared/enums"
	"yomi-admin-backend/domains/sysparam/pkg/shared/input"
	"yomi-admin-backend/domains/sysparam/pkg/shared/output"
	"yomi-admin-backend/domains/sysparam/service_contract"

	"github.com/hwwadmin/yomi-boot/component/common/snowflake"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

const (
	maxKeyLength   = 128
	maxValueRunes  = 2048
	maxNameRunes   = 64
	maxRemarkRunes = 255
	boolValueTrue  = "true"
	boolValueFalse = "false"
)

// keyPattern 约束参数键为小写点分标识，每段以小写字母开头（如 order.auto_cancel_days）
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)

// SysParamAdminService 参数管理实现
type SysParamAdminService struct {
	repo repo_contract.SysParamRepo
	ids  snowflake.Generator
}

var _ service_contract.SysParamAdminService = (*SysParamAdminService)(nil)

// NewSysParamAdminService 创建参数管理服务
func NewSysParamAdminService(repo repo_contract.SysParamRepo, ids snowflake.Generator) *SysParamAdminService {
	return &SysParamAdminService{repo: repo, ids: ids}
}

// PageParams 分页查询参数
func (s *SysParamAdminService) PageParams(ctx context.Context, in *input.PageParamsInput) (*output.PageParamsOutput, error) {
	if in == nil {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.Status != "" && !enums.IsValidStatus(in.Status) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.ValueType != "" && !enums.IsValidParamValueType(in.ValueType) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	items, total, err := s.repo.Page(ctx, &repo_contract.ParamFilter{
		Keyword:   in.Keyword,
		Status:    in.Status,
		ValueType: in.ValueType,
		Page:      in.Page,
		PageSize:  in.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return &output.PageParamsOutput{Items: items, Total: total}, nil
}

// GetParam 查询参数详情
func (s *SysParamAdminService) GetParam(ctx context.Context, id int64) (*entity.Param, error) {
	return s.mustFindParam(ctx, id)
}

// CreateParam 创建参数；键全局唯一且创建后不可变
func (s *SysParamAdminService) CreateParam(ctx context.Context, in *input.CreateParamInput) (*entity.Param, error) {
	if in == nil {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	key := strings.TrimSpace(in.Key)
	value := in.Value
	name := strings.TrimSpace(in.Name)
	remark := strings.TrimSpace(in.Remark)
	if err := validateKey(key); err != nil {
		return nil, err
	}
	if err := validateValue(value, in.ValueType); err != nil {
		return nil, err
	}
	if !validName(name) || !validRemark(remark) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	if _, found, err := s.repo.FindByKey(ctx, key); err != nil {
		return nil, err
	} else if found {
		return nil, errors.Sentinel(service_contract.ErrParamKeyExists)
	}

	id, err := s.ids.NextID()
	if err != nil {
		return nil, errors.Wrap(err, "生成参数 ID 失败")
	}
	param := &entity.Param{
		ID:        id,
		Key:       key,
		Value:     value,
		ValueType: in.ValueType,
		Name:      name,
		Remark:    remark,
		Status:    enums.StatusEnabled,
	}
	if err := s.repo.CreateParam(ctx, param, in.OperatorID); err != nil {
		if errors.Is(err, repo_contract.ErrDuplicatedKey) {
			return nil, errors.Sentinel(service_contract.ErrParamKeyExists)
		}
		return nil, err
	}
	return param, nil
}

// UpdateParam 更新参数值、名称与备注；键与值类型不可变
func (s *SysParamAdminService) UpdateParam(ctx context.Context, in *input.UpdateParamInput) error {
	if in == nil {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	param, err := s.mustFindParam(ctx, in.ID)
	if err != nil {
		return err
	}

	value := in.Value
	name := strings.TrimSpace(in.Name)
	remark := strings.TrimSpace(in.Remark)
	if err := validateValue(value, param.ValueType); err != nil {
		return err
	}
	if !validName(name) || !validRemark(remark) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	param.Value = value
	param.Name = name
	param.Remark = remark
	return s.repo.UpdateParam(ctx, param, in.OperatorID)
}

// SetParamStatus 设置参数启停；禁用即覆盖值失效，读取方回退默认值
func (s *SysParamAdminService) SetParamStatus(ctx context.Context, in *input.SetParamStatusInput) error {
	if in == nil {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if !enums.IsValidStatus(in.Status) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	if _, err := s.mustFindParam(ctx, in.ID); err != nil {
		return err
	}
	return s.repo.UpdateParamStatus(ctx, in.ID, in.Status, in.OperatorID)
}

// DeleteParam 物理删除参数；删除后读取方回退默认值
func (s *SysParamAdminService) DeleteParam(ctx context.Context, in *input.DeleteParamInput) error {
	if in == nil {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	if _, err := s.mustFindParam(ctx, in.ID); err != nil {
		return err
	}
	return s.repo.DeleteParam(ctx, in.ID)
}

// mustFindParam 按主键查询参数，不存在时返回业务错误
func (s *SysParamAdminService) mustFindParam(ctx context.Context, id int64) (*entity.Param, error) {
	if id <= 0 {
		return nil, errors.Sentinel(service_contract.ErrParamNotFound)
	}
	param, found, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.Sentinel(service_contract.ErrParamNotFound)
	}
	return param, nil
}

// validateKey 校验参数键格式与长度
func validateKey(key string) error {
	if key == "" || len(key) > maxKeyLength || !keyPattern.MatchString(key) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	return nil
}

// validateValue 按值类型校验取值与长度；类型未知视为非法
func validateValue(value string, valueType enums.ParamValueType) error {
	if !enums.IsValidParamValueType(valueType) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if utf8.RuneCountInString(value) > maxValueRunes {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	switch valueType {
	case enums.ParamTypeBool:
		if value != boolValueTrue && value != boolValueFalse {
			return errors.Sentinel(service_contract.ErrInvalidInput)
		}
	case enums.ParamTypeInt:
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return errors.Sentinel(service_contract.ErrInvalidInput)
		}
	case enums.ParamTypeJSON:
		if !json.Valid([]byte(value)) {
			return errors.Sentinel(service_contract.ErrInvalidInput)
		}
	}
	return nil
}

// validName 校验参数名称：1..64 字符
func validName(name string) bool {
	runes := utf8.RuneCountInString(name)
	return runes > 0 && runes <= maxNameRunes
}

// validRemark 校验备注：不超过 255 字符
func validRemark(remark string) bool {
	return utf8.RuneCountInString(remark) <= maxRemarkRunes
}
