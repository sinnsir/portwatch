package transformer_test

import (
	"sync"
	"testing"
	"time"

	"github.com/example/portwatch/internal/transformer"
)

func TestTransformer_FullChain(t *testing.T) {
	tr := transformer.New(
		transformer.NormaliseHost,
		transformer.EnsureTimestamp,
		transformer.RequirePort,
		transformer.AddMeta(map[string]string{"source": "integration"}),
	)

	e := &transformer.Event{
		Host:  "  MYHOST.LOCAL  ",
		Port:  443,
		State: "open",
	}

	if err := tr.Apply(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Host != "myhost.local" {
		t.Fatalf("host not normalised: %q", e.Host)
	}
	if e.Timestamp.IsZero() {
		t.Fatal("timestamp should have been set")
	}
	if e.Meta["source"] != "integration" {
		t.Fatalf("meta not set: %v", e.Meta)
	}
}

func TestTransformer_ConcurrentSafety(t *testing.T) {
	tr := transformer.New(
		transformer.NormaliseHost,
		transformer.EnsureTimestamp,
		transformer.RequirePort,
	)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			e := &transformer.Event{
				Host:  "HOST.EXAMPLE",
				Port:  8080,
				State: "open",
				Timestamp: time.Time{},
			}
			if err := tr.Apply(e); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()
}
