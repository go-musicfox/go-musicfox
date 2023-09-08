//go:build darwin

package state_handler

import (
	"sync"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/mediaplayer"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils"
)

var stateMap = map[types.State]mediaplayer.MPNowPlayingPlaybackState{
	types.Unknown:     mediaplayer.MPNowPlayingPlaybackStateUnknown,
	types.Playing:     mediaplayer.MPNowPlayingPlaybackStatePlaying,
	types.Paused:      mediaplayer.MPNowPlayingPlaybackStatePaused,
	types.Stopped:     mediaplayer.MPNowPlayingPlaybackStateStopped,
	types.Interrupted: mediaplayer.MPNowPlayingPlaybackStateInterrupted,
}

type Handler struct {
	nowPlayingCenter    *mediaplayer.MPNowPlayingInfoCenter
	remoteCommandCenter *mediaplayer.MPRemoteCommandCenter
	commandHandler      *remoteCommandHandler
	curArtwork          mediaplayer.MPMediaItemArtwork
	curArtworkUrl       string
	l                   sync.Mutex
}

const (
	MediaTypeNone = iota
	MediaTypeAudio
	MediaTypeVedio
)

func NewHandler(p Controller, _ PlayingInfo) *Handler {
	playingCenter := mediaplayer.MPNowPlayingInfoCenter_defaultCenter()
	commandCenter := mediaplayer.MPRemoteCommandCenter_sharedCommandCenter()
	commandHandler := remoteCommandHandler_new(p)

	handler := &Handler{
		nowPlayingCenter:    &playingCenter,
		remoteCommandCenter: &commandCenter,
		commandHandler:      &commandHandler,
	}
	handler.registerCommands()
	handler.nowPlayingCenter.SetPlaybackState(mediaplayer.MPNowPlayingPlaybackStateStopped)
	return handler
}

func (s *Handler) registerCommands() {
	number := core.NSNumber_numberWithDouble(15.0)
	defer number.Release()
	intervals := core.NSArray_arrayWithObject(number.NSObject)
	defer intervals.Release()

	s.remoteCommandCenter.SkipBackwardCommand().SetPreferredIntervals(intervals)
	s.remoteCommandCenter.SkipForwardCommand().SetPreferredIntervals(intervals)

	s.remoteCommandCenter.PlayCommand().AddTargetAction(s.commandHandler.ID, sel_handlePlayCommand)
	s.remoteCommandCenter.PauseCommand().AddTargetAction(s.commandHandler.ID, sel_handlePauseCommand)
	s.remoteCommandCenter.StopCommand().AddTargetAction(s.commandHandler.ID, sel_handleStopCommand)
	s.remoteCommandCenter.TogglePlayPauseCommand().AddTargetAction(s.commandHandler.ID, sel_handleTogglePlayPauseCommand)
	s.remoteCommandCenter.NextTrackCommand().AddTargetAction(s.commandHandler.ID, sel_handleNextTrackCommand)
	s.remoteCommandCenter.PreviousTrackCommand().AddTargetAction(s.commandHandler.ID, sel_handlePreviousTrackCommand)
	//s.remoteCommandCenter.ChangeRepeatModeCommand().AddTargetAction(s.commandHandler.ID, sel_handleChangeRepeatModeCommand)
	//s.remoteCommandCenter.ChangeShuffleModeCommand().AddTargetAction(s.commandHandler.ID, sel_handleChangeShuffleModeCommand)
	//s.remoteCommandCenter.ChangePlaybackRateCommand().AddTargetAction(s.commandHandler.ID, sel_handleChangePlaybackRateCommand)
	//s.remoteCommandCenter.SeekBackwardCommand().AddTargetAction(s.commandHandler.ID, sel_handleSeekBackwardCommand)
	//s.remoteCommandCenter.SeekForwardCommand().AddTargetAction(s.commandHandler.ID, sel_handleSeekForwardCommand)
	//s.remoteCommandCenter.SkipForwardCommand().AddTargetAction(s.commandHandler.ID, sel_handleSkipForwardCommand)
	//s.remoteCommandCenter.SkipBackwardCommand().AddTargetAction(s.commandHandler.ID, sel_handleSkipBackwardCommand)
	s.remoteCommandCenter.ChangePlaybackPositionCommand().AddTargetAction(s.commandHandler.ID, sel_handleChangePlaybackPositionCommand)
	//s.remoteCommandCenter.LikeCommand().AddTargetAction(s.commandHandler.ID, sel_handleLikeCommand)
	//s.remoteCommandCenter.DislikeCommand().AddTargetAction(s.commandHandler.ID, sel_handleDislikeCommand)
	//s.remoteCommandCenter.BookmarkCommand().AddTargetAction(s.commandHandler.ID, sel_handleBookmarkCommand)
	//s.remoteCommandCenter.EnableLanguageOptionCommand().AddTargetAction(s.commandHandler.ID, sel_handleEnableLanguageOptionCommand)
	//s.remoteCommandCenter.DisableLanguageOptionCommand().AddTargetAction(s.commandHandler.ID, sel_handleDisableLanguageOptionCommand)

	workspaceNC := cocoa.NSWorkspace_sharedWorkspace().NotificationCenter()
	workspaceNC.AddObserverSelectorNameObject(s.commandHandler.ID, sel_handleWillSleepOrPowerOff, core.String("NSWorkspaceWillSleepNotification"), core.NSObject{})
	workspaceNC.AddObserverSelectorNameObject(s.commandHandler.ID, sel_handleWillSleepOrPowerOff, core.String("NSWorkspaceWillPowerOffNotification"), core.NSObject{})
	workspaceNC.AddObserverSelectorNameObject(s.commandHandler.ID, sel_handleDidWake, core.String("NSWorkspaceDidWakeNotification"), core.NSObject{})
}

