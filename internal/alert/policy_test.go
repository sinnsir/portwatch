package alert_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
)

func TestDefaultPolicy_HasPositiveCooldown(t *testing.T) {
	p := alert.DefaultPolicy()
	if p.Cooldown <= 0 {
		t.Fatalf("expected positive cooldown, got %v", p.Cooldown)
	}
}

func TestDefaultPolicy_CooldownAtLeastOneSecond(t *testing.T) {
	p := alert.DefaultPolicy()
	if p.Cooldown < time.Second {
		t.Fatalf("default cooldown too short: %v", p.Cooldown)
	}
}

func TestCustomPolicy_Respected(t *testing.T) {
	custom := alert.Policy{Cooldown: 5 * time.Minute}
	if custom.Cooldown != 5*time.Minute {
		t.Fatalf("unexpected cooldown: %v", custom.Cooldown)
	}
}
