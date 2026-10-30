// Package service_contract 系统参数域服务契约（共享内核：零域间依赖，业务域可直接依赖）
package service_contract

import "gitee.com/hq_project/yomi-boot/component/core/errors"

var (
	// ErrParamNotFound 参数不存在
	ErrParamNotFound = errors.NewBiz("ADMIN_PARAM_NOT_FOUND", "参数不存在")
	// ErrParamKeyExists 参数键已存在
	ErrParamKeyExists = errors.NewBiz("ADMIN_PARAM_KEY_EXISTS", "参数键已存在")
	// ErrInvalidInput 参数不合法
	ErrInvalidInput = errors.NewBiz("ADMIN_INVALID_INPUT", "参数不合法")
)
