//go:build darwin

package state_handler

import (
	"time"

	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/mediaplayer"
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

func init() {
	var err error
	class_RemoteCommandHandler, err = objc.RegisterClass(
		"RemoteCommandHandler",
		objc.GetClass("NSObject"),
		[]*objc.Protocol{},
		[]objc.FieldDef{},
		[]objc.MethodDef{
			{
				Cmd: sel_handlePlayCommand,
				Fn:  handlePlayCommand,
			},
			{
				Cmd: sel_handlePauseCommand,
				Fn:  handlePauseCommand,
			},
			{
				Cmd: sel_handleStopCommand,
				Fn:  handleStopCommand,
			},
			{
				Cmd: sel_handleTogglePlayPauseCommand,
				Fn:  handleTogglePlayPauseCommand,
			},
			{
				Cmd: sel_handleNextTrackCommand,
				Fn:  handleNextTrackCommand,
			},
			{
				Cmd: sel_handlePreviousTrackCommand,
				Fn:  handlePreviousTrackCommand,
			},
			{
				Cmd: sel_handleChangeRepeatModeCommand,
				Fn:  handleChangeRepeatModeCommand,
			},
			{
				Cmd: sel_handleChangeShuffleModeCommand,
				Fn:  handleChangeShuffleModeCommand,
			},
			{
				Cmd: sel_handleChangePlaybackRateCommand,
				Fn:  handleChangePlaybackRateCommand,
			},
			{
				Cmd: sel_handleSeekBackwardCommand,
				Fn:  handleSeekBackwardCommand,
			},
			{
				Cmd: sel_handleSeekForwardCommand,
				Fn:  handleSeekForwardCommand,
			},
			{
				Cmd: sel_handleSkipForwardCommand,
				Fn:  handleSkipForwardCommand,
			},
			{
				Cmd: sel_handleSkipBackwardCommand,
				Fn:  handleSkipBackwardCommand,
			},
			{
				Cmd: sel_handleChangePlaybackPositionCommand,
				Fn:  handleChangePlaybackPositionCommand,
			},
			{
				Cmd: sel_handleLikeCommand,
				Fn:  handleLikeCommand,
			},
			{
				Cmd: sel_handleDislikeCommand,
				Fn:  handleDislikeCommand,
			},
			{
				Cmd: sel_handleBookmarkCommand,
				Fn:  handleBookmarkCommand,
			},
			{
				Cmd: sel_handleEnableLanguageOptionCommand,
				Fn:  handleEnableLanguageOptionCommand,
			},
			{
				Cmd: sel_handleDisableLanguageOptionCommand,
				Fn:  handleDisableLanguageOptionCommand,
			},
			{
				Cmd: sel_handleWillSleepOrPowerOff,
				Fn:  handleWillSleepOrPowerOff,
			},
			{
				Cmd: sel_handleDidWake,
				Fn:  handleDidWake,
			},
		},
	)
	if err != nil {
		panic(err)
	}
}

var (
	class_RemoteCommandHandler objc.Class
	_playerController          Controller
)

func handlePlayCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.CtrlResume()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handlePauseCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.CtrlPaused()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleStopCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.CtrlPaused()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleTogglePlayPauseCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.CtrlToggle()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleNextTrackCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.CtrlNext()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handlePreviousTrackCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.CtrlPrevious()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleChangeRepeatModeCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleChangeShuffleModeCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleChangePlaybackRateCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleSeekBackwardCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleSeekForwardCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleSkipForwardCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleSkipBackwardCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleChangePlaybackPositionCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	// event MPChangePlaybackPositionCommandEvent
	var position time.Duration
	core.Autorelease(func() {
		pos := objc.Send[float64](event, objc.RegisterName("positionTime"))
		position = time.Duration(pos) * time.Second
	})
	_playerController.CtrlSeek(position)
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleLikeCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleDislikeCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleBookmarkCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleEnableLanguageOptionCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleDisableLanguageOptionCommand(id objc.ID, cmd objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func handleWillSleepOrPowerOff(id objc.ID, cmd objc.SEL, notification objc.ID) {
	if _playerController == nil {
		return
	}
	_playerController.CtrlPaused()
}

func handleDidWake(id objc.ID, cmd objc.SEL, notification objc.ID) {
}

type remoteCommandHandler struct {
	core.NSObject
}

func remoteCommandHandler_new(c Controller) remoteCommandHandler {
	_playerController = c
	return remoteCommandHandler{
		core.NSObject{
			ID: objc.ID(class_RemoteCommandHandler).Send(macdriver.SEL_new),
		},
	}
}
