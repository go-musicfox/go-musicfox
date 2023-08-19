package utils

import (
	"sync"
	"time"
)

// Options represents configuration for timer.
type Options struct {
	Duration       time.Duration
	Passed         time.Duration
	TickerInternal time.Duration
	OnPaused       func()
	OnDone         func(stopped bool)
	OnTick         func()
	OnRun          func(started bool)
}

// Timer represents timer with pause/resume features.
type Timer struct {
	options  Options
	ticker   *time.Ticker
	started  bool
	passed   time.Duration
	lastTick time.Time
	done     chan struct{}
	l        sync.Mutex
}

// Passed returns how much done is already passed.
func (t *Timer) Passed() time.Duration {
	return t.passed
}

// SetPassed update passed.
func (t *Timer) SetPassed(passed time.Duration) {
	t.l.Lock()
	defer t.l.Unlock()
	t.passed = passed
}

// Remaining returns how much time is left to end.
func (t *Timer) Remaining() time.Duration {
	return t.options.Duration - t.Passed()
}

// Run starts just created timer and resumes paused.
func (t *Timer) Run() {
	if t.started && t.ticker != nil {
		t.options.OnRun(t.started)
		return
	}
	t.l.Lock()
	if t.started && t.ticker != nil {
		t.l.Unlock()
		t.options.OnRun(t.started)
		return
	}

	//c.active = true
	t.ticker = time.NewTicker(t.options.TickerInternal)
	t.lastTick = time.Now()
	t.started = true
	t.l.Unlock()

	t.options.OnRun(true)
	t.options.OnTick()
	t.done = make(chan struct{})

	for {
		if t.ticker == nil {
			return
		}
		select {
		case tickAt := <-t.ticker.C:
			t.passed += tickAt.Sub(t.lastTick)
			t.lastTick = time.Now()
			t.options.OnTick()
			if t.Remaining() <= 0 {
				t.pushDone()
				t.options.OnDone(false)
			} else if t.Remaining() <= t.options.TickerInternal {
				t.pushDone()
				time.Sleep(t.Remaining())
				t.passed = t.options.Duration
				t.options.OnTick()
				t.options.OnDone(false)

			}
		case <-t.done:
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
	t.lastTick = time.Now()
	t.options.OnPaused()
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
