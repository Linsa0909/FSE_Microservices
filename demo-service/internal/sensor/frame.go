package sensor

import (
	"math/rand"
	"time"
)

// Detection 检测到的目标
type Detection struct {
	Type       string  `json:"type"`       // person/vehicle/animal/aircraft
	Confidence float64 `json:"confidence"` // 0.0-1.0
	BBox       BBox    `json:"bbox"`
}

// BBox 边界框 (像素坐标)
type BBox struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// Frame 一帧图像数据
type Frame struct {
	FrameID    int64       `json:"frame_id"`
	Timestamp  time.Time   `json:"timestamp"`
	Resolution string      `json:"resolution"`
	Mode       string      `json:"mode"`
	Detections []Detection `json:"detections"`
}

// ---------------------------------------------------------------------------
// 帧缓冲区 (环形)
// ---------------------------------------------------------------------------

// FrameBuffer 线程安全的环形帧缓冲区
type FrameBuffer struct {
	frames   []*Frame
	head     int
	size     int
	capacity int
}

// NewFrameBuffer 创建环形缓冲区
func NewFrameBuffer(capacity int) *FrameBuffer {
	return &FrameBuffer{
		frames:   make([]*Frame, capacity),
		capacity: capacity,
	}
}

// Push 添加一帧 (覆盖最旧的)
func (fb *FrameBuffer) Push(f *Frame) {
	fb.frames[fb.head] = f
	fb.head = (fb.head + 1) % fb.capacity
	if fb.size < fb.capacity {
		fb.size++
	}
}

// Latest 返回最新一帧
func (fb *FrameBuffer) Latest() *Frame {
	if fb.size == 0 {
		return nil
	}
	idx := (fb.head - 1 + fb.capacity) % fb.capacity
	return fb.frames[idx]
}

// LastN 返回最近 N 帧
func (fb *FrameBuffer) LastN(n int) []*Frame {
	if n > fb.size {
		n = fb.size
	}
	result := make([]*Frame, n)
	for i := 0; i < n; i++ {
		idx := (fb.head - n + i + fb.capacity) % fb.capacity
		result[i] = fb.frames[idx]
	}
	return result
}

// Size 返回缓冲区当前帧数
func (fb *FrameBuffer) Size() int {
	return fb.size
}

// ---------------------------------------------------------------------------
// 帧生成器
// ---------------------------------------------------------------------------

var detectionTypes = []string{"person", "vehicle", "animal", "aircraft"}

// generateFrame 生成一帧模拟图像数据
// - mode 影响检测概率: night=0.4x, thermal=1.5x
// - sensitivity 线性影响检测数量
func generateFrame(frameID int64, resolution, mode string, sensitivity int) *Frame {
	// 解析分辨率 → 像素范围
	w, h := 1280, 720
	switch resolution {
	case "640x480":
		w, h = 640, 480
	case "1280x720":
		w, h = 1280, 720
	case "1920x1080":
		w, h = 1920, 1080
	case "3840x2160":
		w, h = 3840, 2160
	}

	frame := &Frame{
		FrameID:    frameID,
		Timestamp:  time.Now(),
		Resolution: resolution,
		Mode:       mode,
		Detections: []Detection{},
	}

	// 检测概率因子
	probFactor := 1.0
	switch mode {
	case "night":
		probFactor = 0.4
	case "thermal":
		probFactor = 1.5
	}

	// 灵敏度: 5=基准, 1=0.3x, 10=2x
	sensFactor := 0.3 + float64(sensitivity)*0.17

	// 期望检测数 = 基础 3 × 概率 × 灵敏度
	expectedDet := 3.0 * probFactor * sensFactor
	numDet := 0
	// 使用简单泊松近似: 区间 [0, 2*expected]
	for rand.Float64() < 0.6 && numDet < int(expectedDet*2)+1 {
		numDet++
	}

	for i := 0; i < numDet; i++ {
		confidence := 0.3 + rand.Float64()*0.7 // 0.3-1.0
		dw := 20 + rand.Intn(w/4)
		dh := 20 + rand.Intn(h/4)
		frame.Detections = append(frame.Detections, Detection{
			Type:       detectionTypes[rand.Intn(len(detectionTypes))],
			Confidence: confidence,
			BBox: BBox{
				X: rand.Intn(w - dw),
				Y: rand.Intn(h - dh),
				W: dw,
				H: dh,
			},
		})
	}

	return frame
}
