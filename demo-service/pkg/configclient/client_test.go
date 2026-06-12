package configclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetString(t *testing.T) {
	c := &ConfigClient{rawConfig: map[string]string{"db.url": "localhost:3306"}}

	if v := c.GetString("db.url", "fallback"); v != "localhost:3306" {
		t.Errorf("GetString existing = %q, want localhost:3306", v)
	}
	if v := c.GetString("missing.key", "fallback"); v != "fallback" {
		t.Errorf("GetString missing = %q, want fallback", v)
	}
	if v := c.GetString("empty.key", "default"); v != "default" {
		t.Errorf("GetString empty = %q, want default", v)
	}
}

func TestGetInt(t *testing.T) {
	c := &ConfigClient{rawConfig: map[string]string{
		"radar.scan_rate_hz": "60",
		"bad.port":           "notanumber",
		"overflow.port":      "99999",
		"negative":           "-5",
	}}

	tests := []struct {
		key      string
		def, min, max int
		want     int
	}{
		{"radar.scan_rate_hz", 30, 1, 300, 60},
		{"missing.key", 10, 1, 100, 10},
		{"bad.port", 3000, 1, 65535, 3000},    // non-numeric → default
		{"overflow.port", 3000, 1, 65535, 3000}, // out of range → default
		{"negative", 3000, 1, 65535, 3000},      // negative → default
	}

	for _, tt := range tests {
		got := c.GetInt(tt.key, tt.def, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("GetInt(%q, %d, %d, %d) = %d, want %d", tt.key, tt.def, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestGetFloat(t *testing.T) {
	c := &ConfigClient{rawConfig: map[string]string{
		"radar.gain": "2.5",
		"bad.float":  "abc",
	}}

	if v := c.GetFloat("radar.gain", 1.0, 0.1, 10.0); v != 2.5 {
		t.Errorf("GetFloat valid = %f, want 2.5", v)
	}
	if v := c.GetFloat("missing", 3.0, 1.0, 5.0); v != 3.0 {
		t.Errorf("GetFloat missing = %f, want 3.0", v)
	}
	if v := c.GetFloat("bad.float", 1.0, 0.0, 5.0); v != 1.0 {
		t.Errorf("GetFloat bad = %f, want 1.0 (default)", v)
	}
}

func TestGetEnum(t *testing.T) {
	c := &ConfigClient{rawConfig: map[string]string{
		"radar.mode": "track",
		"bad.mode":   "invalid",
	}}

	if v := c.GetEnum("radar.mode", "search", []string{"search", "track", "engage"}); v != "track" {
		t.Errorf("GetEnum valid = %q, want track", v)
	}
	if v := c.GetEnum("missing.mode", "search", []string{"search", "track", "engage"}); v != "search" {
		t.Errorf("GetEnum missing = %q, want search (default)", v)
	}
	if v := c.GetEnum("bad.mode", "search", []string{"search", "track", "engage"}); v != "search" {
		t.Errorf("GetEnum invalid = %q, want search (default)", v)
	}
}

func TestPullConfig_Success(t *testing.T) {
	// Mock 配置中心
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/configs/test-svc/dev/published" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"config": map[string]string{"db.url": "localhost:3306", "log.level": "debug"},
			})
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	c := New(server.URL, "test-svc", "dev")
	err := c.PullConfig()

	if err != nil {
		t.Fatalf("PullConfig failed: %v", err)
	}
	if c.GetString("db.url", "") != "localhost:3306" {
		t.Errorf("db.url = %q", c.GetString("db.url", ""))
	}
	if c.GetString("log.level", "") != "debug" {
		t.Errorf("log.level = %q", c.GetString("log.level", ""))
	}
	if len(c.GetRawConfig()) != 2 {
		t.Errorf("want 2 keys, got %d", len(c.GetRawConfig()))
	}
}

func TestPullConfig_Retry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(503)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"config": map[string]string{"status": "ok"},
		})
	}))
	defer server.Close()

	c := New(server.URL, "test-svc", "dev")
	c.maxRetries = 5
	c.retryDelay = 0 // 测试时不需要等

	err := c.PullConfig()
	if err != nil {
		t.Fatalf("PullConfig should succeed after retries: %v", err)
	}
	if attempts != 3 {
		t.Errorf("want 3 attempts (2 fail + 1 success), got %d", attempts)
	}
}

func TestPullConfig_AllRetriesExhausted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	c := New(server.URL, "test-svc", "dev")
	c.maxRetries = 3
	c.retryDelay = 0

	err := c.PullConfig()
	if err == nil {
		t.Error("PullConfig should fail after max retries")
	}
}

func TestPullConfig_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"config": map[string]string{}})
	}))
	defer server.Close()

	c := New(server.URL, "test-svc", "dev")
	err := c.PullConfig()
	if err != nil {
		t.Fatalf("empty response should not error: %v", err)
	}
	if len(c.GetRawConfig()) != 0 {
		t.Errorf("want 0 keys, got %d", len(c.GetRawConfig()))
	}
}

func TestPrintDiagnostics(t *testing.T) {
	raw := map[string]string{"db.url": "localhost", "radar.mode": "search"}
	// 仅检查不 panic
	PrintDiagnostics("test-svc", raw, []string{"db.url"})
}
