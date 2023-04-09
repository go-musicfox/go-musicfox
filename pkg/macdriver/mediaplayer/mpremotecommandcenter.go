//go:build darwin

package mediaplayer

import (
	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
)

func init() {
	importFramework()
	class_MPRemoteCommandCenter = objc.GetClass("MPRemoteCommandCenter")
}

var (
	class_MPRemoteCommandCenter objc.Class
)

var (
	sel_sharedCommandCenter           = objc.RegisterName("sharedCommandCenter")
	sel_pauseCommand                  = objc.RegisterName("pauseCommand")
	sel_playCommand                   = objc.RegisterName("playCommand")
	sel_stopCommand                   = objc.RegisterName("stopCommand")
	sel_togglePlayPauseCommand        = objc.RegisterName("togglePlayPauseCommand")
	sel_nextTrackCommand              = objc.RegisterName("nextTrackCommand")
	sel_previousTrackCommand          = objc.RegisterName("previousTrackCommand")
	sel_changeRepeatModeCommand       = objc.RegisterName("changeRepeatModeCommand")
	sel_changeShuffleModeCommand      = objc.RegisterName("changeShuffleModeCommand")
	sel_changePlaybackRateCommand     = objc.RegisterName("changePlaybackRateCommand")
	sel_seekBackwardCommand           = objc.RegisterName("seekBackwardCommand")
	sel_seekForwardCommand            = objc.RegisterName("seekForwardCommand")
	sel_skipBackwardCommand           = objc.RegisterName("skipBackwardCommand")
	sel_skipForwardCommand            = objc.RegisterName("skipForwardCommand")
	sel_changePlaybackPositionCommand = objc.RegisterName("changePlaybackPositionCommand")
	sel_ratingCommand                 = objc.RegisterName("ratingCommand")
	sel_likeCommand                   = objc.RegisterName("likeCommand")
	sel_dislikeCommand                = objc.RegisterName("dislikeCommand")
	sel_bookmarkCommand               = objc.RegisterName("bookmarkCommand")
	sel_enableLanguageOptionCommand   = objc.RegisterName("enableLanguageOptionCommand")
	sel_disableLanguageOptionCommand  = objc.RegisterName("disableLanguageOptionCommand")
)

type MPRemoteCommandCenter struct {
	core.NSObject
}

func MPRemoteCommandCenter_sharedCommandCenter() MPRemoteCommandCenter {
	return MPRemoteCommandCenter{
		core.NSObject{ID: objc.ID(class_MPRemoteCommandCenter).Send(sel_sharedCommandCenter)},
	}
}

func (c MPRemoteCommandCenter) PauseCommand() (cmd MPRemoteCommand) {
	cmd.SetObjcID(c.Send(sel_pauseCommand))
	return
}

func (c MPRemoteCommandCenter) PlayCommand() (cmd MPRemoteCommand) {
	cmd.SetObjcID(c.Send(sel_playCommand))
	return
}

func (c MPRemoteCommandCenter) StopCommand() (cmd MPRemoteCommand) {
	cmd.SetObjcID(c.Send(sel_stopCommand))
	return
}

func (c MPRemoteCommandCenter) TogglePlayPauseCommand() (cmd MPRemoteCommand) {
	cmd.SetObjcID(c.Send(sel_togglePlayPauseCommand))
	return
}

func (c MPRemoteCommandCenter) NextTrackCommand() (cmd MPRemoteCommand) {
	cmd.SetObjcID(c.Send(sel_nextTrackCommand))
	return
}

func (c MPRemoteCommandCenter) PreviousTrackCommand() (cmd MPRemoteCommand) {
	cmd.SetObjcID(c.Send(sel_previousTrackCommand))
	return
}

func (c MPRemoteCommandCenter) ChangeRepeatModeCommand() (cmd MPChangeRepeatModeCommand) {
	cmd.SetObjcID(c.Send(sel_changeRepeatModeCommand))
	return
}

func (c MPRemoteCommandCenter) ChangeShuffleModeCommand() (cmd MPChangeShuffleModeCommand) {
	cmd.SetObjcID(c.Send(sel_changeShuffleModeCommand))
	return
}

func (c MPRemoteCommandCenter) ChangePlaybackRateCommand() (cmd MPChangePlaybackRateCommand) {
	cmd.SetObjcID(c.Send(sel_changePlaybackRateCommand))
	return
}

func (c MPRemoteCommandCenter) SeekBackwardCommand() (cmd MPChangePlaybackRateCommand) {
	cmd.SetObjcID(c.Send(sel_seekBackwardCommand))
	return
}

func (c MPRemoteCommandCenter) SeekForwardCommand() (cmd MPChangePlaybackRateCommand) {
	cmd.SetObjcID(c.Send(sel_seekForwardCommand))
	return
}

func (c MPRemoteCommandCenter) SkipBackwardCommand() (cmd MPSkipIntervalCommand) {
	cmd.SetObjcID(c.Send(sel_skipBackwardCommand))
	return
}

func (c MPRemoteCommandCenter) SkipForwardCommand() (cmd MPSkipIntervalCommand) {
	cmd.SetObjcID(c.Send(sel_skipForwardCommand))
	return
}

func (c MPRemoteCommandCenter) ChangePlaybackPositionCommand() (cmd MPChangePlaybackPositionCommand) {
	cmd.SetObjcID(c.Send(sel_changePlaybackPositionCommand))
	return
}

func (c MPRemoteCommandCenter) RatingCommand() (cmd MPRatingCommand) {
	cmd.SetObjcID(c.Send(sel_ratingCommand))
	return
}

func (c MPRemoteCommandCenter) LikeCommand() (cmd MPFeedbackCommand) {
	cmd.SetObjcID(c.Send(sel_likeCommand))
	return
}

func (c MPRemoteCommandCenter) DislikeCommand() (cmd MPFeedbackCommand) {
	cmd.SetObjcID(c.Send(sel_dislikeCommand))
	return
}

func (c MPRemoteCommandCenter) BookmarkCommand() (cmd MPFeedbackCommand) {
	cmd.SetObjcID(c.Send(sel_bookmarkCommand))
	return
}

func (c MPRemoteCommandCenter) EnableLanguageOptionCommand() (cmd MPRemoteCommand) {
	cmd.SetObjcID(c.Send(sel_enableLanguageOptionCommand))
	return
}

func (c MPRemoteCommandCenter) DisableLanguageOptionCommand() (cmd MPRemoteCommand) {
	cmd.SetObjcID(c.Send(sel_disableLanguageOptionCommand))
	return
}
