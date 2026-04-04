//go:build darwin

package player

import (
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/avcore"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/timex"
)

type osxPlayer struct {
	l sync.Mutex

	player  *avcore.AVPlayer
	handler *playerHandler

	curMusic URLMusic
	timer    *timex.Timer

	volume    int
	state     types.State
	timeChan  chan time.Duration
	stateChan chan types.State
	musicChan chan URLMusic

	close chan struct{}
}

func NewOsxPlayer() *osxPlayer {
	p := &osxPlayer{
		state:     types.Stopped,
		timeChan:  make(chan time.Duration, 1),
		stateChan: make(chan types.State, 10),
		musicChan: make(chan URLMusic, 1),
		close:     make(chan struct{}),
		volume:    100,
	}

	handler := playerHandler_new(p)
	p.handler = &handler

	avPlayer := avcore.AVPlayer_alloc().Init()
	p.player = &avPlayer
	p.player.SetActionAtItemEnd(2) // do nothing => https://developer.apple.com/documentation/avfoundation/avplayeractionatitemend/avplayeractionatitemendnone?language=objc
	p.player.SetVolume(float32(p.volume) / 100.0)
	cocoa.NSNotificationCenter_defaultCenter().
		AddObserverSelectorNameObject(p.handler.ID, sel_handleFinish, core.String("AVPlayerItemDidPlayToEndTimeNotification"), p.player.CurrentItem().NSObject)
	cocoa.NSNotificationCenter_defaultCenter().
		AddObserverSelectorNameObject(p.handler.ID, sel_handleFailed, core.String("AVPlayerItemFailedToPlayToEndTimeNotification"), p.player.CurrentItem().NSObject)

	errorx.WaitGoStart(p.listen)

	return p
}

// listen 开始监听
func (p *osxPlayer) listen() {
	for {
		select {
		case <-p.close:
			return
		case p.curMusic = <-p.musicChan:
			core.Autorelease(func() {
				p.Pause()
				if p.timer != nil {
					p.timer.SetPassed(0)
				}
				if p.timer != nil {
					p.timer.Stop()
					// p.timer = nil
				}

				item := avcore.AVPlayerItem_playerItemWithURL(core.NSURL_URLWithString(core.String(p.curMusic.URL)))
				p.player.ReplaceCurrentItemWithPlayerItem(item)

				// 重新注册通知监听器，因为CurrentItem已经改变
				cocoa.NSNotificationCenter_defaultCenter().
					AddObserverSelectorNameObject(p.handler.ID, sel_handleFinish, core.String("AVPlayerItemDidPlayToEndTimeNotification"), p.player.CurrentItem().NSObject)
				cocoa.NSNotificationCenter_defaultCenter().
					AddObserverSelectorNameObject(p.handler.ID, sel_handleFailed, core.String("AVPlayerItemFailedToPlayToEndTimeNotification"), p.player.CurrentItem().NSObject)

				// 计时器
				p.timer = timex.NewTimer(timex.Options{
					Duration:       8760 * time.Hour,
					TickerInternal: configs.AppConfig.Main.FrameRate.Interval(),
					OnRun:          func(started bool) {},
					OnPause:        func() {},
					OnDone:         func(stopped bool) {},
					OnTick: func() {
						var curTime time.Duration
						core.Autorelease(func() {
							t := p.player.CurrentTime()
							if t.Timescale <= 0 {
								return
							}
							curTime = time.Duration(t.Value/int64(t.Timescale)) * time.Second
						})
						select {
						// osx_player存在一点延迟
						case p.timeChan <- curTime + time.Millisecond*800:
						// case p.timeChan <- p.timer.Passed():
						default:
						}
					},
				})
				p.Resume()
			})
		}
	}
}

func (p *osxPlayer) setState(state types.State) {
	p.state = state
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
	}
}

func (p *osxPlayer) Play(music URLMusic) {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case p.musicChan <- music:
	case <-timer.C:
	}
}

func (p *osxPlayer) CurMusic() URLMusic {
	return p.curMusic
}

func (p *osxPlayer) Pause() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state != types.Playing {
		return
	}
	p.player.Pause()
	p.timer.Pause()
	p.setState(types.Paused)
}

func (p *osxPlayer) Resume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state == types.Playing {
		return
	}

	go p.timer.Run()
	p.player.Play()

	p.setState(types.Playing)
}

func (p *osxPlayer) Stop() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state == types.Stopped {
		return
	}
	p.player.Pause()
	p.timer.Pause()
	p.setState(types.Stopped)
}

func (p *osxPlayer) Toggle() {
	switch p.State() {
	case types.Paused, types.Stopped:
		p.Resume()
	case types.Playing:
		p.Pause()
	default:
		p.Resume()
	}
}

func (p *osxPlayer) Seek(duration time.Duration) {
	p.l.Lock()
	defer p.l.Unlock()
	scale := p.player.CurrentItem().Duration().Timescale
	if scale == 0 {
		return
	}
	p.player.SeekToTime(avcore.CMTime{
		Value:     int64(float64(scale) * duration.Seconds()),
		Timescale: scale,
		Flags:     1,
	})
	p.timer.SetPassed(duration)
}

func (p *osxPlayer) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	var curTime time.Duration
	core.Autorelease(func() {
		t := p.player.CurrentTime()
		if t.Timescale <= 0 {
			return
		}
		curTime = time.Duration(float64(t.Value*1000.0)/float64(t.Timescale)) * time.Millisecond
	})
	return curTime
}

func (p *osxPlayer) PlayedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.ActualRuntime()
}

func (p *osxPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

func (p *osxPlayer) State() types.State {
	return p.state
}

func (p *osxPlayer) StateChan() <-chan types.State {
	return p.stateChan
}

func (p *osxPlayer) UpVolume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.volume+5 >= 100 {
		p.volume = 100
	} else {
		p.volume += 5
	}
	if p.player != nil {
		p.player.SetVolume(float32(p.volume) / 100.0)
	}
}

func (p *osxPlayer) DownVolume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.volume-5 <= 0 {
		p.volume = 0
	} else {
		p.volume -= 5
	}
	if p.player != nil {
		p.player.SetVolume(float32(p.volume) / 100.0)
	}
}

func (p *osxPlayer) Volume() int {
	return p.volume
}

func (p *osxPlayer) SetVolume(volume int) {
	if volume > 100 {
		volume = 100
	}
	if volume < 0 {
		volume = 0
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.volume = volume
	if p.player != nil {
		p.player.SetVolume(float32(p.volume) / 100.0)
	}
}

func (p *osxPlayer) Close() {
	p.l.Lock()
	defer p.l.Unlock()

	if p.timer != nil {
		p.timer.Stop()
	}

	if p.close != nil {
		close(p.close)
		p.close = nil
	}
	p.handler.Release()
	if p.player != nil {
		p.player.Release()
	}
}
