package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "portwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	raw := `
interval: 10s
ports:
  - host: localhost
    port: 8080
    actions:
      - on: open
        webhook: http://example.com/hook
`
	path := writeTempConfig(t, raw)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 10*time.Second {
		t.Errorf("expected 10s interval, got %v", cfg.Interval)
	}
	if len(cfg.Ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(cfg.Ports))
	}
	if cfg.Ports[0].Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Ports[0].Port)
	}
}

func TestLoad_DefaultInterval(t *testing.T) {
	raw := `
ports:
  - host: localhost
    port: 9090
    actions:
      - on: close
        command: echo closed
`
	path := writeTempConfig(t, raw)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 30*time.Second {
		t.Errorf("expected default 30s, got %v", cfg.Interval)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	raw := `
ports:
  - host: localhost
    port: 99999
    actions:
      - on: open
        command: echo hi
`
	path := writeTempConfig(t, raw)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid port, got nil")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
