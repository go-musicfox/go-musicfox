package remote_control

import (
	"time"

	"github.com/go-musicfox/go-musicfox/internal/macdriver/mediaplayer"
)

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
	CtrlSetRepeat(mode mediaplayer.MPRepeatType)
	CtrlSetShuffle(mode mediaplayer.MPShuffleType)
}
