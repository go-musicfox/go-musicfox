//go:build darwin
// +build darwin

package state

import (
	"fmt"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/mediaplayer"
	"github.com/progrium/macdriver/objc"
)

var stateMap = map[uint8]core.NSUInteger{
	Unknown:     mediaplayer.MPNowPlayingPlaybackStateUnknown,
	Playing:     mediaplayer.MPNowPlayingPlaybackStatePlaying,
	Paused:      mediaplayer.MPNowPlayingPlaybackStatePaused,
	Stopped:     mediaplayer.MPNowPlayingPlaybackStateStopped,
	Interrupted: mediaplayer.MPNowPlayingPlaybackStateInterrupted,
}

type Handler struct {
	nowPlayingCenter    *mediaplayer.MPNowPlayingInfoCenter
	remoteCommandCenter *mediaplayer.MPRemoteCommandCenter
	commandHandler      *remoteCommandHandler
}

func NewHandler(player Player) *Handler {
	playingCenter := mediaplayer.MPNowPlayingInfoCenter_defaultCenter()
	commandCenter := mediaplayer.MPRemoteCommandCenter_sharedCommandCenter()
	commandHandler := &remoteCommandHandler{
		player: player,
	}

	handler := &Handler{
		nowPlayingCenter:    &playingCenter,
		remoteCommandCenter: &commandCenter,
		commandHandler:      commandHandler,
	}
	handler.registerCommands()
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
	s.remoteCommandCenter.ChangeRepeatModeCommand().AddTarget_action_(h, objc.Sel("handleChangeRepeatModeCommand:"))
	s.remoteCommandCenter.ChangeShuffleModeCommand().AddTarget_action_(h, objc.Sel("handleChangeShuffleModeCommand:"))
	s.remoteCommandCenter.ChangePlaybackRateCommand().AddTarget_action_(h, objc.Sel("handleChangePlaybackRateCommand:"))
	s.remoteCommandCenter.SeekBackwardCommand().AddTarget_action_(h, objc.Sel("handleSeekBackwardCommand:"))
	s.remoteCommandCenter.SeekForwardCommand().AddTarget_action_(h, objc.Sel("handleSeekForwardCommand:"))
	s.remoteCommandCenter.SkipForwardCommand().AddTarget_action_(h, objc.Sel("handleSkipForwardCommand:"))
	s.remoteCommandCenter.SkipBackwardCommand().AddTarget_action_(h, objc.Sel("handleSkipBackwardCommand:"))
	s.remoteCommandCenter.ChangePlaybackPositionCommand().AddTarget_action_(h, objc.Sel("handleChangePlaybackPositionCommand:"))
	s.remoteCommandCenter.LikeCommand().AddTarget_action_(h, objc.Sel("handleLikeCommand:"))
	s.remoteCommandCenter.DislikeCommand().AddTarget_action_(h, objc.Sel("handleDisLikeCommand:"))
	s.remoteCommandCenter.BookmarkCommand().AddTarget_action_(h, objc.Sel("handleBookmarkCommand:"))
	s.remoteCommandCenter.EnableLanguageOptionCommand().AddTarget_action_(h, objc.Sel("handleEnableLanguageOptionCommand:"))
	s.remoteCommandCenter.DisableLanguageOptionCommand().AddTarget_action_(h, objc.Sel("handleDisableLanguageOptionCommand:"))
}

func (s *Handler) SetPlaybackState(state uint8) {
	s.nowPlayingCenter.SetPlaybackState_(stateMap[state])
}

func (s *Handler) SetPlayingInfo(info PlayingInfo) {
	total := info.TotalDuration.Seconds()
	ur := info.PassedDuration.Seconds()
	rate := info.Rate

	values, keys := core.NSArray_array(), core.NSArray_array()
	values = values.ArrayByAddingObject_(core.NSNumber_numberWithInt_(int32(total)))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPMediaItemPropertyPlaybackDuration))

	values = values.ArrayByAddingObject_(core.NSNumber_numberWithInt_(int32(ur)))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPNowPlayingInfoPropertyElapsedPlaybackTime))

	values = values.ArrayByAddingObject_(core.NSNumber_numberWithFloat_(rate))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPNowPlayingInfoPropertyPlaybackRate))

	values = values.ArrayByAddingObject_(core.NSNumber_numberWithFloat_(1.0))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPNowPlayingInfoPropertyDefaultPlaybackRate))
	dict := core.NSDictionary_dictionaryWithObjects_forKeys_(values, keys)

	s.nowPlayingCenter.SetNowPlayingInfo_(dict)
}

func (s *Handler) Release() {
	s.nowPlayingCenter.Release()
	s.remoteCommandCenter.Release()
}

type remoteCommandHandler struct {
	player Player
}

func (r *remoteCommandHandler) handlePlayCommand(_ objc.Object) core.NSInteger {
	r.player.Resume()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handlePauseCommand(_ objc.Object) core.NSInteger {
	r.player.Paused()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleStopCommand(_ objc.Object) core.NSInteger {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleTogglePlayPauseCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleNextTrackCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handlePreviousTrackCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleChangeRepeatModeCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleChangeShuffleModeCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleChangePlaybackRateCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleSeekBackwardCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleSeekForwardCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleSkipForwardCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleSkipBackwardCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleChangePlaybackPositionCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleLikeCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleDisLikeCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleBookmarkCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleEnableLanguageOptionCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleDisableLanguageOptionCommand(event objc.Object) core.NSInteger {
	fmt.Printf("playing: %#v\n", event)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}
