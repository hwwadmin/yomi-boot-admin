// Package orm 是本仓库自己的表约定与仓储辅助。
//
// yomi-boot v1.0.2 起不再预置表结构，也不再提供通用 BaseModel / BaseRepository。
// 管理后台各域表仍共用同一套主键、审计字段与软删除列，这些约定收在这里，由各 domain 的仓储复用。
package orm
