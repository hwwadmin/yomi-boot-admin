// Package service_contract HTTP 访问记录域服务契约
package service_contract

import "github.com/hwwadmin/yomi-boot/component/core/errors"

var (
	// ErrRecordNotFound 访问记录不存在
	ErrRecordNotFound = errors.NewBiz("ADMIN_HTTP_RECORD_NOT_FOUND", "访问记录不存在")
	// ErrInvalidInput 访问记录参数不合法
	ErrInvalidInput = errors.NewBiz("ADMIN_INVALID_INPUT", "参数不合法")
)
