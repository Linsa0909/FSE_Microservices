package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"demo-service/internal/navigation"
	"demo-service/internal/radar"
	"demo-service/internal/sensor"
	"demo-service/pkg/configclient"

	"github.com/gin-gonic/gin"
)

func main() {
	serviceType := getEnv("SERVICE_TYPE", "default")
	serviceName := getEnv("CONFIG_SERVICE", "order-service")
	env := getEnv("CONFIG_ENV", "dev")
	configURL := getEnv("CONFIG_URL", "http://localhost:8080")

	log.Printf("[demo-service] 启动中 — type=%s service=%s env=%s config-center=%s",
		serviceType, serviceName, env, configURL)

	// 1. 拉取配置
	client := configclient.New(configURL, serviceName, env)
	if err := client.PullConfig(); err != nil {
		log.Printf("[demo-service] ⚠️ %v — 使用全部默认配置", err)
	}

	// 2. 确定端口
	port := determinePort(serviceType)

	// 3. 启动 Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4. 按服务类型分发
	switch serviceType {
	case "radar":
		runRadarService(ctx, r, client)
	case "sensor":
		runSensorService(ctx, r, client)
	case "navigation":
		runNavigationService(ctx, r, client)
	default:
		runDefaultService(r, client, serviceName, env)
	}

	// 5. 公共 /health 端点
	registerHealth(r, client)

	// 6. 优雅关闭
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("[demo-service] 收到终止信号，正在关闭...")
		cancel()
		os.Exit(0)
	}()

	log.Printf("[demo-service] ✅ 服务启动于 :%s (type=%s)", port, serviceType)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// 各服务启动器
// ---------------------------------------------------------------------------

func runRadarService(ctx context.Context, r *gin.Engine, client *configclient.ConfigClient) {
	cfg := radar.ParseRadarConfig(client)
	configclient.PrintDiagnostics("radar-service", client.GetRawConfig(), radar.KnownKeys())

	svc := radar.NewRadarService(cfg)
	svc.Start(ctx)

	// 通过环境变量 PORT 覆盖 server.port (如有)
	applyPortOverride(&cfg.ScanRateHz) // no-op for now, radar doesn't have port config

	radar.RegisterRoutes(&r.RouterGroup, svc)
}

func runSensorService(ctx context.Context, r *gin.Engine, client *configclient.ConfigClient) {
	cfg := sensor.ParseSensorConfig(client)
	configclient.PrintDiagnostics("sensor-service", client.GetRawConfig(), sensor.KnownKeys())

	svc := sensor.NewSensorService(cfg)
	svc.Start(ctx)

	sensor.RegisterRoutes(&r.RouterGroup, svc)
}

func runNavigationService(ctx context.Context, r *gin.Engine, client *configclient.ConfigClient) {
	cfg := navigation.ParseNavConfig(client)
	configclient.PrintDiagnostics("navigation-service", client.GetRawConfig(), navigation.KnownKeys())

	svc := navigation.NewNavService(cfg)
	svc.Start(ctx)

	navigation.RegisterRoutes(&r.RouterGroup, svc)
}

func runDefaultService(r *gin.Engine, client *configclient.ConfigClient, service, env string) {
	raw := client.GetRawConfig()
	knownKeys := []string{"db.url", "server.port", "log.level"}
	configclient.PrintDiagnostics("default-service", raw, knownKeys)

	// 向后兼容: 原 GET / 端点
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": service,
			"env":     env,
			"config": gin.H{
				"server.port": client.GetInt("server.port", 3000, 1, 65535),
				"db.url":      client.GetString("db.url", "localhost:3306"),
				"log.level":   client.GetEnum("log.level", "info", []string{"debug", "info", "warn", "error"}),
			},
			"raw": raw,
		})
	})
}

// ---------------------------------------------------------------------------
// 公共端点
// ---------------------------------------------------------------------------

func registerHealth(r *gin.Engine, client *configclient.ConfigClient) {
	r.GET("/health", func(c *gin.Context) {
		fromCenter := len(client.GetRawConfig()) > 0
		status := "degraded"
		if fromCenter {
			status = "ok"
		}
		c.JSON(http.StatusOK, gin.H{
			"status":              status,
			"config_from_center":  fromCenter,
		})
	})
}

// ---------------------------------------------------------------------------
// 工具函数
// ---------------------------------------------------------------------------

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func determinePort(serviceType string) string {
	// 领域专用端口环境变量
	envMap := map[string]string{
		"radar":      "RADAR_PORT",
		"sensor":     "SENSOR_PORT",
		"navigation": "NAV_PORT",
	}
	if envKey, ok := envMap[serviceType]; ok {
		if p := os.Getenv(envKey); p != "" {
			return p
		}
	}
	// 通用 PORT
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	// 默认端口
	defaults := map[string]string{
		"radar": "3001", "sensor": "3002", "navigation": "3003", "default": "3000",
	}
	if d, ok := defaults[serviceType]; ok {
		return d
	}
	return "3000"
}

func applyPortOverride(_ *int) {
	// port override 已在 determinePort 中处理
	// 此处为将来扩展预留
	fmt.Print("") // suppress unused import
}

// corsMiddleware 允许前端跨域调用各领域服务的 /health 端点
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Accept")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
