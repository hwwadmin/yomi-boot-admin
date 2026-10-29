// Package snapshot 提供 RBAC 授权的不可变快照与内存求值。
//
// 快照是六张表鉴权所需字段的全量映射：既可以直接序列化到 Redis，也可以反序列化后求值。
// 求值索引在首次使用时惰性构建，构建后只读，可安全并发。
package snapshot

import (
	"encoding/json"
	"fmt"
	"sync"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"
)

// Schema 是投影结构版本；结构变更时递增，存量投影会被判定为不可用并触发重建
const Schema = 1

// Role 是快照中的角色
type Role struct {
	ID      int64        `json:"id"`
	Code    string       `json:"code"`
	Builtin bool         `json:"builtin,omitempty"`
	Status  enums.Status `json:"status"`
}

// Permission 是快照中的权限
type Permission struct {
	ID         int64                `json:"id"`
	Code       string               `json:"code"`
	Name       string               `json:"name"`
	Type       enums.PermissionType `json:"type"`
	UIPath     string               `json:"uiPath,omitempty"`
	HTTPMethod string               `json:"httpMethod,omitempty"`
	HTTPPath   string               `json:"httpPath,omitempty"`
	Status     enums.Status         `json:"status"`
}

// PermissionGroup 是快照中的权限组
type PermissionGroup struct {
	ID     int64        `json:"id"`
	Code   string       `json:"code"`
	Name   string       `json:"name"`
	Status enums.Status `json:"status"`
}

// Snapshot 是某一时刻 RBAC 六张表的全量快照
type Snapshot struct {
	Schema           int               `json:"schema"`
	Version          string            `json:"version"`
	BuiltAt          int64             `json:"builtAt"`
	Roles            []Role            `json:"roles"`
	Permissions      []Permission      `json:"permissions"`
	PermissionGroups []PermissionGroup `json:"permissionGroups"`
	UserRoles        map[int64][]int64 `json:"userRoles"`
	RoleGroups       map[int64][]int64 `json:"roleGroups"`
	GroupPermissions map[int64][]int64 `json:"groupPermissions"`

	once  sync.Once
	index *index
}

// Build 由六表数据构建快照；version 与 builtAt 仅用于观测与排障
func Build(version string, builtAt int64, graph *repo_contract.AuthGraph) *Snapshot {
	snap := &Snapshot{
		Schema:           Schema,
		Version:          version,
		BuiltAt:          builtAt,
		Roles:            make([]Role, 0, len(graph.Roles)),
		Permissions:      make([]Permission, 0, len(graph.Permissions)),
		PermissionGroups: make([]PermissionGroup, 0, len(graph.PermissionGroups)),
		UserRoles:        graph.UserRoles,
		RoleGroups:       graph.RoleGroups,
		GroupPermissions: graph.GroupPermissions,
	}

	for _, role := range graph.Roles {
		snap.Roles = append(snap.Roles, Role{
			ID:      role.ID,
			Code:    role.Code,
			Builtin: role.Builtin,
			Status:  role.Status,
		})
	}
	for _, permission := range graph.Permissions {
		snap.Permissions = append(snap.Permissions, Permission{
			ID:         permission.ID,
			Code:       permission.Code,
			Name:       permission.Name,
			Type:       permission.Type,
			UIPath:     permission.UIPath,
			HTTPMethod: permission.HTTPMethod,
			HTTPPath:   permission.HTTPPath,
			Status:     permission.Status,
		})
	}
	for _, group := range graph.PermissionGroups {
		snap.PermissionGroups = append(snap.PermissionGroups, PermissionGroup{
			ID:     group.ID,
			Code:   group.Code,
			Name:   group.Name,
			Status: group.Status,
		})
	}

	return snap
}

// Encode 将快照序列化为投影载荷
func Encode(snap *Snapshot) ([]byte, error) {
	payload, err := json.Marshal(snap)
	if err != nil {
		return nil, fmt.Errorf("序列化 RBAC 授权快照失败: %w", err)
	}
	return payload, nil
}

// Decode 反序列化投影载荷；结构版本不一致时返回错误，由调用方按未命中处理
func Decode(payload []byte) (*Snapshot, error) {
	var snap Snapshot
	if err := json.Unmarshal(payload, &snap); err != nil {
		return nil, fmt.Errorf("反序列化 RBAC 授权快照失败: %w", err)
	}
	if snap.Schema != Schema {
		return nil, fmt.Errorf("RBAC 授权快照结构版本不匹配: 期望 %d，实际 %d", Schema, snap.Schema)
	}
	return &snap, nil
}

// RoleCount 返回快照中的角色数
func (s *Snapshot) RoleCount() int {
	return len(s.Roles)
}

// PermissionCount 返回快照中的权限数
func (s *Snapshot) PermissionCount() int {
	return len(s.Permissions)
}

// PermissionGroupCount 返回快照中的权限组数
func (s *Snapshot) PermissionGroupCount() int {
	return len(s.PermissionGroups)
}

// RelationCount 返回快照中三类关系的条目数之和
func (s *Snapshot) RelationCount() int {
	return linkCount(s.UserRoles) + linkCount(s.RoleGroups) + linkCount(s.GroupPermissions)
}

// linkCount 统计 ID → ID 列表的总条目数
func linkCount(links map[int64][]int64) int {
	total := 0
	for _, ids := range links {
		total += len(ids)
	}
	return total
}

// ensureIndex 返回求值索引；首次调用时构建，之后复用
func (s *Snapshot) ensureIndex() *index {
	s.once.Do(func() {
		s.index = buildIndex(s)
	})
	return s.index
}

// PrebuildIndex 预构建求值索引；幂等。
// 供后台同步在发布快照前调用，把索引构建成本从请求路径挪到后台。
func (s *Snapshot) PrebuildIndex() {
	if s == nil {
		return
	}
	s.ensureIndex()
}
