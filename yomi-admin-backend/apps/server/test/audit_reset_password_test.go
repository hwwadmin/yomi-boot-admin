package test

import (
	"strconv"
	"testing"
	"time"
	userusecase "yomi-admin-backend/apps/server/internal/usecase/user"

	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// TestAuditResetPassword 覆盖重置他人密码的两段式审计
func TestAuditResetPassword(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)
	startTime := auditStartTime()
	suffix := time.Now().UnixNano()

	userID, targetUsername, _ := createAuditUser(t, adminToken, suffix)
	body, _ := Post(API("/system/user/reset-password"), adminToken, userusecase.ResetUserPasswordInput{
		ID:          dto.JSONInt64(userID),
		NewPassword: "auditpass3",
	})
	parseSuccess(t, body)

	targetID := strconv.FormatInt(userID, 10)
	logs := findAuditLogs(t, adminToken, auditLogFilter{
		EventCode: "auth.reset_password",
		StartTime: startTime,
	})
	started, finished := takeAuditPair(t, filterAuditByTargetID(logs, targetID))
	if started.Result != "PENDING" || started.OperatorName != adminUsername ||
		started.TargetType != "user" || started.TargetID != targetID {
		t.Fatalf("重置密码开始记录错误: %+v", started)
	}
	if finished.Result != "SUCCESS" || finished.OperatorName != adminUsername ||
		finished.TargetType != "user" || finished.TargetID != targetID || finished.TargetName != targetUsername {
		t.Fatalf("重置密码结束记录错误: %+v", finished)
	}
}
