package navigation

import (
	"context"
	"hash/fnv"
	"log"
	"math"
	"sync"
	"time"
)

// NavService 船舶航海服务
type NavService struct {
	cfg          NavConfig
	currentPos   GeoPosition
	waypoints    []Waypoint
	currentWpIdx int
	startTime    time.Time
	heading      int // 当前实际航向

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewNavService 创建航海服务实例
func NewNavService(cfg NavConfig) *NavService {
	// 初始位置: 上海附近
	pos := GeoPosition{
		Lat:       31.23,
		Lng:       121.47,
		Timestamp: time.Now(),
		Accuracy:  2.0,
	}

	// 将航点名称转为坐标 (hash-based 确定性映射)
	waypoints := make([]Waypoint, len(cfg.Waypoints))
	for i, name := range cfg.Waypoints {
		waypoints[i] = Waypoint{
			Name:    name,
			Lat:     waypointLat(name),
			Lng:     waypointLng(name),
			Reached: false,
		}
	}

	return &NavService{
		cfg:          cfg,
		currentPos:   pos,
		waypoints:    waypoints,
		currentWpIdx: 0,
		startTime:    time.Now(),
		heading:      cfg.HeadingDeg,
	}
}

// Start 启动后台位置更新协程
func (s *NavService) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)

	go func() {
		interval := time.Duration(s.cfg.UpdateIntervalS) * time.Second
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.tick()
			}
		}
	}()

	log.Printf("[NavService] 启动完成 — interval=%ds speed=%dkn heading=%d° dest=%s auto=%v waypoints=%v",
		s.cfg.UpdateIntervalS, s.cfg.SpeedKnots, s.cfg.HeadingDeg, s.cfg.Destination, s.cfg.AutoPilot, s.cfg.Waypoints)
}

// Stop 停止后台协程
func (s *NavService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// tick 单次位置更新
func (s *NavService) tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 自动导航模式: 计算航向指向下一目标
	if s.cfg.AutoPilot {
		target := s.currentTarget()
		s.heading = computeHeading(s.currentPos, target)
	} else {
		s.heading = s.cfg.HeadingDeg
	}

	// 推进位置
	dt := float64(s.cfg.UpdateIntervalS)
	s.currentPos = advancePosition(s.currentPos, s.heading, s.cfg.SpeedKnots, dt)

	// 检查是否接近下一航点 (10km 以内视为到达)
	if s.currentWpIdx < len(s.waypoints) {
		target := geoPosFromWaypoint(s.waypoints[s.currentWpIdx])
		dist := distanceKm(s.currentPos, target)
		if dist < 10.0 {
			s.waypoints[s.currentWpIdx].Reached = true
			s.currentWpIdx++
			log.Printf("[NavService] ✅ 到达航点: %s", s.waypoints[s.currentWpIdx-1].Name)
		}
	}
}

// currentTarget 返回当前导航目标坐标
func (s *NavService) currentTarget() GeoPosition {
	// 如果有未到达的航点，取第一个
	if s.currentWpIdx < len(s.waypoints) {
		return geoPosFromWaypoint(s.waypoints[s.currentWpIdx])
	}
	// 否则指向最终目的地
	return waypointToGeo(s.cfg.Destination)
}

// geoPosFromWaypoint 从 Waypoint 提取 GeoPosition
func geoPosFromWaypoint(wp Waypoint) GeoPosition {
	return GeoPosition{Lat: wp.Lat, Lng: wp.Lng}
}

// SetHeading 手动设置航向 (仅 auto_pilot=false)
func (s *NavService) SetHeading(deg int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cfg.AutoPilot {
		return &ErrAutoPilotLocked{}
	}

	if deg < 0 || deg > 359 {
		return &ErrInvalidHeading{Deg: deg}
	}

	s.cfg.HeadingDeg = deg
	s.heading = deg
	log.Printf("[NavService] 航向变更: → %d°", deg)
	return nil
}

// GetPosition 返回当前位置
func (s *NavService) GetPosition() GeoPosition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentPos
}

// GetHeading 返回当前航向
func (s *NavService) GetHeading() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.heading
}

// GetRoute 返回路线信息
func (s *NavService) GetRoute() ([]Waypoint, int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// 拷贝避免并发修改
	wps := make([]Waypoint, len(s.waypoints))
	copy(wps, s.waypoints)
	return wps, s.currentWpIdx, len(s.waypoints)
}

// GetCfg 返回当前配置快照
func (s *NavService) GetCfg() NavConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

// ETA 计算到最终目的地的预计时间
func (s *NavService) ETA() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	target := waypointToGeo(s.cfg.Destination)
	dist := distanceKm(s.currentPos, target)
	return etaString(dist, s.cfg.SpeedKnots)
}

// UptimeSeconds 返回运行时长
func (s *NavService) UptimeSeconds() float64 {
	return time.Since(s.startTime).Seconds()
}

// ---------------------------------------------------------------------------
// 航点名 → 坐标映射 (hash-based 确定性)
// ---------------------------------------------------------------------------

func waypointLat(name string) float64 {
	h := hash(name)
	return 20.0 + float64(h%7000)/100.0 // 20.00 - 89.99
}

func waypointLng(name string) float64 {
	h := hash(name + "_lng")
	return 100.0 + float64(h%8000)/100.0 // 100.00 - 179.99
}

func waypointToGeo(name string) GeoPosition {
	return GeoPosition{Lat: waypointLat(name), Lng: waypointLng(name)}
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// ---------------------------------------------------------------------------
// 错误类型
// ---------------------------------------------------------------------------

type ErrAutoPilotLocked struct{}

func (e *ErrAutoPilotLocked) Error() string {
	return "cannot steer manually when auto_pilot is enabled"
}

type ErrInvalidHeading struct{ Deg int }

func (e *ErrInvalidHeading) Error() string {
	return "invalid heading, must be 0-359"
}

func init() {
	// ensure math is used (avoid import cycle note)
	_ = math.Pi
}
