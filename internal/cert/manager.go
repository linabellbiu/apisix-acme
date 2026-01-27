package cert

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/robfig/cron/v3"

	"apisix-acme/internal/apisix"
	"apisix-acme/internal/config"
	"apisix-acme/internal/dns"
)

// Manager 证书生命周期管理器
type Manager struct {
	Cfg          *config.Config
	ApisixClient *apisix.Client
}

// User implements acme.User
type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// UserStorage helps in persisting user data
type UserStorage struct {
	Email        string
	Registration *registration.Resource
	PrivateKey   []byte
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		Cfg:          cfg,
		ApisixClient: apisix.NewClient(cfg.Apisix),
	}
}

func (m *Manager) Run() {
	// Ensure data directory exists
	if err := os.MkdirAll(m.Cfg.DataDir, 0755); err != nil {
		log.Fatalf("无法创建数据目录: %v", err)
	}

	log.Println("证书管家已启动，正在初始化任务...")

	// 立即执行一次
	m.safeProcess()

	// 启动定时任务
	c := cron.New(
		cron.WithLocation(time.Local),
	)
	if _, err := c.AddFunc(m.Cfg.CronSchedule, func() {
		m.safeProcess()
	}); err != nil {
		log.Fatalf("无法添加定时任务: %v", err)
	}

	c.Start()
	log.Printf("定时任务已启动，Cron 表达式: %s", m.Cfg.CronSchedule)

	// 阻塞主程
	select {}
}

func (m *Manager) safeProcess() {
	var hasError bool
	for _, certCfg := range m.Cfg.Certificates {
		log.Printf("[开始] 处理证书组: %v", certCfg.Domains)
		if err := m.process(certCfg); err != nil {
			log.Printf("[错误] 域名 %v 处理失败: %v", certCfg.Domains, err)
			hasError = true
		} else {
			log.Printf("[成功] 域名 %v 处理完成", certCfg.Domains)
		}
	}

	if hasError {
		log.Println("[完成] 证书检查流程结束，但部分证书处理失败")
	} else {
		log.Println("[完成] 所有证书检查流程成功结束")
	}
}

func (m *Manager) process(certCfg config.Certificate) error {
	// 初始化用户
	user, err := m.getOrCreateUser()
	if err != nil {
		return fmt.Errorf("用户初始化失败: %v", err)
	}

	// 配置 ACME 客户端
	legoConfig := lego.NewConfig(user)
	if m.Cfg.LetsEncryptEnv == "staging" {
		legoConfig.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	} else {
		legoConfig.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	}
	legoConfig.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(legoConfig)
	if err != nil {
		return fmt.Errorf("创建 lego 客户端失败: %v", err)
	}

	// 配置 DNS Provider
	// 使用证书指定的 Provider 类型，配合全局的 Provider 配置
	provider, err := dns.NewDNSProvider(certCfg.DNSProvider, m.Cfg.DNSProviderConfig)
	if err != nil {
		return fmt.Errorf("DNS Provider 初始化失败: %v", err)
	}

	if err := client.Challenge.SetDNS01Provider(provider); err != nil {
		return err
	}

	// 注册账号
	if user.Registration == nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return fmt.Errorf("账号注册失败: %v", err)
		}
		user.Registration = reg
		m.saveUser(user)
	}

	// 检查证书状态
	certPath := filepath.Join(m.Cfg.DataDir, certCfg.Domains[0]+".crt")
	keyPath := filepath.Join(m.Cfg.DataDir, certCfg.Domains[0]+".key")

	needsRenew := true
	if exists(certPath) {
		valid, err := isCertValid(certPath, certCfg.RenewBeforeExpiryDays)
		if err == nil && valid {
			needsRenew = false
			log.Printf("证书有效期充足 (大于%d天)，无需更新", certCfg.RenewBeforeExpiryDays)
		}
	}

	var crt, key []byte

	if needsRenew {
		log.Println("正在申请新证书...")
		req := certificate.ObtainRequest{
			Domains: certCfg.Domains,
			Bundle:  true,
		}
		certRes, err := client.Certificate.Obtain(req)
		if err != nil {
			return fmt.Errorf("证书申请失败: %v", err)
		}

		crt = certRes.Certificate
		key = certRes.PrivateKey

		_ = os.WriteFile(certPath, crt, 0644)
		_ = os.WriteFile(keyPath, key, 0600)
		log.Println("新证书已保存到本地")
	} else {
		crt, _ = os.ReadFile(certPath)
		key, _ = os.ReadFile(keyPath)
	}

	// 更新 APISIX
	log.Println("正在推送证书到 APISIX...")
	if err := m.ApisixClient.UpdateSSL(certCfg.Domains, crt, key); err != nil {
		return fmt.Errorf("更新 APISIX 失败: %v", err)
	}

	log.Println("APISIX 证书更新成功")
	return nil
}

func (m *Manager) getOrCreateUser() (*User, error) {
	path := filepath.Join(m.Cfg.DataDir, "user.json")
	if exists(path) {
		data, _ := os.ReadFile(path)
		var s UserStorage
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, err
		}
		block, _ := pem.Decode(s.PrivateKey)
		k, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return &User{Email: s.Email, Registration: s.Registration, key: k}, nil
	}

	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return &User{Email: m.Cfg.Email, key: k}, nil
}

func (m *Manager) saveUser(u *User) {
	kBytes, _ := x509.MarshalECPrivateKey(u.key.(*ecdsa.PrivateKey))
	pemBlock := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: kBytes})
	s := UserStorage{Email: u.Email, Registration: u.Registration, PrivateKey: pemBlock}
	data, _ := json.MarshalIndent(s, "", "  ")
	_ = os.WriteFile(filepath.Join(m.Cfg.DataDir, "user.json"), data, 0600)
}

func isCertValid(path string, renewBeforeDays int) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return false, fmt.Errorf("无法解析 PEM 格式")
	}
	crt, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, err
	}

	if time.Until(crt.NotAfter) < time.Duration(renewBeforeDays)*24*time.Hour {
		return false, nil
	}
	return true, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
