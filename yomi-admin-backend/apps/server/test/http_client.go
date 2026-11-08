package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// httpClient 是集成测试用的 HTTP 客户端。
// yomi-boot v1.0.2 移除了 component/common/testclient，请求辅助改由本包持有。
var httpClient = &http.Client{Timeout: 15 * time.Second}

// Get 发起 GET。HTTP 状态码为 200 时 ok 为 true。
func Get(url, token string) (body string, ok bool) {
	return do(http.MethodGet, url, token, nil)
}

// Post 发起 POST。reqBody 非 nil 时序列化为 JSON。HTTP 状态码为 200 时 ok 为 true。
func Post(url, token string, reqBody any) (body string, ok bool) {
	return do(http.MethodPost, url, token, reqBody)
}

func do(method, url, token string, reqBody any) (string, bool) {
	if strings.TrimSpace(url) == "" {
		return "请求地址不能为空", false
	}

	var bodyReader io.Reader
	if reqBody != nil {
		payload, err := json.Marshal(reqBody)
		if err != nil {
			return "序列化请求体失败: " + err.Error(), false
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return "创建请求失败: " + err.Error(), false
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	setAuthorization(req, token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "请求失败: " + err.Error(), false
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "读取响应失败: " + err.Error(), false
	}
	return string(respBody), resp.StatusCode == http.StatusOK
}

func setAuthorization(req *http.Request, token string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}
	if len(token) >= 7 && strings.EqualFold(token[:7], "Bearer ") {
		req.Header.Set("Authorization", token)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
}
