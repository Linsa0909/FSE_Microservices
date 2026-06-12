package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"config-center/store"

	"github.com/gin-gonic/gin"
)

// setup 创建测试环境
func setup() (*gin.Engine, *store.ConfigStore) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	cs := store.New()
	h := New(cs)

	api := r.Group("/api")
	configs := api.Group("/configs")
	{
		configs.GET("", h.ListAll)
		configs.GET("/watch", h.Watch)
		configs.GET("/:service/:env", h.GetOne)
		configs.GET("/:service/:env/published", h.GetPublished)
		configs.PUT("/:service/:env/keys/:key", h.SetKey)
		configs.DELETE("/:service/:env/keys/:key", h.DeleteKey)
		configs.POST("/:service/:env/publish", h.Publish)
		configs.POST("/:service/:env/push", h.Push)
	}

	return r, cs
}

// ===================== 1. GET /api/configs =====================

func TestHandler_ListAll_Empty(t *testing.T) {
	r, _ := setup()
	req := httptest.NewRequest(http.MethodGet, "/api/configs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	var body struct {
		Configs []store.ConfigGroup `json:"configs"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Configs == nil {
		t.Error("configs should be empty array, not nil")
	}
}

func TestHandler_ListAll_WithData(t *testing.T) {
	r, cs := setup()
	cs.SetKey("svc", "dev", "k", "v")

	req := httptest.NewRequest(http.MethodGet, "/api/configs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var body struct {
		Configs []store.ConfigGroup `json:"configs"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Configs) != 1 {
		t.Fatalf("want 1 config, got %d", len(body.Configs))
	}
	if body.Configs[0].Service != "svc" {
		t.Errorf("service = %q, want %q", body.Configs[0].Service, "svc")
	}
}

// ===================== 2. GET /api/configs/:service/:env =====================

