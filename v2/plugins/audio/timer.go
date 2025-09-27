//go:build darwin

package audio

import (
	"sync"
	"time"
)

// Timer 简单的计时器实现
type Timer struct {
	mutex          sync.RWMutex
	startTime      time.Time
	pausedTime     time.Time
	totalPaused    time.Duration
	running        bool
	paused         bool
	duration       time.Duration
	ticker         *time.Ticker
	stopChan       chan bool
	onTick         func()
	onRun          func(bool)
	onPause        func()
	onDone         func(bool)
	tickerInterval time.Duration
	passed         time.Duration
}

// TimerOptions 计时器选项
type TimerOptions struct {
	Duration       time.Duration
	TickerInternal time.Duration
	OnRun          func(bool)
	OnPause        func()
	OnDone         func(bool)
	OnTick         func()
}

// NewTimer 创建新的计时器
func NewTimer(options TimerOptions) *Timer {
	t := &Timer{
		duration:       options.Duration,
		tickerInterval: options.TickerInternal,
		onTick:         options.OnTick,
		onRun:          options.OnRun,
		onPause:        options.OnPause,
		onDone:         options.OnDone,
		stopChan:       make(chan bool, 1),
	}
	return t
}

// Start 启动计时器
func (t *Timer) Start() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.running {
		return
	}

	t.startTime = time.Now()
	t.running = true
	t.paused = false
	t.totalPaused = 0

	if t.onRun != nil {
		t.onRun(true)
	}

	// 启动ticker
	if t.tickerInterval > 0 && t.onTick != nil {
		t.ticker = time.NewTicker(t.tickerInterval)
		go t.tickerLoop()
	}
}

// Stop 停止计时器
func (t *Timer) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.running {
		return
	}

	t.running = false
	t.paused = false

	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
	}

	select {
	case t.stopChan <- true:
	default:
	}

	if t.onDone != nil {
		t.onDone(false)
	}
}

// Pause 暂停计时器
func (t *Timer) Pause() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.running || t.paused {
		return
	}

	t.paused = true
	t.pausedTime = time.Now()

	if t.onPause != nil {
		t.onPause()
	}
}

// Resume 恢复计时器
func (t *Timer) Resume() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.running || !t.paused {
		return
	}

	t.totalPaused += time.Since(t.pausedTime)
	t.paused = false

	if t.onRun != nil {
		t.onRun(false)
	}
}

// Passed 获取已经过的时间
func (t *Timer) Passed() time.Duration {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.passed
}

// SetPassed 设置已经过的时间
func (t *Timer) SetPassed(duration time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.passed = duration
}

// ActualRuntime 获取实际运行时间
func (t *Timer) ActualRuntime() time.Duration {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if !t.running {
		return 0
	}

	now := time.Now()
	if t.paused {
		return t.pausedTime.Sub(t.startTime) - t.totalPaused
	}
	return now.Sub(t.startTime) - t.totalPaused
}

// IsRunning 检查是否正在运行
func (t *Timer) IsRunning() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.running
}

// IsPaused 检查是否已暂停
func (t *Timer) IsPaused() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.paused
}

// tickerLoop ticker循环
func (t *Timer) tickerLoop() {
	for {
		select {
		case <-t.stopChan:
			return
		case <-t.ticker.C:
			t.mutex.RLock()
			if t.running && !t.paused && t.onTick != nil {
				t.onTick()
			}
			t.mutex.RUnlock()
		}
	}
}
