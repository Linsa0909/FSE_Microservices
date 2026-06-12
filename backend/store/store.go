// Package store provides the in-memory data model and thread-safe storage for the config center.
// It defines ConfigGroup (a service+env configuration), ChangeRecord (audit log), and ConfigStore (the main repository).
package store

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ConfigGroup 配置组 — 一个服务在某个环境下的所有配置
type ConfigGroup struct {
	Service          string            `json:"service"`
	Env              string            `json:"env"`
	PublishedVersion string            `json:"publishedVersion"` // "1.0.0" 三位版本号
	PublishedData    map[string]string `json:"publishedData"`    // 线上生效的配置
	DraftData        map[string]string `json:"draftData"`        // 编辑中、未发布的配置
	LastPublishedAt  time.Time         `json:"lastPublishedAt"`
}

// ChangeRecord 变更记录
type ChangeRecord struct {
	Time     time.Time `json:"time"`
	Action   string    `json:"action"` // "新增" / "修改" / "删除" / "发布"
	Key      string    `json:"key"`
	Version  string    `json:"version"`  // 当时版本号, e.g. "1.0.2"
	Operator string    `json:"operator"` // MVP 固定 "admin"
}

// ConfigStore 内存存储层
type ConfigStore struct {
	sync.RWMutex
	Configs map[string]*ConfigGroup   // key: "service:env"
	Logs    map[string][]ChangeRecord // key: "service:env"，最近 10 条
}

// New 创建 ConfigStore
func New() *ConfigStore {
	return &ConfigStore{
		Configs: make(map[string]*ConfigGroup),
		Logs:    make(map[string][]ChangeRecord),
	}
}

// Key 生成存储 key
func Key(service, env string) string {
	return fmt.Sprintf("%s:%s", service, env)
}

// HasDraft 判断是否有未发布的变更（状态由数据驱动，无 Status 字段）
func (g *ConfigGroup) HasDraft() bool {
	return !reflect.DeepEqual(g.DraftData, g.PublishedData)
}

// CloneMap 深拷贝 map，避免引用污染
func CloneMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// BumpVersion 递增补丁版本号: "1.0.0" → "1.0.1"
func BumpVersion(v string) string {
	if v == "" {
		return "1.0.0"
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return "1.0.0"
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "1.0.0"
	}
	return fmt.Sprintf("%s.%s.%d", parts[0], parts[1], patch+1)
}

// --- ConfigStore 方法 ---

// GetOrCreate 获取或创建 ConfigGroup
func (cs *ConfigStore) GetOrCreate(service, env string) *ConfigGroup {
	k := Key(service, env)
	g, ok := cs.Configs[k]
	if !ok {
		g = &ConfigGroup{
			Service:       service,
			Env:           env,
			PublishedData: make(map[string]string),
			DraftData:     make(map[string]string),
		}
		cs.Configs[k] = g
	}
	return g
}

// Get 获取 ConfigGroup（不存在返回 nil）
func (cs *ConfigStore) Get(service, env string) *ConfigGroup {
	return cs.Configs[Key(service, env)]
}

// SetKey 新增/修改 DraftData 中的单个配置项
func (cs *ConfigStore) SetKey(service, env, key, value string) {
	g := cs.GetOrCreate(service, env)
	oldVal, existed := g.DraftData[key]
	g.DraftData[key] = value

	if !existed {
		cs.addLog(service, env, ChangeRecord{
			Time:     time.Now(),
			Action:   "新增",
			Key:      key,
			Version:  g.PublishedVersion,
			Operator: "admin",
		})
	} else if oldVal != value {
		cs.addLog(service, env, ChangeRecord{
			Time:     time.Now(),
			Action:   "修改",
			Key:      key,
			Version:  g.PublishedVersion,
			Operator: "admin",
		})
	}
}

// DeleteKey 删除 DraftData 中的单个配置项
func (cs *ConfigStore) DeleteKey(service, env, key string) {
	g := cs.Get(service, env)
	if g == nil {
		return
	}
	if _, ok := g.DraftData[key]; !ok {
		return
	}
	delete(g.DraftData, key)

	cs.addLog(service, env, ChangeRecord{
		Time:     time.Now(),
		Action:   "删除",
		Key:      key,
		Version:  g.PublishedVersion,
		Operator: "admin",
	})
}

// Publish 发布配置：DraftData DeepCopy → PublishedData，递增补丁版本号
func (cs *ConfigStore) Publish(service, env string) {
	g := cs.GetOrCreate(service, env)
	g.PublishedData = CloneMap(g.DraftData)
	g.PublishedVersion = BumpVersion(g.PublishedVersion)
	g.LastPublishedAt = time.Now()

	cs.addLog(service, env, ChangeRecord{
		Time:     time.Now(),
		Action:   "发布",
		Key:      "",
		Version:  g.PublishedVersion,
		Operator: "admin",
	})
}

// GetLogs 获取某个服务+环境的变更记录
func (cs *ConfigStore) GetLogs(service, env string) []ChangeRecord {
	return cs.Logs[Key(service, env)]
}

// addLog 追加一条变更记录，保留最近 10 条
func (cs *ConfigStore) addLog(service, env string, record ChangeRecord) {
	k := Key(service, env)
	logs := cs.Logs[k]
	logs = append(logs, record)
	if len(logs) > 10 {
		logs = logs[len(logs)-10:]
	}
	cs.Logs[k] = logs
}
