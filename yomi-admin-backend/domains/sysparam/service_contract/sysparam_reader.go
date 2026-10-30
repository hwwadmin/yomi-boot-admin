// Package service_contract 系统参数域服务契约（共享内核：零域间依赖，业务域可直接依赖）
package service_contract

import "context"

// SysParamReader 业务代码读取系统参数的契约。
//
// 语义（选用系统参数即接受）：
//   - 读取直接查库，参数保存即对业务生效；
//   - 默认值由读方自带：参数不存在、被禁用、被删除，甚至数据库故障，
//     一律返回 fallback——读参数永不返回错误、永不阻塞业务；
//   - 参数键与默认值建议在业务模块内以常量文件维护（params.go），
//     避免同一键在多处写出不一致的 fallback。
type SysParamReader interface {
	// GetBool 读取布尔参数（值类型须为 BOOL）
	GetBool(ctx context.Context, key string, fallback bool) bool
	// GetInt64 读取整数参数（值类型须为 INT）
	GetInt64(ctx context.Context, key string, fallback int64) int64
	// GetString 读取字符串参数（值类型须为 STRING）
	GetString(ctx context.Context, key, fallback string) string
	// GetJSON 读取 JSON 参数并反序列化到 out（值类型须为 JSON）；
	// 命中且反序列化成功返回 true，未配置/禁用/损坏返回 false（out 保持原值）
	GetJSON(ctx context.Context, key string, out any) bool
	// Exists 判断参数是否已配置且启用（不看类型与取值）
	Exists(ctx context.Context, key string) bool
	// GetMany 一次读取多个参数，返回"参数键 → 原始取值"的映射。
	// 只包含已配置且启用的键；未配置、被禁用、数据库故障的键不出现在结果里，
	// 由调用方按自己的默认值兜底。空键或重复键会被忽略。
	// 用于一次请求要读多个参数的场景，避免逐个键各查一次库；取值仍按参数自身类型自行解释。
	GetMany(ctx context.Context, keys []string) map[string]string
}
