package state_handler

import "time"

type Controller interface {
	CtrlPaused()
	CtrlResume()
	CtrlStop()
	CtrlToggle()
	CtrlNext()
	CtrlPrevious()
	CtrlSeek(duration time.Duration)
	CtrlSetVolume(volume int)
}
