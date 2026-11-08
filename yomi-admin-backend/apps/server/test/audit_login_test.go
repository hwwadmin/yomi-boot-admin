package test

import (
	"fmt"
	"strconv"
	"testing"
	"time"
	authusecase "yomi-admin-backend/apps/server/internal/usecase/auth"
)

// TestAuditLogin 覆盖登录失败与成功的两段式审计
func TestAuditLogin(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)
	startTime := auditStartTime()
	suffix := time.Now().UnixNano()

	// 登录失败：开始/结束都记录尝试的用户名
	ghost := fmt.Sprintf("audit_ghost_%d", suffix)
	body, _ := Post(API("/auth/login"), "", authusecase.LoginInput{
		Username: ghost,
		Password: "wrong-password",
	})
	assertFailure(t, body)

	logs := findAuditLogs(t, adminToken, auditLogFilter{
		EventCode: "auth.login",
		Keyword:   ghost,
		StartTime: startTime,
	})
	started, finished := takeAuditPair(t, logs)
	if started.Result != "PENDING" || started.OperatorID != "0" || started.OperatorName != ghost {
		t.Fatalf("登录失败开始记录错误: %+v", started)
	}
	if finished.Result != "FAILURE" || finished.OperatorID != "0" || finished.OperatorName != ghost {
		t.Fatalf("登录失败结束记录错误: %+v", finished)
	}

	// 登录成功：结束记录补真实用户 ID
	userID, username, password := createAuditUser(t, adminToken, suffix)
	login(t, username, password)

	logs = findAuditLogs(t, adminToken, auditLogFilter{
		EventCode: "auth.login",
		Keyword:   username,
		StartTime: startTime,
	})
	started, finished = takeAuditPair(t, logs)
	if started.OperatorID != "0" || started.OperatorName != username {
		t.Fatalf("登录成功开始记录错误: %+v", started)
	}
	if finished.OperatorID != strconv.FormatInt(userID, 10) || finished.OperatorName != username ||
		finished.Result != "SUCCESS" || finished.BizCode != "OK" {
		t.Fatalf("登录成功结束记录错误: %+v", finished)
	}
}
