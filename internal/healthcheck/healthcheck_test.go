package healthcheck_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/healthcheck"
)

func TestCheck_AllHealthy(t *testing.T) {
	probe := healthcheck.Probe{
		Name: "test_component",
		Fn:   func() error { return nil },
	}
	c := healthcheck.New(probe)
	s := c.Check()

	if !s.Healthy {
		t.Error("expected healthy status")
	}
	if s.Components["test_component"] != "ok" {
		t.Errorf("expected ok, got %q", s.Components["test_component"])
	}
}

func TestCheck_UnhealthyProbe(t *testing.T) {
	probe := healthcheck.Probe{
		Name: "broken",
		Fn:   func() error { return errors.New("db down") },
	}
	c := healthcheck.New(probe)
	s := c.Check()

	if s.Healthy {
		t.Error("expected unhealthy status")
	}
	if !strings.Contains(s.Components["broken"], "db down") {
		t.Errorf("expected error detail in component status, got %q", s.Components["broken"])
	}
}

func TestCheck_NoProbes(t *testing.T) {
	c := healthcheck.New()
	s := c.Check()

	if !s.Healthy {
		t.Error("expected healthy when no probes registered")
	}
	if len(s.Components) != 0 {
		t.Errorf("expected empty components, got %v", s.Components)
	}
}

func TestCheck_MetadataPopulated(t *testing.T) {
	c := healthcheck.New()
	before := time.Now()
	s := c.Check()

	if s.GoVersion == "" {
		t.Error("expected GoVersion to be set")
	}
	if s.NumCPU <= 0 {
		t.Errorf("expected positive NumCPU, got %d", s.NumCPU)
	}
	if s.StartedAt.Before(before.Add(-time.Second)) {
		t.Error("StartedAt should be recent")
	}
	if s.Uptime == "" {
		t.Error("expected Uptime to be set")
	}
}

func TestCheck_MixedProbes(t *testing.T) {
	probes := []healthcheck.Probe{
		{Name: "ok_one", Fn: func() error { return nil }},
		{Name: "bad_one", Fn: func() error { return errors.New("fail") }},
		{Name: "ok_two", Fn: func() error { return nil }},
	}
	c := healthcheck.New(probes...)
	s := c.Check()

	if s.Healthy {
		t.Error("expected unhealthy when at least one probe fails")
	}
	if s.Components["ok_one"] != "ok" || s.Components["ok_two"] != "ok" {
		t.Error("expected passing probes to report ok")
	}
}
