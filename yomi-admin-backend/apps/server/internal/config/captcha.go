// Package config 编排层业务配置段。
//
// 只承载本服务自有的 YAML 段；基础设施配置归 yomi-boot 各插件所有。
package config

import (
	"strings"
	"time"

	"gitee.com/hq_project/yomi-boot/component/common/captcha"
	"gitee.com/hq_project/yomi-boot/component/core/config"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// SectionKey 是业务配置段 YAML 根 key
const SectionKey = "biz"

// defaultCaptchaKeyPrefix 是登录验证码默认 Redis 键前缀
const defaultCaptchaKeyPrefix = "yomi:captcha:login"

// defaultCaptchaRefreshInterval 是同 IP 生成验证码的默认最小间隔
const defaultCaptchaRefreshInterval = 10 * time.Second

// maxCaptchaRefreshIntervalSeconds 是刷新间隔上限（秒）
const maxCaptchaRefreshIntervalSeconds = 3600

// BootConfig 是 biz 配置段的 YAML 绑定
type BootConfig struct {
	Captcha CaptchaConfig `mapstructure:"captcha"`
	Rbac    RbacConfig    `mapstructure:"rbac"`
}

// CaptchaConfig 是登录图像验证码配置；零值字段按默认值处理
type CaptchaConfig struct {
	Enabled                *bool  `mapstructure:"enabled"`                  // Enabled 是登录是否强制校验图像验证码，未配置默认 true
	Driver                 string `mapstructure:"driver"`                   // Driver 是验证码驱动：digit / string / math，空则 digit
	Length                 int    `mapstructure:"length"`                   // Length 是答案字符数，digit/string 有效，0 则 4
	Width                  int    `mapstructure:"width"`                    // Width 是图片宽度（像素），0 则 240
	Height                 int    `mapstructure:"height"`                   // Height 是图片高度（像素），0 则 80
	TTLSeconds             int    `mapstructure:"ttl_seconds"`              // TTLSeconds 是验证码有效期（秒），0 则 300
	KeyPrefix              string `mapstructure:"key_prefix"`               // KeyPrefix 是 Redis 键前缀，空则 yomi:captcha:login
	RefreshIntervalSeconds *int   `mapstructure:"refresh_interval_seconds"` // RefreshIntervalSeconds 是同 IP 生成验证码的最小间隔（秒），未配置默认 10，0 表示不限流
}

// EnabledOrDefault 返回登录是否强制校验图像验证码，未配置默认 true
func (c CaptchaConfig) EnabledOrDefault() bool {
	return c.Enabled == nil || *c.Enabled
}

// RefreshInterval 返回同 IP 生成验证码的最小间隔，未配置默认 10s
func (c CaptchaConfig) RefreshInterval() time.Duration {
	if c.RefreshIntervalSeconds == nil {
		return defaultCaptchaRefreshInterval
	}
	return time.Duration(*c.RefreshIntervalSeconds) * time.Second
}

// KeyPrefixOrDefault 返回 Redis 键前缀，空则使用登录场景默认前缀
func (c CaptchaConfig) KeyPrefixOrDefault() string {
	prefix := strings.TrimSpace(c.KeyPrefix)
	if prefix == "" {
		return defaultCaptchaKeyPrefix
	}
	return prefix
}

// ToConfig 转换为 yomi-boot captcha 运行时配置；零值字段由 captcha 包填默认值
func (c CaptchaConfig) ToConfig() captcha.Config {
	return captcha.Config{
		Driver: captcha.Driver(c.Driver),
		Length: c.Length,
		Width:  c.Width,
		Height: c.Height,
		TTL:    time.Duration(c.TTLSeconds) * time.Second,
	}
}

// Validate 校验验证码配置语义
func (c CaptchaConfig) Validate() error {
	if c.RefreshIntervalSeconds != nil {
		seconds := *c.RefreshIntervalSeconds
		if seconds < 0 || seconds > maxCaptchaRefreshIntervalSeconds {
			return errors.Newf("%s.captcha: refresh_interval_seconds 必须在 0..%d 之间", SectionKey, maxCaptchaRefreshIntervalSeconds)
		}
	}
	if err := c.ToConfig().Validate(); err != nil {
		return errors.Wrap(err, "%s.captcha", SectionKey)
	}
	return nil
}

// ConfigSection 返回 biz 配置段
func ConfigSection() config.Section {
	return config.NewSection(SectionKey, &BootConfig{}, func(c *BootConfig, _ *config.Config) error {
		if err := c.Captcha.Validate(); err != nil {
			return err
		}
		return c.Rbac.Validate()
	})
}
