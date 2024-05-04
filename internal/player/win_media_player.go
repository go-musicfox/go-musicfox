//go:build windows

package player

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils"
	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go"
	"github.com/saltosystems/winrt-go/windows/foundation"
	"github.com/saltosystems/winrt-go/windows/media"
	"github.com/saltosystems/winrt-go/windows/media/core"
	"github.com/saltosystems/winrt-go/windows/media/playback"
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

func init() {
	utils.Must(ole.RoInitialize(1))
}

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

	cleaner func()
	close   chan struct{}
}

func NewWinMediaPlayer() *winMediaPlayer {
	p := &winMediaPlayer{
		state:     types.Stopped,
		timeChan:  make(chan time.Duration),
		stateChan: make(chan types.State),
		musicChan: make(chan UrlMusic),
		close:     make(chan struct{}),
		volume:    100,
	}

	p.player = utils.Must1(playback.NewMediaPlayer())
	utils.Must(p.player.SetVolume(float64(p.volume / 100.0)))
	utils.Must(p.player.SetAudioCategory(playback.MediaPlayerAudioCategoryMedia))
	//playbackSession := utils.Must1(p.player.GetPlaybackSession())

	smtc := utils.Must1(p.player.GetSystemMediaTransportControls())
	utils.Must(smtc.SetIsEnabled(true))
	utils.Must(smtc.SetIsPauseEnabled(true))
	utils.Must(smtc.SetIsPlayEnabled(true))

	eventReceivedGuid3 := winrt.ParameterizedInstanceGUID(
		foundation.GUIDTypedEventHandler,
		media.SignatureSystemMediaTransportControls,
		media.SignatureSystemMediaTransportControlsButtonPressedEventArgs,
	)
	h4 := foundation.NewTypedEventHandler(
		ole.NewGUID(eventReceivedGuid3), func(_ *foundation.TypedEventHandler, sender, args unsafe.Pointer) {
			ctrl := (*media.SystemMediaTransportControls)(sender)
			arg := (*media.SystemMediaTransportControlsButtonPressedEventArgs)(args)
			fmt.Println("button pressed", ctrl, arg)
		})
	_ = utils.Must1(smtc.AddButtonPressed(h4))

	// state changed
	//stateHandler := foundation.NewTypedEventHandler(
	//	ole.NewGUID(playbackSessionEventGUID),
	//	func(_ *foundation.TypedEventHandler, sender, _ unsafe.Pointer) {
	//		session := (*playback.MediaPlaybackSession)(sender)
	//		switch utils.Must1(session.GetPlaybackState()) {
	//		case playback.MediaPlaybackStatePlaying:
	//			p.Resume()
	//			p.setState(types.Playing)
	//		case playback.MediaPlaybackStatePaused:
	//			p.Paused()
	//			p.setState(types.Paused)
	//		}
	//	},
	//)
	//stateChangedToken := utils.Must1(playbackSession.AddPlaybackStateChanged(stateHandler))
	//p.appendCleaner(func() {
	//	session := utils.Must1(p.player.GetPlaybackSession())
	//	utils.Must(session.RemovePlaybackStateChanged(stateChangedToken))
	//})

	// // current state changed(old version)
	//curStateHandler := foundation.NewTypedEventHandler(
	//	ole.NewGUID(playerEventGUID),
	//	func(_ *foundation.TypedEventHandler, sender, _ unsafe.Pointer) {
	//		player := (*playback.MediaPlayer)(sender)
	//		switch utils.Must1(player.GetCurrentState()) {
	//		case playback.MediaPlayerStatePlaying:
	//			p.Resume()
	//			p.setState(types.Playing)
	//		case playback.MediaPlayerStatePaused:
	//			p.Paused()
	//			p.setState(types.Paused)
	//		case playback.MediaPlayerStateStopped:
	//			p.Stop()
	//			p.setState(types.Stopped)
	//		}
	//	},
	//)
	//curStateChangedToken := utils.Must1(p.player.AddCurrentStateChanged(curStateHandler))
	//p.appendCleaner(func() {
	//	utils.Must(p.player.RemoveCurrentStateChanged(curStateChangedToken))
	//})

	// finished
	//finishedHandler := foundation.NewTypedEventHandler(
	//	ole.NewGUID(playerEventGUID),
	//	func(_ *foundation.TypedEventHandler, _, _ unsafe.Pointer) {
	//		p.Stop()
	//		p.setState(types.Stopped)
	//	},
	//)
	//failedHandler := foundation.NewTypedEventHandler(
	//	ole.NewGUID(playerFailedEventGUID),
	//	func(_ *foundation.TypedEventHandler, _, _ unsafe.Pointer) {
	//		p.Stop()
	//		p.setState(types.Stopped)
	//	},
	//)
	//finishedToken := utils.Must1(p.player.AddMediaEnded(finishedHandler))
	//p.appendCleaner(func() { utils.Must(p.player.RemoveMediaEnded(finishedToken)) })
	//failedToken := utils.Must1(p.player.AddMediaFailed(failedHandler))
	//p.appendCleaner(func() { utils.Must(p.player.RemoveMediaFailed(failedToken)) })

	utils.WaitGoStart(p.listen)

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
					session := utils.Must1(p.player.GetPlaybackSession())
					t := utils.Must1(session.GetPosition())
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

func (p *winMediaPlayer) Play(music UrlMusic) {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case p.musicChan <- music:
	case <-timer.C:
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
}

func (p *winMediaPlayer) Resume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state == types.Playing {
		return
	}
	go p.timer.Run()
	utils.Must(p.player.Play())
}

func (p *winMediaPlayer) Stop() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.state == types.Stopped {
		return
	}
	utils.Must(p.player.Pause())
	p.timer.Pause()
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
	session := utils.Must1(p.player.GetPlaybackSession())
	utils.Must(session.SetPosition(foundation.TimeSpan{Duration: duration.Milliseconds() * TicksPerMillisecond}))
	if p.timer != nil {
		p.timer.SetPassed(duration)
	}
}

func (p *winMediaPlayer) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	var curTime time.Duration
	session := utils.Must1(p.player.GetPlaybackSession())
	t := utils.Must1(session.GetPosition())
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
	if p.cleaner != nil {
		p.cleaner()
	}
	if p.player != nil {
		p.player.Release()
	}
}

func (p *winMediaPlayer) appendCleaner(cleaner func()) {
	pre := p.cleaner
	p.cleaner = func() {
		if pre != nil {
			pre()
		}
		cleaner()
	}
}
