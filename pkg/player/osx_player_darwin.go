//go:build darwin
// +build darwin

package player

import (
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
	"sync"
	"time"

	"github.com/progrium/macdriver/avcore"
	"go-musicfox/utils"
)

type osxPlayer struct {
	l sync.Mutex

	player  *avcore.AVPlayer
	handler objc.Object

	curMusic  UrlMusic
	curSongId int
	timer     *utils.Timer
	//latestPlayTime time.Time //避免切歌时产生的stop信号造成影响

	volume    int
	state     State
	timeChan  chan time.Duration
	stateChan chan State
	musicChan chan UrlMusic

	close chan struct{}
}

func NewOsxPlayer() Player {
	p := &osxPlayer{
		state:     Stopped,
		timeChan:  make(chan time.Duration),
		stateChan: make(chan State),
		musicChan: make(chan UrlMusic),
		close:     make(chan struct{}),
	}

	clsName := "AVPlayerHandler"
	handlerCls := objc.NewClass(clsName, "NSObject")
	handlerCls.AddMethod("handle:", func(_ objc.Object, ns objc.Object) {
		//_ = core.NSNotification_fromRef(ns)
		//_ = mediaplayer.NSNotif(ns)
		//utils.Logger().Println(n)
		p.Stop()
	})
	objc.RegisterClass(handlerCls)
	p.handler = objc.Get(clsName).Alloc().Init()

	avPlayer := avcore.AVPlayer_alloc().Init_asAVPlayer()
	p.player = &avPlayer
	p.player.SetActionAtItemEnd_(2) // do nothing => https://developer.apple.com/documentation/avfoundation/avplayeractionatitemend/avplayeractionatitemendnone?language=objc

	go func() {
		defer utils.Recover(false)
		p.listen()
	}()

	return p
}

// listen 开始监听
func (p *osxPlayer) listen() {
	for {
		select {
		case <-p.close:
			return
		case p.curMusic = <-p.musicChan:
			p.Paused()
			// 重置
			{
				if p.timer != nil {
					p.timer.Stop()
				}
			}

			item := avcore.AVPlayerItem_playerItemWithURL_(core.NSURL_URLWithString_(core.String(p.curMusic.Url)))
			p.player.ReplaceCurrentItemWithPlayerItem_(item)

			core.NSNotificationCenter_defaultCenter().
				AddObserver_selector_name_object_(p.handler, objc.Sel("handle:"), core.String("AVPlayerItemDidPlayToEndTimeNotification"), p.player.CurrentItem())

			// 计时器
			p.timer = utils.NewTimer(utils.Options{
				Duration:       8760 * time.Hour,
				TickerInternal: 500 * time.Millisecond,
				OnRun:          func(started bool) {},
				OnPaused:       func() {},
				OnDone:         func(stopped bool) {},
				OnTick: func() {
					t := p.player.CurrentTime()
					var d time.Duration
					if t.Timescale > 0 {
						d = time.Second * time.Duration(t.Value/int64(t.Timescale))
					}
					select {
					case p.timeChan <- d:
					default:
					}
				},
			})
			p.Resume()
		}
	}
}

func (p *osxPlayer) setState(state State) {
	p.state = state
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
	}
}

func (p *osxPlayer) Play(music UrlMusic) {
	select {
	case p.musicChan <- music:
	default:
	}
}

func (p *osxPlayer) CurMusic() UrlMusic {
	return p.curMusic
}

func (p *osxPlayer) Paused() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state != Playing {
		return
	}
	p.player.Pause()
	p.timer.Pause()
	p.setState(Paused)
}

func (p *osxPlayer) Resume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state == Playing {
		return
	}
	go p.timer.Run()
	p.player.Play()
	p.setState(Playing)
}

func (p *osxPlayer) Stop() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state == Stopped {
		return
	}
	p.player.Pause()
	p.timer.Pause()
	p.setState(Stopped)
}

func (p *osxPlayer) Toggle() {
	switch p.State() {
	case Paused, Stopped:
		p.Resume()
	case Playing:
		p.Paused()
	}
}

func (p *osxPlayer) Seek(duration time.Duration) {
	p.l.Lock()
	defer p.l.Unlock()
	scale := p.player.CurrentItem().Duration().Timescale
	if scale == 0 {
		return
	}

	p.player.SeekToTime_(core.CMTime{
		Value:     int64(scale) * int64(duration.Seconds()),
		Timescale: scale,
		Flags:     1,
	})
	p.timer.SetPassed(duration)
}

func (p *osxPlayer) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.Passed()
}

func (p *osxPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

func (p *osxPlayer) State() State {
	return p.state
}

func (p *osxPlayer) StateChan() <-chan State {
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
	p.player.SetVolume_(float32(p.volume) / 100.0)
}

func (p *osxPlayer) DownVolume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.volume-5 <= 0 {
		p.volume = 0
	} else {
		p.volume -= 5
	}
	p.player.SetVolume_(float32(p.volume) / 100.0)
}

func (p *osxPlayer) Close() {
	if p.timer != nil {
		p.timer.Stop()
	}

	p.close <- struct{}{}
	p.player.Release()
}
