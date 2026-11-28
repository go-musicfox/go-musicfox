package player

import (
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/timex"
)

// newTestMpvPlayer 构造一个不启动 mpv 守护进程的 mpvPlayer，仅用于状态/计时逻辑测试。
func newTestMpvPlayer() *mpvPlayer {
	p := &mpvPlayer{
		state:     types.Playing,
		timeChan:  make(chan time.Duration, 1),
		stateChan: make(chan types.State, 10),
		closeCh:   make(chan struct{}),
	}
	p.timer = timex.NewTimer(timex.Options{
		Duration:       8760 * time.Hour,
		TickerInternal: 2 * time.Millisecond,
		OnRun:          func(bool) {},
		OnPause:        func() {},
		OnDone:         func(bool) {},
		OnTick: func() {
			select {
			case p.timeChan <- p.timer.Passed():
			default:
			}
		},
	})
	return p
}

// TestMpvNaturalEndKeepsPlayedTimeForReporting 回归测试：
// mpv 自然结束时不能把 timer 置空，否则 UI 在切歌前读取 PlayedTime() 会得到 0，
// 导致网易云 / Last.fm 上报被「不足 20 秒」门槛静默丢弃。
func TestMpvNaturalEndKeepsPlayedTimeForReporting(t *testing.T) {
	p := newTestMpvPlayer()
	go p.timer.Run()
	defer p.timer.Stop()

	deadline := time.Now().Add(time.Second)
	for p.PlayedTime() <= 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	before := p.PlayedTime()
	if before <= 0 {
		t.Fatalf("timer did not advance before end-file: PlayedTime=%v", before)
	}

	p.handleEndFile(true, false)

	if p.State() != types.Stopped {
		t.Fatalf("state = %v, want Stopped", p.State())
	}
	if p.timer == nil {
		t.Fatal("timer was cleared on natural end; PlayedTime would be 0 and the report dropped")
	}
	if after := p.PlayedTime(); after <= 0 {
		t.Fatalf("PlayedTime = %v after natural end, want > 0 (before=%v)", after, before)
	}
}

// TestMpvEndFileIgnoresNonNatural 非自然结束（stop/new-file）不应触发 Stopped。
func TestMpvEndFileIgnoresNonNatural(t *testing.T) {
	p := newTestMpvPlayer()
	defer p.timer.Stop()

	p.handleEndFile(false, false)

	if p.State() != types.Playing {
		t.Fatalf("state = %v, want Playing for non-natural end", p.State())
	}
}
