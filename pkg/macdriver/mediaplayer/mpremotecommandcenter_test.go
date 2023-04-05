//go:build darwin

package mediaplayer

import (
	"testing"
)

func assertNotEmpty(command MPRemoteCommand, err string) {
	if command.ID == 0 {
		panic(err)
	}
}

func TestMPRemoteCommandCenter(t *testing.T) {
	center := MPRemoteCommandCenter_sharedCommandCenter()
	if center.ID == 0 {
		panic("get command center failed")
	}

	assertNotEmpty(center.PauseCommand(), "get pause command failed")
	assertNotEmpty(center.PlayCommand(), "get play command failed")
	assertNotEmpty(center.StopCommand(), "get stop command failed")
	assertNotEmpty(center.TogglePlayPauseCommand(), "get toggle command failed")
	assertNotEmpty(center.NextTrackCommand(), "get next command failed")
	assertNotEmpty(center.PreviousTrackCommand(), "get previous command failed")
	assertNotEmpty(center.ChangeRepeatModeCommand().MPRemoteCommand, "get change repeat mode command failed")
	assertNotEmpty(center.ChangeShuffleModeCommand().MPRemoteCommand, "get change shuffle mode command failed")
	assertNotEmpty(center.ChangePlaybackRateCommand().MPRemoteCommand, "get change playback rate command failed")
	assertNotEmpty(center.SeekBackwardCommand().MPRemoteCommand, "get seek backward command failed")
	assertNotEmpty(center.SeekForwardCommand().MPRemoteCommand, "get seek forward command failed")
	assertNotEmpty(center.SkipBackwardCommand().MPRemoteCommand, "get skip backward command failed")
	assertNotEmpty(center.SkipForwardCommand().MPRemoteCommand, "get skip forward command failed")
	assertNotEmpty(center.ChangePlaybackPositionCommand().MPRemoteCommand, "get change playback position command failed")
	assertNotEmpty(center.RatingCommand().MPRemoteCommand, "get rating command failed")
	assertNotEmpty(center.LikeCommand().MPRemoteCommand, "get like command failed")
	assertNotEmpty(center.DislikeCommand().MPRemoteCommand, "get dislike command failed")
	assertNotEmpty(center.BookmarkCommand().MPRemoteCommand, "get bookmark command failed")
	assertNotEmpty(center.EnableLanguageOptionCommand(), "get enable language command failed")
	assertNotEmpty(center.DisableLanguageOptionCommand(), "get disable language command failed")
}
