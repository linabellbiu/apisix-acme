package apisix

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/linabellbiu/apisix-acme/internal/config"
)

type Client struct {
	client *resty.Client
}

func NewClient(cfg config.ApisixConfig) *Client {
	r := resty.New()
	r.SetTimeout(30 * time.Second)
	r.SetBaseURL(strings.TrimRight(cfg.AdminURL, "/"))
	r.SetHeader("Content-Type", "application/json")
	r.SetHeader("X-API-KEY", cfg.AdminKey)

	return &Client{
		client: r,
	}
}

func (c *Client) UpdateSSL(domains []string, cert, key []byte) error {
	// Create an ID based on the primary domain
	sslID := strings.ReplaceAll(domains[0], ".", "_")
	sslID = strings.ReplaceAll(sslID, "*", "wildcard")

	payload := map[string]interface{}{
		"cert": string(cert),
		"key":  string(key),
		"snis": domains,
		"id":   sslID,
	}

	resp, err := c.client.R().
		SetBody(payload).
		Put(fmt.Sprintf("/apisix/admin/ssls/%s", sslID))

	if err != nil {
		return fmt.Errorf("请求 APISIX 失败: %v", err)
	}

	if resp.IsError() {
		return fmt.Errorf("调用 APISIX 失败: %s - %s", resp.Status(), resp.String())
	}

	return nil
}
