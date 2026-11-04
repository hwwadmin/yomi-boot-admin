package test

import (
	"strconv"
	"testing"
	"time"
	authusecase "yomi-admin-backend/apps/server/internal/usecase/auth"

	"github.com/hwwadmin/yomi-boot/component/common/testclient"
)

// TestAuditChangePassword 覆盖修改密码失败与成功的两段式审计
func TestAuditChangePassword(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)
	startTime := auditStartTime()
	suffix := time.Now().UnixNano()

	userID, username, password := createAuditUser(t, adminToken, suffix)
	userToken := login(t, username, password)
	operatorID := strconv.FormatInt(userID, 10)

	// 原密码错误：结束记录失败，开始/结束共用 audit_id
	body, _ := testclient.Post(API("/auth/password"), userToken, authusecase.ChangePasswordInput{
		OldPassword: "wrong-old",
		NewPassword: "auditpass2",
	})
	assertFailure(t, body)

	failureLogs := findAuditLogs(t, adminToken, auditLogFilter{
		EventCode: "auth.change_password",
		Keyword:   username,
		Result:    "FAILURE",
		StartTime: startTime,
	})
	if len(failureLogs) != 1 {
		t.Fatalf("改密失败应有 1 条结束记录，实际 %d: %+v", len(failureLogs), failureLogs)
	}
	started, finished := takeAuditPair(t, findAuditLogs(t, adminToken, auditLogFilter{
		AuditID: failureLogs[0].AuditID,
	}))
	if started.Result != "PENDING" || started.OperatorID != operatorID {
		t.Fatalf("改密失败开始记录错误: %+v", started)
	}
	if finished.Result != "FAILURE" || finished.BizCode != "ADMIN_OLD_PASSWORD_MISMATCH" {
		t.Fatalf("改密失败结束记录错误: %+v", finished)
	}

	// 改密成功：结束记录成功
	body, _ = testclient.Post(API("/auth/password"), userToken, authusecase.ChangePasswordInput{
		OldPassword: password,
		NewPassword: "auditpass2",
	})
	parseSuccess(t, body)

	successLogs := findAuditLogs(t, adminToken, auditLogFilter{
		EventCode: "auth.change_password",
		Keyword:   username,
		Result:    "SUCCESS",
		StartTime: startTime,
	})
	if len(successLogs) != 1 {
		t.Fatalf("改密成功应有 1 条结束记录，实际 %d: %+v", len(successLogs), successLogs)
	}
	started, finished = takeAuditPair(t, findAuditLogs(t, adminToken, auditLogFilter{
		AuditID: successLogs[0].AuditID,
	}))
	if started.Result != "PENDING" || started.OperatorID != operatorID {
		t.Fatalf("改密成功开始记录错误: %+v", started)
	}
	if finished.Result != "SUCCESS" || finished.BizCode != "OK" {
		t.Fatalf("改密成功结束记录错误: %+v", finished)
	}
}
