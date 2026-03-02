//go:build windows

package player

import (
	"sync"
	"time"
	"unsafe"

	. "github.com/go-musicfox/go-musicfox/utils/errorx"
	. "github.com/go-musicfox/go-musicfox/utils/timex"
	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go"
	"github.com/saltosystems/winrt-go/windows/foundation"
	"github.com/saltosystems/winrt-go/windows/media/core"
	"github.com/saltosystems/winrt-go/windows/media/playback"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	control "github.com/go-musicfox/go-musicfox/internal/remote_control"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

const (
	TicksPerMicrosecond int64 = 10
	TicksPerMillisecond       = TicksPerMicrosecond * 1000
	TicksPerSecond            = TicksPerMillisecond * 1000

	SignatureIInspectable = "cinterface(IInspectable)"
)

var (
	playbackSessionEventGUID = winrt.ParameterizedInstanceGUID(
		foundation.GUIDTypedEventHandler,
		playback.SignatureMediaPlaybackSession,
		SignatureIInspectable,
	)
	playerEventGUID = winrt.ParameterizedInstanceGUID(
		foundation.GUIDTypedEventHandler,
		playback.SignatureMediaPlayer,
		SignatureIInspectable,
	)
	playerFailedEventGUID = winrt.ParameterizedInstanceGUID(
		foundation.GUIDTypedEventHandler,
		playback.SignatureMediaPlayer,
		playback.SignatureMediaPlayerFailedEventArgs,
	)
)

type winMediaPlayer struct {
	l sync.Mutex

	player *playback.MediaPlayer

	curMusic URLMusic
	timer    *Timer

	volume    int
	state     types.State
	timeChan  chan time.Duration
	stateChan chan types.State
	musicChan chan URLMusic

	close chan struct{}
}

func NewWinMediaPlayer() *winMediaPlayer {
	p := &winMediaPlayer{
		state:     types.Stopped,
		timeChan:  make(chan time.Duration, 1),
		stateChan: make(chan types.State, 10),
		musicChan: make(chan URLMusic, 1),
		close:     make(chan struct{}),
		volume:    100,
	}

	p.buildWinPlayer()

	WaitGoStart(p.listen)

	return p
}

func (p *winMediaPlayer) buildWinPlayer() {
	_ = ole.RoInitialize(1)
	p.player = Must1(playback.NewMediaPlayer())
	Must(p.player.SetVolume(float64(p.volume / 100.0)))
	Must(p.player.SetAudioCategory(playback.MediaPlayerAudioCategoryMedia))

	cmdManager := Must1(p.player.GetCommandManager())
	defer cmdManager.Release()
	Must(cmdManager.SetIsEnabled(false))

	control.SMTC = Must1(p.player.GetSystemMediaTransportControls())

	playbackSession := Must1(p.player.GetPlaybackSession())
	defer playbackSession.Release()

	// state changed
	stateHandler := foundation.NewTypedEventHandler(
		ole.NewGUID(playbackSessionEventGUID),
		func(h *foundation.TypedEventHandler, sender, _ unsafe.Pointer) {
			session := (*playback.MediaPlaybackSession)(sender)
			switch Must1(session.GetPlaybackState()) {
			case playback.MediaPlaybackStatePlaying:
				p.Resume()
				p.setState(types.Playing)
			case playback.MediaPlaybackStatePaused:
				p.Pause()
				p.setState(types.Paused)
			}
		},
	)
	defer stateHandler.Release()
	Must1(playbackSession.AddPlaybackStateChanged(stateHandler))

	// current state changed(old version)
	curStateHandler := foundation.NewTypedEventHandler(
		ole.NewGUID(playerEventGUID),
		func(_ *foundation.TypedEventHandler, sender, _ unsafe.Pointer) {
			player := (*playback.MediaPlayer)(sender)
			switch Must1(player.GetCurrentState()) {
			case playback.MediaPlayerStatePlaying:
				p.Resume()
				p.setState(types.Playing)
			case playback.MediaPlayerStatePaused:
				p.Pause()
				p.setState(types.Paused)
			case playback.MediaPlayerStateStopped:
				p.Stop()
				p.setState(types.Stopped)
			}
		},
	)
	defer curStateHandler.Release()
	Must1(p.player.AddCurrentStateChanged(curStateHandler))

	// finished
	finishedHandler := foundation.NewTypedEventHandler(
		ole.NewGUID(playerEventGUID),
		func(_ *foundation.TypedEventHandler, _, _ unsafe.Pointer) {
			p.Stop()
			p.setState(types.Stopped)
		},
	)
	defer finishedHandler.Release()
	failedHandler := foundation.NewTypedEventHandler(
		ole.NewGUID(playerFailedEventGUID),
		func(_ *foundation.TypedEventHandler, _, _ unsafe.Pointer) {
			p.Stop()
			p.setState(types.Stopped)
		},
	)
	defer failedHandler.Release()
	Must1(p.player.AddMediaEnded(finishedHandler))
	Must1(p.player.AddMediaFailed(failedHandler))
}

// listen 开始监听
func (p *winMediaPlayer) listen() {
	var (
		uri          *foundation.Uri
		mediaSource  *core.MediaSource
		playbackItem *playback.MediaPlaybackItem
		reset        = func() {
			if p.timer != nil {
				p.timer.SetPassed(0)
			}
			if p.timer != nil {
				p.timer.Stop()
				p.timer = nil
			}
			if uri != nil {
				uri.Release()
			}
			if mediaSource != nil {
				mediaSource.Release()
			}
			if playbackItem != nil {
				playbackItem.Release()
			}
		}
	)
	for {
		select {
		case <-p.close:
			reset()
			return
		case p.curMusic = <-p.musicChan:
			p.Pause()
			reset()

			uri = Must1(foundation.UriCreateUri(p.curMusic.URL))
			mediaSource = Must1(core.MediaSourceCreateFromUri(uri))
			Must(p.player.SetSource((*playback.IMediaPlaybackSource)(unsafe.Pointer(mediaSource))))

			// 计时器
			p.timer = NewTimer(Options{
				Duration:       8760 * time.Hour,
				TickerInternal: configs.AppConfig.Main.FrameRate.Interval(),
				OnRun:          func(started bool) {},
				OnPause:        func() {},
				OnDone:         func(stopped bool) {},
				OnTick: func() {
					var curTime time.Duration
					session := Must1(p.player.GetPlaybackSession())
					t := Must1(session.GetPosition())
					if t.Duration <= 0 {
						return
					}
					curTime = time.Duration(t.Duration/TicksPerMillisecond) * time.Millisecond
					select {
					case p.timeChan <- curTime:
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

func (p *winMediaPlayer) Play(music URLMusic) {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case p.musicChan <- music:
	case <-timer.C:
	}
}

func (p *winMediaPlayer) CurMusic() URLMusic {
	return p.curMusic
}

func (p *winMediaPlayer) Pause() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state != types.Playing {
		return
	}
	Must(p.player.Pause())
	p.timer.Pause()
}

func (p *winMediaPlayer) Resume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state == types.Playing {
		return
	}
	go p.timer.Run()
	Must(p.player.Play())
}

func (p *winMediaPlayer) Stop() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state == types.Stopped {
		return
	}
	Must(p.player.Pause())
	p.timer.Pause()
}

func (p *winMediaPlayer) Toggle() {
	switch p.State() {
	case types.Paused, types.Stopped:
		p.Resume()
	case types.Playing:
		p.Pause()
	default:
		p.Resume()
	}
}

func (p *winMediaPlayer) Seek(duration time.Duration) {
	p.l.Lock()
	defer p.l.Unlock()
	session := Must1(p.player.GetPlaybackSession())
	defer session.Release()
	Must(session.SetPosition(foundation.TimeSpan{Duration: duration.Milliseconds() * TicksPerMillisecond}))
	if p.timer != nil {
		p.timer.SetPassed(duration)
	}
}

func (p *winMediaPlayer) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	var curTime time.Duration
	session := Must1(p.player.GetPlaybackSession())
	defer session.Release()
	t := Must1(session.GetPosition())
	if t.Duration <= 0 {
		return curTime
	}
	curTime = time.Duration(t.Duration/TicksPerMillisecond) * time.Millisecond
	return curTime
}

func (p *winMediaPlayer) PlayedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.ActualRuntime()
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
		Must(p.player.SetVolume(float64(p.volume) / 100.0))
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
		Must(p.player.SetVolume(float64(p.volume) / 100.0))
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
		Must(p.player.SetVolume(float64(p.volume) / 100.0))
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
