package alert_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/portcheck"
)

func newFilter(cooldown time.Duration) (*alert.Filter, *time.Time) {
	f := alert.New(alert.Policy{Cooldown: cooldown})
	now := time.Now()
	// Inject controllable clock via exported field — use the returned pointer to advance time.
	return f, &now
}

func TestAllow_FirstCallAlwaysPasses(t *testing.T) {
	f := alert.New(alert.DefaultPolicy())
	if !f.Allow(8080, portcheck.StateOpen) {
		t.Fatal("first alert should always be allowed")
	}
}

func TestAllow_SuppressedWithinCooldown(t *testing.T) {
	f := alert.New(alert.Policy{Cooldown: 10 * time.Second})
	f.Allow(8080, portcheck.StateOpen) // record
	if f.Allow(8080, portcheck.StateOpen) {
		t.Fatal("second alert within cooldown should be suppressed")
	}
}

func TestAllow_DifferentStatesIndependent(t *testing.T) {
	f := alert.New(alert.Policy{Cooldown: 10 * time.Second})
	f.Allow(8080, portcheck.StateOpen)
	if !f.Allow(8080, portcheck.StateClosed) {
		t.Fatal("different state should not be suppressed")
	}
}

func TestAllow_DifferentPortsIndependent(t *testing.T) {
	f := alert.New(alert.Policy{Cooldown: 10 * time.Second})
	f.Allow(8080, portcheck.StateOpen)
	if !f.Allow(9090, portcheck.StateOpen) {
		t.Fatal("different port should not be suppressed")
	}
}

func TestReset_ClearsSuppressionForKey(t *testing.T) {
	f := alert.New(alert.Policy{Cooldown: 10 * time.Second})
	f.Allow(8080, portcheck.StateOpen)
	f.Reset(8080, portcheck.StateOpen)
	if !f.Allow(8080, portcheck.StateOpen) {
		t.Fatal("after Reset, alert should be allowed again")
	}
}

func TestAllow_ZeroCooldownAlwaysPasses(t *testing.T) {
	f := alert.New(alert.Policy{Cooldown: 0})
	f.Allow(8080, portcheck.StateOpen)
	if !f.Allow(8080, portcheck.StateOpen) {
		t.Fatal("zero cooldown should always allow")
	}
}
