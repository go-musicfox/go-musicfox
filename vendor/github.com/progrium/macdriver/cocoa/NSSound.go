//go:build darwin
// +build darwin

package cocoa

import (
	"github.com/progrium/macdriver/core"
)

type NSSound struct {
	gen_NSSound
}

func NSSound_InitWithData(data core.NSDataRef) NSSound {
	return NSSound_alloc().InitWithData__asNSSound(data)
}

func NSSound_InitWithURL(url core.NSURLRef, byRef bool) NSSound {
	return NSSound_alloc().InitWithContentsOfURL_byReference__asNSSound(url, byRef)
}

func (sound NSSound) Play() {
	sound.gen_NSSound.Play()
}

func (sound NSSound) Pause() {
	sound.gen_NSSound.Pause()
}

func (sound NSSound) Resume() {
	sound.gen_NSSound.Resume()
}
