package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/response"

	"github.com/gin-gonic/gin"
)

// fakeAuthService 只实现接口访问判定，其余方法由嵌入接口兜底
type fakeAuthService struct {
	rbaccontract.RbacAuthService
	allowed  bool
	calls    int
	lastUser int64
	lastPath string
}

func (f *fakeAuthService) CanAccessAPI(_ context.Context, userID int64, _, path string) (bool, error) {
	f.calls++
	f.lastUser = userID
	f.lastPath = path
	return f.allowed, nil
}

// newRBACEngine 构造带令牌会话与 RBAC 中间件的测试引擎
func newRBACEngine(t *testing.T, auth rbaccontract.RbacAuthService, skip map[string]struct{}) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	response.RegisterDefaultForTest(t)

	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		session := &token.TokenSession{UserID: 7}
		c.Request = c.Request.WithContext(token.WithSession(c.Request.Context(), session))
		c.Next()
	})
	engine.Use(RBAC(auth, skip))

	group := engine.Group("/api")
	group.GET("/system/role", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	group.POST("/system/role", func(c *gin.Context) { c.String(http.StatusOK, "posted") })
	return engine
}

// serve 执行一次请求并返回响应
func serve(engine *gin.Engine, method, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, nil)
	engine.ServeHTTP(recorder, request)
	return recorder
}

// bodyCode 解析统一响应体里的业务码
func bodyCode(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应体失败: %v body=%s", err, recorder.Body.String())
	}
	return body.Code
}

// TestRBACDeniesUnmatchedRoute 校验未匹配路由按默认拒绝处理，不再放行
func TestRBACDeniesUnmatchedRoute(t *testing.T) {
	auth := &fakeAuthService{allowed: true}
	engine := newRBACEngine(t, auth, nil)

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "未注册路径", method: http.MethodGet, path: "/api/not-exists"},
		{name: "方法不允许", method: http.MethodDelete, path: "/api/system/role"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := serve(engine, tc.method, tc.path)
			if code := bodyCode(t, recorder); code != "ADMIN_PERMISSION_DENIED" {
				t.Fatalf("期望默认拒绝，实际 code=%s body=%s", code, recorder.Body.String())
			}
			if auth.calls != 0 {
				t.Fatalf("未匹配路由不应查询权限: calls=%d", auth.calls)
			}
		})
	}
}

// TestRBACChecksMatchedRoute 校验已注册路由按权限判定放行或拒绝
func TestRBACChecksMatchedRoute(t *testing.T) {
	denied := &fakeAuthService{allowed: false}
	engine := newRBACEngine(t, denied, nil)

	recorder := serve(engine, http.MethodGet, "/api/system/role")
	if code := bodyCode(t, recorder); code != "ADMIN_PERMISSION_DENIED" {
		t.Fatalf("无权限期望拒绝，实际 code=%s body=%s", code, recorder.Body.String())
	}
	if denied.calls != 1 || denied.lastUser != 7 || denied.lastPath != "/api/system/role" {
		t.Fatalf("权限查询入参不符: calls=%d user=%d path=%s", denied.calls, denied.lastUser, denied.lastPath)
	}

	allowed := &fakeAuthService{allowed: true}
	engine = newRBACEngine(t, allowed, nil)
	recorder = serve(engine, http.MethodGet, "/api/system/role")
	if recorder.Code != http.StatusOK || recorder.Body.String() != "ok" {
		t.Fatalf("有权限应放行: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

// TestRBACSkipsPublicPaths 校验跳过名单与预检请求不参与鉴权
func TestRBACSkipsPublicPaths(t *testing.T) {
	auth := &fakeAuthService{allowed: false}
	engine := newRBACEngine(t, auth, map[string]struct{}{"GET /api/system/role": {}})

	// 命中跳过名单：交给后续处理，不查权限
	if recorder := serve(engine, http.MethodGet, "/api/system/role"); recorder.Code != http.StatusOK || recorder.Body.String() != "ok" {
		t.Fatalf("跳过名单应交给后续处理: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if recorder := serve(engine, http.MethodOptions, "/api/not-exists"); recorder.Code != http.StatusNotFound {
		t.Fatalf("预检请求应放行: status=%d", recorder.Code)
	}
	if auth.calls != 0 {
		t.Fatalf("跳过名单与预检不应查询权限: calls=%d", auth.calls)
	}

	// 同一路径的其它方法不在跳过名单里：必须查询权限
	if recorder := serve(engine, http.MethodPost, "/api/system/role"); bodyCode(t, recorder) != "ADMIN_PERMISSION_DENIED" {
		t.Fatalf("同一路径的其它方法不应被跳过: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if auth.calls != 1 {
		t.Fatalf("同一路径的其它方法应查询权限: calls=%d", auth.calls)
	}
}
