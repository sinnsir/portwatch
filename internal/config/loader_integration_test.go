package config_test

import (
	"testing"

	"github.com/user/portwatch/internal/config"
)

func TestLoad_ExampleFile(t *testing.T) {
	cfg, err := config.Load("example.yaml")
	if err != nil {
		t.Fatalf("failed to load example config: %v", err)
	}
	if len(cfg.Ports) == 0 {
		t.Error("expected at least one port in example config")
	}
	for i, p := range cfg.Ports {
		if p.Host == "" {
			t.Errorf("port[%d]: empty host", i)
		}
		if p.Port == 0 {
			t.Errorf("port[%d]: zero port", i)
		}
		if len(p.Actions) == 0 {
			t.Errorf("port[%d]: no actions defined", i)
		}
	}
}

func TestLoad_InvalidActionOn(t *testing.T) {
	raw := `
ports:
  - host: localhost
    port: 3000
    actions:
      - on: maybe
        command: echo hi
`
	path := writeTempConfig(t, raw)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for unknown 'on' value")
	}
}

func TestLoad_ActionMissingWebhookAndCommand(t *testing.T) {
	raw := `
ports:
  - host: localhost
    port: 3000
    actions:
      - on: open
`
	path := writeTempConfig(t, raw)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error when both webhook and command are missing")
	}
}
