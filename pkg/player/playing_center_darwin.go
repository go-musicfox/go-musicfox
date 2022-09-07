//go:build darwin
// +build darwin

package player

import (
	"fmt"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/mediaplayer"
	"github.com/progrium/macdriver/objc"
)

type centerPackage struct {
	nowPlayingCenter    *mediaplayer.MPNowPlayingInfoCenter
	remoteCommandCenter *mediaplayer.MPRemoteCommandCenter
}

func init() {
	playingCenter := mediaplayer.MPNowPlayingInfoCenter_defaultCenter()
	commandCenter := mediaplayer.MPRemoteCommandCenter_sharedCommandCenter()

	centers = &centerPackage{
		nowPlayingCenter: &playingCenter,
		remoteCommandCenter: &commandCenter,
	}
}

func registerCommands(player *Player) {
	centers.remoteCommandCenter.SkipBackwardCommand().SetPreferredIntervals_(core.NSArray_arrayWithObject_(core.NSNumber_numberWithFloat_(15.0)))
	centers.remoteCommandCenter.SkipForwardCommand().SetPreferredIntervals_(core.NSArray_arrayWithObject_(core.NSNumber_numberWithFloat_(15.0)))

	handler := NewRemoteCommandHandler(player)
	cls := objc.NewClass("RemoteCommandHandler", "NSObject")
	cls.AddMethod("handlePlayCommand:", handler.handlePlayCommand)
	cls.AddMethod("handlePauseCommand:", handler.handlePauseCommand)
	cls.AddMethod("handleStopCommand:", handler.handleStopCommand)
	cls.AddMethod("handleTogglePlayPauseCommand:", handler.handleTogglePlayPauseCommand)
	cls.AddMethod("handleNextTrackCommand:", handler.handleNextTrackCommand)
	cls.AddMethod("handlePreviousTrackCommand:", handler.handlePreviousTrackCommand)
	cls.AddMethod("handleChangeRepeatModeCommand:", handler.handleChangeRepeatModeCommand)
	cls.AddMethod("handleChangeShuffleModeCommand:", handler.handleChangeShuffleModeCommand)
	cls.AddMethod("handleChangePlaybackRateCommand:", handler.handleChangePlaybackRateCommand)
	cls.AddMethod("handleSeekBackwardCommand:", handler.handleSeekBackwardCommand)
	cls.AddMethod("handleSeekForwardCommand:", handler.handleSeekForwardCommand)
	cls.AddMethod("handleSkipForwardCommand:", handler.handleSkipForwardCommand)
	cls.AddMethod("handleSkipBackwardCommand:", handler.handleSkipBackwardCommand)
	cls.AddMethod("handleChangePlaybackPositionCommand:", handler.handleChangePlaybackPositionCommand)
	cls.AddMethod("handleLikeCommand:", handler.handleLikeCommand)
	cls.AddMethod("handleDisLikeCommand:", handler.handleDisLikeCommand)
	cls.AddMethod("handleBookmarkCommand:", handler.handleBookmarkCommand)
	cls.AddMethod("handleEnableLanguageOptionCommand:", handler.handleEnableLanguageOptionCommand)
	cls.AddMethod("handleDisableLanguageOptionCommand:", handler.handleDisableLanguageOptionCommand)

	objc.RegisterClass(cls)
	h := objc.Get("RemoteCommandHandler").Alloc().Init()
	centers.remoteCommandCenter.PlayCommand().AddTarget_action_(h, objc.Sel("handlePlayCommand:"))
	centers.remoteCommandCenter.PauseCommand().AddTarget_action_(h, objc.Sel("handlePauseCommand:"))
	centers.remoteCommandCenter.StopCommand().AddTarget_action_(h, objc.Sel("handleStopCommand:"))
	centers.remoteCommandCenter.TogglePlayPauseCommand().AddTarget_action_(h, objc.Sel("handleTogglePlayPauseCommand:"))
	centers.remoteCommandCenter.NextTrackCommand().AddTarget_action_(h, objc.Sel("handleNextTrackCommand:"))
	centers.remoteCommandCenter.PreviousTrackCommand().AddTarget_action_(h, objc.Sel("handlePreviousTrackCommand:"))
	centers.remoteCommandCenter.ChangeRepeatModeCommand().AddTarget_action_(h, objc.Sel("handleChangeRepeatModeCommand:"))
	centers.remoteCommandCenter.ChangeShuffleModeCommand().AddTarget_action_(h, objc.Sel("handleChangeShuffleModeCommand:"))
	centers.remoteCommandCenter.ChangePlaybackRateCommand().AddTarget_action_(h, objc.Sel("handleChangePlaybackRateCommand:"))
	centers.remoteCommandCenter.SeekBackwardCommand().AddTarget_action_(h, objc.Sel("handleSeekBackwardCommand:"))
	centers.remoteCommandCenter.SeekForwardCommand().AddTarget_action_(h, objc.Sel("handleSeekForwardCommand:"))
	centers.remoteCommandCenter.SkipForwardCommand().AddTarget_action_(h, objc.Sel("handleSkipForwardCommand:"))
	centers.remoteCommandCenter.SkipBackwardCommand().AddTarget_action_(h, objc.Sel("handleSkipBackwardCommand:"))
	centers.remoteCommandCenter.ChangePlaybackPositionCommand().AddTarget_action_(h, objc.Sel("handleChangePlaybackPositionCommand:"))
	centers.remoteCommandCenter.LikeCommand().AddTarget_action_(h, objc.Sel("handleLikeCommand:"))
	centers.remoteCommandCenter.DislikeCommand().AddTarget_action_(h, objc.Sel("handleDisLikeCommand:"))
	centers.remoteCommandCenter.BookmarkCommand().AddTarget_action_(h, objc.Sel("handleBookmarkCommand:"))
	centers.remoteCommandCenter.EnableLanguageOptionCommand().AddTarget_action_(h, objc.Sel("handleEnableLanguageOptionCommand:"))
	centers.remoteCommandCenter.DisableLanguageOptionCommand().AddTarget_action_(h, objc.Sel("handleDisableLanguageOptionCommand:"))
}

