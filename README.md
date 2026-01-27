# APISIX ACME Service (Aliyun DNS)

这是一个使用 Go 编写的自动 https 证书续期服务，专门适配 Aliyun DNS (txt记录解析) 和 APISIX 网关。

## 功能
1. 自动使用 Let's Encrypt 申请/续期证书。
2. 使用 Aliyun DNS API 自动添加验证 TXT 记录。
3. 申请成功后自动将证书推送到 APISIX Admin API。
4. 每天检查一次证书过期时间，小于 30 天自动续期。

## 技术栈
- Go 1.23+
- [lego](https://github.com/go-acme/lego) (ACME 客户端)
- Docker
- YAML 配置

## 快速开始

### 1. 准备配置文件
在本地创建一个 `config.yaml` 文件，并填入你的信息：

```yaml
email: "your-email@example.com"
domains:
  - "example.com"
  - "*.example.com"
ali_access_key: "LTAI4......"
ali_secret_key: "secret......"
apisix_admin_url: "http://127.0.0.1:9180"
apisix_admin_key: "your-admin-token"
data_dir: "./data"
run_interval: "24h"
lets_encrypt_env: "production" # 测试时可改为 "staging"
```

### 2. 构建 Docker 镜像

```bash
docker build -t apisix-acme-service .
```

### 3. 运行

#### 使用 Docker Compose (推荐)

确保你的 APISIX 已经运行，并且知道它所在的 Docker 网络名称（默认为 `apisix`）。

1. 修改 `config.yaml` 中的 `apisix_admin_url`。如果容器在同一个网络下，通常可以使用容器名访问，例如：`http://apisix:9180`。

2. 启动服务：
```bash
docker-compose up -d --build
```

#### 使用 Docker CLI
将你的 `config.yaml` 和数据目录挂载到容器中：

```bash
docker run -d \
  --name apisix-acme \
  --restart always \
  --network apisix \
  -v $(pwd)/config.yaml:/root/data/config.yaml \
  -v $(pwd)/data:/root/data \
  apisix-acme-service
```

注意：默认 Dockerfile 中 CMD 指令设定为读取 `/root/data/config.yaml`。请确保护载路径正确。

## 本地运行

```bash
go run main.go -c config.yaml
```
