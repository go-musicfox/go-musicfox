package playlist

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func TestInfiniteRandomPlayMode_GetMode(t *testing.T) {
	mode := NewInfiniteRandomPlayMode()
	if mode.GetMode() != types.PmInfRandom {
		t.Errorf("Expected mode %v, got %v", types.PmInfRandom, mode.GetMode())
	}
}

func TestInfiniteRandomPlayMode_GetModeName(t *testing.T) {
	mode := NewInfiniteRandomPlayMode()
	expected := "无限随机"
	if mode.GetModeName() != expected {
		t.Errorf("Expected mode name %s, got %s", expected, mode.GetModeName())
	}
}

func TestInfiniteRandomPlayMode_Initialize(t *testing.T) {
	tests := []struct {
		name         string
		currentIndex int
		playlist     []structs.Song
		wantErr      bool
	}{
		{
			name:         "empty playlist",
			currentIndex: 0,
			playlist:     []structs.Song{},
			wantErr:      true,
		},
		{
			name:         "valid playlist",
			currentIndex: 1,
			playlist:     createTestPlaylist(5),
			wantErr:      false,
		},
		{
			name:         "invalid current index",
			currentIndex: 10,
			playlist:     createTestPlaylist(3),
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := NewInfiniteRandomPlayMode()
			err := mode.Initialize(tt.currentIndex, tt.playlist)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInfiniteRandomPlayMode_NextSong(t *testing.T) {
	tests := []struct {
		name         string
		currentIndex int
		playlist     []structs.Song
		manual       bool
		wantErr      bool
	}{
		{
			name:         "empty playlist",
			currentIndex: 0,
			playlist:     []structs.Song{},
			manual:       false,
			wantErr:      true,
		},
		{
			name:         "valid playlist",
			currentIndex: 0,
			playlist:     createTestPlaylist(5),
			manual:       false,
			wantErr:      false,
		},
		{
			name:         "single song playlist",
			currentIndex: 0,
			playlist:     createTestPlaylist(1),
			manual:       false,
			wantErr:      false, // 无限随机可以重复播放同一首歌
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := NewInfiniteRandomPlayMode()
			if len(tt.playlist) > 0 {
				_ = mode.Initialize(tt.currentIndex, tt.playlist)
			}

			nextIndex, err := mode.NextSong(tt.currentIndex, tt.playlist, tt.manual)
			if (err != nil) != tt.wantErr {
				t.Errorf("NextSong() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if nextIndex < 0 || nextIndex >= len(tt.playlist) {
					t.Errorf("NextSong() returned invalid index %d for playlist length %d", nextIndex, len(tt.playlist))
				}
			}
		})
	}
}

func TestInfiniteRandomPlayMode_PreviousSong(t *testing.T) {
	tests := []struct {
		name         string
		currentIndex int
		playlist     []structs.Song
		manual       bool
		wantErr      bool
		setupHistory bool
	}{
		{
			name:         "empty playlist",
			currentIndex: 0,
			playlist:     []structs.Song{},
			manual:       false,
			wantErr:      true,
			setupHistory: false,
		},
		{
			name:         "no history",
			currentIndex: 0,
			playlist:     createTestPlaylist(5),
			manual:       false,
			wantErr:      true,
			setupHistory: false,
		},
		{
			name:         "with history",
			currentIndex: 2,
			playlist:     createTestPlaylist(5),
			manual:       false,
			wantErr:      false,
			setupHistory: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := NewInfiniteRandomPlayMode().(*InfiniteRandomPlayMode)
			if len(tt.playlist) > 0 {
				_ = mode.Initialize(tt.currentIndex, tt.playlist)
				if tt.setupHistory {
					// 添加一些历史记录
					mode.addToHistory(0)
					mode.addToHistory(1)
				}
			}

			prevIndex, err := mode.PreviousSong(tt.currentIndex, tt.playlist, tt.manual)
			if (err != nil) != tt.wantErr {
				t.Errorf("PreviousSong() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if prevIndex < 0 || prevIndex >= len(tt.playlist) {
					t.Errorf("PreviousSong() returned invalid index %d for playlist length %d", prevIndex, len(tt.playlist))
				}
			}
		})
	}
}

func TestInfiniteRandomPlayMode_OnPlaylistChanged(t *testing.T) {
	tests := []struct {
		name         string
		currentIndex int
		playlist     []structs.Song
		wantErr      bool
	}{
		{
			name:         "empty playlist",
			currentIndex: 0,
			playlist:     []structs.Song{},
			wantErr:      false,
		},
		{
			name:         "valid playlist",
			currentIndex: 1,
			playlist:     createTestPlaylist(5),
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := NewInfiniteRandomPlayMode()
			err := mode.OnPlaylistChanged(tt.currentIndex, tt.playlist)
			if (err != nil) != tt.wantErr {
				t.Errorf("OnPlaylistChanged() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInfiniteRandomPlayMode_HistoryManagement(t *testing.T) {
	// 测试历史管理功能
	playlist := createTestPlaylist(5)
	mode := NewInfiniteRandomPlayMode().(*InfiniteRandomPlayMode)
	
	// 初始化
	err := mode.Initialize(0, playlist)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 验证初始历史记录
	if len(mode.history) != 1 || mode.history[0] != 0 {
		t.Errorf("Expected initial history [0], got %v", mode.history)
	}
	if mode.currentPos != 0 {
		t.Errorf("Expected initial currentPos 0, got %d", mode.currentPos)
	}

	// 生成几首下一首歌曲
	currentIndex := 0
	for i := 0; i < 3; i++ {
		nextIndex, err := mode.NextSong(currentIndex, playlist, false)
		if err != nil {
			t.Fatalf("NextSong failed: %v", err)
		}
		currentIndex = nextIndex
	}

	// 验证历史记录增长
	if len(mode.history) != 4 { // 初始 + 3次NextSong
		t.Errorf("Expected history length 4, got %d", len(mode.history))
	}
	if mode.currentPos != 3 {
		t.Errorf("Expected currentPos 3, got %d", mode.currentPos)
	}

	// 测试返回上一首
	prevIndex, err := mode.PreviousSong(currentIndex, playlist, false)
	if err != nil {
		t.Fatalf("PreviousSong failed: %v", err)
	}
	if prevIndex != mode.history[2] {
		t.Errorf("Expected previous index %d, got %d", mode.history[2], prevIndex)
	}
	if mode.currentPos != 2 {
		t.Errorf("Expected currentPos 2 after PreviousSong, got %d", mode.currentPos)
	}

	// 测试在历史中导航后再次NextSong
	nextIndex, err := mode.NextSong(prevIndex, playlist, false)
	if err != nil {
		t.Fatalf("NextSong after PreviousSong failed: %v", err)
	}
	if nextIndex != mode.history[3] {
		t.Errorf("Expected next index from history %d, got %d", mode.history[3], nextIndex)
	}
	if mode.currentPos != 3 {
		t.Errorf("Expected currentPos 3 after NextSong in history, got %d", mode.currentPos)
	}
}

func TestInfiniteRandomPlayMode_AvoidRecentlyPlayed(t *testing.T) {
	// 测试避免重复播放最近播放的歌曲
	playlist := createTestPlaylist(10) // 足够大的播放列表
	mode := NewInfiniteRandomPlayMode().(*InfiniteRandomPlayMode)
	
	// 初始化
	err := mode.Initialize(0, playlist)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 生成多首歌曲，检查是否避免重复
	currentIndex := 0
	playedIndices := make(map[int]int) // index -> count
	playedIndices[currentIndex] = 1

	for i := 0; i < 20; i++ {
		nextIndex, err := mode.NextSong(currentIndex, playlist, false)
		if err != nil {
			t.Fatalf("NextSong failed at iteration %d: %v", i, err)
		}
		
		playedIndices[nextIndex]++
		currentIndex = nextIndex
	}

	// 验证随机性：不应该有太多重复
	uniqueCount := len(playedIndices)
	if uniqueCount < len(playlist)/2 {
		t.Errorf("Expected at least %d unique songs, got %d", len(playlist)/2, uniqueCount)
	}

	// 验证没有歌曲被过度重复播放
	for index, count := range playedIndices {
		if count > 5 { // 允许一定程度的重复，但不应该过度
			t.Errorf("Song at index %d was played %d times, which seems excessive", index, count)
		}
	}
}

func TestInfiniteRandomPlayMode_PlaylistChangedHistoryCleanup(t *testing.T) {
	// 测试播放列表变化时历史记录的清理
	mode := NewInfiniteRandomPlayMode().(*InfiniteRandomPlayMode)
	
	// 初始播放列表
	initialPlaylist := createTestPlaylist(5)
	err := mode.Initialize(0, initialPlaylist)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 添加一些历史记录
	mode.addToHistory(1)
	mode.addToHistory(2)
	mode.addToHistory(4)

	initialHistoryLen := len(mode.history)

	// 缩小播放列表（移除一些歌曲）
	newPlaylist := createTestPlaylist(3) // 只保留前3首歌
	err = mode.OnPlaylistChanged(1, newPlaylist)
	if err != nil {
		t.Fatalf("OnPlaylistChanged failed: %v", err)
	}

	// 验证无效的历史记录被清理
	for _, index := range mode.history {
		if index < 0 || index >= len(newPlaylist) {
			t.Errorf("Invalid index %d in history after playlist change", index)
		}
	}

	// 验证历史记录长度减少（因为索引4被移除）
	if len(mode.history) >= initialHistoryLen {
		t.Errorf("Expected history length to decrease after playlist change, was %d, now %d", initialHistoryLen, len(mode.history))
	}

	// 验证当前位置被正确调整
	if mode.currentPos >= len(mode.history) {
		t.Errorf("currentPos %d is out of range for history length %d", mode.currentPos, len(mode.history))
	}
}

func TestInfiniteRandomPlayMode_MaxHistoryLimit(t *testing.T) {
	// 测试历史记录长度限制
	playlist := createTestPlaylist(5)
	mode := NewInfiniteRandomPlayMode().(*InfiniteRandomPlayMode)
	mode.maxHistory = 3 // 设置较小的历史限制用于测试
	
	// 初始化
	err := mode.Initialize(0, playlist)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 添加超过限制的历史记录
	for i := 1; i < 6; i++ {
		mode.addToHistory(i % len(playlist))
	}

	// 验证历史记录长度不超过限制
	if len(mode.history) > mode.maxHistory {
		t.Errorf("History length %d exceeds max limit %d", len(mode.history), mode.maxHistory)
	}

	// 验证当前位置仍然有效
	if mode.currentPos < 0 || mode.currentPos >= len(mode.history) {
		t.Errorf("currentPos %d is out of range for history length %d", mode.currentPos, len(mode.history))
	}
}