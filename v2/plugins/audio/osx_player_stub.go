//go:build !darwin
// +build !darwin

package audio

import "fmt"

// newOSXPlayer 在非macOS平台创建OSX播放器（返回错误）
func newOSXPlayer(config map[string]interface{}) (PlayerBackend, error) {
	return nil, fmt.Errorf("osx player is only available on macOS platform")
}
