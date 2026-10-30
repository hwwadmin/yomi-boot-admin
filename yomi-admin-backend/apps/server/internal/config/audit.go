package config

// AuditConfig 操作审计日志配置
type AuditConfig struct {
	// Enabled 是审计写入开关，默认开启；关闭后仍可查询历史
	Enabled *bool `mapstructure:"enabled"`
}

// EnabledOrDefault 审计写入是否开启，默认 true
func (c AuditConfig) EnabledOrDefault() bool {
	return c.Enabled == nil || *c.Enabled
}

// Validate 校验审计配置
func (c AuditConfig) Validate() error {
	return nil
}
