package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置对象
type Config struct {
	Email             string            `yaml:"email"`
	Domains           []string          `yaml:"domains"`
	DNSProvider       string            `yaml:"dns_provider"`        // 例如: "alidns", "cloudflare", "dnspod"
	DNSProviderConfig map[string]string `yaml:"dns_provider_config"` // 对应 Provider 的环境变量映射

	Apisix ApisixConfig `yaml:"apisix"`

	DataDir        string `yaml:"data_dir"`
	CronSchedule   string `yaml:"cron_schedule"`
	LetsEncryptEnv string `yaml:"lets_encrypt_env"`
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

	if cfg.LetsEncryptEnv == "" {
		cfg.LetsEncryptEnv = "production"
	}

	return &cfg, nil
}
