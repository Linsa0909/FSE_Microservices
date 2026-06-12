package sensor

import (
	"context"
	"log"
	"sync/atomic"
	"time"
)

// SensorService 光电传感器服务
type SensorService struct {
	cfg         SensorConfig
	buffer      *FrameBuffer
	startTime   time.Time
	totalFrames atomic.Int64
	frameID     atomic.Int64

	ctx    context.Context
	cancel context.CancelFunc
}

// NewSensorService 创建传感器服务实例
func NewSensorService(cfg SensorConfig) *SensorService {
	return &SensorService{
		cfg:       cfg,
		buffer:    NewFrameBuffer(60), // 保留最近 60 帧
		startTime: time.Now(),
	}
}

// Start 启动后台帧生成协程
func (s *SensorService) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)

	go func() {
		// 保护: 最多 30 实际 fps，避免 CPU 过载
		actualFPS := s.cfg.FPS
		if actualFPS > 30 {
			actualFPS = 30
		}
		interval := time.Duration(1000/float64(actualFPS)) * time.Millisecond
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				frameID := s.frameID.Add(1)
				frame := generateFrame(frameID, s.cfg.Resolution, s.cfg.Mode, s.cfg.Sensitivity)
				s.buffer.Push(frame)
				s.totalFrames.Add(1)
			}
		}
	}()

	log.Printf("[SensorService] 启动完成 — wavelength=%dnm fps=%d resolution=%s sensitivity=%d mode=%s",
		s.cfg.WavelengthNm, s.cfg.FPS, s.cfg.Resolution, s.cfg.Sensitivity, s.cfg.Mode)
}

// Stop 停止后台协程
func (s *SensorService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// GetBuffer 返回帧缓冲区
func (s *SensorService) GetBuffer() *FrameBuffer {
	return s.buffer
}

// GetCfg 返回当前配置
func (s *SensorService) GetCfg() SensorConfig {
	return s.cfg
}

// UptimeSeconds 返回运行时长
func (s *SensorService) UptimeSeconds() float64 {
	return time.Since(s.startTime).Seconds()
}

// TotalFrames 返回累计生成帧数
func (s *SensorService) TotalFrames() int64 {
	return s.totalFrames.Load()
}

// TotalDetections 统计缓冲区中所有帧的检测总数
func (s *SensorService) TotalDetections() int64 {
	var count int64
	for _, f := range s.buffer.LastN(s.buffer.Size()) {
		count += int64(len(f.Detections))
	}
	return count
}
