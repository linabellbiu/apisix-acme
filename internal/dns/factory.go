package dns

import (
	"fmt"
	"os"

	"github.com/go-acme/lego/v4/challenge"

	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/dnspod"
	"github.com/go-acme/lego/v4/providers/dns/huaweicloud"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
)

// NewDNSProvider 基于配置的类型和环境变量映射创建一个 DNS Provider
// 这是一个工厂方法，通过字符串 switch case 来决定实例化哪个 Provider
func NewDNSProvider(providerType string, config map[string]string) (challenge.Provider, error) {
	// Lego 的 Provider 大多通过读取环境变量来初始化
	// 这里将配置文件映射覆盖到环境变量中
	for k, v := range config {
		os.Setenv(k, v)
	}

	switch providerType {
	case "alidns":
		return alidns.NewDNSProvider()
	case "cloudflare":
		return cloudflare.NewDNSProvider()
	case "dnspod":
		return dnspod.NewDNSProvider()
	case "tencentcloud":
		return tencentcloud.NewDNSProvider()
	case "huaweicloud":
		return huaweicloud.NewDNSProvider()
	// Add more providers here as needed
	default:
		return nil, fmt.Errorf("unsupported dns provider: %s", providerType)
	}
}
