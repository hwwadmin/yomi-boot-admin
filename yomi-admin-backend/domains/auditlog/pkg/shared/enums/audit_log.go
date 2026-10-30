// Package enums 操作审计日志枚举
package enums

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Result 表示审计事件的结果
type Result string

const (
	// ResultPending 表示审计事件已开始、结果未出
	ResultPending Result = "PENDING"
	// ResultSuccess 表示操作成功
	ResultSuccess Result = "SUCCESS"
	// ResultFailure 表示操作失败或被拒绝
	ResultFailure Result = "FAILURE"
)

// IsValidResult 判断结果是否为合法枚举
func IsValidResult(result Result) bool {
	return result == ResultPending || result == ResultSuccess || result == ResultFailure
}

// ParseResult 解析结果值：去空白、转大写并校验；空串返回空
func ParseResult(raw string) (Result, error) {
	value := Result(strings.ToUpper(strings.TrimSpace(raw)))
	if value == "" {
		return "", nil
	}
	if !IsValidResult(value) {
		return "", fmt.Errorf("invalid Result: %s", raw)
	}
	return value, nil
}

// UnmarshalJSON 实现 json.Unmarshaler，反序列化时归一化并校验
func (r *Result) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	value, err := ParseResult(raw)
	if err != nil {
		return err
	}
	*r = value
	return nil
}

// Phase 表示审计记录所处的阶段
type Phase string

const (
	// PhaseStarted 表示业务开始前写入的开始记录
	PhaseStarted Phase = "STARTED"
	// PhaseFinished 表示业务结束后写入的结束记录
	PhaseFinished Phase = "FINISHED"
)

// IsValidPhase 判断阶段是否为合法枚举
func IsValidPhase(phase Phase) bool {
	return phase == PhaseStarted || phase == PhaseFinished
}

// ParsePhase 解析阶段值：去空白、转大写并校验；空串返回空
func ParsePhase(raw string) (Phase, error) {
	value := Phase(strings.ToUpper(strings.TrimSpace(raw)))
	if value == "" {
		return "", nil
	}
	if !IsValidPhase(value) {
		return "", fmt.Errorf("invalid Phase: %s", raw)
	}
	return value, nil
}

// UnmarshalJSON 实现 json.Unmarshaler，反序列化时归一化并校验
func (p *Phase) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	value, err := ParsePhase(raw)
	if err != nil {
		return err
	}
	*p = value
	return nil
}
