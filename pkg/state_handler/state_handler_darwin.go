//go:build darwin
// +build darwin

package state_handler

import (
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/mediaplayer"
	"github.com/progrium/macdriver/objc"
	"go-musicfox/pkg/player"
	"time"
)

var stateMap = map[player.State]core.NSUInteger{
	player.Unknown:     mediaplayer.MPNowPlayingPlaybackStateUnknown,
	player.Playing:     mediaplayer.MPNowPlayingPlaybackStatePlaying,
	player.Paused:      mediaplayer.MPNowPlayingPlaybackStatePaused,
	player.Stopped:     mediaplayer.MPNowPlayingPlaybackStateStopped,
	player.Interrupted: mediaplayer.MPNowPlayingPlaybackStateInterrupted,
}

type Handler struct {
	nowPlayingCenter    *mediaplayer.MPNowPlayingInfoCenter
	remoteCommandCenter *mediaplayer.MPRemoteCommandCenter
	commandHandler      *remoteCommandHandler
}

const (
	MediaTypeNone = iota
	MediaTypeAudio
	MediaTypeVedio
)

func NewHandler(p Controller) *Handler {
	playingCenter := mediaplayer.MPNowPlayingInfoCenter_defaultCenter()
	commandCenter := mediaplayer.MPRemoteCommandCenter_sharedCommandCenter()
	commandHandler := &remoteCommandHandler{
		player: p,
	}

	handler := &Handler{
		nowPlayingCenter:    &playingCenter,
		remoteCommandCenter: &commandCenter,
		commandHandler:      commandHandler,
	}
	handler.registerCommands()
	handler.nowPlayingCenter.SetPlaybackState_(mediaplayer.MPNowPlayingPlaybackStateStopped)
	return handler
}

func (s *Handler) registerCommands() {
	s.remoteCommandCenter.SkipBackwardCommand().SetPreferredIntervals_(core.NSArray_arrayWithObject_(core.NSNumber_numberWithFloat_(15.0)))
	s.remoteCommandCenter.SkipForwardCommand().SetPreferredIntervals_(core.NSArray_arrayWithObject_(core.NSNumber_numberWithFloat_(15.0)))

	cls := objc.NewClass("RemoteCommandHandler", "NSObject")
	cls.AddMethod("handlePlayCommand:", s.commandHandler.handlePlayCommand)
	cls.AddMethod("handlePauseCommand:", s.commandHandler.handlePauseCommand)
	cls.AddMethod("handleStopCommand:", s.commandHandler.handleStopCommand)
	cls.AddMethod("handleTogglePlayPauseCommand:", s.commandHandler.handleTogglePlayPauseCommand)
	cls.AddMethod("handleNextTrackCommand:", s.commandHandler.handleNextTrackCommand)
	cls.AddMethod("handlePreviousTrackCommand:", s.commandHandler.handlePreviousTrackCommand)
	cls.AddMethod("handleChangeRepeatModeCommand:", s.commandHandler.handleChangeRepeatModeCommand)
	cls.AddMethod("handleChangeShuffleModeCommand:", s.commandHandler.handleChangeShuffleModeCommand)
	cls.AddMethod("handleChangePlaybackRateCommand:", s.commandHandler.handleChangePlaybackRateCommand)
	cls.AddMethod("handleSeekBackwardCommand:", s.commandHandler.handleSeekBackwardCommand)
	cls.AddMethod("handleSeekForwardCommand:", s.commandHandler.handleSeekForwardCommand)
	cls.AddMethod("handleSkipForwardCommand:", s.commandHandler.handleSkipForwardCommand)
	cls.AddMethod("handleSkipBackwardCommand:", s.commandHandler.handleSkipBackwardCommand)
	cls.AddMethod("handleChangePlaybackPositionCommand:", s.commandHandler.handleChangePlaybackPositionCommand)
	cls.AddMethod("handleLikeCommand:", s.commandHandler.handleLikeCommand)
	cls.AddMethod("handleDisLikeCommand:", s.commandHandler.handleDisLikeCommand)
	cls.AddMethod("handleBookmarkCommand:", s.commandHandler.handleBookmarkCommand)
	cls.AddMethod("handleEnableLanguageOptionCommand:", s.commandHandler.handleEnableLanguageOptionCommand)
	cls.AddMethod("handleDisableLanguageOptionCommand:", s.commandHandler.handleDisableLanguageOptionCommand)

	objc.RegisterClass(cls)
	h := objc.Get("RemoteCommandHandler").Alloc().Init()
	s.remoteCommandCenter.PlayCommand().AddTarget_action_(h, objc.Sel("handlePlayCommand:"))
	s.remoteCommandCenter.PauseCommand().AddTarget_action_(h, objc.Sel("handlePauseCommand:"))
	s.remoteCommandCenter.StopCommand().AddTarget_action_(h, objc.Sel("handleStopCommand:"))
	s.remoteCommandCenter.TogglePlayPauseCommand().AddTarget_action_(h, objc.Sel("handleTogglePlayPauseCommand:"))
	s.remoteCommandCenter.NextTrackCommand().AddTarget_action_(h, objc.Sel("handleNextTrackCommand:"))
	s.remoteCommandCenter.PreviousTrackCommand().AddTarget_action_(h, objc.Sel("handlePreviousTrackCommand:"))
	s.remoteCommandCenter.ChangePlaybackPositionCommand().AddTarget_action_(h, objc.Sel("handleChangePlaybackPositionCommand:"))
	//s.remoteCommandCenter.ChangeRepeatModeCommand().AddTarget_action_(h, objc.Sel("handleChangeRepeatModeCommand:"))
	//s.remoteCommandCenter.ChangeShuffleModeCommand().AddTarget_action_(h, objc.Sel("handleChangeShuffleModeCommand:"))
	//s.remoteCommandCenter.ChangePlaybackRateCommand().AddTarget_action_(h, objc.Sel("handleChangePlaybackRateCommand:"))
	//s.remoteCommandCenter.SeekBackwardCommand().AddTarget_action_(h, objc.Sel("handleSeekBackwardCommand:"))
	//s.remoteCommandCenter.SeekForwardCommand().AddTarget_action_(h, objc.Sel("handleSeekForwardCommand:"))
	//s.remoteCommandCenter.SkipForwardCommand().AddTarget_action_(h, objc.Sel("handleSkipForwardCommand:"))
	//s.remoteCommandCenter.SkipBackwardCommand().AddTarget_action_(h, objc.Sel("handleSkipBackwardCommand:"))
	//s.remoteCommandCenter.LikeCommand().AddTarget_action_(h, objc.Sel("handleLikeCommand:"))
	//s.remoteCommandCenter.DislikeCommand().AddTarget_action_(h, objc.Sel("handleDisLikeCommand:"))
	//s.remoteCommandCenter.BookmarkCommand().AddTarget_action_(h, objc.Sel("handleBookmarkCommand:"))
	//s.remoteCommandCenter.EnableLanguageOptionCommand().AddTarget_action_(h, objc.Sel("handleEnableLanguageOptionCommand:"))
	//s.remoteCommandCenter.DisableLanguageOptionCommand().AddTarget_action_(h, objc.Sel("handleDisableLanguageOptionCommand:"))
}

