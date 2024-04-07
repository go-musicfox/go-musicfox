//go:build windows

package player

import (
	"sync"
	"time"
	"unsafe"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils"
	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go/windows/foundation"
	"github.com/saltosystems/winrt-go/windows/media/core"
	"github.com/saltosystems/winrt-go/windows/media/playback"
)

const (
	TicksPerMicrosecond int64 = 10
	TicksPerMillisecond       = TicksPerMicrosecond * 1000
	TicksPerSecond            = TicksPerMillisecond * 1000
)

type winMediaPlayer struct {
	l sync.Mutex

	player *playback.MediaPlayer

	curMusic UrlMusic
	timer    *utils.Timer

	volume    int
	state     types.State
	timeChan  chan time.Duration
	stateChan chan types.State
	musicChan chan UrlMusic

	close chan struct{}
}

func NewWinMediaPlayer() Player {
	p := &winMediaPlayer{
		state:     types.Stopped,
		timeChan:  make(chan time.Duration),
		stateChan: make(chan types.State),
		musicChan: make(chan UrlMusic),
		close:     make(chan struct{}),
		volume:    100,
	}

	utils.Must(ole.CoInitialize(0))

	p.player = utils.Must1(playback.NewMediaPlayer())
	utils.Must(p.player.SetVolume(float64(p.volume / 100.0)))
	utils.Must(p.player.SetAudioCategory(playback.MediaPlayerAudioCategoryMedia))

	// TODO: add 监听播放状态

	go utils.PanicRecoverWrapper(false, p.listen)

	return p
}

// listen 开始监听
func (p *winMediaPlayer) listen() {
	var (
		uri         *foundation.Uri
		mediaSource *core.MediaSource
	)
	for {
		select {
		case <-p.close:
			return
		case p.curMusic = <-p.musicChan:
			p.Paused()
			if p.timer != nil {
				p.timer.SetPassed(0)
			}
			if p.timer != nil {
				p.timer.Stop()
				p.timer = nil
			}
			if mediaSource != nil {
				mediaSource.Release()
			}
			if uri != nil {
				uri.Release()
			}

			uri = utils.Must1(foundation.UriCreateUri(p.curMusic.Url))
			mediaSource = utils.Must1(core.MediaSourceCreateFromUri(uri))
			utils.Must(p.player.SetSource((*playback.IMediaPlaybackSource)(unsafe.Pointer(mediaSource))))

			// 计时器
			p.timer = utils.NewTimer(utils.Options{
				Duration:       8760 * time.Hour,
				TickerInternal: 500 * time.Millisecond,
				OnRun:          func(started bool) {},
				OnPaused:       func() {},
				OnDone:         func(stopped bool) {},
				OnTick: func() {
					var curTime time.Duration
					t := utils.Must1(p.player.GetTimelineControllerPositionOffset())
					if t.Duration <= 0 {
						return
					}
					curTime = time.Duration(t.Duration/TicksPerMillisecond) * time.Millisecond
					select {
					//osx_player存在一点延迟
					case p.timeChan <- curTime + time.Millisecond*800:
					//case p.timeChan <- p.timer.Passed():
					default:
					}
				},
			})
			p.Resume()
		}
	}
}

func (p *winMediaPlayer) setState(state types.State) {
	p.state = state
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
	}
}

func (p *winMediaPlayer) Play(music UrlMusic) {
	select {
	case p.musicChan <- music:
	default:
	}
}

func (p *winMediaPlayer) CurMusic() UrlMusic {
	return p.curMusic
}

func (p *winMediaPlayer) Paused() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state != types.Playing {
		return
	}
	utils.Must(p.player.Pause())
	p.timer.Pause()
	p.setState(types.Paused)
}

func (p *winMediaPlayer) Resume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state == types.Playing {
		return
	}
	go p.timer.Run()
	utils.Must(p.player.Play())
	p.setState(types.Playing)
}

func (p *winMediaPlayer) Stop() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state == types.Stopped {
		return
	}
	utils.Must(p.player.Pause())
	p.timer.Pause()
	p.setState(types.Stopped)
}

func (p *winMediaPlayer) Toggle() {
	switch p.State() {
	case types.Paused, types.Stopped:
		p.Resume()
	case types.Playing:
		p.Paused()
	default:
		p.Resume()
	}
}

func (p *winMediaPlayer) Seek(duration time.Duration) {
	p.l.Lock()
	defer p.l.Unlock()
	//scale := p.player.CurrentItem().Duration().Timescale
	//if scale == 0 {
	//	return
	//}
	//p.player.SeekToTime(avcore.CMTime{
	//	Value:     int64(float64(scale) * duration.Seconds()),
	//	Timescale: scale,
	//	Flags:     1,
	//})
	p.timer.SetPassed(duration)
}

func (p *winMediaPlayer) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	var curTime time.Duration
	t := utils.Must1(p.player.GetTimelineControllerPositionOffset())
	if t.Duration <= 0 {
		return curTime
	}
	curTime = time.Duration(t.Duration/TicksPerMillisecond) * time.Millisecond
	return curTime
}

func (p *winMediaPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

func (p *winMediaPlayer) State() types.State {
	return p.state
}

func (p *winMediaPlayer) StateChan() <-chan types.State {
	return p.stateChan
}

func (p *winMediaPlayer) UpVolume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.volume+5 >= 100 {
		p.volume = 100
	} else {
		p.volume += 5
	}
	if p.player != nil {
		utils.Must(p.player.SetVolume(float64(p.volume) / 100.0))
	}
}

func (p *winMediaPlayer) DownVolume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.volume-5 <= 0 {
		p.volume = 0
	} else {
		p.volume -= 5
	}
	if p.player != nil {
		utils.Must(p.player.SetVolume(float64(p.volume) / 100.0))
	}
}

func (p *winMediaPlayer) Volume() int {
	return p.volume
}

func (p *winMediaPlayer) SetVolume(volume int) {
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
		utils.Must(p.player.SetVolume(float64(p.volume) / 100.0))
	}
}

func (p *winMediaPlayer) Close() {
	p.l.Lock()
	defer p.l.Unlock()

	if p.timer != nil {
		p.timer.Stop()
	}

	if p.close != nil {
		close(p.close)
		p.close = nil
	}
	if p.player != nil {
		p.player.Release()
	}
}
