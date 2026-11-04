// Package config 编排层业务配置段。
//
// 只承载本服务自有的 YAML 段；基础设施配置归 yomi-boot 各插件所有。
package config

import (
	"github.com/hwwadmin/yomi-boot/component/core/config"
)

// SectionKey 是业务配置段 YAML 根 key
const SectionKey = "biz"

// BootConfig 是 biz 配置段的 YAML 绑定
type BootConfig struct {
	Captcha    CaptchaConfig    `mapstructure:"captcha"`
	Rbac       RbacConfig       `mapstructure:"rbac"`
	LoginGuard LoginGuardConfig `mapstructure:"login_guard"`
	Docs       DocsConfig       `mapstructure:"docs"`
	Audit      AuditConfig      `mapstructure:"audit"`
	Auth       AuthConfig       `mapstructure:"auth"`
}

// ConfigSection 返回 biz 配置段
func ConfigSection() config.Section {
	return config.NewSection(SectionKey, &BootConfig{}, func(c *BootConfig, _ *config.Config) error {
		if err := c.Captcha.Validate(); err != nil {
			return err
		}
		if err := c.Rbac.Validate(); err != nil {
			return err
		}
		if err := c.LoginGuard.Validate(); err != nil {
			return err
		}
		if err := c.Docs.Validate(); err != nil {
			return err
		}
		if err := c.Audit.Validate(); err != nil {
			return err
		}
		return c.Auth.Validate()
	})
}
