package state_handler

import (
	"time"

	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/core"
	"github.com/go-musicfox/go-musicfox/pkg/macdriver/mediaplayer"
)

func init() {
	var err error
	class_RemoteCommandHandler, err = objc.RegisterClass(&remoteCommandHandlerBinding{})
	if err != nil {
		panic(err)
	}
}

var (
	class_RemoteCommandHandler objc.Class
	_playerController          Controller
)

var (
	sel_handlePlayCommand                   = objc.RegisterName("handlePlayCommand:")
	sel_handlePauseCommand                  = objc.RegisterName("handlePauseCommand:")
	sel_handleStopCommand                   = objc.RegisterName("handleStopCommand:")
	sel_handleTogglePlayPauseCommand        = objc.RegisterName("handleTogglePlayPauseCommand:")
	sel_handleNextTrackCommand              = objc.RegisterName("handleNextTrackCommand:")
	sel_handlePreviousTrackCommand          = objc.RegisterName("handlePreviousTrackCommand:")
	sel_handleChangeRepeatModeCommand       = objc.RegisterName("handleChangeRepeatModeCommand:")
	sel_handleChangeShuffleModeCommand      = objc.RegisterName("handleChangeShuffleModeCommand:")
	sel_handleChangePlaybackRateCommand     = objc.RegisterName("handleChangePlaybackRateCommand:")
	sel_handleSeekBackwardCommand           = objc.RegisterName("handleSeekBackwardCommand:")
	sel_handleSeekForwardCommand            = objc.RegisterName("handleSeekForwardCommand:")
	sel_handleSkipForwardCommand            = objc.RegisterName("handleSkipForwardCommand:")
	sel_handleSkipBackwardCommand           = objc.RegisterName("handleSkipBackwardCommand:")
	sel_handleChangePlaybackPositionCommand = objc.RegisterName("handleChangePlaybackPositionCommand:")
	sel_handleLikeCommand                   = objc.RegisterName("handleLikeCommand:")
	sel_handleDislikeCommand                = objc.RegisterName("handleDislikeCommand:")
	sel_handleBookmarkCommand               = objc.RegisterName("handleBookmarkCommand:")
	sel_handleEnableLanguageOptionCommand   = objc.RegisterName("handleEnableLanguageOptionCommand:")
	sel_handleDisableLanguageOptionCommand  = objc.RegisterName("handleDisableLanguageOptionCommand:")
	sel_handleWillSleepOrPowerOff           = objc.RegisterName("handleWillSleepOrPowerOff:")
	sel_handleDidWake                       = objc.RegisterName("handleDidWake:")
)

type remoteCommandHandlerBinding struct {
	isa objc.Class `objc:"RemoteCommandHandler : NSObject"`
}

func (remoteCommandHandlerBinding) HandlePlayCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.Resume()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandlePauseCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.Paused()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleStopCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.Paused()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleTogglePlayPauseCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.Toggle()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleNextTrackCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.Next()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandlePreviousTrackCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.Previous()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleChangeRepeatModeCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleChangeShuffleModeCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleChangePlaybackRateCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleSeekBackwardCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleSeekForwardCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleSkipForwardCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleSkipBackwardCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleChangePlaybackPositionCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	// event MPChangePlaybackPositionCommandEvent
	position := objc.Send[float64](event, objc.RegisterName("positionTime"))
	_playerController.Seek(time.Duration(position) * time.Second)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleLikeCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleDislikeCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleBookmarkCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleEnableLanguageOptionCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleDisableLanguageOptionCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleWillSleepOrPowerOff(_ objc.SEL, notification objc.ID) {
	if _playerController == nil {
		return
	}
	_playerController.Paused()
}

func (remoteCommandHandlerBinding) HandleDidWake(_ objc.SEL, notification objc.ID) {
}

func (remoteCommandHandlerBinding) Selector(metName string) objc.SEL {
	switch metName {
	case "HandlePlayCommand":
		return sel_handlePlayCommand
	case "HandlePauseCommand":
		return sel_handlePauseCommand
	case "HandleStopCommand":
		return sel_handleStopCommand
	case "HandleTogglePlayPauseCommand":
		return sel_handleTogglePlayPauseCommand
	case "HandleNextTrackCommand":
		return sel_handleNextTrackCommand
	case "HandlePreviousTrackCommand":
		return sel_handlePreviousTrackCommand
	case "HandleChangeRepeatModeCommand":
		return sel_handleChangeRepeatModeCommand
	case "HandleChangeShuffleModeCommand":
		return sel_handleChangeShuffleModeCommand
	case "HandleChangePlaybackRateCommand":
		return sel_handleChangePlaybackRateCommand
	case "HandleSeekBackwardCommand":
		return sel_handleSeekBackwardCommand
	case "HandleSeekForwardCommand":
		return sel_handleSeekForwardCommand
	case "HandleSkipForwardCommand":
		return sel_handleSkipForwardCommand
	case "HandleSkipBackwardCommand":
		return sel_handleSkipBackwardCommand
	case "HandleChangePlaybackPositionCommand":
		return sel_handleChangePlaybackPositionCommand
	case "HandleLikeCommand":
		return sel_handleLikeCommand
	case "HandleDislikeCommand":
		return sel_handleDislikeCommand
	case "HandleBookmarkCommand":
		return sel_handleBookmarkCommand
	case "HandleEnableLanguageOptionCommand":
		return sel_handleEnableLanguageOptionCommand
	case "HandleDisableLanguageOptionCommand":
		return sel_handleDisableLanguageOptionCommand
	case "HandleWillSleepOrPowerOff":
		return sel_handleWillSleepOrPowerOff
	case "HandleDidWake":
		return sel_handleDidWake
	default:
		return 0
	}
}

type remoteCommandHandler struct {
	core.NSObject
}

func remoteCommandHandler_new(c Controller) remoteCommandHandler {
	_playerController = c
	return remoteCommandHandler{
		NSObject: core.NSObject{
			ID: objc.ID(class_RemoteCommandHandler).Send(macdriver.SEL_new),
		},
	}
}
