// Package service 系统参数读取服务实现：业务读取的唯一入口
package service

import (
	"context"
	"encoding/json"
	"strconv"
	"sync/atomic"
	"time"
	"yomi-admin-backend/domains/sysparam/internal/repo_contract"
	"yomi-admin-backend/domains/sysparam/pkg/shared/entity"
	"yomi-admin-backend/domains/sysparam/pkg/shared/enums"
	"yomi-admin-backend/domains/sysparam/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

const (
	// fallbackWarnInterval 是降级告警的最小间隔，避免数据库故障期间刷日志
	fallbackWarnInterval = 30 * time.Second
)

// SysParamReaderService 类型化参数读取实现。
//
// 语义：读参数永不返回错误、永不阻塞业务——参数不存在、被禁用、被删除，
// 甚至数据库故障，都返回调用方自带的 fallback（查询失败限流告警）。
// 读取直接查库：参数保存即对业务生效。
type SysParamReaderService struct {
	repo repo_contract.SysParamRepo
	// lastWarn 记录上次降级告警的时间（Unix 纳秒）
	lastWarn atomic.Int64
	now      func() time.Time
}

var _ service_contract.SysParamReader = (*SysParamReaderService)(nil)

// NewSysParamReaderService 创建参数读取服务
func NewSysParamReaderService(repo repo_contract.SysParamRepo) *SysParamReaderService {
	return &SysParamReaderService{
		repo: repo,
		now:  time.Now,
	}
}

// GetBool 读取布尔参数；未配置/禁用/删除/故障返回 fallback
func (s *SysParamReaderService) GetBool(ctx context.Context, key string, fallback bool) bool {
	param, ok := s.getParam(ctx, key)
	if !ok {
		return fallback
	}
	if param.ValueType != enums.ParamTypeBool {
		s.warnValueType(ctx, key, param.ValueType, enums.ParamTypeBool)
		return fallback
	}
	switch param.Value {
	case "true":
		return true
	case "false":
		return false
	default:
		s.warnCorrupt(ctx, key)
		return fallback
	}
}

// GetInt64 读取整数参数；未配置/禁用/删除/故障返回 fallback
func (s *SysParamReaderService) GetInt64(ctx context.Context, key string, fallback int64) int64 {
	param, ok := s.getParam(ctx, key)
	if !ok {
		return fallback
	}
	if param.ValueType != enums.ParamTypeInt {
		s.warnValueType(ctx, key, param.ValueType, enums.ParamTypeInt)
		return fallback
	}
	parsed, err := strconv.ParseInt(param.Value, 10, 64)
	if err != nil {
		s.warnCorrupt(ctx, key)
		return fallback
	}
	return parsed
}

// GetString 读取字符串参数；未配置/禁用/删除/故障返回 fallback
func (s *SysParamReaderService) GetString(ctx context.Context, key, fallback string) string {
	param, ok := s.getParam(ctx, key)
	if !ok {
		return fallback
	}
	if param.ValueType != enums.ParamTypeString {
		s.warnValueType(ctx, key, param.ValueType, enums.ParamTypeString)
		return fallback
	}
	return param.Value
}

// GetJSON 读取 JSON 参数并反序列化到 out；命中且反序列化成功返回 true，
// 未配置/禁用/删除/损坏返回 false（out 保持原值）
func (s *SysParamReaderService) GetJSON(ctx context.Context, key string, out any) bool {
	param, ok := s.getParam(ctx, key)
	if !ok {
		return false
	}
	if param.ValueType != enums.ParamTypeJSON {
		s.warnValueType(ctx, key, param.ValueType, enums.ParamTypeJSON)
		return false
	}
	if err := json.Unmarshal([]byte(param.Value), out); err != nil {
		s.warnCorrupt(ctx, key)
		return false
	}
	return true
}

// Exists 判断参数是否已配置且启用（不看类型与取值）
func (s *SysParamReaderService) Exists(ctx context.Context, key string) bool {
	_, ok := s.getParam(ctx, key)
	return ok
}

// getParam 读取启用参数：单键直查数据库。
// 返回 ok=false 表示参数未配置、已禁用或数据库不可用，调用方使用 fallback。
func (s *SysParamReaderService) getParam(ctx context.Context, key string) (*entity.Param, bool) {
	param, found, err := s.repo.GetEnabledByKey(ctx, key)
	if err != nil {
		// 数据库不可用：回退读方默认值，业务不受阻
		s.warnFallback(ctx, "查询系统参数失败", err)
		return nil, false
	}
	if !found {
		return nil, false
	}
	return param, true
}

// warnFallback 按固定间隔输出查询失败告警，防止故障期间日志被单个请求刷满
func (s *SysParamReaderService) warnFallback(ctx context.Context, msg string, err error) {
	now := s.now().UnixNano()
	last := s.lastWarn.Load()
	if now-last < int64(fallbackWarnInterval) {
		return
	}
	if !s.lastWarn.CompareAndSwap(last, now) {
		return
	}
	logger.Warn(ctx, "[SysParam] "+msg, "error", err)
}

// warnValueType 输出类型不匹配告警：读取方期望的类型与参数定义不一致
func (s *SysParamReaderService) warnValueType(ctx context.Context, key string, actual, expected enums.ParamValueType) {
	logger.Warn(ctx, "[SysParam] 参数值类型与读取方式不匹配，返回默认值",
		"key", key, "valueType", actual, "expected", expected)
}

// warnCorrupt 输出取值损坏告警：写入校验理论已拦截，防御脏数据
func (s *SysParamReaderService) warnCorrupt(ctx context.Context, key string) {
	logger.Warn(ctx, "[SysParam] 参数取值损坏，返回默认值", "key", key)
}
