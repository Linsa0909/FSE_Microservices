package navigation

import (
	"fmt"
	"math"
	"time"
)

// GeoPosition GPS 位置
type GeoPosition struct {
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Timestamp time.Time `json:"timestamp"`
	Accuracy  float64   `json:"accuracy_m"` // 模拟 GPS 精度 (米)
}

// Waypoint 航点
type Waypoint struct {
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Reached bool    `json:"reached"`
}

// ---------------------------------------------------------------------------
// 位置计算
// ---------------------------------------------------------------------------

// 1 节 = 1 海里/小时 ≈ 1852 米/小时
// 1 纬度 ≈ 111,320 米
// 1 经度 ≈ 111,320 × cos(lat) 米
const metersPerLat = 111320.0

func metersPerLng(lat float64) float64 {
	return 111320.0 * math.Cos(lat*math.Pi/180.0)
}

// advancePosition 沿航向以指定速度推进位置
// heading: 0=北, 90=东, 180=南, 270=西
// speedKnots: 节
// dtSec: 时间步长 (秒)
func advancePosition(pos GeoPosition, headingDeg, speedKnots int, dtSec float64) GeoPosition {
	// 速度转换: 节 → 米/秒
	speedMS := float64(speedKnots) * 1852.0 / 3600.0
	distance := speedMS * dtSec

	// 航向转弧度 (0=北, 顺时针)
	rad := float64(headingDeg) * math.Pi / 180.0

	// 北向分量 → 纬度增量
	dlat := distance * math.Cos(rad) / metersPerLat
	// 东向分量 → 经度增量
	dlng := distance * math.Sin(rad) / metersPerLng(pos.Lat)

	pos.Lat += dlat
	pos.Lng += dlng
	pos.Timestamp = time.Now()
	// 模拟 GPS 精度: 2-15 米，受速度影响
	pos.Accuracy = 2.0 + float64(speedKnots)*0.2 + (float64(time.Now().UnixNano()%1000) / 1000.0 * 3.0)

	// 边界检查
	if pos.Lat > 90 {
		pos.Lat = 90
	}
	if pos.Lat < -90 {
		pos.Lat = -90
	}
	if pos.Lng > 180 {
		pos.Lng -= 360
	}
	if pos.Lng < -180 {
		pos.Lng += 360
	}

	return pos
}

// computeHeading 计算从 src 指向 dst 的航向角 (度)
func computeHeading(src, dst GeoPosition) int {
	dlat := (dst.Lat - src.Lat) * metersPerLat
	dlng := (dst.Lng - src.Lng) * metersPerLng(src.Lat)

	deg := math.Atan2(dlng, dlat) * 180.0 / math.Pi
	if deg < 0 {
		deg += 360
	}
	return int(deg)
}

// distanceKm 计算两点距离 (km, 简化球面近似)
func distanceKm(a, b GeoPosition) float64 {
	dlat := (b.Lat - a.Lat) * metersPerLat
	dlng := (b.Lng - a.Lng) * metersPerLng((a.Lat+b.Lat)/2)
	return math.Sqrt(dlat*dlat+dlng*dlng) / 1000.0
}

// normalizeHeading 归一化航向到 [0, 360)
func normalizeHeading(deg int) int {
	deg = deg % 360
	if deg < 0 {
		deg += 360
	}
	return deg
}

// etaString 计算预计到达时间字符串
func etaString(distKm float64, speedKnots int) string {
	if speedKnots == 0 {
		return "--"
	}
	speedKmH := float64(speedKnots) * 1.852
	hours := distKm / speedKmH
	if hours < 0.016 {
		return "< 1min"
	}
	h := int(hours)
	m := int((hours - float64(h)) * 60)
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
