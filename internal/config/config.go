package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置对象
type Config struct {
	Email        string        `yaml:"email"`
	Certificates []Certificate `yaml:"certificates"`

	// 新版：多 DNS Provider 配置（推荐）
	// 每个 key 是一个自定义名称，证书配置通过 dns_provider 字段引用
	DNSProviders map[string]DNSProviderEntry `yaml:"dns_providers"`

	// 旧版：全局单一 DNS Provider 配置（向后兼容）
	// 如果同时配置了 dns_providers 和 dns_provider_config，优先使用 dns_providers
	DNSProviderConfig map[string]string `yaml:"dns_provider_config"`

	Apisix ApisixConfig `yaml:"apisix"`

	DataDir        string `yaml:"data_dir"`
	CronSchedule   string `yaml:"cron_schedule"`
	LetsEncryptEnv string `yaml:"lets_encrypt_env"`
}

// DNSProviderEntry 定义一个命名的 DNS Provider 配置
type DNSProviderEntry struct {
	Type string            `yaml:"type"` // provider 类型，如 alidns, cloudflare
	Env  map[string]string `yaml:"env"`  // 该 provider 所需的环境变量
}

// Certificate 定义一组需要申请证书的域名及其 DNS 服务商配置
type Certificate struct {
	Domains               []string `yaml:"domains"`
	DNSProvider           string   `yaml:"dns_provider"`
	RenewBeforeExpiryDays int      `yaml:"renew_before_expiry_days"`
}

type ApisixConfig struct {
	AdminURL string `yaml:"admin_url"`
	AdminKey string `yaml:"admin_key"`
}

// ResolveDNSProvider 根据证书配置中的 dns_provider 字段，解析出实际的 provider 类型和环境变量
// 查找顺序：
//  1. 在 dns_providers 中按名称匹配（新版格式）
//  2. 将 dns_provider 视为 provider 类型，使用全局 dns_provider_config（旧版兼容）
func (c *Config) ResolveDNSProvider(providerName string) (providerType string, env map[string]string, err error) {
	// 优先从新版 dns_providers 中查找
	if c.DNSProviders != nil {
		if entry, ok := c.DNSProviders[providerName]; ok {
			return entry.Type, entry.Env, nil
		}
	}

	// 向后兼容：将 providerName 作为 provider 类型，使用全局配置
	if c.DNSProviderConfig != nil {
		return providerName, c.DNSProviderConfig, nil
	}

	return "", nil, fmt.Errorf("DNS provider %q 未找到：请在 dns_providers 中定义该名称，或配置全局 dns_provider_config", providerName)
}

// LoadConfig 读取并解析 YAML 配置文件
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Defaults
	if cfg.DataDir == "" {
		cfg.DataDir = "./data"
	}
	if cfg.CronSchedule == "" {
		cfg.CronSchedule = "0 0 * * *" // 默认每天凌晨0点执行
	}

	// 为每个证书配置设置默认续期时间
	for i := range cfg.Certificates {
		if cfg.Certificates[i].RenewBeforeExpiryDays <= 0 {
			cfg.Certificates[i].RenewBeforeExpiryDays = 80 // 默认提前80天续期
		}
	}

	if cfg.LetsEncryptEnv == "" {
		cfg.LetsEncryptEnv = "production"
	}

	// 校验证书配置中引用的 dns_provider 是否存在
	for _, certCfg := range cfg.Certificates {
		if _, _, err := cfg.ResolveDNSProvider(certCfg.DNSProvider); err != nil {
			return nil, fmt.Errorf("证书 %v 的配置无效: %v", certCfg.Domains, err)
		}
	}

	return &cfg, nil
}
