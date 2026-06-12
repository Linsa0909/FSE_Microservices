package store

import (
	"reflect"
	"testing"
)

func TestSetKeyDraftOnly(t *testing.T) {
	cs := New()
	cs.Lock()
	cs.SetKey("test-svc", "dev", "key1", "val1")
	cs.Unlock()

	g := cs.Get("test-svc", "dev")
	if g == nil {
		t.Fatal("group should exist")
	}
	if g.DraftData["key1"] != "val1" {
		t.Errorf("DraftData[key1] = %q, want %q", g.DraftData["key1"], "val1")
	}
	// PublishedData should still be empty
	if len(g.PublishedData) != 0 {
		t.Error("PublishedData should be empty before publish")
	}
}

func TestPublishVersionIncrease(t *testing.T) {
	cs := New()
	cs.Lock()
	cs.SetKey("test-svc", "dev", "k", "v1")
	cs.Publish("test-svc", "dev")
	cs.Unlock()

	g := cs.Get("test-svc", "dev")
	if g.PublishedVersion != "1.0.0" {
		t.Errorf("version = %q, want %q", g.PublishedVersion, "1.0.0")
	}
	if g.PublishedData["k"] != "v1" {
		t.Errorf("PublishedData[k] = %q, want %q", g.PublishedData["k"], "v1")
	}

	// Publish again — patch version should increase
	cs.Lock()
	cs.Publish("test-svc", "dev")
	cs.Unlock()

	if g.PublishedVersion != "1.0.1" {
		t.Errorf("version after second publish = %q, want %q", g.PublishedVersion, "1.0.1")
	}
}

