package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/portwatch/internal/snapshot"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "state.json")
}

func TestSnapshot_SetAndGet(t *testing.T) {
	s, err := snapshot.New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := s.Set("localhost", 8080, "open"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	e, ok := s.Get("localhost", 8080)
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.State != "open" {
		t.Errorf("state = %q, want %q", e.State, "open")
	}
	if e.Port != 8080 {
		t.Errorf("port = %d, want 8080", e.Port)
	}
	if e.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestSnapshot_GetMissing(t *testing.T) {
	s, _ := snapshot.New(tempPath(t))
	_, ok := s.Get("localhost", 9999)
	if ok {
		t.Error("expected missing entry")
	}
}

func TestSnapshot_PersistsAcrossReload(t *testing.T) {
	path := tempPath(t)

	s1, _ := snapshot.New(path)
	_ = s1.Set("127.0.0.1", 3000, "closed")

	s2, err := snapshot.New(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	e, ok := s2.Get("127.0.0.1", 3000)
	if !ok {
		t.Fatal("entry missing after reload")
	}
	if e.State != "closed" {
		t.Errorf("state = %q, want %q", e.State, "closed")
	}
}

func TestSnapshot_All(t *testing.T) {
	s, _ := snapshot.New(tempPath(t))
	_ = s.Set("localhost", 80, "open")
	_ = s.Set("localhost", 443, "open")

	all := s.All()
	if len(all) != 2 {
		t.Errorf("len(All) = %d, want 2", len(all))
	}
}

func TestSnapshot_UpdatedAtIsRecent(t *testing.T) {
	s, _ := snapshot.New(tempPath(t))
	before := time.Now().UTC()
	_ = s.Set("localhost", 8080, "open")
	after := time.Now().UTC()

	e, _ := s.Get("localhost", 8080)
	if e.UpdatedAt.Before(before) || e.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt %v not in [%v, %v]", e.UpdatedAt, before, after)
	}
}

func TestSnapshot_MissingFileIsOK(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	_, err := snapshot.New(path)
	if err != nil {
		t.Errorf("expected no error for missing file, got %v", err)
	}
}

func TestSnapshot_OverwriteEntry(t *testing.T) {
	s, _ := snapshot.New(tempPath(t))
	_ = s.Set("localhost", 8080, "open")
	_ = s.Set("localhost", 8080, "closed")

	e, _ := s.Get("localhost", 8080)
	if e.State != "closed" {
		t.Errorf("state = %q, want %q", e.State, "closed")
	}
	if len(s.All()) != 1 {
		t.Error("expected only one entry after overwrite")
	}
}

func TestSnapshot_InvalidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	_ = os.WriteFile(path, []byte("not-json{"), 0o644)
	_, err := snapshot.New(path)
	if err == nil {
		t.Error("expected error for invalid JSON file")
	}
}
