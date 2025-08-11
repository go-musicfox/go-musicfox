package playlist

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func TestListLoopPlayMode_NextSong(t *testing.T) {
	mode := NewListLoopPlayMode()
	
	// 测试空播放列表
	t.Run("empty playlist", func(t *testing.T) {
		_, err := mode.NextSong(0, []structs.Song{}, false)
		if err != ErrEmptyPlaylist {
			t.Errorf("expected ErrEmptyPlaylist, got %v", err)
		}
	})
	
	// 创建测试播放列表
	playlist := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
		{Id: 3, Name: "Song 3"},
	}
	
	// 测试正常情况
	t.Run("normal next", func(t *testing.T) {
		nextIndex, err := mode.NextSong(0, playlist, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if nextIndex != 1 {
			t.Errorf("expected index 1, got %d", nextIndex)
		}
	})
	
	// 测试循环到开头
	t.Run("loop to beginning", func(t *testing.T) {
		nextIndex, err := mode.NextSong(2, playlist, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if nextIndex != 0 {
			t.Errorf("expected index 0, got %d", nextIndex)
		}
	})
	
	// 测试无效索引
	t.Run("invalid index", func(t *testing.T) {
		nextIndex, err := mode.NextSong(-1, playlist, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if nextIndex != 0 {
			t.Errorf("expected index 0, got %d", nextIndex)
		}
		
		nextIndex, err = mode.NextSong(10, playlist, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if nextIndex != 0 {
			t.Errorf("expected index 0, got %d", nextIndex)
		}
	})
	
	// 测试手动和自动切换（列表循环模式下行为相同）
	t.Run("manual vs automatic", func(t *testing.T) {
		manualNext, err1 := mode.NextSong(1, playlist, true)
		autoNext, err2 := mode.NextSong(1, playlist, false)
		
		if err1 != nil || err2 != nil {
			t.Errorf("unexpected errors: %v, %v", err1, err2)
		}
		if manualNext != autoNext {
			t.Errorf("manual and auto should be same in list loop mode, got %d vs %d", manualNext, autoNext)
		}
	})
}

func TestListLoopPlayMode_PreviousSong(t *testing.T) {
	mode := NewListLoopPlayMode()
	
	// 测试空播放列表
	t.Run("empty playlist", func(t *testing.T) {
		_, err := mode.PreviousSong(0, []structs.Song{}, false)
		if err != ErrEmptyPlaylist {
			t.Errorf("expected ErrEmptyPlaylist, got %v", err)
		}
	})
	
	// 创建测试播放列表
	playlist := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
		{Id: 3, Name: "Song 3"},
	}
	
	// 测试正常情况
	t.Run("normal previous", func(t *testing.T) {
		prevIndex, err := mode.PreviousSong(2, playlist, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if prevIndex != 1 {
			t.Errorf("expected index 1, got %d", prevIndex)
		}
	})
	
	// 测试循环到末尾
	t.Run("loop to end", func(t *testing.T) {
		prevIndex, err := mode.PreviousSong(0, playlist, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if prevIndex != 2 {
			t.Errorf("expected index 2, got %d", prevIndex)
		}
	})
	
	// 测试无效索引
	t.Run("invalid index", func(t *testing.T) {
		prevIndex, err := mode.PreviousSong(-1, playlist, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if prevIndex != 2 {
			t.Errorf("expected index 2, got %d", prevIndex)
		}
		
		prevIndex, err = mode.PreviousSong(10, playlist, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if prevIndex != 2 {
			t.Errorf("expected index 2, got %d", prevIndex)
		}
	})
}

func TestListLoopPlayMode_GetMode(t *testing.T) {
	mode := NewListLoopPlayMode()
	if mode.GetMode() != types.PmListLoop {
		t.Errorf("expected PmListLoop, got %v", mode.GetMode())
	}
}

func TestListLoopPlayMode_GetModeName(t *testing.T) {
	mode := NewListLoopPlayMode()
	expected := "列表循环"
	if mode.GetModeName() != expected {
		t.Errorf("expected %s, got %s", expected, mode.GetModeName())
	}
}

func TestListLoopPlayMode_Initialize(t *testing.T) {
	mode := NewListLoopPlayMode()
	playlist := []structs.Song{
		{Id: 1, Name: "Song 1"},
	}
	
	err := mode.Initialize(0, playlist)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestListLoopPlayMode_OnPlaylistChanged(t *testing.T) {
	mode := NewListLoopPlayMode()
	playlist := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
	}
	
	err := mode.OnPlaylistChanged(0, playlist)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}