package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ConfigClient 配置客户端 SDK 雏形
type ConfigClient struct {
	baseURL    string
	service    string
	env        string
	httpClient *http.Client
	config     map[string]string
	maxRetries int
	retryDelay time.Duration
}

// NewConfigClient 创建配置客户端
func NewConfigClient(baseURL, service, env string) *ConfigClient {
	return &ConfigClient{
		baseURL:    baseURL,
		service:    service,
		env:        env,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		config:     make(map[string]string),
		maxRetries: 5,
		retryDelay: time.Second,
	}
}

// pullConfig 拉取线上配置（含重试机制，容忍后端未就绪）
func (c *ConfigClient) pullConfig() error {
	url := fmt.Sprintf("%s/api/configs/%s/%s/published", c.baseURL, c.service, c.env)
	log.Printf("[ConfigClient] 正在拉取配置: %s", url)

	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		resp, err := c.httpClient.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("拉取配置失败: %w", err)
			if attempt < c.maxRetries {
				log.Printf("[ConfigClient] 第 %d/%d 次尝试失败: %v, %s 后重试...", attempt, c.maxRetries, err, c.retryDelay)
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
			lastErr = fmt.Errorf("解析配置失败: %w", err)
			if attempt < c.maxRetries {
				log.Printf("[ConfigClient] 第 %d/%d 次尝试解析失败: %v, %s 后重试...", attempt, c.maxRetries, err, c.retryDelay)
				time.Sleep(c.retryDelay)
				continue
			}
			return lastErr
		}

		c.config = result.Config
		log.Printf("[ConfigClient] 拉取成功 (第 %d 次尝试): %d 个配置项", attempt, len(c.config))
		for k, v := range c.config {
			log.Printf("  %s = %s", k, v)
		}
		return nil
	}
	return fmt.Errorf("拉取配置失败 (已重试 %d 次): %w", c.maxRetries, lastErr)
}

// GetConfig 获取本地缓存的配置
func (c *ConfigClient) GetConfig() map[string]string {
	return c.config
}

func main() {
	client := NewConfigClient("http://localhost:8080", "order-service", "dev")

	// 启动时拉取配置
	log.Println("[demo-service] 启动中，拉取配置...")
	if err := client.pullConfig(); err != nil {
		log.Printf("[demo-service] ⚠️ %v — 使用默认端口 3000", err)
	}

	cfg := client.GetConfig()
	port := cfg["server.port"]
	if port == "" {
		port = "3000"
	}

	// 启动 HTTP 服务
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "order-service",
			"config":  client.GetConfig(),
		})
	})

	log.Printf("[demo-service] ✅ 服务启动于 :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
