// Package auth 认证相关编排
package auth

// builtinAdminUsername 是 P0 的内置管理员账号。
// P0 没有 RBAC，此账号在登录/资料接口中返回全量权限占位。
const builtinAdminUsername = "admin"

// permissionsOf 返回 P0 的权限占位；P1 接入 RBAC 后替换。
func permissionsOf(username string) []string {
	if username == builtinAdminUsername {
		return []string{"*"}
	}
	return []string{}
}
