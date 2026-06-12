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
	if g.PublishedVersion != 1 {
		t.Errorf("version = %d, want 1", g.PublishedVersion)
	}
	if g.PublishedData["k"] != "v1" {
		t.Errorf("PublishedData[k] = %q, want %q", g.PublishedData["k"], "v1")
	}

	// Publish again without changes — version should still increase
	cs.Lock()
	cs.Publish("test-svc", "dev")
	cs.Unlock()

	if g.PublishedVersion != 2 {
		t.Errorf("version after second publish = %d, want 2", g.PublishedVersion)
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
