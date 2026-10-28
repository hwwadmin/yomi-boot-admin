// Package repo_contract 用户域仓储契约
package repo_contract

import "gitee.com/hq_project/yomi-boot/component/core/errors"

// ErrDuplicatedKey 表示唯一键冲突
var ErrDuplicatedKey = errors.NewNoStack("REPO_DUPLICATED_KEY")
