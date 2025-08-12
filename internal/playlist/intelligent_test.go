package playlist

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func TestIntelligentPlayMode_GetMode(t *testing.T) {
	mode := NewIntelligentPlayMode()
	if mode.GetMode() != types.PmIntelligent {
		t.Errorf("Expected mode %v, got %v", types.PmIntelligent, mode.GetMode())
	}
}

func TestIntelligentPlayMode_GetModeName(t *testing.T) {
	mode := NewIntelligentPlayMode()
	expected := "心动模式"
	if mode.GetModeName() != expected {
		t.Errorf("Expected mode name %s, got %s", expected, mode.GetModeName())
	}
}

func TestIntelligentPlayMode_Initialize(t *testing.T) {
	mode := NewIntelligentPlayMode().(*IntelligentPlayMode)
	playlist := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
		{Id: 3, Name: "Song 3"},
	}

	// 测试正常初始化
	err := mode.Initialize(1, playlist)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if mode.currentIndex != 1 {
		t.Errorf("Expected currentIndex 1, got %d", mode.currentIndex)
	}
	if mode.originalSong.Id != 2 {
		t.Errorf("Expected originalSong.Id 2, got %d", mode.originalSong.Id)
	}

	// 测试空播放列表
	err = mode.Initialize(0, []structs.Song{})
	if err != ErrEmptyPlaylist {
		t.Errorf("Expected ErrEmptyPlaylist, got %v", err)
	}

	// 测试无效索引
	err = mode.Initialize(-1, playlist)
	if err != ErrInvalidIndex {
		t.Errorf("Expected ErrInvalidIndex, got %v", err)
	}

	err = mode.Initialize(10, playlist)
	if err != ErrInvalidIndex {
		t.Errorf("Expected ErrInvalidIndex, got %v", err)
	}
}

func TestIntelligentPlayMode_NextSong(t *testing.T) {
	mode := NewIntelligentPlayMode().(*IntelligentPlayMode)
	playlist := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
		{Id: 3, Name: "Song 3"},
	}

	// 测试空播放列表
	index, err := mode.NextSong(0, []structs.Song{}, false)
	if err != ErrEmptyPlaylist {
		t.Errorf("Expected ErrEmptyPlaylist, got %v", err)
	}
	if index != -1 {
		t.Errorf("Expected index -1, got %d", index)
	}

	// 测试无效当前索引
	index, err = mode.NextSong(-1, playlist, false)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if index != 0 {
		t.Errorf("Expected index 0, got %d", index)
	}

	// 测试正常下一首
	index, err = mode.NextSong(0, playlist, false)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if index != 1 {
		t.Errorf("Expected index 1, got %d", index)
	}

	// 测试到达列表末尾
	index, err = mode.NextSong(2, playlist, false)
	if err != ErrNoNextSong {
		t.Errorf("Expected ErrNoNextSong, got %v", err)
	}
	if index != -1 {
		t.Errorf("Expected index -1, got %d", index)
	}
}

func TestIntelligentPlayMode_PreviousSong(t *testing.T) {
	mode := NewIntelligentPlayMode().(*IntelligentPlayMode)
	playlist := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
		{Id: 3, Name: "Song 3"},
	}

	// 测试空播放列表
	index, err := mode.PreviousSong(0, []structs.Song{}, false)
	if err != ErrEmptyPlaylist {
		t.Errorf("Expected ErrEmptyPlaylist, got %v", err)
	}
	if index != -1 {
		t.Errorf("Expected index -1, got %d", index)
	}

	// 测试无效当前索引
	index, err = mode.PreviousSong(-1, playlist, false)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if index != 2 {
		t.Errorf("Expected index 2, got %d", index)
	}

	// 测试正常上一首
	index, err = mode.PreviousSong(2, playlist, false)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if index != 1 {
		t.Errorf("Expected index 1, got %d", index)
	}

	// 测试到达列表开头
	index, err = mode.PreviousSong(0, playlist, false)
	if err != ErrNoPreviousSong {
		t.Errorf("Expected ErrNoPreviousSong, got %v", err)
	}
	if index != -1 {
		t.Errorf("Expected index -1, got %d", index)
	}
}

func TestIntelligentPlayMode_OnPlaylistChanged(t *testing.T) {
	mode := NewIntelligentPlayMode().(*IntelligentPlayMode)
	playlist := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
	}

	// 测试空播放列表
	err := mode.OnPlaylistChanged(0, []structs.Song{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if mode.currentIndex != -1 {
		t.Errorf("Expected currentIndex -1, got %d", mode.currentIndex)
	}
	if mode.recommendedSongs != nil {
		t.Errorf("Expected recommendedSongs to be nil")
	}

	// 测试正常播放列表变化
	err = mode.OnPlaylistChanged(1, playlist)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if mode.currentIndex != 1 {
		t.Errorf("Expected currentIndex 1, got %d", mode.currentIndex)
	}
	if mode.originalSong.Id != 2 {
		t.Errorf("Expected originalSong.Id 2, got %d", mode.originalSong.Id)
	}

	// 测试无效索引
	err = mode.OnPlaylistChanged(-1, playlist)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// 无效索引时不应该更新currentIndex和originalSong
	if mode.currentIndex != 1 {
		t.Errorf("Expected currentIndex to remain 1, got %d", mode.currentIndex)
	}
}