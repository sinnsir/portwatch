package dedup_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/user/portwatch/internal/dedup"
)

func TestIsDuplicate_ConcurrentSafety(t *testing.T) {
	d := dedup.New(dedup.Policy{Window: 200 * time.Millisecond})
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("host:%d:open", i%10)
			d.IsDuplicate(key)
		}(i)
	}
	wg.Wait()
	// No race detected — primary assertion is the race detector itself.
}

func TestPurge_ConcurrentWithIsDuplicate(t *testing.T) {
	d := dedup.New(dedup.Policy{Window: 20 * time.Millisecond})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			d.IsDuplicate(fmt.Sprintf("k%d", i))
		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(25 * time.Millisecond)
		d.Purge()
	}()
	wg.Wait()
}

func TestIsDuplicate_WindowBoundary(t *testing.T) {
	window := 60 * time.Millisecond
	d := dedup.New(dedup.Policy{Window: window})
	const key = "boundary:1234:closed"

	if d.IsDuplicate(key) {
		t.Fatal("first call must not be duplicate")
	}
	if !d.IsDuplicate(key) {
		t.Fatal("second call within window must be duplicate")
	}
	time.Sleep(window + 10*time.Millisecond)
	if d.IsDuplicate(key) {
		t.Fatal("call after window expiry must not be duplicate")
	}
}
