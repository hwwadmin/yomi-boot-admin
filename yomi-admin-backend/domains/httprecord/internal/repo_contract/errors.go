// Package repo_contract HTTP 访问记录仓储契约
package repo_contract

import "github.com/hwwadmin/yomi-boot/component/core/errors"

// ErrInvalidRecord 表示访问记录参数不合法
var ErrInvalidRecord = errors.NewNoStack("REPO_INVALID_HTTP_RECORD")