func TestBumpVersion(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "1.0.0"},
		{"1.0.0", "1.0.1"},
		{"2.3.9", "2.3.10"},
		{"bad", "1.0.0"},
	}
	for _, tt := range tests {
		got := BumpVersion(tt.in)
		if got != tt.want {
			t.Errorf("BumpVersion(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestPublishDeepCopy(t *testing.T) {
	cs := New()
	cs.Lock()
	cs.SetKey("svc", "env", "k", "v")
	cs.Publish("svc", "env")
	cs.Unlock()

	g := cs.Get("svc", "env")

	// After publish, modify DraftData — PublishedData should stay unchanged
	cs.Lock()
	g.DraftData["k"] = "hacked"
	cs.Unlock()

	if g.PublishedData["k"] != "v" {
		t.Errorf("PublishedData corrupted by DraftData mutation: got %q, want %q", g.PublishedData["k"], "v")
	}
	if g.DraftData["k"] != "hacked" {
		t.Errorf("DraftData should reflect mutation: got %q", g.DraftData["k"])
	}
}

func TestHasDraft(t *testing.T) {
	cs := New()
	cs.Lock()
	cs.SetKey("svc", "env", "k", "v")
	cs.Publish("svc", "env")
	cs.Unlock()

	g := cs.Get("svc", "env")
	if g.HasDraft() {
		t.Error("should not have draft after publish")
	}

	cs.Lock()
	cs.SetKey("svc", "env", "k", "v2")
	cs.Unlock()

	if !g.HasDraft() {
		t.Error("should have draft after editing")
	}
}

func TestChangeLogLimit(t *testing.T) {
	cs := New()
	cs.Lock()
	// Add 15 records
	for i := 0; i < 15; i++ {
		cs.SetKey("svc", "env", "k", string(rune('a'+i)))
	}
	cs.Unlock()

	logs := cs.GetLogs("svc", "env")
	if len(logs) > 10 {
		t.Errorf("logs length = %d, want <= 10", len(logs))
	}
	if len(logs) != 10 {
		t.Errorf("logs length = %d, want 10", len(logs))
	}
}

func TestCloneMap(t *testing.T) {
	src := map[string]string{"a": "1", "b": "2"}
	dst := CloneMap(src)

	if !reflect.DeepEqual(src, dst) {
		t.Error("CloneMap should produce equal map")
	}

	// Mutate original — clone should not change
	src["a"] = "changed"
	if dst["a"] != "1" {
		t.Error("CloneMap mutation leaked")
	}
}

func TestGetOrCreate(t *testing.T) {
	cs := New()
	// 不存在的 service:env 应自动创建
	g := cs.GetOrCreate("radar-service", "dev")
	if g == nil {
		t.Fatal("GetOrCreate should create group")
	}
	if g.Service != "radar-service" || g.Env != "dev" {
		t.Errorf("group identity = %s:%s, want radar-service:dev", g.Service, g.Env)
	}
	// 再次获取应返回同一个
	g2 := cs.GetOrCreate("radar-service", "dev")
	if g2 != g {
		t.Error("GetOrCreate should return same pointer for existing group")
	}
}

func TestMultiServiceIsolation(t *testing.T) {
	cs := New()
	cs.Lock()
	cs.SetKey("radar-service", "dev", "radar.scan_rate_hz", "60")
	cs.SetKey("radar-service", "prod", "radar.scan_rate_hz", "120")
	cs.SetKey("sensor-service", "dev", "sensor.fps", "30")
	cs.Unlock()

	radarDev := cs.Get("radar-service", "dev")
	radarProd := cs.Get("radar-service", "prod")
	sensorDev := cs.Get("sensor-service", "dev")

	if radarDev == nil || radarProd == nil || sensorDev == nil {
		t.Fatal("all groups should exist")
	}
	if radarDev.DraftData["radar.scan_rate_hz"] != "60" {
		t.Errorf("radar-dev scan_rate = %q, want 60", radarDev.DraftData["radar.scan_rate_hz"])
	}
	if radarProd.DraftData["radar.scan_rate_hz"] != "120" {
		t.Errorf("radar-prod scan_rate = %q, want 120", radarProd.DraftData["radar.scan_rate_hz"])
	}
	if sensorDev.DraftData["sensor.fps"] != "30" {
		t.Errorf("sensor-dev fps = %q, want 30", sensorDev.DraftData["sensor.fps"])
	}
	// radar-dev 不应有 sensor.fps
	if _, ok := radarDev.DraftData["sensor.fps"]; ok {
		t.Error("radar-dev should not contain sensor.fps — cross-service leak")
	}
}

func TestChangeLogActions(t *testing.T) {
	cs := New()
	cs.Lock()
	cs.SetKey("svc", "env", "newkey", "v") // 新增
	cs.Publish("svc", "env")                // 发布
	cs.SetKey("svc", "env", "newkey", "v2") // 修改
	cs.Unlock()

	logs := cs.GetLogs("svc", "env")
	if len(logs) != 3 {
		t.Fatalf("want 3 log entries, got %d", len(logs))
	}
	// 顺序: 新增 → 发布 → 修改
	if logs[0].Action != "新增" {
		t.Errorf("log[0] action = %q, want 新增", logs[0].Action)
	}
	if logs[1].Action != "发布" {
		t.Errorf("log[1] action = %q, want 发布", logs[1].Action)
	}
	if logs[2].Action != "修改" {
		t.Errorf("log[2] action = %q, want 修改", logs[2].Action)
	}
}

func TestDomainSeedDataCount(t *testing.T) {
	// 模拟 seed data: 4 通用 + 6 领域 = 10 组
	cs := New()
	seedGroups := []struct {
		service, env string
	}{
		{"order-service", "dev"}, {"order-service", "test"},
		{"user-service", "dev"}, {"user-service", "prod"},
		{"radar-service", "dev"}, {"radar-service", "prod"},
		{"sensor-service", "dev"}, {"sensor-service", "prod"},
		{"navigation-service", "dev"}, {"navigation-service", "prod"},
	}
	for _, s := range seedGroups {
		cs.GetOrCreate(s.service, s.env)
	}

	if len(cs.Configs) != 10 {
		t.Errorf("seed should create 10 config groups, got %d", len(cs.Configs))
	}

	// 验证服务列表
	svcSet := make(map[string]bool)
	for _, g := range cs.Configs {
		svcSet[g.Service] = true
	}
	expectedSvcs := []string{"order-service", "user-service", "radar-service", "sensor-service", "navigation-service"}
	for _, s := range expectedSvcs {
		if !svcSet[s] {
			t.Errorf("missing service: %s", s)
		}
	}
}

func TestDeleteKey(t *testing.T) {
	cs := New()
	cs.Lock()
	cs.SetKey("svc", "env", "k1", "v1")
	cs.SetKey("svc", "env", "k2", "v2")
	cs.DeleteKey("svc", "env", "k1")
	cs.Unlock()

	g := cs.Get("svc", "env")
	if _, ok := g.DraftData["k1"]; ok {
		t.Error("k1 should be deleted")
	}
	if g.DraftData["k2"] != "v2" {
		t.Error("k2 should still exist")
	}
}
