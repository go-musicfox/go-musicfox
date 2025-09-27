//go:build !windows
// +build !windows

package audio

import "fmt"

// newWindowsPlayer 在非Windows平台创建Windows播放器（返回错误）
func newWindowsPlayer(config map[string]interface{}) (PlayerBackend, error) {
	return nil, fmt.Errorf("windows player is only available on Windows platform")
}
