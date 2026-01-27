# APISIX ACME Automatic renewal

这是一个使用 Go 编写的自动 https 证书续期服务，适配 APISIX 网关。

## 功能
1. 自动使用 Let's Encrypt 申请/续期证书。
2. 使用云厂商 DNS API 自动添加验证 TXT 记录。
3. 申请成功后自动将证书推送到 APISIX Admin API。
4. 每天检查一次证书过期时间，默认小于 80 天自动续期。

## 技术栈
- Go 1.23+
- [lego](https://github.com/go-acme/lego) (ACME 客户端)

## 支持的 DNS 厂商及配置

在 `config.yaml` 的 `dns_provider_config` 块中配置对应的环境变量。

| 厂商 | dns_provider | 必需环境变量 | 说明 |
|------|--------------|--------------|------|
| **阿里云 (Aliyun)** | `alidns` | `ALICLOUD_ACCESS_KEY`<br>`ALICLOUD_SECRET_KEY` | 阿里云的 AccessKey ID 和 Secret |
| **腾讯云 (Tencent Cloud)** | `tencentcloud` | `TENCENTCLOUD_SECRET_ID`<br>`TENCENTCLOUD_SECRET_KEY` | 腾讯云 API 密钥 |
| **DNSPod (国内版)** | `dnspod` | `DNSPOD_API_KEY` | 格式为 `ID,Token` (例如 `12345,abcdef...`) |
| **华为云 (Huawei Cloud)** | `huaweicloud` | `HUAWEICLOUD_ACCESS_KEY_ID`<br>`HUAWEICLOUD_SECRET_ACCESS_KEY`<br>`HUAWEICLOUD_REGION` | Region 例如 `cn-north-4` |
| **Cloudflare** | `cloudflare` | `CLOUDFLARE_DNS_API_TOKEN` | 推荐使用 API Token |

*更多厂商支持请参考 [Lego DNS Providers](https://go-acme.github.io/lego/dns/) 文档，本项目代码已预埋扩展接口，修改 `internal/dns/factory.go` 即可添加新厂商。*

## 配置文件示例 (config.yaml)

```yaml
email: "your-email@example.com"

# 证书配置列表
certificates:
  - domains: [ "example.com", "*.example.com" ]
    dns_provider: "alidns"
    # 证书过期前多少天进行续期，默认 80 天
    # 建议每个申请的证书使用不同的过期时间，防止集中申请触发限流
    renew_before_expiry_days: 80

  - domains: [ "other-domain.com" ]
    dns_provider: "tencentcloud"
    renew_before_expiry_days: 75

# 全局环境变量配置（集合所有厂商需要的 Key）
dns_provider_config:
  # Aliyun
  ALICLOUD_ACCESS_KEY: "LTAI4..."
  ALICLOUD_SECRET_KEY: "secret..."
  
  # Tencent Cloud
  TENCENTCLOUD_SECRET_ID: "AKID..."
  TENCENTCLOUD_SECRET_KEY: "secret..."

apisix:
  admin_url: "http://apisix:9180" # APISIX Admin API 地址
  admin_key: "your-admin-token"   # APISIX Admin Token

data_dir: "./data"
cron_schedule: "0 0 * * *" # 每天 0点 执行
lets_encrypt_env: "production" # "staging" 用于测试，"production" 用于生产
```

## 快速开始

### 1. 准备环境
确保 APISIX 已经运行，并且知道它所在的 Docker 网络名称（默认为 `apisix`）。

### 2. 启动服务
使用 Docker Compose 启动：

```bash
docker-compose up -d --build
```

### 3. 数据持久化
`data/` 目录会保存：
- `user.json`: ACME 账户凭证（重要，请备份）
- `*.crt`, `*.key`: 申请到的证书文件

## 开发运行
```bash
go mod tidy
go run main.go -c config.yaml
```