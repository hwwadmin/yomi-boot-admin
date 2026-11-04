package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	bizerr "github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/response"
	sharederr "github.com/hwwadmin/yomi-boot/component/core/shared/error"

	"github.com/gin-gonic/gin"
)

// TestMain 注册默认编码器：生产环境由 boot.Run 完成，handler 单测需要显式注册。
func TestMain(m *testing.M) {
	if !response.Registered() {
		response.MustRegister(response.DefaultEncoder{})
	}
	os.Exit(m.Run())
}

// TestHandleFailureCodes 验证用例出口的错误码收敛：
// 领域 BizError 保留 code，未知内部错误统一 INTERNAL_ERROR，超时映射 REQUEST_TIMEOUT。
func TestHandleFailureCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/unknown", func(c *gin.Context) {
		handle(c, func(context.Context) (any, error) {
			return nil, errors.New("repo: query failed")
		})
	})
	router.GET("/biz", func(c *gin.Context) {
		handle(c, func(context.Context) (any, error) {
			return nil, bizerr.NewBiz("ADMIN_ROLE_NOT_FOUND", "角色不存在")
		})
	})
	router.GET("/timeout", func(c *gin.Context) {
		handleVoid(c, func(context.Context) error {
			return context.DeadlineExceeded
		})
	})
	router.GET("/invalid-input", func(c *gin.Context) {
		handleVoid(c, func(context.Context) error {
			return sharederr.ErrInvalidInput
		})
	})

	tests := []struct {
		path string
		code string
	}{
		{path: "/unknown", code: "INTERNAL_ERROR"},
		{path: "/biz", code: "ADMIN_ROLE_NOT_FOUND"},
		{path: "/timeout", code: "REQUEST_TIMEOUT"},
		{path: "/invalid-input", code: "INVALID_INPUT"},
	}

	for _, tt := range tests {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.path, nil))

		var body struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s 响应不是合法 JSON：%v", tt.path, err)
		}
		if body.Code != tt.code {
			t.Fatalf("%s 期望 code=%s，实际 code=%s", tt.path, tt.code, body.Code)
		}
	}
}
