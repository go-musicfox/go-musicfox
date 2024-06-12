//go:build windows

package remote_control

import (
	"sync"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go"
	"github.com/saltosystems/winrt-go/windows/foundation"
	"github.com/saltosystems/winrt-go/windows/media"
	"github.com/saltosystems/winrt-go/windows/media/playback"
	"github.com/saltosystems/winrt-go/windows/storage/streams"

	"github.com/go-musicfox/go-musicfox/internal/types"
	. "github.com/go-musicfox/go-musicfox/utils/errorx"
)

const (
	TicksPerMicrosecond int64 = 10
	TicksPerMillisecond       = TicksPerMicrosecond * 1000
	TicksPerSecond            = TicksPerMillisecond * 1000
)

var (
	SMTC     *media.SystemMediaTransportControls
	__player *playback.MediaPlayer // Note: Do not use it!!! useless player, just for get smtc

	buttonPressedEventGUID = winrt.ParameterizedInstanceGUID(
		foundation.GUIDTypedEventHandler,
		media.SignatureSystemMediaTransportControls,
		media.SignatureSystemMediaTransportControlsButtonPressedEventArgs,
	)

	stateMap = map[types.State]media.MediaPlaybackStatus{
		types.Unknown:     media.MediaPlaybackStatusClosed,
		types.Playing:     media.MediaPlaybackStatusPlaying,
		types.Paused:      media.MediaPlaybackStatusPaused,
		types.Stopped:     media.MediaPlaybackStatusStopped,
		types.Interrupted: media.MediaPlaybackStatusClosed,
	}
)

type RemoteControl struct {
	p    Controller
	smtc *media.SystemMediaTransportControls

	lastPlayingInfo  PlayingInfo
	lastPlayingInfoL sync.Mutex
}

func NewRemoteControl(p Controller, _ PlayingInfo) *RemoteControl {
	_ = ole.RoInitialize(1)

	if SMTC == nil {
		__player = Must1(playback.NewMediaPlayer())
		SMTC = Must1(__player.GetSystemMediaTransportControls())
	}

	c := &RemoteControl{
		p:    p,
		smtc: SMTC,
	}

	c.registerEventHandlers()

	return c
}

func (c *RemoteControl) registerEventHandlers() {
	if c.smtc == nil {
		return
	}

	Must(c.smtc.SetIsEnabled(true))
	Must(c.smtc.SetIsPauseEnabled(true))
	Must(c.smtc.SetIsPlayEnabled(true))
	Must(c.smtc.SetIsNextEnabled(true))
	Must(c.smtc.SetIsPreviousEnabled(true))

	pressedHandler := foundation.NewTypedEventHandler(
		ole.NewGUID(buttonPressedEventGUID),
		func(_ *foundation.TypedEventHandler, _ unsafe.Pointer, args unsafe.Pointer) {
			eventArgs := (*media.SystemMediaTransportControlsButtonPressedEventArgs)(args)
			defer eventArgs.Release()
			switch Must1(eventArgs.GetButton()) {
			case media.SystemMediaTransportControlsButtonPlay:
				c.p.CtrlResume()
			case media.SystemMediaTransportControlsButtonPause:
				c.p.CtrlPaused()
			case media.SystemMediaTransportControlsButtonNext:
				c.p.CtrlNext()
			case media.SystemMediaTransportControlsButtonPrevious:
				c.p.CtrlPrevious()
			}
		},
	)
	defer pressedHandler.Release()
	Must1(c.smtc.AddButtonPressed(pressedHandler))
}

func (c *RemoteControl) SetPosition(pos time.Duration) {
	if c.smtc == nil {
		return
	}
	c.lastPlayingInfoL.Lock()
	c.lastPlayingInfo.PassedDuration = pos
	c.lastPlayingInfoL.Unlock()
	c.setTimelineProps(c.lastPlayingInfo)
}

func (c *RemoteControl) SetPlayingInfo(info PlayingInfo) {
	if c.smtc == nil {
		return
	}
	Must(c.smtc.SetPlaybackStatus(stateMap[info.State]))

	updater := Must1(c.smtc.GetDisplayUpdater())
	defer updater.Release()
	if info.PicUrl != "" {
		imgUri := Must1(foundation.UriCreateUri(info.PicUrl))
		defer imgUri.Release()
		stream := Must1(streams.RandomAccessStreamReferenceCreateFromUri(imgUri))
		defer stream.Release()
		Must(updater.SetThumbnail(stream))
	}
	Must(updater.SetType(media.MediaPlaybackTypeMusic))

	musicProps := Must1(updater.GetMusicProperties())
	defer musicProps.Release()
	Must(musicProps.SetTitle(info.Name))
	Must(musicProps.SetArtist(info.Artist))
	Must(musicProps.SetAlbumTitle(info.Album))
	Must(musicProps.SetAlbumArtist(info.AlbumArtist))
	Must(updater.Update())

	c.lastPlayingInfoL.Lock()
	c.lastPlayingInfo = info
	c.lastPlayingInfoL.Unlock()
	c.setTimelineProps(info)
}

func (c *RemoteControl) Release() {
	if c.smtc != nil {
		c.smtc.Release()
	}
}

func (c *RemoteControl) setTimelineProps(info PlayingInfo) {
	timelineProps := Must1(media.NewSystemMediaTransportControlsTimelineProperties())
	defer timelineProps.Release()
	Must(timelineProps.SetStartTime(foundation.TimeSpan{}))
	Must(timelineProps.SetEndTime(foundation.TimeSpan{Duration: info.TotalDuration.Milliseconds() * TicksPerMillisecond}))
	Must(timelineProps.SetMinSeekTime(foundation.TimeSpan{}))
	Must(timelineProps.SetMaxSeekTime(foundation.TimeSpan{Duration: info.TotalDuration.Milliseconds() * TicksPerMillisecond}))
	Must(timelineProps.SetPosition(foundation.TimeSpan{Duration: info.PassedDuration.Milliseconds() * TicksPerMillisecond}))
	Must(c.smtc.UpdateTimelineProperties(timelineProps))
}
