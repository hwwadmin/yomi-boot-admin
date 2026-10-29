package config

import (
	"strings"
	"testing"
)

// TestAuthConfigPublicRoutes 校验公开路由按声明原样返回（path 已是完整 URI）
func TestAuthConfigPublicRoutes(t *testing.T) {
	cfg := AuthConfig{PublicPaths: []PublicRouteConfig{
		{Method: "POST", Path: "/api/auth/login", Token: false},
		{Method: "GET", Path: "/api/auth/profile", Token: true},
	}}

	routes := cfg.PublicRoutes()
	if len(routes) != 2 {
		t.Fatalf("公开路由数量不符: %+v", routes)
	}
	if routes[0].Path != "/api/auth/login" || routes[0].Token {
		t.Fatalf("完全公开路由不符: %+v", routes[0])
	}
	if routes[1].Path != "/api/auth/profile" || !routes[1].Token {
		t.Fatalf("登录即可路由不符: %+v", routes[1])
	}
}

// TestAuthConfigValidate 校验公开路由声明的非法输入
func TestAuthConfigValidate(t *testing.T) {
	valid := AuthConfig{PublicPaths: []PublicRouteConfig{{Method: "POST", Path: "/api/auth/login"}}}
	if err := valid.Validate(); err != nil {
		t.Fatalf("合法声明应通过: %v", err)
	}

	cases := []struct {
		name    string
		cfg     AuthConfig
		wantSub string
	}{
		{
			name:    "空清单",
			cfg:     AuthConfig{},
			wantSub: "不能为空",
		},
		{
			name:    "非法方法",
			cfg:     AuthConfig{PublicPaths: []PublicRouteConfig{{Method: "PUT", Path: "/api/auth/login"}}},
			wantSub: "仅支持 GET / POST",
		},
		{
			name:    "小写方法",
			cfg:     AuthConfig{PublicPaths: []PublicRouteConfig{{Method: "get", Path: "/api/auth/login"}}},
			wantSub: "仅支持 GET / POST",
		},
		{
			name:    "缺少前导斜杠",
			cfg:     AuthConfig{PublicPaths: []PublicRouteConfig{{Method: "GET", Path: "api/auth/login"}}},
			wantSub: "必须以 / 开头",
		},
		{
			name:    "尾部带斜杠",
			cfg:     AuthConfig{PublicPaths: []PublicRouteConfig{{Method: "GET", Path: "/api/auth/login/"}}},
			wantSub: "不以 / 结尾",
		},
		{
			name: "重复路由",
			cfg: AuthConfig{PublicPaths: []PublicRouteConfig{
				{Method: "GET", Path: "/api/auth/login"},
				{Method: "GET", Path: "/api/auth/login"},
			}},
			wantSub: "重复路由",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if err == nil {
				t.Fatal("非法声明应校验失败")
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("错误信息应包含 %q，实际 %v", tc.wantSub, err)
			}
		})
	}
}
