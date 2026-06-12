package sensor

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册传感器服务路由
func RegisterRoutes(rg *gin.RouterGroup, svc *SensorService) {
	rg.GET("/sensor/status", handleSensorStatus(svc))
	rg.GET("/sensor/frame", handleSensorFrame(svc))
	rg.GET("/sensor/frames", handleSensorFrames(svc))
	rg.POST("/sensor/calibrate", handleSensorCalibrate(svc))
}

func handleSensorStatus(svc *SensorService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := svc.GetCfg()
		c.JSON(http.StatusOK, gin.H{
			"service": "sensor-service",
			"config": gin.H{
				"wavelength_nm": cfg.WavelengthNm,
				"fps":           cfg.FPS,
				"resolution":    cfg.Resolution,
				"sensitivity":   cfg.Sensitivity,
				"mode":          cfg.Mode,
			},
			"uptime_seconds":  svc.UptimeSeconds(),
			"total_frames":    svc.TotalFrames(),
			"total_detections": svc.TotalDetections(),
			"buffer_size":     svc.GetBuffer().Size(),
		})
	}
}

func handleSensorFrame(svc *SensorService) gin.HandlerFunc {
	return func(c *gin.Context) {
		frame := svc.GetBuffer().Latest()
		if frame == nil {
			c.JSON(http.StatusOK, gin.H{"frame": nil, "message": "no frames captured yet"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"frame": frame})
	}
}

func handleSensorFrames(svc *SensorService) gin.HandlerFunc {
	return func(c *gin.Context) {
		countStr := c.DefaultQuery("count", "10")
		count, err := strconv.Atoi(countStr)
		if err != nil || count < 1 {
			count = 10
		}
		if count > 60 {
			count = 60
		}
		frames := svc.GetBuffer().LastN(count)
		// 每帧只返回摘要 (frame_id + detection count + timestamp)
		type FrameSummary struct {
			FrameID     int64  `json:"frame_id"`
			Timestamp   string `json:"timestamp"`
			NumDet      int    `json:"num_detections"`
			Mode        string `json:"mode"`
		}
		summaries := make([]FrameSummary, len(frames))
		for i, f := range frames {
			summaries[i] = FrameSummary{
				FrameID:   f.FrameID,
				Timestamp: f.Timestamp.Format("15:04:05.000"),
				NumDet:    len(f.Detections),
				Mode:      f.Mode,
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"frames": summaries,
			"count":  len(summaries),
		})
	}
}

func handleSensorCalibrate(svc *SensorService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 模拟校准参数返回
		c.JSON(http.StatusOK, gin.H{
			"message": "calibration complete",
			"params": gin.H{
				"dark_current":  0.023,
				"gain_level":    svc.GetCfg().Sensitivity,
				"exposure_ms":   33.3,
				"temperature_c": 22.5,
				"wavelength_nm": svc.GetCfg().WavelengthNm,
			},
		})
	}
}
