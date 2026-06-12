package radar

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// TrackState 目标跟踪状态
type TrackState string

const (
	StateDetected  TrackState = "detected"
	StateTracking  TrackState = "tracking"
	StateLocked    TrackState = "locked"
	StateEngaged   TrackState = "engaged"
)

// Position 三维位置 (km, 以雷达为原点)
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// Velocity 三维速度 (km/s)
type Velocity struct {
	Vx float64 `json:"vx"`
	Vy float64 `json:"vy"`
	Vz float64 `json:"vz"`
}

// Target 雷达目标
type Target struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`       // aircraft/missile/drone/ship
	Position  Position   `json:"position"`
	Velocity  Velocity   `json:"velocity"`
	State     TrackState `json:"state"`
	FirstSeen time.Time  `json:"first_seen"`
	LastSeen  time.Time  `json:"last_seen"`
	ticks     int        // 内部计数器，用于状态机推进
}

// ---------------------------------------------------------------------------
// 目标存储
// ---------------------------------------------------------------------------

// TargetStore 线程安全的目标存储
type TargetStore struct {
	mu      sync.RWMutex
	targets map[string]*Target
}

func NewTargetStore() *TargetStore {
	return &TargetStore{targets: make(map[string]*Target)}
}

func (ts *TargetStore) Add(t *Target) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.targets[t.ID] = t
}

func (ts *TargetStore) Remove(id string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.targets, id)
}

func (ts *TargetStore) Get(id string) (*Target, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.targets[id]
	return t, ok
}

func (ts *TargetStore) All() []*Target {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	result := make([]*Target, 0, len(ts.targets))
	for _, t := range ts.targets {
		result = append(result, t)
	}
	return result
}

func (ts *TargetStore) Count() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.targets)
}

func (ts *TargetStore) CountByState(state TrackState) int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	n := 0
	for _, t := range ts.targets {
		if t.State == state {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------------------
// 目标生成器 & 状态机
// ---------------------------------------------------------------------------

var targetTypes = []string{"aircraft", "missile", "drone", "ship"}
var idChars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func randomID() string {
	b := make([]rune, 8)
	for i := range b {
		b[i] = idChars[rand.Intn(len(idChars))]
	}
	return "TGT-" + string(b)
}

// generateTarget 在雷达范围内随机生成一个目标
func generateTarget(rangeKm int) *Target {
	r := rand.Float64() * float64(rangeKm) * 0.9 // 不生成在边缘
	theta := rand.Float64() * 2 * math.Pi
	phi := rand.Float64() * math.Pi / 3 // 仰角 0-60°

	// 速度量级: 0.1 - 2.0 km/s (Mach 0.3 - 6)
	speed := 0.1 + rand.Float64()*1.9

	return &Target{
		ID:   randomID(),
		Type: targetTypes[rand.Intn(len(targetTypes))],
		Position: Position{
			X: r * math.Cos(theta) * math.Cos(phi),
			Y: r * math.Sin(theta) * math.Cos(phi),
			Z: r * math.Sin(phi),
		},
		Velocity: Velocity{
			Vx: speed * (rand.Float64()*2 - 1),
			Vy: speed * (rand.Float64()*2 - 1),
			Vz: speed * (rand.Float64() - 0.5),
		},
		State:     StateDetected,
		FirstSeen: time.Now(),
		LastSeen:  time.Now(),
		ticks:     0,
	}
}

// advanceState 推进目标状态机
// 每次 tick (100ms) 递增计数器。状态推进速率受 mode 影响:
//   - search 模式: 只到 tracking，不锁定不交战
//   - track 模式: 到 locked
//   - engage 模式: 可到 engaged
func (t *Target) advanceState(mode string) {
	t.ticks++
	// 大约每 20 ticks (2 秒) 进阶一次
	switch {
	case t.State == StateDetected && t.ticks >= 20:
		t.State = StateTracking
	case t.State == StateTracking && t.ticks >= 40 && (mode == "track" || mode == "engage"):
		t.State = StateLocked
	case t.State == StateLocked && t.ticks >= 60 && mode == "engage":
		t.State = StateEngaged
	}
}

// updatePosition 更新时间 & 位置
func (t *Target) updatePosition(dt float64) {
	t.Position.X += t.Velocity.Vx * dt
	t.Position.Y += t.Velocity.Vy * dt
	t.Position.Z += t.Velocity.Vz * dt
	t.LastSeen = time.Now()
}

// distance 计算目标离雷达原点的距离
func (t *Target) distance() float64 {
	return math.Sqrt(t.Position.X*t.Position.X + t.Position.Y*t.Position.Y + t.Position.Z*t.Position.Z)
}
