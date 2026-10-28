// Package support 编排层浅层错误投影。
//
// 这里只放与入口/会话相关的错误，不承载业务规则，也不依赖任何领域包。
package support

import "gitee.com/hq_project/yomi-boot/component/core/errors"

// ErrCannotOperateSelf 不能操作当前登录账号
var ErrCannotOperateSelf = errors.NewBiz("ADMIN_CANNOT_OPERATE_SELF", "不能操作当前登录账号")
