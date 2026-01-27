# APISIX ACME Service (Aliyun DNS)

这是一个使用 Go 编写的自动 https 证书续期服务，适配APISIX 网关。

## 功能
1. 自动使用 Let's Encrypt 申请/续期证书。
2. 使用 Aliyun DNS API 自动添加验证 TXT 记录。
3. 申请成功后自动将证书推送到 APISIX Admin API。
4. 每天检查一次证书过期时间，默认小于80天自动续期。

## 技术栈
- Go 1.25+
- [lego](https://github.com/go-acme/lego) (ACME 客户端)

## 快速开始

### 1. 准备配置文件
在本地创建一个 `config.yaml` 文件，并填入你的信息：

```yaml
email: "your-email@example.com"

# 证书配置列表
certificates:
  - domains: [ "example.com" ]
    dns_provider: "alidns"
    # 证书过期前多少天进行续期，默认 80 天
  #    renew_before_expiry_days: 80
  - domains: [ "example.com" ]
    dns_provider: "alidns"
#    renew_before_expiry_days: 89
dns_provider_config:
  ALICLOUD_ACCESS_KEY: "LTAI4..."
  ALICLOUD_SECRET_KEY: "secret..."

apisix:
  admin_url: "http://apisix:9180"
  admin_key: "your-admin-token"

data_dir: "./data"
cron_schedule: "0 0 * * *"
lets_encrypt_env: "production"
```

### 2. 运行

#### 使用 Docker Compose

确保 APISIX 已经运行，并且知道它所在的 Docker 网络名称。

1. 修改 `config.yaml` 中的 `apisix_admin_url`。如果容器在同一个网络下，通常可以使用容器名访问，例如：`http://apisix:9180`。

2. 启动服务：
```bash
docker-compose up -d --build
```
