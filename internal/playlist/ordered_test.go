package playlist

import (
	"fmt"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// createTestPlaylist 创建测试用的播放列表
func createTestPlaylist(size int) []structs.Song {
	playlist := make([]structs.Song, size)
	for i := 0; i < size; i++ {
		playlist[i] = structs.Song{
			Id:   int64(i + 1),
			Name: fmt.Sprintf("Song %d", i+1),
		}
	}
	return playlist
}

func TestOrderedPlayMode_NextSong(t *testing.T) {
	mode := NewOrderedPlayMode()
	
	t.Run("空播放列表", func(t *testing.T) {
		emptyPlaylist := []structs.Song{}
		index, err := mode.NextSong(0, emptyPlaylist, false)
		if err == nil {
			t.Error("期望返回错误，但没有错误")
		}
		if index != -1 {
			t.Errorf("期望索引为-1，实际为%d", index)
		}
	})
	
	t.Run("正常情况 - 从第一首到第二首", func(t *testing.T) {
		playlist := createTestPlaylist(3)
		index, err := mode.NextSong(0, playlist, false)
		if err != nil {
			t.Errorf("不期望错误，但得到: %v", err)
		}
		if index != 1 {
			t.Errorf("期望索引为1，实际为%d", index)
		}
	})
	
	t.Run("正常情况 - 从中间歌曲到下一首", func(t *testing.T) {
		playlist := createTestPlaylist(3)
		index, err := mode.NextSong(1, playlist, false)
		if err != nil {
			t.Errorf("不期望错误，但得到: %v", err)
		}
		if index != 2 {
			t.Errorf("期望索引为2，实际为%d", index)
		}
	})
	
	t.Run("边界情况 - 最后一首歌的下一首", func(t *testing.T) {
		playlist := createTestPlaylist(3)
		index, err := mode.NextSong(2, playlist, false)
		if err == nil {
			t.Error("期望返回错误，但没有错误")
		}
		if index != -1 {
			t.Errorf("期望索引为-1，实际为%d", index)
		}
	})
	
	t.Run("无效索引 - 负数", func(t *testing.T) {
		playlist := createTestPlaylist(3)
		index, err := mode.NextSong(-1, playlist, false)
		if err != nil {
			t.Errorf("不期望错误，但得到: %v", err)
		}
		if index != 0 {
			t.Errorf("期望索引为0，实际为%d", index)
		}
	})
	
	t.Run("无效索引 - 超出范围", func(t *testing.T) {
		playlist := createTestPlaylist(3)
		index, err := mode.NextSong(5, playlist, false)
		if err != nil {
			t.Errorf("不期望错误，但得到: %v", err)
		}
		if index != 0 {
			t.Errorf("期望索引为0，实际为%d", index)
		}
	})
}

func TestOrderedPlayMode_PreviousSong(t *testing.T) {
	mode := NewOrderedPlayMode()
	
	t.Run("空播放列表", func(t *testing.T) {
		emptyPlaylist := []structs.Song{}
		index, err := mode.PreviousSong(0, emptyPlaylist, false)
		if err == nil {
			t.Error("期望返回错误，但没有错误")
		}
		if index != -1 {
			t.Errorf("期望索引为-1，实际为%d", index)
		}
	})
	
	t.Run("正常情况 - 从第二首到第一首", func(t *testing.T) {
		playlist := createTestPlaylist(3)
		index, err := mode.PreviousSong(1, playlist, false)
		if err != nil {
			t.Errorf("不期望错误，但得到: %v", err)
		}
		if index != 0 {
			t.Errorf("期望索引为0，实际为%d", index)
		}
	})
	
	t.Run("正常情况 - 从最后一首到倒数第二首", func(t *testing.T) {
		playlist := createTestPlaylist(3)
		index, err := mode.PreviousSong(2, playlist, false)
		if err != nil {
			t.Errorf("不期望错误，但得到: %v", err)
		}
		if index != 1 {
			t.Errorf("期望索引为1，实际为%d", index)
		}
	})
	
	t.Run("边界情况 - 第一首歌的上一首", func(t *testing.T) {
		playlist := createTestPlaylist(3)
		index, err := mode.PreviousSong(0, playlist, false)
		if err == nil {
			t.Error("期望返回错误，但没有错误")
		}
		if index != -1 {
			t.Errorf("期望索引为-1，实际为%d", index)
		}
	})
	
	t.Run("无效索引 - 负数", func(t *testing.T) {
		playlist := createTestPlaylist(3)
		index, err := mode.PreviousSong(-1, playlist, false)
		if err != nil {
			t.Errorf("不期望错误，但得到: %v", err)
		}
		if index != 2 {
			t.Errorf("期望索引为2，实际为%d", index)
		}
	})
	
	t.Run("无效索引 - 超出范围", func(t *testing.T) {
		playlist := createTestPlaylist(3)
		index, err := mode.PreviousSong(5, playlist, false)
		if err != nil {
			t.Errorf("不期望错误，但得到: %v", err)
		}
		if index != 2 {
			t.Errorf("期望索引为2，实际为%d", index)
		}
	})
}

func TestOrderedPlayMode_Initialize(t *testing.T) {
	mode := NewOrderedPlayMode()
	playlist := createTestPlaylist(3)
	
	err := mode.Initialize(1, playlist)
	if err != nil {
		t.Errorf("不期望错误，但得到: %v", err)
	}
}

func TestOrderedPlayMode_GetMode(t *testing.T) {
	mode := NewOrderedPlayMode()
	if mode.GetMode() != types.PmOrdered {
		t.Errorf("期望模式为PmOrdered，实际为%v", mode.GetMode())
	}
}

func TestOrderedPlayMode_GetModeName(t *testing.T) {
	mode := NewOrderedPlayMode()
	expected := "顺序播放"
	if mode.GetModeName() != expected {
		t.Errorf("期望模式名称为%s，实际为%s", expected, mode.GetModeName())
	}
}

func TestOrderedPlayMode_OnPlaylistChanged(t *testing.T) {
	mode := NewOrderedPlayMode()
	playlist := createTestPlaylist(3)
	
	err := mode.OnPlaylistChanged(1, playlist)
	if err != nil {
		t.Errorf("不期望错误，但得到: %v", err)
	}
}