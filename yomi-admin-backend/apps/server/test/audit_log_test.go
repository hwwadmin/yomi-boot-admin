// Package test 操作审计日志集成测试辅助（须先启动 yomi-admin-backend 服务）
package test

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"
	"time"
	userusecase "yomi-admin-backend/apps/server/internal/usecase/user"
	userenums "yomi-admin-backend/domains/user/pkg/shared/enums"

	"gitee.com/hq_project/yomi-boot/component/common/testclient"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// auditLogItem 是审计日志列表项
type auditLogItem struct {
	ID           string `json:"id"`
	AuditID      string `json:"auditId"`
	Phase        string `json:"phase"`
	EventCode    string `json:"eventCode"`
	OperatorID   string `json:"operatorId"`
	OperatorName string `json:"operatorName"`
	ClientIP     string `json:"clientIp"`
	TargetType   string `json:"targetType"`
	TargetID     string `json:"targetId"`
	TargetName   string `json:"targetName"`
	Result       string `json:"result"`
	BizCode      string `json:"bizCode"`
	CreateTime   int64  `json:"createTime"`
}

// auditLogFilter 审计日志查询条件
type auditLogFilter struct {
	EventCode string
	Phase     string
	Keyword   string
	Result    string
	AuditID   string
	StartTime string
}

// findAuditLogs 按条件查询审计日志
func findAuditLogs(t *testing.T, token string, filter auditLogFilter) []auditLogItem {
	t.Helper()

	query := url.Values{}
	query.Set("page", "1")
	query.Set("pageSize", "100")
	query.Set("eventCode", filter.EventCode)
	query.Set("phase", filter.Phase)
	query.Set("keyword", filter.Keyword)
	query.Set("result", filter.Result)
	query.Set("auditId", filter.AuditID)
	query.Set("startTime", filter.StartTime)
	// 列表查询必须带完整时间范围；只按 auditId 点查时不需要
	if filter.StartTime != "" {
		query.Set("endTime", fmt.Sprintf("%d", time.Now().Add(time.Minute).UnixMilli()))
	}

	body, ok := testclient.Get(API("/system/audit-log?"+query.Encode()), token)
	if !ok {
		t.Fatalf("审计日志接口不可达: %s", body)
	}
	return parsePage[auditLogItem](t, body).Items
}

// auditStartTime 返回一分钟前的时间戳，作为审计查询下界
func auditStartTime() string {
	return fmt.Sprintf("%d", time.Now().Add(-time.Minute).UnixMilli())
}

// takeAuditPair 从结果中取同一 audit_id 的开始/结束两条记录
func takeAuditPair(t *testing.T, logs []auditLogItem) (auditLogItem, auditLogItem) {
	t.Helper()
	if len(logs) != 2 {
		t.Fatalf("应写开始/结束两条审计，实际 %d: %+v", len(logs), logs)
	}

	var started, finished auditLogItem
	for _, item := range logs {
		switch item.Phase {
		case "STARTED":
			started = item
		case "FINISHED":
			finished = item
		}
	}
	if started.AuditID == "" || started.AuditID != finished.AuditID {
		t.Fatalf("开始/结束未共用 audit_id: %+v", logs)
	}
	return started, finished
}

// filterAuditByTargetID 从结果中筛出指定目标 ID 的记录
func filterAuditByTargetID(logs []auditLogItem, targetID string) []auditLogItem {
	items := make([]auditLogItem, 0, len(logs))
	for _, item := range logs {
		if item.TargetID == targetID {
			items = append(items, item)
		}
	}
	return items
}

// createAuditUser 创建临时审计测试用户，返回用户 ID、用户名与密码
func createAuditUser(t *testing.T, adminToken string, suffix int64) (int64, string, string) {
	t.Helper()

	username := fmt.Sprintf("audit_user_%d", suffix)
	password := "auditpass1"
	body, ok := testclient.Post(API("/system/user/create"), adminToken, userusecase.CreateUserInput{
		Username: username,
		Password: password,
		Nickname: "审计测试用户",
		Status:   userenums.UserStatusEnabled,
	})
	if !ok {
		t.Fatalf("创建用户接口不可达: %s", body)
	}
	parseSuccess(t, body)

	rawID := findUserID(t, adminToken, username)
	userID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		t.Fatalf("解析用户 ID 失败: %v", err)
	}
	t.Cleanup(func() {
		_, _ = testclient.Post(API("/system/user/delete"), adminToken, userusecase.DeleteUserInput{
			ID: dto.JSONInt64(userID),
		})
	})
	return userID, username, password
}
