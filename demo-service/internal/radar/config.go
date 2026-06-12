package radar

import "demo-service/pkg/configclient"

// RadarConfig 火控雷达配置契约
type RadarConfig struct {
	ScanRateHz int    // radar.scan_rate_hz — 扫描频率 (Hz)
	RangeKm    int    // radar.range_km — 探测距离 (km)
	MaxTargets int    // radar.max_targets — 最大同时跟踪目标数
	Mode       string // radar.mode — 工作模式: search/track/engage
	Band       string // radar.band — 雷达波段: X/S/C/Ku
}

// DefaultRadarConfig 返回默认配置（配置中心不可用时的降级值）
func DefaultRadarConfig() RadarConfig {
	return RadarConfig{
		ScanRateHz: 60,
		RangeKm:    150,
		MaxTargets: 32,
		Mode:       "search",
		Band:       "X",
	}
}

// ParseRadarConfig 从 ConfigClient 解析类型安全的雷达配置
func ParseRadarConfig(client *configclient.ConfigClient) RadarConfig {
	cfg := DefaultRadarConfig()

	cfg.ScanRateHz = client.GetInt("radar.scan_rate_hz", cfg.ScanRateHz, 1, 300)
	cfg.RangeKm = client.GetInt("radar.range_km", cfg.RangeKm, 1, 500)
	cfg.MaxTargets = client.GetInt("radar.max_targets", cfg.MaxTargets, 1, 256)
	cfg.Mode = client.GetEnum("radar.mode", cfg.Mode, []string{"search", "track", "engage"})
	cfg.Band = client.GetEnum("radar.band", cfg.Band, []string{"X", "S", "C", "Ku"})

	return cfg
}

// KnownKeys 返回本服务认识的所有配置 key（诊断用）
func KnownKeys() []string {
	return []string{"radar.scan_rate_hz", "radar.range_km", "radar.max_targets", "radar.mode", "radar.band"}
}
