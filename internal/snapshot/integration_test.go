package snapshot_test

import (
	"sync"
	"testing"

	"github.com/user/portwatch/internal/snapshot"
)

func TestSnapshot_ConcurrentSafety(t *testing.T) {
	s, err := snapshot.New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		port := 8000 + i
		go func(p int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				state := "open"
				if j%2 == 0 {
					state = "closed"
				}
				_ = s.Set("localhost", p, state)
				_, _ = s.Get("localhost", p)
				_ = s.All()
			}
		}(port)
	}

	wg.Wait()

	all := s.All()
	if len(all) != goroutines {
		t.Errorf("expected %d entries, got %d", goroutines, len(all))
	}
}

func TestSnapshot_MultiplePortsSameHost(t *testing.T) {
	s, _ := snapshot.New(tempPath(t))
	ports := []int{80, 443, 8080, 8443, 9090}

	for _, p := range ports {
		if err := s.Set("example.com", p, "open"); err != nil {
			t.Fatalf("Set port %d: %v", p, err)
		}
	}

	for _, p := range ports {
		e, ok := s.Get("example.com", p)
		if !ok {
			t.Errorf("missing entry for port %d", p)
		}
		if e.State != "open" {
			t.Errorf("port %d: state = %q, want open", p, e.State)
		}
	}

	if len(s.All()) != len(ports) {
		t.Errorf("expected %d entries, got %d", len(ports), len(s.All()))
	}
}
