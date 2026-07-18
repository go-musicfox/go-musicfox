//go:build darwin

package avcore

import "github.com/go-musicfox/go-musicfox/internal/macdriver/core"

// AVAssetTrack identifies a single asset track for AVAudioMix parameters.
type AVAssetTrack struct {
	core.NSObject
}
