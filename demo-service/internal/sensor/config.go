package sensor

import "demo-service/pkg/configclient"

// SensorConfig 光电传感器配置契约
type SensorConfig struct {
	WavelengthNm int    // sensor.wavelength_nm — 红外波长 (nm)
	FPS          int    // sensor.fps — 帧率
	Resolution   string // sensor.resolution — 分辨率
	Sensitivity  int    // sensor.sensitivity — 灵敏度 (1-10)
	Mode         string // sensor.mode — 工作模式: day/night/thermal
}

// DefaultSensorConfig 返回默认配置
func DefaultSensorConfig() SensorConfig {
	return SensorConfig{
		WavelengthNm: 1550,
		FPS:          30,
		Resolution:   "1280x720",
		Sensitivity:  5,
		Mode:         "day",
	}
}

// ParseSensorConfig 从 ConfigClient 解析类型安全的传感器配置
func ParseSensorConfig(client *configclient.ConfigClient) SensorConfig {
	cfg := DefaultSensorConfig()

	cfg.WavelengthNm = client.GetInt("sensor.wavelength_nm", cfg.WavelengthNm, 400, 14000)
	cfg.FPS = client.GetInt("sensor.fps", cfg.FPS, 1, 240)
	cfg.Resolution = client.GetEnum("sensor.resolution", cfg.Resolution, []string{"640x480", "1280x720", "1920x1080", "3840x2160"})
	cfg.Sensitivity = client.GetInt("sensor.sensitivity", cfg.Sensitivity, 1, 10)
	cfg.Mode = client.GetEnum("sensor.mode", cfg.Mode, []string{"day", "night", "thermal"})

	return cfg
}

// KnownKeys 返回本服务认识的所有配置 key
func KnownKeys() []string {
	return []string{"sensor.wavelength_nm", "sensor.fps", "sensor.resolution", "sensor.sensitivity", "sensor.mode"}
}