func TestHandler_GetOne_Existing(t *testing.T) {
	r, cs := setup()
	cs.SetKey("svc", "dev", "k", "v")

	req := httptest.NewRequest(http.MethodGet, "/api/configs/svc/dev", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	var body struct {
		Config store.ConfigGroup   `json:"config"`
		Logs   []store.ChangeRecord `json:"logs"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Config.Service != "svc" {
		t.Errorf("service = %q, want svc", body.Config.Service)
	}
	if body.Logs == nil {
		t.Error("logs should be []")
	}
}

func TestHandler_GetOne_NonExisting(t *testing.T) {
	r, _ := setup()

	req := httptest.NewRequest(http.MethodGet, "/api/configs/nosvc/noenv", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200 for non-existing, got %d", w.Code)
	}
}

// ===================== 3. PUT /api/configs/:service/:env/keys/:key =====================

func TestHandler_SetKey_Success(t *testing.T) {
	r, _ := setup()
	body := `{"value":"testval"}`
	req := httptest.NewRequest(http.MethodPut, "/api/configs/svc/dev/keys/mykey", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_SetKey_MissingBody(t *testing.T) {
	r, _ := setup()
	req := httptest.NewRequest(http.MethodPut, "/api/configs/svc/dev/keys/mykey", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400 for missing value, got %d", w.Code)
	}
}

func TestHandler_SetKey_URLEncodedKey(t *testing.T) {
	r, cs := setup()
	// key with dots and hyphens
	body := `{"value":"v"}`
	req := httptest.NewRequest(http.MethodPut, "/api/configs/svc/dev/keys/db.url-v2_test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	g := cs.Get("svc", "dev")
	if g.DraftData["db.url-v2_test"] != "v" {
		t.Errorf("draft key = %q, want v", g.DraftData["db.url-v2_test"])
	}
}

// ===================== 4. DELETE /api/configs/:service/:env/keys/:key =====================

func TestHandler_DeleteKey_Existing(t *testing.T) {
	r, cs := setup()
	cs.SetKey("svc", "dev", "key1", "val1")
	cs.SetKey("svc", "dev", "key2", "val2")

	req := httptest.NewRequest(http.MethodDelete, "/api/configs/svc/dev/keys/key1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	g := cs.Get("svc", "dev")
	if _, ok := g.DraftData["key1"]; ok {
		t.Error("key1 should be deleted")
	}
	if g.DraftData["key2"] != "val2" {
		t.Error("key2 should still exist")
	}
}

func TestHandler_DeleteKey_NonExisting(t *testing.T) {
	r, _ := setup()
	req := httptest.NewRequest(http.MethodDelete, "/api/configs/nosvc/noenv/keys/nokey", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("delete non-existing should be 200 (idempotent), got %d", w.Code)
	}
}

// ===================== 5. POST /api/configs/:service/:env/publish =====================

func TestHandler_Publish(t *testing.T) {
	r, cs := setup()
	cs.SetKey("svc", "dev", "k", "v")

	req := httptest.NewRequest(http.MethodPost, "/api/configs/svc/dev/publish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	g := cs.Get("svc", "dev")
	if g.PublishedVersion != 1 {
		t.Errorf("version = %d, want 1", g.PublishedVersion)
	}
	if g.PublishedData["k"] != "v" {
		t.Errorf("published key = %q, want v", g.PublishedData["k"])
	}
}

// ===================== 6. GET /api/configs/:service/:env/published =====================

func TestHandler_GetPublished(t *testing.T) {
	r, cs := setup()
	cs.SetKey("svc", "dev", "k", "v")
	cs.Publish("svc", "dev")

	req := httptest.NewRequest(http.MethodGet, "/api/configs/svc/dev/published", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}

	var body struct {
		Config map[string]string `json:"config"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Config["k"] != "v" {
		t.Errorf("config.k = %q, want v", body.Config["k"])
	}
}

func TestHandler_GetPublished_NonExisting(t *testing.T) {
	r, _ := setup()

	req := httptest.NewRequest(http.MethodGet, "/api/configs/nosvc/noenv/published", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

// ===================== 7 & 8. Reserved endpoints =====================

func TestHandler_Push_Reserved(t *testing.T) {
	r, _ := setup()

	req := httptest.NewRequest(http.MethodPost, "/api/configs/svc/dev/push", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "reserved") {
		t.Errorf("should mention reserved, got %s", w.Body.String())
	}
}

func TestHandler_Watch_Reserved(t *testing.T) {
	r, _ := setup()

	req := httptest.NewRequest(http.MethodGet, "/api/configs/watch", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

// ===================== Full flow: Edit → Publish → Verify =====================

func TestHandler_FullFlow_EditPublishVerify(t *testing.T) {
	r, cs := setup()

	// Step 1: Set some keys
	steps := []struct {
		method string
		url    string
		body   string
		want   int
	}{
		{http.MethodPut, "/api/configs/order-svc/prod/keys/db.url", `{"value":"localhost"}`, 200},
		{http.MethodPut, "/api/configs/order-svc/prod/keys/port", `{"value":"8080"}`, 200},
	}

	for _, s := range steps {
		req := httptest.NewRequest(s.method, s.url, strings.NewReader(s.body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != s.want {
			t.Errorf("step %s %s: want %d, got %d", s.method, s.url, s.want, w.Code)
		}
	}

	// Step 2: Verify DraftData not in PublishedData yet
	g := cs.Get("order-svc", "prod")
	if len(g.PublishedData) != 0 {
		t.Error("PublishedData should be empty before publish")
	}
	if g.DraftData["db.url"] != "localhost" {
		t.Error("db.url should be in DraftData")
	}

	// Step 3: Publish
	req := httptest.NewRequest(http.MethodPost, "/api/configs/order-svc/prod/publish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("publish failed: %d", w.Code)
	}

	// Step 4: Verify PublishedData now has the config
	if g.PublishedVersion != 1 {
		t.Errorf("version = %d, want 1", g.PublishedVersion)
	}
	if g.PublishedData["db.url"] != "localhost" {
		t.Errorf("PublishedData[db.url] = %q, want localhost", g.PublishedData["db.url"])
	}
	if g.HasDraft() {
		t.Error("should not have draft after publish")
	}

	// Step 5: Edit one key → should have draft
	req2 := httptest.NewRequest(http.MethodPut, "/api/configs/order-svc/prod/keys/db.url",
		strings.NewReader(`{"value":"10.0.0.1"}`))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if !g.HasDraft() {
		t.Error("should have draft after edit")
	}

	// Step 6: Fetch published — should still return original value (not draft)
	req3 := httptest.NewRequest(http.MethodGet, "/api/configs/order-svc/prod/published", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	var pub struct {
		Config map[string]string `json:"config"`
	}
	json.Unmarshal(w3.Body.Bytes(), &pub)
	if pub.Config["db.url"] != "localhost" {
		t.Errorf("published db.url = %q, want localhost (draft not published yet)", pub.Config["db.url"])
	}
}

// ===================== Concurrent safety =====================

func TestHandler_ConcurrentWrites(t *testing.T) {
	r, _ := setup()
	var wg sync.WaitGroup
	errCh := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			url := "/api/configs/svc/dev/keys/k" + string(rune('0'+idx))
			body := `{"value":"v` + string(rune('0'+idx)) + `"}`
			req := httptest.NewRequest(http.MethodPut, url, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != 200 {
				errCh <- nil // just a marker
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	// Concurrent publish should also work
	req := httptest.NewRequest(http.MethodPost, "/api/configs/svc/dev/publish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("concurrent publish failed: %d", w.Code)
	}
}