func (s *Handler) SetPlayingInfo(info PlayingInfo) {
	total := info.TotalDuration.Seconds()
	ur := info.PassedDuration.Seconds()

	values, keys := core.NSArray_array(), core.NSArray_array()
	values = values.ArrayByAddingObject_(core.NSNumber_numberWithInt_(int32(total)))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPMediaItemPropertyPlaybackDuration))

	values = values.ArrayByAddingObject_(core.NSNumber_numberWithInt_(int32(ur)))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPNowPlayingInfoPropertyElapsedPlaybackTime))

	values = values.ArrayByAddingObject_(core.NSNumber_numberWithFloat_(1.0))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPNowPlayingInfoPropertyDefaultPlaybackRate))

	values = values.ArrayByAddingObject_(core.NSNumber_numberWithFloat_(float32(ur / total)))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPNowPlayingInfoPropertyPlaybackProgress))

	values = values.ArrayByAddingObject_(core.NSNumber_numberWithInt_(MediaTypeAudio))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPNowPlayingInfoPropertyMediaType))

	values = values.ArrayByAddingObject_(core.NSNumber_numberWithInt_(mediaplayer.MPMediaTypeMusic))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPMediaItemPropertyMediaType))

	values = values.ArrayByAddingObject_(core.String("测试"))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPMediaItemPropertyTitle))

	dict := core.NSDictionary_dictionaryWithObjects_forKeys_(values, keys)

	s.nowPlayingCenter.SetNowPlayingInfo_(dict)
	s.nowPlayingCenter.SetPlaybackState_(stateMap[info.State])
}

func (s *Handler) Release() {
	s.nowPlayingCenter.Release()
	s.remoteCommandCenter.Release()
}

type remoteCommandHandler struct {
	player Controller
}

func (r *remoteCommandHandler) handlePlayCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	r.player.Resume()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handlePauseCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	r.player.Paused()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleStopCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	r.player.Paused()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleTogglePlayPauseCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	r.player.Toggle()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleNextTrackCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	r.player.Next()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handlePreviousTrackCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	r.player.Previous()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleChangeRepeatModeCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleChangeShuffleModeCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleChangePlaybackRateCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleSeekBackwardCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleSeekForwardCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleSkipForwardCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleSkipBackwardCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleChangePlaybackPositionCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	event := mediaplayer.MPChangePlaybackPositionCommandEvent_fromRef(eventObj)
	r.player.Seek(time.Duration(event.PositionTime()) * time.Second)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleLikeCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleDisLikeCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleBookmarkCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleEnableLanguageOptionCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleDisableLanguageOptionCommand(self objc.Object, eventObj objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}
