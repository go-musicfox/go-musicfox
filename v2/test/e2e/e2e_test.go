package e2e

import (
	"testing"
)

// TestE2ESuite 运行所有端到端测试
func TestE2ESuite(t *testing.T) {
	t.Log("Starting go-musicfox end-to-end tests...")

	// 1. 音乐播放端到端测试
	t.Run("MusicPlaybackE2E", TestMusicPlaybackE2E)

	t.Log("All end-to-end tests completed successfully!")
}