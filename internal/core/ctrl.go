// Package core implements the playback engine coordination layer. It is
// frontend-agnostic: no TUI (foxful-cli / bubbletea / internal/ui) imports are
// allowed. Frontends embed the coordinator and observe playback events through
// the Observer seam.
package core

import "time"

// PlayDirection 下首歌的方向
type PlayDirection uint8

const (
	DurationNext PlayDirection = iota
	DurationPrev
)

type CtrlType string

type CtrlSignal struct {
	Type        CtrlType
	Duration    time.Duration
	RepeatType  any
	ShuffleType any
}

const (
	CtrlResume   CtrlType = "Resume"
	CtrlPaused   CtrlType = "Paused"
	CtrlStop     CtrlType = "Stop"
	CtrlToggle   CtrlType = "Toggle"
	CtrlPrevious CtrlType = "Previous"
	CtrlNext     CtrlType = "Next"
	CtrlSeek     CtrlType = "Seek"
	CtrlRerender CtrlType = "Rerender"
	CtrlShuffle  CtrlType = "Shuffle"
	CtrlRepeat   CtrlType = "Repeat"
)
