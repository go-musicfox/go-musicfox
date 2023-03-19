//go:build darwin

package player

import (
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/progrium/macdriver/avcore"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
)

type osxPlayer struct {
	l sync.Mutex

	player  *avcore.AVPlayer
	handler objc.Object

	curMusic UrlMusic
	timer    *utils.Timer

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
		volume:    100,
	}

	clsName := "AVPlayerHandler"
	handlerCls := objc.NewClass(clsName, "NSObject")
	handlerCls.AddMethod("handleFinish:", func(_ objc.Object, ns objc.Object) {
		// 这里会出现两次通知
		url := avcore.AVPlayerItem_fromRef(ns.Get("object")).Asset().Get("URL")
		curUrl := p.player.CurrentItem().Asset().Get("URL")
		if url == curUrl {
			p.Stop()
		}
	})
	objc.RegisterClass(handlerCls)
	p.handler = objc.Get(clsName).Alloc().Init()

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
			objc.Autorelease(func() {
				p.Paused()
				if p.timer != nil {
					p.timer.SetPassed(0)
				}
				// 重置
				{
					if p.timer != nil {
						p.timer.Stop()
						p.timer = nil
					}
					if p.player != nil {
						//p.player.ReplaceCurrentItemWithPlayerItem_(nil)
						p.player.Release()
					}
				}

				// 有内存释放问题，所以每次都重新创建Player
				avPlayer := avcore.AVPlayer_alloc().Init_asAVPlayer()
				item := avcore.AVPlayerItem_playerItemWithURL_(core.NSURL_URLWithString_(core.String(p.curMusic.Url)))
				p.player = &avPlayer
				p.player.InitWithPlayerItem__asAVPlayer(item)
				p.player.SetActionAtItemEnd_(2) // do nothing => https://developer.apple.com/documentation/avfoundation/avplayeractionatitemend/avplayeractionatitemendnone?language=objc
				p.player.SetVolume_(float32(p.volume) / 100.0)

				core.NSNotificationCenter_defaultCenter().
					AddObserver_selector_name_object_(p.handler, objc.Sel("handleFinish:"), core.String("AVPlayerItemDidPlayToEndTimeNotification"), p.player.CurrentItem())

				// 计时器
				p.timer = utils.NewTimer(utils.Options{
					Duration:       8760 * time.Hour,
					TickerInternal: 500 * time.Millisecond,
					OnRun:          func(started bool) {},
					OnPaused:       func() {},
					OnDone:         func(stopped bool) {},
					OnTick: func() {
						// FIXME 获取osx播放器有内存释放问题，先用timer的时间
						//var curTime time.Duration
						//objc.Autorelease(func() {
						//	t := p.player.CurrentTime()
						//	curTime = time.Duration(t.Value/int64(t.Timescale)) * time.Second
						//})
						select {
						//osx_player存在一点延迟
						//case p.timeChan <- curTime + time.Millisecond*800:
						case p.timeChan <- p.timer.Passed():
						default:
						}
					},
				})
				p.Resume()
			})
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
	if p.player != nil {
		p.player.SetVolume_(float32(p.volume) / 100.0)
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
		p.player.SetVolume_(float32(p.volume) / 100.0)
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
		p.player.SetVolume_(float32(p.volume) / 100.0)
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
	objc.Autorelease(func() {
		p.handler.Release()
		if p.player != nil {
			p.player.Release()
		}
	})
}
