package timex

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTimerTracksPassedTime(t *testing.T) {
	var tickCount atomic.Int64
	timer := NewTimer(Options{
		Duration:       time.Hour,
		TickerInternal: 10 * time.Millisecond,
		OnRun:          func(bool) {},
		OnPause:        func() {},
		OnDone:         func(bool) {},
		OnTick:         func() { tickCount.Add(1) },
	})
	go timer.Run()
	defer timer.Stop()

	time.Sleep(100 * time.Millisecond)
	passed := timer.Passed()
	if passed < 50*time.Millisecond || passed > 300*time.Millisecond {
		t.Fatalf("Passed = %v, want ~100ms", passed)
	}
	if tickCount.Load() < 3 {
		t.Fatalf("tick count = %d, want >= 3", tickCount.Load())
	}
}

func TestTimerStopHaltsTicks(t *testing.T) {
	ticks := make(chan struct{}, 16)
	timer := NewTimer(Options{
		Duration:       time.Hour,
		TickerInternal: 5 * time.Millisecond,
		OnRun:          func(bool) {},
		OnPause:        func() {},
		OnDone:         func(bool) {},
		OnTick:         func() { ticks <- struct{}{} },
	})
	go timer.Run()

	select {
	case <-ticks:
	case <-time.After(time.Second):
		t.Fatal("timer never ticked")
	}

	timer.Stop()
	// Allow an in-flight tick to settle, then require silence.
	time.Sleep(10 * time.Millisecond)
	for {
		select {
		case <-ticks:
		default:
			goto drained
		}
	}
drained:
	select {
	case <-ticks:
		t.Fatal("tick delivered after Stop")
	case <-time.After(30 * time.Millisecond):
	}
}

// TestTimerConcurrentAccessNoRace hammers Passed/ActualRuntime/SetPassed from
// several goroutines while the timer goroutine keeps ticking. Run with -race.
func TestTimerConcurrentAccessNoRace(t *testing.T) {
	var tickCount atomic.Int64
	timer := NewTimer(Options{
		Duration:       time.Hour,
		TickerInternal: 2 * time.Millisecond,
		OnRun:          func(bool) {},
		OnPause:        func() {},
		OnDone:         func(bool) {},
		OnTick:         func() { tickCount.Add(1) },
	})
	go timer.Run()
	defer timer.Stop()

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				_ = timer.Passed()
				_ = timer.ActualRuntime()
				timer.SetPassed(time.Duration(j) * time.Millisecond)
			}
		}()
	}
	wg.Wait()

	before := tickCount.Load()
	time.Sleep(20 * time.Millisecond)
	if tickCount.Load() <= before {
		t.Fatal("timer stopped ticking after concurrent access")
	}
}
