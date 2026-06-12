package navigation

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册航海服务路由
func RegisterRoutes(rg *gin.RouterGroup, svc *NavService) {
	rg.GET("/nav/status", handleNavStatus(svc))
	rg.GET("/nav/position", handleNavPosition(svc))
	rg.GET("/nav/route", handleNavRoute(svc))
	rg.POST("/nav/steer", handleNavSteer(svc))
}

func handleNavStatus(svc *NavService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := svc.GetCfg()
		pos := svc.GetPosition()
		wps, curIdx, total := svc.GetRoute()
		currentWP := "--"
		if curIdx < len(wps) {
			currentWP = wps[curIdx].Name
		}

		c.JSON(http.StatusOK, gin.H{
			"service": "navigation-service",
			"config": gin.H{
				"update_interval_s": cfg.UpdateIntervalS,
				"speed_knots":       cfg.SpeedKnots,
				"heading_deg":       svc.GetHeading(),
				"configured_heading": cfg.HeadingDeg,
				"destination":       cfg.Destination,
				"auto_pilot":        cfg.AutoPilot,
				"waypoints":         cfg.Waypoints,
			},
			"current_position":  pos,
			"heading_deg":       svc.GetHeading(),
			"speed_knots":       cfg.SpeedKnots,
			"destination":       cfg.Destination,
			"eta":               svc.ETA(),
			"auto_pilot":        cfg.AutoPilot,
			"current_waypoint":  currentWP,
			"waypoint_progress": strconv.Itoa(curIdx) + "/" + strconv.Itoa(total),
			"uptime_seconds":    svc.UptimeSeconds(),
		})
	}
}

func handleNavPosition(svc *NavService) gin.HandlerFunc {
	return func(c *gin.Context) {
		pos := svc.GetPosition()
		c.JSON(http.StatusOK, gin.H{
			"position": pos,
		})
	}
}

func handleNavRoute(svc *NavService) gin.HandlerFunc {
	return func(c *gin.Context) {
		wps, curIdx, total := svc.GetRoute()
		cfg := svc.GetCfg()
		dest := waypointToGeo(cfg.Destination)
		eta := svc.ETA()

		c.JSON(http.StatusOK, gin.H{
			"waypoints":       wps,
			"total_waypoints": total,
			"completed":       curIdx,
			"current":         func() string {
				if curIdx < len(wps) {
					return wps[curIdx].Name
				}
				return cfg.Destination
			}(),
			"destination": gin.H{
				"name": cfg.Destination,
				"lat":  dest.Lat,
				"lng":  dest.Lng,
			},
			"eta": eta,
		})
	}
}

func handleNavSteer(svc *NavService) gin.HandlerFunc {
	return func(c *gin.Context) {
		headingStr := c.Query("heading")
		if headingStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'heading' query parameter"})
			return
		}
		heading, err := strconv.Atoi(headingStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "heading must be an integer"})
			return
		}

		old := svc.GetHeading()
		if err := svc.SetHeading(heading); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":         "heading changed",
			"previous_heading": old,
			"current_heading":  svc.GetHeading(),
		})
	}
}
