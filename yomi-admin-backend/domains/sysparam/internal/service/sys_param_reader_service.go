// Package service 系统参数读取服务实现：业务读取的唯一入口
package service

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"yomi-admin-backend/domains/sysparam/internal/repo_contract"
	"yomi-admin-backend/domains/sysparam/pkg/shared/entity"
	"yomi-admin-backend/domains/sysparam/pkg/shared/enums"
	"yomi-admin-backend/domains/sysparam/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/logger"
)

const (
	// fallbackWarnInterval 是同类告警按参数键去重的最小间隔，避免持续状态把日志刷满
	fallbackWarnInterval = 30 * time.Second
	// maxKeysPerQuery 是批量读取单次 IN 查询的键上限，调用方传入超长列表时自动分批
	maxKeysPerQuery = 100
)

// SysParamReaderService 类型化参数读取实现。
//
// 语义：读参数永不返回错误、永不阻塞业务——参数不存在、被禁用、被删除，
// 甚至数据库故障，都返回调用方自带的 fallback（查询失败限流告警）。
// 读取直接查库：参数保存即对业务生效。
type SysParamReaderService struct {
	repo repo_contract.SysParamRepo
	// lastWarn 记录上次数据库故障告警的时间（Unix 纳秒），失败告警全局一个窗口
	lastWarn atomic.Int64
	// keyWarns 按参数键记录上次告警时间（Unix 纳秒），供类型不符/取值损坏限流
	keyWarns sync.Map
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

// GetJSON 读取 JSON 参数并反序列化到 out；out 须为非空指针。
// 命中且反序列化成功返回 true，未配置/禁用/目标类型不对/取值损坏返回 false（out 保持原值）
func (s *SysParamReaderService) GetJSON(ctx context.Context, key string, out any) bool {
	if v := reflect.ValueOf(out); v.Kind() != reflect.Pointer || v.IsNil() {
		// 调用方把目标写成值而不是指针属代码问题，别记成"取值损坏"，否则排障方向会被带偏
		s.warnByKey(key, func() {
			logger.Warn(ctx, "[SysParam] GetJSON 的目标须为非空指针，本次返回 false", "key", key)
		})
		return false
	}
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

// GetMany 一次读取多个参数，返回"参数键 → 原始取值"。
// 只包含已配置且启用的键；未配置、禁用、数据库故障的键不出现在结果里，由调用方兜底。
// 空键与重复键会被忽略，一次 IN 查询的键数上限为 maxKeysPerQuery，超出自动分批。
func (s *SysParamReaderService) GetMany(ctx context.Context, keys []string) map[string]string {
	result := make(map[string]string, len(keys))
	pending := normalizeParamKeys(keys)
	for start := 0; start < len(pending); start += maxKeysPerQuery {
		end := min(start+maxKeysPerQuery, len(pending))
		rows, err := s.repo.FindEnabledByKeys(ctx, pending[start:end])
		if err != nil {
			// 数据库不可用：与单键读取一致，回退调用方默认值，业务不受阻
			s.warnFallback(ctx, "批量查询系统参数失败", err)
			return result
		}
		for _, param := range rows {
			result[param.Key] = param.Value
		}
	}
	return result
}

// normalizeParamKeys 去掉空键并按首次出现顺序去重，避免重复键放大 IN 查询
func normalizeParamKeys(keys []string) []string {
	seen := make(map[string]struct{}, len(keys))
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
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

// warnFallback 按固定间隔输出数据库故障告警，防止故障期间日志被单个请求刷满
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

// warnByKey 按参数键限流输出告警：同一键在一个窗口内只提示一次，不同键互不影响
func (s *SysParamReaderService) warnByKey(key string, log func()) {
	stamp, _ := s.keyWarns.LoadOrStore(key, &atomic.Int64{})
	last := stamp.(*atomic.Int64)
	now := s.now().UnixNano()
	prev := last.Load()
	if now-prev < int64(fallbackWarnInterval) {
		return
	}
	if !last.CompareAndSwap(prev, now) {
		return
	}
	log()
}

// warnValueType 输出类型不匹配告警：读取方期望的类型与参数定义不一致
func (s *SysParamReaderService) warnValueType(ctx context.Context, key string, actual, expected enums.ParamValueType) {
	s.warnByKey(key, func() {
		logger.Warn(ctx, "[SysParam] 参数值类型与读取方式不匹配，返回默认值",
			"key", key, "valueType", actual, "expected", expected)
	})
}

// warnCorrupt 输出取值损坏告警：写入校验理论已拦截，防御脏数据
func (s *SysParamReaderService) warnCorrupt(ctx context.Context, key string) {
	s.warnByKey(key, func() {
		logger.Warn(ctx, "[SysParam] 参数取值损坏，返回默认值", "key", key)
	})
}
