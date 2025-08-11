package playlist

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func TestSingleLoopPlayMode_NextSong(t *testing.T) {
	mode := NewSingleLoopPlayMode()
	
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
	
	// 测试自动播放（应该返回当前索引）
	t.Run("automatic next - repeat current", func(t *testing.T) {
		nextIndex, err := mode.NextSong(1, playlist, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if nextIndex != 1 {
			t.Errorf("expected index 1 (current), got %d", nextIndex)
		}
	})
	
	// 测试手动切换（应该正常切换到下一首）
	t.Run("manual next - normal progression", func(t *testing.T) {
		nextIndex, err := mode.NextSong(1, playlist, true)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if nextIndex != 2 {
			t.Errorf("expected index 2, got %d", nextIndex)
		}
	})
	
	// 测试手动切换循环到开头
	t.Run("manual next - loop to beginning", func(t *testing.T) {
		nextIndex, err := mode.NextSong(2, playlist, true)
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
		
		nextIndex, err = mode.NextSong(10, playlist, true)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if nextIndex != 0 {
			t.Errorf("expected index 0, got %d", nextIndex)
		}
	})
}

func TestSingleLoopPlayMode_PreviousSong(t *testing.T) {
	mode := NewSingleLoopPlayMode()
	
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
	
	// 测试自动播放（应该返回当前索引）
	t.Run("automatic previous - repeat current", func(t *testing.T) {
		prevIndex, err := mode.PreviousSong(1, playlist, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if prevIndex != 1 {
			t.Errorf("expected index 1 (current), got %d", prevIndex)
		}
	})
	
	// 测试手动切换（应该正常切换到上一首）
	t.Run("manual previous - normal progression", func(t *testing.T) {
		prevIndex, err := mode.PreviousSong(2, playlist, true)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if prevIndex != 1 {
			t.Errorf("expected index 1, got %d", prevIndex)
		}
	})
	
	// 测试手动切换循环到末尾
	t.Run("manual previous - loop to end", func(t *testing.T) {
		prevIndex, err := mode.PreviousSong(0, playlist, true)
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
		
		prevIndex, err = mode.PreviousSong(10, playlist, true)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if prevIndex != 2 {
			t.Errorf("expected index 2, got %d", prevIndex)
		}
	})
}

func TestSingleLoopPlayMode_ManualVsAutomatic(t *testing.T) {
	mode := NewSingleLoopPlayMode()
	playlist := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
		{Id: 3, Name: "Song 3"},
	}
	
	// 测试手动和自动切换的区别
	t.Run("manual vs automatic behavior", func(t *testing.T) {
		// 自动播放应该重复当前歌曲
		autoNext, err1 := mode.NextSong(1, playlist, false)
		autoPrev, err2 := mode.PreviousSong(1, playlist, false)
		
		if err1 != nil || err2 != nil {
			t.Errorf("unexpected errors: %v, %v", err1, err2)
		}
		if autoNext != 1 || autoPrev != 1 {
			t.Errorf("automatic should repeat current (1), got next=%d, prev=%d", autoNext, autoPrev)
		}
		
		// 手动切换应该正常切换
		manualNext, err3 := mode.NextSong(1, playlist, true)
		manualPrev, err4 := mode.PreviousSong(1, playlist, true)
		
		if err3 != nil || err4 != nil {
			t.Errorf("unexpected errors: %v, %v", err3, err4)
		}
		if manualNext != 2 || manualPrev != 0 {
			t.Errorf("manual should progress normally, got next=%d (expected 2), prev=%d (expected 0)", manualNext, manualPrev)
		}
	})
}

func TestSingleLoopPlayMode_GetMode(t *testing.T) {
	mode := NewSingleLoopPlayMode()
	if mode.GetMode() != types.PmSingleLoop {
		t.Errorf("expected PmSingleLoop, got %v", mode.GetMode())
	}
}

func TestSingleLoopPlayMode_GetModeName(t *testing.T) {
	mode := NewSingleLoopPlayMode()
	expected := "单曲循环"
	if mode.GetModeName() != expected {
		t.Errorf("expected %s, got %s", expected, mode.GetModeName())
	}
}

func TestSingleLoopPlayMode_Initialize(t *testing.T) {
	mode := NewSingleLoopPlayMode()
	playlist := []structs.Song{
		{Id: 1, Name: "Song 1"},
	}
	
	err := mode.Initialize(0, playlist)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSingleLoopPlayMode_OnPlaylistChanged(t *testing.T) {
	mode := NewSingleLoopPlayMode()
	playlist := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
	}
	
	err := mode.OnPlaylistChanged(0, playlist)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}