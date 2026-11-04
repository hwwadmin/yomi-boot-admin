// Package service_contract 操作审计日志域服务契约（共享内核：零域间依赖）
package service_contract

import "github.com/hwwadmin/yomi-boot/component/core/errors"

// ErrInvalidInput 审计日志参数不合法，复用全局 INVALID_INPUT 错误码
var ErrInvalidInput = errors.NewBiz("ADMIN_INVALID_INPUT", "审计日志参数不合法")
