package timex

import (
	"sync"
	"time"
)

// Options represents configuration for timer.
type Options struct {
	Duration       time.Duration
	Passed         time.Duration
	TickerInternal time.Duration
	OnPause        func()
	OnDone         func(stopped bool)
	OnTick         func()
	OnRun          func(started bool)
}

// Timer represents timer with pause/resume features.
type Timer struct {
	options       Options
	ticker        *time.Ticker
	started       bool
	passed        time.Duration
	actualRuntime time.Duration
	lastTick      time.Time
	done          chan struct{}
	l             sync.RWMutex
}

// Passed returns how much done is already passed.
func (t *Timer) Passed() time.Duration {
	t.l.RLock()
	defer t.l.RUnlock()
	return t.passed
}

// ActualRuntime returns how much time the timer has actually spent running.
func (t *Timer) ActualRuntime() time.Duration {
	t.l.RLock()
	defer t.l.RUnlock()
	return t.actualRuntime
}

// SetPassed update passed.
func (t *Timer) SetPassed(passed time.Duration) {
	t.l.Lock()
	defer t.l.Unlock()
	t.passed = passed
}

// Reset starts elapsed-time accounting for a new item without interrupting the ticker.
func (t *Timer) Reset() {
	t.l.Lock()
	defer t.l.Unlock()
	t.passed = 0
	t.actualRuntime = 0
	t.lastTick = time.Now()
}

// Remaining returns how much time is left to end.
func (t *Timer) Remaining() time.Duration {
	return t.options.Duration - t.Passed()
}

// Run starts just created timer and resumes paused.
func (t *Timer) Run() {
	if t == nil {
		return
	}
	t.l.Lock()
	if t.started && t.ticker != nil {
		t.l.Unlock()
		t.options.OnRun(t.started)
		return
	}

	t.ticker = time.NewTicker(t.options.TickerInternal)
	t.lastTick = time.Now()
	t.started = true
	// Keep local references: the loop must never re-read t.ticker/t.done
	// lock-free — pushDone (Stop/Pause) nils them under t.l, and a racy read
	// could panic on a nil ticker dereference.
	ticker := t.ticker
	done := make(chan struct{})
	t.done = done
	t.l.Unlock()

	t.options.OnRun(true)
	t.options.OnTick()

	for {
		select {
		case tickAt := <-ticker.C:
			// Field updates under lock: Passed/ActualRuntime/SetPassed read
			// and write these fields concurrently from other goroutines.
			t.l.Lock()
			t.passed += tickAt.Sub(t.lastTick)
			t.actualRuntime += tickAt.Sub(t.lastTick)
			t.lastTick = tickAt
			t.l.Unlock()

			t.options.OnTick()

			if t.Remaining() <= 0 {
				t.l.Lock()
				t.pushDone()
				t.l.Unlock()
				t.options.OnDone(false)
				return
			}
			if t.Remaining() <= t.options.TickerInternal {
				t.l.Lock()
				t.pushDone()
				t.l.Unlock()
				time.Sleep(t.Remaining())
				t.l.Lock()
				t.passed = t.options.Duration
				t.l.Unlock()
				t.options.OnTick()
				t.options.OnDone(false)
				return
			}
		case <-done:
			return
		}
	}
}

// Pause temporarily pauses active timer.
func (t *Timer) Pause() {
	t.l.Lock()
	defer t.l.Unlock()

	t.pushDone()
	t.passed += time.Since(t.lastTick)
	t.actualRuntime += time.Since(t.lastTick)
	t.lastTick = time.Now()
	t.options.OnPause()
}

// Stop finishes the timer.
func (t *Timer) Stop() {
	t.l.Lock()
	defer t.l.Unlock()

	t.pushDone()
	t.options.OnDone(true)
}

// NewTimer creates instance of timer.
func NewTimer(options Options) *Timer {
	return &Timer{
		options: options,
	}
}

// pushDone stops the internal ticker and closes the done channel.
// Callers must hold t.l.
func (t *Timer) pushDone() {
	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
	}
	if t.done != nil {
		close(t.done)
		t.done = nil
	}
}
