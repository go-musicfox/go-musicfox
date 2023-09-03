//go:build darwin

package state_handler

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/mediaplayer"
)

func TestRemoteCommandHandler(t *testing.T) {
	handler := remoteCommandHandler_new(nil)
	if handler.ID == 0 {
		panic("new remote command handler failed")
	}

	center := mediaplayer.MPRemoteCommandCenter_sharedCommandCenter()
	center.PlayCommand().AddTargetAction(handler.ID, sel_handlePlayCommand)
	center.PauseCommand().AddTargetAction(handler.ID, sel_handlePauseCommand)
	center.StopCommand().AddTargetAction(handler.ID, sel_handleStopCommand)
	center.TogglePlayPauseCommand().AddTargetAction(handler.ID, sel_handleTogglePlayPauseCommand)
	center.NextTrackCommand().AddTargetAction(handler.ID, sel_handleNextTrackCommand)
	center.PreviousTrackCommand().AddTargetAction(handler.ID, sel_handlePreviousTrackCommand)
	center.ChangeRepeatModeCommand().AddTargetAction(handler.ID, sel_handleChangeRepeatModeCommand)
	center.ChangeShuffleModeCommand().AddTargetAction(handler.ID, sel_handleChangeShuffleModeCommand)
	center.ChangePlaybackRateCommand().AddTargetAction(handler.ID, sel_handleChangePlaybackRateCommand)
	center.SeekBackwardCommand().AddTargetAction(handler.ID, sel_handleSeekBackwardCommand)
	center.SeekForwardCommand().AddTargetAction(handler.ID, sel_handleSeekForwardCommand)
	center.SkipForwardCommand().AddTargetAction(handler.ID, sel_handleSkipForwardCommand)
	center.SkipBackwardCommand().AddTargetAction(handler.ID, sel_handleSkipBackwardCommand)
	center.ChangePlaybackPositionCommand().AddTargetAction(handler.ID, sel_handleChangePlaybackPositionCommand)
	center.LikeCommand().AddTargetAction(handler.ID, sel_handleLikeCommand)
	center.DislikeCommand().AddTargetAction(handler.ID, sel_handleDislikeCommand)
	center.BookmarkCommand().AddTargetAction(handler.ID, sel_handleBookmarkCommand)
	center.EnableLanguageOptionCommand().AddTargetAction(handler.ID, sel_handleEnableLanguageOptionCommand)
	center.DisableLanguageOptionCommand().AddTargetAction(handler.ID, sel_handleDisableLanguageOptionCommand)
}
