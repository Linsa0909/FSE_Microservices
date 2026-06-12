package configclient

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// ConfigClient 配置中心客户端 SDK
type ConfigClient struct {
	baseURL    string
	Service    string
	Env        string
	httpClient *http.Client
	rawConfig  map[string]string
	maxRetries int
	retryDelay time.Duration
}

// New 创建一个配置中心客户端
func New(baseURL, service, env string) *ConfigClient {
	return &ConfigClient{
		baseURL:    baseURL,
		Service:    service,
		Env:        env,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		rawConfig:  make(map[string]string),
		maxRetries: 5,
		retryDelay: time.Second,
	}
}

// PullConfig 拉取线上 PublishedData（含重试机制）
func (c *ConfigClient) PullConfig() error {
	url := fmt.Sprintf("%s/api/configs/%s/%s/published", c.baseURL, c.Service, c.Env)
	log.Printf("[ConfigClient] pulling config: %s", url)

	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		resp, err := c.httpClient.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("pull config failed: %w", err)
			if attempt < c.maxRetries {
				log.Printf("[ConfigClient] attempt %d/%d failed: %v, retrying in %s...", attempt, c.maxRetries, err, c.retryDelay)
				time.Sleep(c.retryDelay)
				continue
			}
			return lastErr
		}

		result := struct {
			Config map[string]string `json:"config"`
		}{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("parse config failed: %w", err)
			if attempt < c.maxRetries {
				log.Printf("[ConfigClient] attempt %d/%d parse failed: %v, retrying in %s...", attempt, c.maxRetries, err, c.retryDelay)
				time.Sleep(c.retryDelay)
				continue
			}
			return lastErr
		}

		c.rawConfig = result.Config
		log.Printf("[ConfigClient] pull success (attempt %d): %d keys", attempt, len(c.rawConfig))
		for k, v := range c.rawConfig {
			log.Printf("  %s = %s", k, v)
		}
		return nil
	}
	return fmt.Errorf("pull config failed (retried %d times): %w", c.maxRetries, lastErr)
}

// GetRawConfig 返回原始配置 map
func (c *ConfigClient) GetRawConfig() map[string]string {
	return c.rawConfig
}

// GetString 获取字符串配置值，不存在或为空则返回默认值
func (c *ConfigClient) GetString(key, defaultVal string) string {
	if v, ok := c.rawConfig[key]; ok && v != "" {
		return v
	}
	return defaultVal
}

// GetInt 获取整数配置值，带范围校验
// 校验失败时打印警告并返回默认值（降级启动）
func (c *ConfigClient) GetInt(key string, defaultVal, min, max int) int {
	v, ok := c.rawConfig[key]
	if !ok || v == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(v)
	if err != nil || val < min || val > max {
		log.Printf("[ConfigClient] ⚠️ %s=%q invalid (want %d-%d), using default %d", key, v, min, max, defaultVal)
		return defaultVal
	}
	return val
}

// GetFloat 获取浮点配置值，带范围校验
func (c *ConfigClient) GetFloat(key string, defaultVal, min, max float64) float64 {
	v, ok := c.rawConfig[key]
	if !ok || v == "" {
		return defaultVal
	}
	val, err := strconv.ParseFloat(v, 64)
	if err != nil || val < min || val > max {
		log.Printf("[ConfigClient] ⚠️ %s=%q invalid (want %.1f-%.1f), using default %.1f", key, v, min, max, defaultVal)
		return defaultVal
	}
	return val
}

// GetEnum 获取枚举配置值，不在允许列表中则返回默认值
func (c *ConfigClient) GetEnum(key, defaultVal string, validValues []string) string {
	v, ok := c.rawConfig[key]
	if !ok || v == "" {
		return defaultVal
	}
	for _, valid := range validValues {
		if v == valid {
			return v
		}
	}
	log.Printf("[ConfigClient] ⚠️ %s=%q invalid (want one of %v), using default %q", key, v, validValues, defaultVal)
	return defaultVal
}
