package config

import (
	"strings"
	"testing"
)

// TestAuthConfigPublicRoutes 校验公开路由拼接路由前缀与令牌标记
func TestAuthConfigPublicRoutes(t *testing.T) {
	cfg := AuthConfig{PublicPaths: []PublicRouteConfig{
		{Method: "POST", Path: "/auth/login", Token: false},
		{Method: "GET", Path: "/auth/profile", Token: true},
	}}

	routes := cfg.PublicRoutes("/api/")
	if len(routes) != 2 {
		t.Fatalf("公开路由数量不符: %+v", routes)
	}
	if routes[0].Path != "/api/auth/login" || routes[0].Token {
		t.Fatalf("完全公开路由拼接不符: %+v", routes[0])
	}
	if routes[1].Path != "/api/auth/profile" || !routes[1].Token {
		t.Fatalf("登录即可路由拼接不符: %+v", routes[1])
	}

	// 空前缀：路径原样使用
	routes = cfg.PublicRoutes("")
	if routes[0].Path != "/auth/login" {
		t.Fatalf("空前缀应保留原路径: %+v", routes[0])
	}
}

// TestAuthConfigValidate 校验公开路由声明的非法输入
func TestAuthConfigValidate(t *testing.T) {
	valid := AuthConfig{PublicPaths: []PublicRouteConfig{{Method: "POST", Path: "/auth/login"}}}
	if err := valid.Validate("/api"); err != nil {
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
			cfg:     AuthConfig{PublicPaths: []PublicRouteConfig{{Method: "PUT", Path: "/auth/login"}}},
			wantSub: "仅支持 GET / POST",
		},
		{
			name:    "小写方法",
			cfg:     AuthConfig{PublicPaths: []PublicRouteConfig{{Method: "get", Path: "/auth/login"}}},
			wantSub: "仅支持 GET / POST",
		},
		{
			name:    "缺少前导斜杠",
			cfg:     AuthConfig{PublicPaths: []PublicRouteConfig{{Method: "GET", Path: "auth/login"}}},
			wantSub: "必须以 / 开头",
		},
		{
			name:    "尾部带斜杠",
			cfg:     AuthConfig{PublicPaths: []PublicRouteConfig{{Method: "GET", Path: "/auth/login/"}}},
			wantSub: "不以 / 结尾",
		},
		{
			name:    "写了路由前缀",
			cfg:     AuthConfig{PublicPaths: []PublicRouteConfig{{Method: "GET", Path: "/api/auth/login"}}},
			wantSub: "不要写路由前缀",
		},
		{
			name: "重复路由",
			cfg: AuthConfig{PublicPaths: []PublicRouteConfig{
				{Method: "GET", Path: "/auth/login"},
				{Method: "GET", Path: "/auth/login"},
			}},
			wantSub: "重复路由",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate("/api")
			if err == nil {
				t.Fatal("非法声明应校验失败")
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("错误信息应包含 %q，实际 %v", tc.wantSub, err)
			}
		})
	}

	// 前缀为空时不校验"是否重复写前缀"
	noPrefix := AuthConfig{PublicPaths: []PublicRouteConfig{{Method: "GET", Path: "/api/auth/login"}}}
	if err := noPrefix.Validate(""); err != nil {
		t.Fatalf("空前缀下路径含 /api 应合法: %v", err)
	}
}
