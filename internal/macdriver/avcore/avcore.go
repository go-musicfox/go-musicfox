//go:build darwin

package avcore

import (
	"sync"

	"github.com/ebitengine/purego"
)

var importOnce sync.Once

func importFramework() {
	importOnce.Do(func() {
		_, err := purego.Dlopen("/System/Library/Frameworks/AVFoundation.framework/AVFoundation", purego.RTLD_GLOBAL)
		if err != nil {
			panic(err)
		}
	})
}

type CMTime struct {
	Value     int64
	Timescale int32
	Flags     uint32
	Epoch     int64
}
