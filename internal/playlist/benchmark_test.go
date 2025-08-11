package playlist

import (
	"fmt"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// createTestSongs 创建测试歌曲列表
func createTestSongs(count int) []structs.Song {
	songs := make([]structs.Song, count)
	for i := 0; i < count; i++ {
		songs[i] = structs.Song{
			Id:   int64(i + 1),
			Name: fmt.Sprintf("Test Song %d", i+1),
		}
	}
	return songs
}

// BenchmarkPlaylistManagerInitialize 测试初始化性能
func BenchmarkPlaylistManagerInitialize(b *testing.B) {
	songs := createTestSongs(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager := NewPlaylistManager()
		_ = manager.Initialize(0, songs)
	}
}

// BenchmarkPlaylistManagerNextSong 测试下一首歌曲性能
func BenchmarkPlaylistManagerNextSong(b *testing.B) {
	manager := NewPlaylistManager()
	songs := createTestSongs(1000)
	_ = manager.Initialize(0, songs)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = manager.NextSong(true)
	}
}

// BenchmarkPlaylistManagerPreviousSong 测试上一首歌曲性能
func BenchmarkPlaylistManagerPreviousSong(b *testing.B) {
	manager := NewPlaylistManager()
	songs := createTestSongs(1000)
	_ = manager.Initialize(500, songs)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = manager.PreviousSong(true)
	}
}

// BenchmarkPlaylistManagerSetPlayMode 测试设置播放模式性能
func BenchmarkPlaylistManagerSetPlayMode(b *testing.B) {
	manager := NewPlaylistManager()
	songs := createTestSongs(1000)
	_ = manager.Initialize(0, songs)
	modes := []types.Mode{types.PmOrdered, types.PmListLoop, types.PmSingleLoop, types.PmListRandom, types.PmInfRandom}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mode := modes[i%len(modes)]
		_ = manager.SetPlayMode(mode)
	}
}

// BenchmarkPlaylistManagerRemoveSong 测试删除歌曲性能
func BenchmarkPlaylistManagerRemoveSong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		manager := NewPlaylistManager()
		songs := createTestSongs(1000)
		_ = manager.Initialize(500, songs)
		b.StartTimer()

		_, _ = manager.RemoveSong(100)
	}
}

// BenchmarkPlaylistManagerRandomMode 测试随机模式性能
func BenchmarkPlaylistManagerRandomMode(b *testing.B) {
	manager := NewPlaylistManager()
	songs := createTestSongs(1000)
	_ = manager.Initialize(0, songs)
	_ = manager.SetPlayMode(types.PmListRandom)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = manager.NextSong(true)
	}
}

// BenchmarkPlaylistManagerLargePlaylist 测试大播放列表性能
func BenchmarkPlaylistManagerLargePlaylist(b *testing.B) {
	songs := createTestSongs(10000) // 10k songs
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager := NewPlaylistManager()
		_ = manager.Initialize(5000, songs)
		_, _ = manager.NextSong(true)
		_, _ = manager.PreviousSong(true)
		_ = manager.SetPlayMode(types.PmListRandom)
		_, _ = manager.NextSong(true)
	}
}