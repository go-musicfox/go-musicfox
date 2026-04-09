//go:build darwin

package remote_control

import "time"

type Controller interface {
	CtrlPause()
	CtrlResume()
	CtrlStop()
	CtrlToggle()
	CtrlNext()
	CtrlPrevious()
	CtrlSeek(duration time.Duration)
	CtrlSetVolume(volume int)
	CtrlLikeNowPlaying()
	CtrlDislikeNowPlaying()
	CtrlShuffle()
	CtrlRepeat()
	CtrlSetRepeat(mode any)
	CtrlSetShuffle(mode any)
}