type remoteCommandHandler struct {
	player *Player
}

func NewRemoteCommandHandler(player *Player) *remoteCommandHandler {
	return &remoteCommandHandler{
		player: player,
	}
}

func nowPlayingInfo(player *Player) core.NSDictionary {
	total := player.CurMusic.Duration.Seconds()
	ur := player.Timer.Passed().Seconds()

	values, keys := core.NSArray_array(), core.NSArray_array()
	values = values.ArrayByAddingObject_(core.NSNumber_numberWithInt_(int32(total)))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPMediaItemPropertyPlaybackDuration))

	values = values.ArrayByAddingObject_(core.NSNumber_numberWithInt_(int32(ur)))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPNowPlayingInfoPropertyElapsedPlaybackTime))

	values = values.ArrayByAddingObject_(core.NSNumber_numberWithFloat_(1.0))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPNowPlayingInfoPropertyPlaybackRate))

	values = values.ArrayByAddingObject_(core.NSNumber_numberWithFloat_(1.0))
	keys = keys.ArrayByAddingObject_(core.String(mediaplayer.MPNowPlayingInfoPropertyDefaultPlaybackRate))
	return core.NSDictionary_dictionaryWithObjects_forKeys_(values, keys)
}

func (r *remoteCommandHandler) handlePlayCommand(event objc.Object) core.NSInteger {
	//nowPlayingCenter.SetNowPlayingInfo_(nowPlayingInfo(r.player))
	//nowPlayingCenter.SetPlaybackState_(mediaplayer.MPNowPlayingPlaybackStatePlaying)
	r.player.Resume()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handlePauseCommand(event objc.Object) core.NSInteger {
	r.player.Paused()
	//nowPlayingCenter.SetPlaybackState_(mediaplayer.MPNowPlayingPlaybackStatePaused)
	//nowPlayingCenter.SetNowPlayingInfo_(nowPlayingInfo(r.player))
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (r *remoteCommandHandler) handleStopCommand(event objc.Object) core.NSInteger {
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
