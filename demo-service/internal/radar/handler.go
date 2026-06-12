package radar

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册雷达服务路由
func RegisterRoutes(rg *gin.RouterGroup, svc *RadarService) {
	rg.GET("/radar/status", handleRadarStatus(svc))
	rg.GET("/radar/targets", handleRadarTargets(svc))
	rg.GET("/radar/targets/:id", handleRadarTarget(svc))
	rg.POST("/radar/mode", handleRadarMode(svc))
}

func handleRadarStatus(svc *RadarService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := svc.GetCfg()
		stats := svc.Stats()
		c.JSON(http.StatusOK, gin.H{
			"service": "radar-service",
			"config": gin.H{
				"scan_rate_hz": cfg.ScanRateHz,
				"range_km":     cfg.RangeKm,
				"max_targets":  cfg.MaxTargets,
				"mode":         svc.GetMode(),
				"band":         cfg.Band,
			},
			"uptime_seconds":    svc.UptimeSeconds(),
			"targets_detected":  svc.TotalDetected(),
			"targets_tracked":   stats.Tracking,
			"targets_locked":    stats.Locked,
			"targets_engaged":   stats.Engaged,
			"active_targets":    svc.GetStore().Count(),
		})
	}
}

func handleRadarTargets(svc *RadarService) gin.HandlerFunc {
	return func(c *gin.Context) {
		targets := svc.GetStore().All()
		c.JSON(http.StatusOK, gin.H{
			"targets": targets,
			"count":   len(targets),
		})
	}
}

func handleRadarTarget(svc *RadarService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		t, ok := svc.GetStore().Get(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "target not found: " + id})
			return
		}
		c.JSON(http.StatusOK, gin.H{"target": t})
	}
}

func handleRadarMode(svc *RadarService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Mode string `json:"mode"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.Mode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'mode' in request body"})
			return
		}
		old := svc.GetMode()
		if err := svc.SetMode(body.Mode); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":       "mode switched to " + body.Mode,
			"previous_mode": old,
			"current_mode":  svc.GetMode(),
		})
	}
}