func (s *Handler) SetPlayingInfo(info PlayingInfo) {
	total := info.TotalDuration.Seconds()
	ur := info.PassedDuration.Seconds()

	dic := core.NSMutableDictionary_init()
	defer dic.Release()

	setKV := func(k core.NSString, v core.NSObject) {
		dic.SetValueForKey(k, v)
		k.Release()
		v.Release()
	}

	setKV(core.String(mediaplayer.MPMediaItemPropertyPlaybackDuration), core.NSNumber_numberWithInt(int32(total)).NSObject)
	setKV(core.String(mediaplayer.MPMediaItemPropertyPersistentID), core.NSNumber_numberWithInt(int32(info.TrackID)).NSObject)
	setKV(core.String(mediaplayer.MPNowPlayingInfoPropertyElapsedPlaybackTime), core.NSNumber_numberWithInt(int32(ur)).NSObject)
	setKV(core.String(mediaplayer.MPNowPlayingInfoPropertyDefaultPlaybackRate), core.NSNumber_numberWithDouble(1.0).NSObject)
	setKV(core.String(mediaplayer.MPNowPlayingInfoPropertyPlaybackProgress), core.NSNumber_numberWithDouble(ur/total).NSObject)
	setKV(core.String(mediaplayer.MPNowPlayingInfoPropertyMediaType), core.NSNumber_numberWithInt(MediaTypeAudio).NSObject)
	setKV(core.String(mediaplayer.MPMediaItemPropertyMediaType), core.NSNumber_numberWithInt(int32(mediaplayer.MPMediaTypeMusic)).NSObject)
	setKV(core.String(mediaplayer.MPMediaItemPropertyTitle), core.String(info.Name).NSObject)
	setKV(core.String(mediaplayer.MPMediaItemPropertyAlbumTitle), core.String(info.Album).NSObject)
	setKV(core.String(mediaplayer.MPMediaItemPropertyArtist), core.String(info.Artist).NSObject)
	setKV(core.String(mediaplayer.MPMediaItemPropertyAlbumArtist), core.String(info.AlbumArtist).NSObject)

	if info.PicUrl != "" {
		picUrl := utils.AddResizeParamForPicUrl(info.PicUrl, 60)
		s.l.Lock()
		if s.curArtworkUrl != picUrl {
			var lastArtwork = s.curArtwork
			defer lastArtwork.Release()

			s.curArtworkUrl = picUrl
			url := core.NSURL_URLWithString(core.String(picUrl))
			defer url.Release()
			image := cocoa.NSImage_alloc()
			if i := image.InitWithContentsOfURL(url); i.ID > 0 {
				image = i
				s.curArtwork = mediaplayer.MPMediaItemArtwork_alloc().InitWithImage(image)
			}
			defer image.Release()
		}
		s.l.Unlock()
		if s.curArtwork.ID > 0 {
			dic.SetValueForKey(core.String(mediaplayer.MPMediaItemPropertyArtwork), s.curArtwork.NSObject)
		}
	}

	s.nowPlayingCenter.SetPlaybackState(stateMap[info.State])
	s.nowPlayingCenter.SetNowPlayingInfo(dic.NSDictionary)
}

func (s *Handler) Release() {
	s.nowPlayingCenter.Release()
	s.remoteCommandCenter.Release()
}
