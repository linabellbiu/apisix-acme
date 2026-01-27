package apisix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"apisix-acme/internal/config"
)

type Client struct {
	BaseURL string
	APIKey  string
}

func NewClient(cfg config.ApisixConfig) *Client {
	return &Client{
		BaseURL: strings.TrimRight(cfg.AdminURL, "/"),
		APIKey:  cfg.AdminKey,
	}
}

func (c *Client) UpdateSSL(domains []string, cert, key []byte) error {
	// Create an ID based on the primary domain
	sslID := strings.ReplaceAll(domains[0], ".", "_")

	url := fmt.Sprintf("%s/apisix/admin/ssls/%s", c.BaseURL, sslID)

	payload := map[string]interface{}{
		"cert": string(cert),
		"key":  string(key),
		"snis": domains,
		"id":   sslID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", c.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("调用 APISIX 失败: %s - %s", resp.Status, string(respBody))
	}

	return nil
}
