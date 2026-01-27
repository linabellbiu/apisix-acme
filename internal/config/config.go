package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置对象
type Config struct {
	Email             string            `yaml:"email"`
	Certificates      []Certificate     `yaml:"certificates"`
	DNSProviderConfig map[string]string `yaml:"dns_provider_config"`

	Apisix ApisixConfig `yaml:"apisix"`

	DataDir        string `yaml:"data_dir"`
	CronSchedule   string `yaml:"cron_schedule"`
	LetsEncryptEnv string `yaml:"lets_encrypt_env"`
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

	return &cfg, nil
}
