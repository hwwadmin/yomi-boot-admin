// Package support 编排层浅层错误投影。
//
// 这里只放与入口/会话相关的错误，不承载业务规则，也不依赖任何领域包。
package support

import "github.com/hwwadmin/yomi-boot/component/core/errors"

// ErrCannotOperateSelf 不能操作当前登录账号
var ErrCannotOperateSelf = errors.NewBiz("ADMIN_CANNOT_OPERATE_SELF", "不能操作当前登录账号")

// ErrTokenRevokeFailed 账号状态已经提交，但该用户的登录令牌还没撤销。
//
// 返回此错误时数据库变更已经生效：停用用户的重试会再次撤销令牌，删除用户的重试
// 会跳过已删除的账号直接补撤，所以提示里要让管理员知道"重试即可补上"。
var ErrTokenRevokeFailed = errors.NewBiz("ADMIN_TOKEN_REVOKE_FAILED",
	"账号状态已变更，但登录令牌未撤销，该账号在令牌过期前仍可访问有权限的接口，请重试")

// ErrPasswordChangeTokenRevokeFailed 改密与撤销登录状态在同一个幂等事务里：撤销失败时密码更新一并回滚。
//
// 返回此错误时密码没有生效（可能已有部分令牌被撤销），重新登录后用原密码重试即可。
var ErrPasswordChangeTokenRevokeFailed = errors.NewBiz("ADMIN_PASSWORD_CHANGE_TOKEN_REVOKE_FAILED",
	"改密未完成：登录状态撤销失败，请重新登录后重试")

// ErrSessionRevokeFailed 强制下线会话时，令牌撤销重试后仍失败。
//
// 与 ErrTokenRevokeFailed 的区别：账号状态没有变化，重试踢人即可，不涉及改密或停用。
var ErrSessionRevokeFailed = errors.NewBiz("ADMIN_SESSION_REVOKE_FAILED",
	"会话未撤销成功，请重试")

// ErrIdempotencyKeyInvalid 幂等键超过 t_sys_biz.biz_id 列宽，无法作为幂等记录保存。
var ErrIdempotencyKeyInvalid = errors.NewBiz("ADMIN_IDEMPOTENCY_KEY_INVALID",
	"幂等键不合法，请刷新页面后重试")
