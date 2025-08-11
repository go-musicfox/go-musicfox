package playlist

import (
	"sync"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// TestNewPlaylistManager 测试创建新的播放列表管理器
func TestNewPlaylistManager(t *testing.T) {
	manager := NewPlaylistManager()
	if manager == nil {
		t.Fatal("NewPlaylistManager() returned nil")
	}

	// 测试初始状态
	if manager.GetCurrentIndex() != -1 {
		t.Errorf("Expected initial index to be -1, got %d", manager.GetCurrentIndex())
	}

	playlist := manager.GetPlaylist()
	if len(playlist) != 0 {
		t.Errorf("Expected empty playlist, got %d songs", len(playlist))
	}

	// 验证默认播放模式为顺序播放
	if manager.GetPlayMode() != types.PmOrdered {
		t.Errorf("Expected default play mode to be PmOrdered, got %v", manager.GetPlayMode())
	}
}

// TestInitialize 测试初始化播放列表
func TestInitialize(t *testing.T) {
	manager := NewPlaylistManager()

	// 测试空播放列表初始化
	err := manager.Initialize(0, []structs.Song{})
	if err != nil {
		t.Errorf("Initialize with empty playlist should not return error, got %v", err)
	}

	if manager.GetCurrentIndex() != -1 {
		t.Errorf("Expected index to be -1 for empty playlist, got %d", manager.GetCurrentIndex())
	}

	// 测试有效播放列表初始化
	songs := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
		{Id: 3, Name: "Song 3"},
	}

	err = manager.Initialize(1, songs)
	if err != nil {
		t.Errorf("Initialize with valid playlist should not return error, got %v", err)
	}

	if manager.GetCurrentIndex() != 1 {
		t.Errorf("Expected index to be 1, got %d", manager.GetCurrentIndex())
	}

	playlist := manager.GetPlaylist()
	if len(playlist) != 3 {
		t.Errorf("Expected playlist length to be 3, got %d", len(playlist))
	}

	// 测试无效索引
	err = manager.Initialize(-1, songs)
	if err == nil {
		t.Error("Initialize with invalid index should return error")
	}

	err = manager.Initialize(3, songs)
	if err == nil {
		t.Error("Initialize with out of range index should return error")
	}
}

// TestGetCurrentSong 测试获取当前歌曲
func TestGetCurrentSong(t *testing.T) {
	manager := NewPlaylistManager()

	// 测试空播放列表
	_, err := manager.GetCurrentSong()
	if err == nil {
		t.Error("GetCurrentSong with empty playlist should return error")
	}

	// 测试有效播放列表
	songs := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
	}

	err = manager.Initialize(0, songs)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	currentSong, err := manager.GetCurrentSong()
	if err != nil {
		t.Errorf("GetCurrentSong should not return error, got %v", err)
	}

	if currentSong.Id != 1 {
		t.Errorf("Expected current song ID to be 1, got %d", currentSong.Id)
	}
}

// TestRemoveSong 测试移除歌曲
func TestRemoveSong(t *testing.T) {
	manager := NewPlaylistManager()

	// 测试空播放列表
	_, err := manager.RemoveSong(0)
	if err == nil {
		t.Error("RemoveSong with empty playlist should return error")
	}

	// 测试有效播放列表
	songs := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
		{Id: 3, Name: "Song 3"},
	}

	err = manager.Initialize(1, songs)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 测试移除当前播放的歌曲
	nextSong, err := manager.RemoveSong(1)
	if err != nil {
		t.Errorf("RemoveSong should not return error, got %v", err)
	}

	if nextSong.Id != 3 {
		t.Errorf("Expected next song ID to be 3, got %d", nextSong.Id)
	}

	if manager.GetCurrentIndex() != 1 {
		t.Errorf("Expected current index to be 1, got %d", manager.GetCurrentIndex())
	}

	playlist := manager.GetPlaylist()
	if len(playlist) != 2 {
		t.Errorf("Expected playlist length to be 2, got %d", len(playlist))
	}

	// 测试无效索引
	_, err = manager.RemoveSong(-1)
	if err == nil {
		t.Error("RemoveSong with invalid index should return error")
	}

	_, err = manager.RemoveSong(10)
	if err == nil {
		t.Error("RemoveSong with out of range index should return error")
	}
}

