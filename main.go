package main

import (
	"flag"
	"log"

	"apisix-acme/internal/cert"
	"apisix-acme/internal/config"
)

func main() {
	configPath := flag.String("c", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}

	log.Printf("服务已启动 | 域名: %v | DNS服务商: %s", cfg.Domains, cfg.DNSProvider)

	// 初始化并运行证书管理器
	manager := cert.NewManager(cfg)
	manager.Run()
}
