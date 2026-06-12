package navigation

import (
	"strings"

	"demo-service/pkg/configclient"
)

// NavConfig 船舶航海配置契约
type NavConfig struct {
	UpdateIntervalS int      // nav.update_interval_s — GPS 更新间隔 (秒)
	SpeedKnots      int      // nav.speed_knots — 巡航速度 (节)
	HeadingDeg      int      // nav.heading_deg — 航向 (度)
	Destination     string   // nav.destination — 目的地
	AutoPilot       bool     // nav.auto_pilot — 自动导航
	Waypoints       []string // nav.waypoints — 航点列表
}

// DefaultNavConfig 返回默认配置
func DefaultNavConfig() NavConfig {
	return NavConfig{
		UpdateIntervalS: 5,
		SpeedKnots:      20,
		HeadingDeg:      0,
		Destination:     "Shanghai",
		AutoPilot:       false,
		Waypoints:       []string{"PortA", "PortB", "PortC"},
	}
}

// ParseNavConfig 从 ConfigClient 解析类型安全的航海配置
func ParseNavConfig(client *configclient.ConfigClient) NavConfig {
	cfg := DefaultNavConfig()

	cfg.UpdateIntervalS = client.GetInt("nav.update_interval_s", cfg.UpdateIntervalS, 1, 300)
	cfg.SpeedKnots = client.GetInt("nav.speed_knots", cfg.SpeedKnots, 1, 60)
	cfg.HeadingDeg = client.GetInt("nav.heading_deg", cfg.HeadingDeg, 0, 359)
	cfg.Destination = client.GetString("nav.destination", cfg.Destination)

	autoStr := client.GetString("nav.auto_pilot", "false")
	cfg.AutoPilot = autoStr == "true"

	wpStr := client.GetString("nav.waypoints", strings.Join(cfg.Waypoints, ","))
	parts := strings.Split(wpStr, ",")
	cfg.Waypoints = nil
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			cfg.Waypoints = append(cfg.Waypoints, p)
		}
	}
	if len(cfg.Waypoints) == 0 {
		cfg.Waypoints = DefaultNavConfig().Waypoints
	}

	return cfg
}

// KnownKeys 返回本服务认识的所有配置 key
func KnownKeys() []string {
	return []string{"nav.update_interval_s", "nav.speed_knots", "nav.heading_deg", "nav.destination", "nav.auto_pilot", "nav.waypoints"}
}
