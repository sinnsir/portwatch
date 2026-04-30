package circuit

import (
	"testing"
	"time"
)

func fastPolicy() Policy {
	return Policy{MaxFailures: 3, Cooldown: 50 * time.Millisecond}
}

func TestBreaker_InitialStateIsClosed(t *testing.T) {
	b := New(fastPolicy())
	if b.State() != StateClosed {
		t.Fatalf("expected closed, got %s", b.State())
	}
}

func TestBreaker_AllowPassesWhenClosed(t *testing.T) {
	b := New(fastPolicy())
	if err := b.Allow(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBreaker_OpensAfterMaxFailures(t *testing.T) {
	p := fastPolicy()
	b := New(p)
	for i := 0; i < p.MaxFailures; i++ {
		b.RecordFailure()
	}
	if b.State() != StateOpen {
		t.Fatalf("expected open, got %s", b.State())
	}
	if err := b.Allow(); err != ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestBreaker_SuccessResetsClosed(t *testing.T) {
	b := New(fastPolicy())
	b.RecordFailure()
	b.RecordFailure()
	b.RecordSuccess()
	if b.State() != StateClosed {
		t.Fatalf("expected closed after success, got %s", b.State())
	}
	if err := b.Allow(); err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
}

func TestBreaker_HalfOpenAfterCooldown(t *testing.T) {
	p := fastPolicy()
	b := New(p)
	for i := 0; i < p.MaxFailures; i++ {
		b.RecordFailure()
	}
	time.Sleep(p.Cooldown + 10*time.Millisecond)
	if err := b.Allow(); err != nil {
		t.Fatalf("expected half-open probe to pass, got %v", err)
	}
	if b.State() != StateHalfOpen {
		t.Fatalf("expected half-open, got %s", b.State())
	}
}

func TestStateString(t *testing.T) {
	cases := []struct {
		s    State
		want string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown"},
	}
	for _, c := range cases {
		if got := c.s.String(); got != c.want {
			t.Errorf("State(%d).String() = %q, want %q", c.s, got, c.want)
		}
	}
}
