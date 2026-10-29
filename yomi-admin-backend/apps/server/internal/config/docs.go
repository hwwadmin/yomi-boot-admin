package config

// DocsConfig API 文档配置
type DocsConfig struct {
	Enabled *bool `mapstructure:"enabled"`
}

// EnabledOrDefault 文档开关，默认关闭；生产环境不暴露文档路由
func (c DocsConfig) EnabledOrDefault() bool {
	return c.Enabled != nil && *c.Enabled
}

// Validate 校验文档配置
func (c DocsConfig) Validate() error {
	return nil
}