// TestSetPlayMode 测试设置播放模式
func TestSetPlayMode(t *testing.T) {
	manager := NewPlaylistManager()

	// 测试设置列表循环模式
	err := manager.SetPlayMode(types.PmListLoop)
	if err != nil {
		t.Errorf("SetPlayMode with PmListLoop should not return error, got %v", err)
	}
	if manager.GetPlayMode() != types.PmListLoop {
		t.Errorf("Expected play mode to be PmListLoop, got %v", manager.GetPlayMode())
	}
	if manager.GetPlayModeName() != "列表循环" {
		t.Errorf("Expected play mode name to be '列表循环', got %s", manager.GetPlayModeName())
	}

	// 测试设置单曲循环模式
	err = manager.SetPlayMode(types.PmSingleLoop)
	if err != nil {
		t.Errorf("SetPlayMode with PmSingleLoop should not return error, got %v", err)
	}
	if manager.GetPlayMode() != types.PmSingleLoop {
		t.Errorf("Expected play mode to be PmSingleLoop, got %v", manager.GetPlayMode())
	}
	if manager.GetPlayModeName() != "单曲循环" {
		t.Errorf("Expected play mode name to be '单曲循环', got %s", manager.GetPlayModeName())
	}

	// 测试设置列表随机播放模式
	err = manager.SetPlayMode(types.PmListRandom)
	if err != nil {
		t.Errorf("SetPlayMode with PmListRandom should not return error, got %v", err)
	}
	if manager.GetPlayMode() != types.PmListRandom {
		t.Errorf("Expected play mode to be PmListRandom, got %v", manager.GetPlayMode())
	}
	if manager.GetPlayModeName() != "列表随机" {
		t.Errorf("Expected play mode name to be '列表随机', got %s", manager.GetPlayModeName())
	}

	// 测试设置无限随机播放模式
	err = manager.SetPlayMode(types.PmInfRandom)
	if err != nil {
		t.Errorf("SetPlayMode with PmInfRandom should not return error, got %v", err)
	}
	if manager.GetPlayMode() != types.PmInfRandom {
		t.Errorf("Expected play mode to be PmInfRandom, got %v", manager.GetPlayMode())
	}
	if manager.GetPlayModeName() != "无限随机" {
		t.Errorf("Expected play mode name to be '无限随机', got %s", manager.GetPlayModeName())
	}

	// 测试设置顺序播放模式
	err = manager.SetPlayMode(types.PmOrdered)
	if err != nil {
		t.Errorf("SetPlayMode with PmOrdered should not return error, got %v", err)
	}
	if manager.GetPlayMode() != types.PmOrdered {
		t.Errorf("Expected play mode to be PmOrdered, got %v", manager.GetPlayMode())
	}

	// 测试设置未知播放模式
	err = manager.SetPlayMode(types.PmUnknown)
	if err == nil {
		t.Error("SetPlayMode with unknown mode should return error")
	}

	// 测试获取播放模式名称
	modeName := manager.GetPlayModeName()
	if modeName == "" {
		t.Error("GetPlayModeName should not return empty string")
	}
}

// TestNextSong 测试下一首歌曲
func TestNextSong(t *testing.T) {
	manager := NewPlaylistManager()

	// 测试空播放列表
	_, err := manager.NextSong(true)
	if err == nil {
		t.Error("NextSong with empty playlist should return error")
	}

	// 测试有效播放列表
	songs := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
		{Id: 3, Name: "Song 3"},
	}

	err = manager.Initialize(0, songs)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 测试下一首歌曲
	nextSong, err := manager.NextSong(true)
	if err != nil {
		t.Errorf("NextSong should not return error, got %v", err)
	}

	if nextSong.Id != 2 {
		t.Errorf("Expected next song ID to be 2, got %d", nextSong.Id)
	}
}

// TestPreviousSong 测试上一首歌曲
func TestPreviousSong(t *testing.T) {
	manager := NewPlaylistManager()

	// 测试空播放列表
	_, err := manager.PreviousSong(true)
	if err == nil {
		t.Error("PreviousSong with empty playlist should return error")
	}

	// 测试有效播放列表
	songs := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
		{Id: 3, Name: "Song 3"},
	}

	err = manager.Initialize(1, songs)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 测试上一首歌曲
	prevSong, err := manager.PreviousSong(true)
	if err != nil {
		t.Errorf("PreviousSong should not return error, got %v", err)
	}

	if prevSong.Id != 1 {
		t.Errorf("Expected previous song ID to be 1, got %d", prevSong.Id)
	}
}

// TestConcurrentAccess 测试并发访问安全性
func TestConcurrentAccess(t *testing.T) {
	manager := NewPlaylistManager()

	songs := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
		{Id: 3, Name: "Song 3"},
		{Id: 4, Name: "Song 4"},
		{Id: 5, Name: "Song 5"},
	}

	err := manager.Initialize(0, songs)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	var wg sync.WaitGroup
	errorChan := make(chan error, 100)

	// 并发读取操作
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, err := manager.GetCurrentSong()
				if err != nil {
					errorChan <- err
					return
				}
				_ = manager.GetPlaylist()
				_ = manager.GetCurrentIndex()
				_ = manager.GetPlayMode()
				_ = manager.GetPlayModeName()
			}
		}()
	}

	// 并发写入操作
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				mode := types.PmOrdered
				if j%2 == 0 {
					mode = types.PmListLoop
				}
				err := manager.SetPlayMode(mode)
				if err != nil {
					errorChan <- err
					return
				}
			}
		}(i)
	}

	// 并发NextSong/PreviousSong操作
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				_, _ = manager.NextSong(false)
				_, _ = manager.PreviousSong(false)
			}
		}()
	}

	wg.Wait()
	close(errorChan)

	// 检查是否有错误
	for err := range errorChan {
		t.Errorf("Concurrent access error: %v", err)
	}
}