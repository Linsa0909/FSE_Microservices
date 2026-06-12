package radar

import (
	"context"
	"log"
	"math/rand"
	"sync/atomic"
	"time"
)

// RadarService 火控雷达服务
type RadarService struct {
	cfg         RadarConfig
	store       *TargetStore
	startTime   time.Time
	mode        string // 当前工作模式 (可从 API 动态切换)
	totalDetect atomic.Int64

	ctx    context.Context
	cancel context.CancelFunc
}

// NewRadarService 创建雷达服务实例
func NewRadarService(cfg RadarConfig) *RadarService {
	return &RadarService{
		cfg:       cfg,
		store:     NewTargetStore(),
		startTime: time.Now(),
		mode:      cfg.Mode,
	}
}

// Start 启动后台模拟协程
func (s *RadarService) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)

	// Goroutine 1: 目标生成器 — 按 scan_rate_hz 频率生成随机目标
	go func() {
		interval := time.Duration(1000/float64(s.cfg.ScanRateHz)) * time.Millisecond
		if interval < 10*time.Millisecond {
			interval = 10 * time.Millisecond // 保护: 最快 100Hz
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				if s.store.Count() >= s.cfg.MaxTargets {
					continue // 达到最大目标数上限
				}
				t := generateTarget(s.cfg.RangeKm)
				s.store.Add(t)
				s.totalDetect.Add(1)
			}
		}
	}()

	// Goroutine 2: 目标更新器 — 每 100ms 更新位置 + 状态机
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.updateTargets()
			}
		}
	}()

	log.Printf("[RadarService] 启动完成 — scan=%dHz range=%dkm maxTargets=%d mode=%s band=%s",
		s.cfg.ScanRateHz, s.cfg.RangeKm, s.cfg.MaxTargets, s.mode, s.cfg.Band)
}

// Stop 停止后台协程
func (s *RadarService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// updateTargets 推进所有目标的物理状态
func (s *RadarService) updateTargets() {
	dt := 0.1 // 100ms → 秒
	mode := s.mode

	for _, t := range s.store.All() {
		t.updatePosition(dt)
		t.advanceState(mode)
		// 超出探测范围则移除
		if t.distance() > float64(s.cfg.RangeKm) {
			s.store.Remove(t.ID)
		}
	}
}

// SetMode 动态切换工作模式
func (s *RadarService) SetMode(mode string) error {
	valid := map[string]bool{"search": true, "track": true, "engage": true}
	if !valid[mode] {
		return &ErrInvalidMode{Mode: mode}
	}
	old := s.mode
	s.mode = mode
	log.Printf("[RadarService] 模式切换: %s → %s", old, mode)
	return nil
}

// GetMode 返回当前模式
func (s *RadarService) GetMode() string {
	return s.mode
}

// GetStore 返回目标存储
func (s *RadarService) GetStore() *TargetStore {
	return s.store
}

// GetCfg 返回当前配置
func (s *RadarService) GetCfg() RadarConfig {
	return s.cfg
}

// UptimeSeconds 返回运行时长
func (s *RadarService) UptimeSeconds() float64 {
	return time.Since(s.startTime).Seconds()
}

// TotalDetected 返回历史累计检测目标数
func (s *RadarService) TotalDetected() int64 {
	return s.totalDetect.Load()
}

// Stats 返回实时的目标统计
type RadarStats struct {
	Detected int `json:"detected"`
	Tracking int `json:"tracking"`
	Locked   int `json:"locked"`
	Engaged  int `json:"engaged"`
}

func (s *RadarService) Stats() RadarStats {
	return RadarStats{
		Detected: s.store.CountByState(StateDetected),
		Tracking: s.store.CountByState(StateTracking),
		Locked:   s.store.CountByState(StateLocked),
		Engaged:  s.store.CountByState(StateEngaged),
	}
}

// ErrInvalidMode 非法模式错误
type ErrInvalidMode struct{ Mode string }

func (e *ErrInvalidMode) Error() string {
	return "invalid mode: " + e.Mode + " (valid: search/track/engage)"
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
