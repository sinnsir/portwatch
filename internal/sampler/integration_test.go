package sampler_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/user/portwatch/internal/sampler"
)

func TestAllow_ConcurrentSafety(t *testing.T) {
	s := sampler.New(sampler.Policy{Rate: 0.5})

	const goroutines = 50
	const callsEach = 200

	var passed atomic.Int64
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < callsEach; i++ {
				if s.Allow() {
					passed.Add(1)
				}
			}
		}()
	}
	wg.Wait()

	total := int64(goroutines * callsEach)
	lo := int64(float64(total) * 0.35)
	hi := int64(float64(total) * 0.65)
	got := passed.Load()
	if got < lo || got > hi {
		t.Fatalf("concurrent sampler: expected %d–%d passed, got %d", lo, hi, got)
	}
}

func TestAllow_FullRateConcurrent(t *testing.T) {
	s := sampler.New(sampler.DefaultPolicy())

	const goroutines = 20
	const callsEach = 500

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < callsEach; i++ {
				if !s.Allow() {
					t.Errorf("expected Allow=true for rate=1.0")
				}
			}
		}()
	}
	wg.Wait()
}
