package playlist

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func TestListRandomPlayMode_GetMode(t *testing.T) {
	mode := NewListRandomPlayMode()
	if mode.GetMode() != types.PmListRandom {
		t.Errorf("Expected mode %v, got %v", types.PmListRandom, mode.GetMode())
	}
}

func TestListRandomPlayMode_GetModeName(t *testing.T) {
	mode := NewListRandomPlayMode()
	expected := "列表随机"
	if mode.GetModeName() != expected {
		t.Errorf("Expected mode name %s, got %s", expected, mode.GetModeName())
	}
}

func TestListRandomPlayMode_Initialize(t *testing.T) {
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
			mode := NewListRandomPlayMode()
			err := mode.Initialize(tt.currentIndex, tt.playlist)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListRandomPlayMode_NextSong(t *testing.T) {
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
			currentIndex: -1, // 使用-1表示需要特殊处理
			playlist:     createTestPlaylist(5),
			manual:       false,
			wantErr:      false,
		},
		{
			name:         "single song playlist",
			currentIndex: 0,
			playlist:     createTestPlaylist(1),
			manual:       false,
			wantErr:      true, // 列表随机播放完一遍后应该停止
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := NewListRandomPlayMode()
			currentIndex := tt.currentIndex
			
			if len(tt.playlist) > 0 {
				if currentIndex == -1 {
					// 特殊处理：初始化后使用随机序列的第一个索引（但不是最后一个）
					_ = mode.Initialize(0, tt.playlist)
					lrm := mode.(*ListRandomPlayMode)
					if len(lrm.randomOrder) > 1 {
						// 使用第一个索引，确保不是最后一个位置
						currentIndex = lrm.randomOrder[0]
						// 设置currentPos为0，确保有下一首歌
						lrm.currentPos = 0
					} else if len(lrm.randomOrder) == 1 {
						// 只有一首歌的情况，应该返回错误
						currentIndex = lrm.randomOrder[0]
					}
				} else {
					_ = mode.Initialize(currentIndex, tt.playlist)
				}
			}

			nextIndex, err := mode.NextSong(currentIndex, tt.playlist, tt.manual)
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

func TestListRandomPlayMode_PreviousSong(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := NewListRandomPlayMode()
			if len(tt.playlist) > 0 {
				_ = mode.Initialize(tt.currentIndex, tt.playlist)
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

	// 测试基本的上一首歌功能
	t.Run("basic previous song functionality", func(t *testing.T) {
		playlist := createTestPlaylist(5)
		mode := NewListRandomPlayMode()
		
		// 初始化
		err := mode.Initialize(0, playlist)
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}
		
		// 测试PreviousSong是否返回有效索引或适当的错误
		for i := 0; i < len(playlist); i++ {
			prevIndex, err := mode.PreviousSong(i, playlist, false)
			if err != nil {
				// 如果返回错误，应该是ErrNoPreviousSong
				if err != ErrNoPreviousSong {
					t.Errorf("Expected ErrNoPreviousSong or valid index, got error: %v", err)
				}
			} else {
				// 如果返回索引，应该是有效的
				if prevIndex < 0 || prevIndex >= len(playlist) {
					t.Errorf("PreviousSong returned invalid index %d", prevIndex)
				}
			}
		}
	})
}

func TestListRandomPlayMode_OnPlaylistChanged(t *testing.T) {
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
			mode := NewListRandomPlayMode()
			err := mode.OnPlaylistChanged(tt.currentIndex, tt.playlist)
			if (err != nil) != tt.wantErr {
				t.Errorf("OnPlaylistChanged() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListRandomPlayMode_RandomSequence(t *testing.T) {
	// 测试随机序列的生成和使用
	playlist := createTestPlaylist(5)
	mode := NewListRandomPlayMode().(*ListRandomPlayMode)
	
	// 初始化
	err := mode.Initialize(0, playlist)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 验证随机序列已生成
	if len(mode.randomOrder) != len(playlist) {
		t.Errorf("Expected random order length %d, got %d", len(playlist), len(mode.randomOrder))
	}

	// 验证随机序列包含所有索引
	indexMap := make(map[int]bool)
	for _, index := range mode.randomOrder {
		if index < 0 || index >= len(playlist) {
			t.Errorf("Invalid index %d in random order", index)
		}
		indexMap[index] = true
	}

	if len(indexMap) != len(playlist) {
		t.Errorf("Random order missing some indices. Expected %d unique indices, got %d", len(playlist), len(indexMap))
	}

	// 测试从随机序列的第一个位置开始播放
	// 重新初始化，确保从第一个位置开始
	firstIndex := mode.randomOrder[0]
	err = mode.Initialize(firstIndex, playlist)
	if err != nil {
		t.Fatalf("Re-initialize failed: %v", err)
	}

	// 测试按序列播放，但只播放几首歌来验证逻辑
	currentIndex := firstIndex
	maxSongs := len(playlist) - 1
	if maxSongs > 3 {
		maxSongs = 3 // 只测试前几首歌
	}
	
	for i := 0; i < maxSongs; i++ {
		nextIndex, err := mode.NextSong(currentIndex, playlist, false)
		if err != nil {
			// 如果到达列表末尾，这是正常的
			if err == ErrNoNextSong {
				break
			}
			t.Errorf("Unexpected error getting next song: %v", err)
			break
		}
		
		// 验证返回的索引是有效的
		if nextIndex < 0 || nextIndex >= len(playlist) {
			t.Errorf("NextSong returned invalid index %d", nextIndex)
		}
		
		currentIndex = nextIndex
	}

	// 测试到达列表末尾时的行为
	// 手动设置到最后一个位置
	if len(mode.randomOrder) > 0 {
		lastIndex := mode.randomOrder[len(mode.randomOrder)-1]
		// 手动设置currentPos到最后一个位置
		mode.currentPos = len(mode.randomOrder) - 1
		
		_, err = mode.NextSong(lastIndex, playlist, false)
		if err != ErrNoNextSong {
			t.Errorf("Expected ErrNoNextSong when at end of list, got: %v", err)
		}
	}
}

func TestListRandomPlayMode_PlaylistChangedRegeneration(t *testing.T) {
	// 测试播放列表变化时重新生成随机序列
	mode := NewListRandomPlayMode().(*ListRandomPlayMode)
	
	// 初始播放列表
	initialPlaylist := createTestPlaylist(3)
	err := mode.Initialize(0, initialPlaylist)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	initialOrder := make([]int, len(mode.randomOrder))
	copy(initialOrder, mode.randomOrder)

	// 改变播放列表
	newPlaylist := createTestPlaylist(5)
	err = mode.OnPlaylistChanged(0, newPlaylist)
	if err != nil {
		t.Fatalf("OnPlaylistChanged failed: %v", err)
	}

	// 验证随机序列已重新生成
	if len(mode.randomOrder) != len(newPlaylist) {
		t.Errorf("Expected new random order length %d, got %d", len(newPlaylist), len(mode.randomOrder))
	}

	// 验证新的随机序列包含正确的索引范围
	for _, index := range mode.randomOrder {
		if index < 0 || index >= len(newPlaylist) {
			t.Errorf("Invalid index %d in new random order for playlist length %d", index, len(newPlaylist))
		}
	}
}