// +build darwin

package player

import (
	"fmt"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/mediaplayer"
	"github.com/progrium/macdriver/objc"
)

func init() {
	//go func() {
	//
	//}()

	playingCenter := mediaplayer.MPNowPlayingInfoCenter_defaultCenter()
	commandCenter := mediaplayer.MPRemoteCommandCenter_sharedCommandCenter()

	nowPlayingCenter = &playingCenter
	remoteCommandCenter = &commandCenter
}

func registerCommands(player *Player) {
	remoteCommandCenter.SkipBackwardCommand().SetPreferredIntervals_(core.NSArray_WithObjects(core.NSNumber_numberWithFloat_(15.0)))
	remoteCommandCenter.SkipForwardCommand().SetPreferredIntervals_(core.NSArray_WithObjects(core.NSNumber_numberWithFloat_(15.0)))

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
	remoteCommandCenter.PlayCommand().AddTarget_action_(h, objc.Sel("handlePlayCommand:"))
	remoteCommandCenter.PauseCommand().AddTarget_action_(h, objc.Sel("handlePauseCommand:"))
	remoteCommandCenter.StopCommand().AddTarget_action_(h, objc.Sel("handleStopCommand:"))
	remoteCommandCenter.TogglePlayPauseCommand().AddTarget_action_(h, objc.Sel("handleTogglePlayPauseCommand:"))
	remoteCommandCenter.NextTrackCommand().AddTarget_action_(h, objc.Sel("handleNextTrackCommand:"))
	remoteCommandCenter.PreviousTrackCommand().AddTarget_action_(h, objc.Sel("handlePreviousTrackCommand:"))
	remoteCommandCenter.ChangeRepeatModeCommand().AddTarget_action_(h, objc.Sel("handleChangeRepeatModeCommand:"))
	remoteCommandCenter.ChangeShuffleModeCommand().AddTarget_action_(h, objc.Sel("handleChangeShuffleModeCommand:"))
	remoteCommandCenter.ChangePlaybackRateCommand().AddTarget_action_(h, objc.Sel("handleChangePlaybackRateCommand:"))
	remoteCommandCenter.SeekBackwardCommand().AddTarget_action_(h, objc.Sel("handleSeekBackwardCommand:"))
	remoteCommandCenter.SeekForwardCommand().AddTarget_action_(h, objc.Sel("handleSeekForwardCommand:"))
	remoteCommandCenter.SkipForwardCommand().AddTarget_action_(h, objc.Sel("handleSkipForwardCommand:"))
	remoteCommandCenter.SkipBackwardCommand().AddTarget_action_(h, objc.Sel("handleSkipBackwardCommand:"))
	remoteCommandCenter.ChangePlaybackPositionCommand().AddTarget_action_(h, objc.Sel("handleChangePlaybackPositionCommand:"))
	remoteCommandCenter.LikeCommand().AddTarget_action_(h, objc.Sel("handleLikeCommand:"))
	remoteCommandCenter.DislikeCommand().AddTarget_action_(h, objc.Sel("handleDisLikeCommand:"))
	remoteCommandCenter.BookmarkCommand().AddTarget_action_(h, objc.Sel("handleBookmarkCommand:"))
	remoteCommandCenter.EnableLanguageOptionCommand().AddTarget_action_(h, objc.Sel("handleEnableLanguageOptionCommand:"))
	remoteCommandCenter.DisableLanguageOptionCommand().AddTarget_action_(h, objc.Sel("handleDisableLanguageOptionCommand:"))
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
	values := core.NSArray_WithObjects(core.NSNumber_numberWithInt_(int32(total)), core.NSNumber_numberWithInt_(int32(ur)), core.NSNumber_numberWithFloat_(1.0))
	values = values.ArrayByAddingObject_(core.NSNumber_numberWithFloat_(1.0))
	keys := core.NSArray_WithObjects(core.String(mediaplayer.MPMediaItemPropertyPlaybackDuration), core.String(mediaplayer.MPNowPlayingInfoPropertyElapsedPlaybackTime), core.String(mediaplayer.MPNowPlayingInfoPropertyPlaybackRate))
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
