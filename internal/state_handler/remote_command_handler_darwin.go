//go:build darwin

package state_handler

import (
	"time"

	"github.com/ebitengine/purego/objc"
	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/mediaplayer"
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

	sels = map[string]objc.SEL{
		"HandlePlayCommand":                   sel_handlePlayCommand,
		"HandlePauseCommand":                  sel_handlePauseCommand,
		"HandleStopCommand":                   sel_handleStopCommand,
		"HandleTogglePlayPauseCommand":        sel_handleTogglePlayPauseCommand,
		"HandleNextTrackCommand":              sel_handleNextTrackCommand,
		"HandlePreviousTrackCommand":          sel_handlePreviousTrackCommand,
		"HandleChangeRepeatModeCommand":       sel_handleChangeRepeatModeCommand,
		"HandleChangeShuffleModeCommand":      sel_handleChangeShuffleModeCommand,
		"HandleChangePlaybackRateCommand":     sel_handleChangePlaybackRateCommand,
		"HandleSeekBackwardCommand":           sel_handleSeekBackwardCommand,
		"HandleSeekForwardCommand":            sel_handleSeekForwardCommand,
		"HandleSkipForwardCommand":            sel_handleSkipForwardCommand,
		"HandleSkipBackwardCommand":           sel_handleSkipBackwardCommand,
		"HandleChangePlaybackPositionCommand": sel_handleChangePlaybackPositionCommand,
		"HandleLikeCommand":                   sel_handleLikeCommand,
		"HandleDislikeCommand":                sel_handleDislikeCommand,
		"HandleBookmarkCommand":               sel_handleBookmarkCommand,
		"HandleEnableLanguageOptionCommand":   sel_handleEnableLanguageOptionCommand,
		"HandleDisableLanguageOptionCommand":  sel_handleDisableLanguageOptionCommand,
		"HandleWillSleepOrPowerOff":           sel_handleWillSleepOrPowerOff,
		"HandleDidWake":                       sel_handleDidWake,
	}
)

type remoteCommandHandlerBinding struct {
	//nolint:golint,unused
	isa objc.Class `objc:"RemoteCommandHandler : NSObject"`
}

func (remoteCommandHandlerBinding) HandlePlayCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.CtrlResume()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandlePauseCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.CtrlPaused()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleStopCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.CtrlPaused()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleTogglePlayPauseCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.CtrlToggle()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandleNextTrackCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.CtrlNext()
	return mediaplayer.MPRemoteCommandHandlerStatusSuccess
}

func (remoteCommandHandlerBinding) HandlePreviousTrackCommand(_ objc.SEL, event objc.ID) mediaplayer.MPRemoteCommandHandlerStatus {
	if _playerController == nil {
		return mediaplayer.MPRemoteCommandHandlerStatusCommandFailed
	}
	_playerController.CtrlPrevious()
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
	var position time.Duration
	core.Autorelease(func() {
		pos := objc.Send[float64](event, objc.RegisterName("positionTime"))
		position = time.Duration(pos) * time.Second
	})
	_playerController.CtrlSeek(position)
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
	_playerController.CtrlPaused()
}

func (remoteCommandHandlerBinding) HandleDidWake(_ objc.SEL, notification objc.ID) {
}

func (remoteCommandHandlerBinding) Selector(metName string) objc.SEL {
	if sel, ok := sels[metName]; ok {
		return sel
	}
	return 0
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
