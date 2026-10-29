package config

import (
	"strings"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// publicRouteMethods 是公开路由允许声明的 HTTP 方法（本服务对外只用 GET / POST）
var publicRouteMethods = map[string]struct{}{
	"GET":  {},
	"POST": {},
}

// PublicRouteConfig 是一条公开路由声明
type PublicRouteConfig struct {
	Method string `mapstructure:"method"` // Method 是 HTTP 方法，仅支持 GET / POST
	Path   string `mapstructure:"path"`   // Path 是路由路径，省略 http.router_prefix，启动时自动拼接
	Token  bool   `mapstructure:"token"`  // Token 表示是否校验登录令牌：false 完全公开，true 登录即可（仅跳过接口权限）
}

// AuthConfig 是认证相关的公开路由配置。
//
// 免令牌与免接口权限的路由只在这里声明，代码里不再硬编码路径：
//   - token=false 的路由同时注入令牌中间件的跳过清单（完全公开）；
//   - 全部条目组成 RBAC 的跳过清单（按 方法 + 完整路径 匹配）。
type AuthConfig struct {
	PublicPaths []PublicRouteConfig `mapstructure:"public_paths"` // PublicPaths 是公开路由清单
}

// PublicRoute 是拼接好路由前缀的公开路由
type PublicRoute struct {
	Method string // Method 是 HTTP 方法
	Path   string // Path 是含 http.router_prefix 的完整路径
	Token  bool   // Token 表示是否校验登录令牌
}

// Validate 校验公开路由声明；routerPrefix 取自 http.router_prefix，用于防止路径重复写前缀
func (c AuthConfig) Validate(routerPrefix string) error {
	if len(c.PublicPaths) == 0 {
		return errors.Newf("%s.auth.public_paths 不能为空：登录、验证码等路由必须在配置里显式声明为公开", SectionKey)
	}

	prefix := normalizeRouterPrefix(routerPrefix)
	seen := make(map[string]struct{}, len(c.PublicPaths))
	for i, item := range c.PublicPaths {
		if _, ok := publicRouteMethods[item.Method]; !ok {
			return errors.Newf("%s.auth.public_paths[%d].method 仅支持 GET / POST，实际为 %q", SectionKey, i, item.Method)
		}
		if !strings.HasPrefix(item.Path, "/") || strings.HasSuffix(item.Path, "/") || strings.ContainsAny(item.Path, " \t") {
			return errors.Newf("%s.auth.public_paths[%d].path 必须以 / 开头、不以 / 结尾且不含空白，实际为 %q", SectionKey, i, item.Path)
		}
		if prefix != "" && (item.Path == prefix || strings.HasPrefix(item.Path, prefix+"/")) {
			return errors.Newf("%s.auth.public_paths[%d].path 不要写路由前缀 %s（启动时会自动拼接），实际为 %q",
				SectionKey, i, prefix, item.Path)
		}

		key := item.Method + " " + item.Path
		if _, dup := seen[key]; dup {
			return errors.Newf("%s.auth.public_paths 存在重复路由 %s %s", SectionKey, item.Method, item.Path)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// PublicRoutes 返回拼接好路由前缀的公开路由；声明已在 Validate 阶段校验
func (c AuthConfig) PublicRoutes(routerPrefix string) []PublicRoute {
	prefix := normalizeRouterPrefix(routerPrefix)
	routes := make([]PublicRoute, 0, len(c.PublicPaths))
	for _, item := range c.PublicPaths {
		routes = append(routes, PublicRoute{Method: item.Method, Path: prefix + item.Path, Token: item.Token})
	}
	return routes
}

// normalizeRouterPrefix 规范化路由前缀：去空白与尾部 /，缺少前导 / 时补上
func normalizeRouterPrefix(prefix string) string {
	prefix = strings.TrimRight(strings.TrimSpace(prefix), "/")
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return prefix
}
