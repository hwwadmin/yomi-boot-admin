// Package support 编排层浅层错误投影。
//
// 这里只放与入口/会话相关的错误，不承载业务规则，也不依赖任何领域包。
package support

import "gitee.com/hq_project/yomi-boot/component/core/errors"

// ErrCannotOperateSelf 不能操作当前登录账号
var ErrCannotOperateSelf = errors.NewBiz("ADMIN_CANNOT_OPERATE_SELF", "不能操作当前登录账号")

// ErrTokenRevokeFailed 账号状态已经提交，但该用户的登录令牌还没撤销。
//
// 返回此错误时数据库变更已经生效：停用用户的重试会再次撤销令牌，删除用户的重试
// 会跳过已删除的账号直接补撤，所以提示里要让管理员知道"重试即可补上"。
var ErrTokenRevokeFailed = errors.NewBiz("ADMIN_TOKEN_REVOKE_FAILED",
	"账号状态已变更，但登录令牌未撤销，该账号在令牌过期前仍可访问有权限的接口，请重试")

// ErrPasswordChangedTokenRevokeFailed 密码已经改好，但旧令牌还没撤销。
//
// 与 ErrTokenRevokeFailed 的区别是"不需要重试改密"：密码已经生效，重试改密会因为
// 原密码不匹配而失败，用户该做的是重新登录。
var ErrPasswordChangedTokenRevokeFailed = errors.NewBiz("ADMIN_PASSWORD_CHANGED_TOKEN_REVOKE_FAILED",
	"密码已修改，但旧登录状态未完全撤销，请重新登录")
