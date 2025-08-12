package playlist

import (
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// TestPlaylistManagerIntegration 端到端集成测试
// 验证完整的播放流程和所有播放模式的正确性
func TestPlaylistManagerIntegration(t *testing.T) {
	// 创建测试播放列表
	songs := []structs.Song{
		{Id: 1, Name: "Song 1", Duration: time.Minute * 3},
		{Id: 2, Name: "Song 2", Duration: time.Minute * 4},
		{Id: 3, Name: "Song 3", Duration: time.Minute * 2},
		{Id: 4, Name: "Song 4", Duration: time.Minute * 5},
		{Id: 5, Name: "Song 5", Duration: time.Minute * 3},
	}

	manager := NewPlaylistManager()

	// 测试初始化
	err := manager.Initialize(2, songs)
	if err != nil {
		t.Fatalf("Failed to initialize playlist: %v", err)
	}

	// 验证初始状态
	currentSong, err := manager.GetCurrentSong()
	if err != nil {
		t.Fatalf("Failed to get current song: %v", err)
	}
	if currentSong.Id != 3 {
		t.Errorf("Expected current song ID to be 3, got %d", currentSong.Id)
	}

	// 测试所有播放模式的完整流程
	t.Run("OrderedPlayMode", func(t *testing.T) {
		testOrderedPlayModeIntegration(t, manager, songs)
	})

	t.Run("ListLoopPlayMode", func(t *testing.T) {
		testListLoopPlayModeIntegration(t, manager, songs)
	})

	t.Run("SingleLoopPlayMode", func(t *testing.T) {
		testSingleLoopPlayModeIntegration(t, manager, songs)
	})

	t.Run("ListRandomPlayMode", func(t *testing.T) {
		testListRandomPlayModeIntegration(t, manager, songs)
	})

	t.Run("InfiniteRandomPlayMode", func(t *testing.T) {
		testInfiniteRandomPlayModeIntegration(t, manager, songs)
	})
}

// testOrderedPlayModeIntegration 测试顺序播放模式的完整流程
func testOrderedPlayModeIntegration(t *testing.T, manager PlaylistManager, songs []structs.Song) {
	err := manager.SetPlayMode(types.PmOrdered)
	if err != nil {
		t.Fatalf("Failed to set ordered play mode: %v", err)
	}

	// 重新初始化到第一首歌
	err = manager.Initialize(0, songs)
	if err != nil {
		t.Fatalf("Failed to reinitialize: %v", err)
	}

	// 验证顺序播放：依次播放每首歌
	for i := 0; i < len(songs)-1; i++ {
		nextSong, err := manager.NextSong(false)
		if err != nil {
			t.Fatalf("Failed to get next song at index %d: %v", i, err)
		}
		expectedId := songs[i+1].Id
		if nextSong.Id != expectedId {
			t.Errorf("Expected song ID %d, got %d at position %d", expectedId, nextSong.Id, i+1)
		}
	}

	// 验证到达末尾后无法继续
	_, err = manager.NextSong(false)
	if err == nil {
		t.Error("Expected error when reaching end of playlist in ordered mode")
	}

	// 验证反向播放
	for i := len(songs) - 1; i > 0; i-- {
		prevSong, err := manager.PreviousSong(false)
		if err != nil {
			t.Fatalf("Failed to get previous song at index %d: %v", i, err)
		}
		expectedId := songs[i-1].Id
		if prevSong.Id != expectedId {
			t.Errorf("Expected song ID %d, got %d at position %d", expectedId, prevSong.Id, i-1)
		}
	}
}

// testListLoopPlayModeIntegration 测试列表循环播放模式的完整流程
func testListLoopPlayModeIntegration(t *testing.T, manager PlaylistManager, songs []structs.Song) {
	err := manager.SetPlayMode(types.PmListLoop)
	if err != nil {
		t.Fatalf("Failed to set list loop play mode: %v", err)
	}

	// 重新初始化到最后一首歌
	err = manager.Initialize(len(songs)-1, songs)
	if err != nil {
		t.Fatalf("Failed to reinitialize: %v", err)
	}

	// 验证从最后一首循环到第一首
	nextSong, err := manager.NextSong(false)
	if err != nil {
		t.Fatalf("Failed to get next song in loop mode: %v", err)
	}
	if nextSong.Id != songs[0].Id {
		t.Errorf("Expected to loop to first song (ID %d), got %d", songs[0].Id, nextSong.Id)
	}

	// 验证从第一首反向循环到最后一首
	prevSong, err := manager.PreviousSong(false)
	if err != nil {
		t.Fatalf("Failed to get previous song in loop mode: %v", err)
	}
	if prevSong.Id != songs[len(songs)-1].Id {
		t.Errorf("Expected to loop to last song (ID %d), got %d", songs[len(songs)-1].Id, prevSong.Id)
	}

	// 验证完整循环
	for i := 0; i < len(songs)*2; i++ {
		_, err := manager.NextSong(false)
		if err != nil {
			t.Fatalf("Failed during complete loop at iteration %d: %v", i, err)
		}
	}
}

// testSingleLoopPlayModeIntegration 测试单曲循环播放模式的完整流程
func testSingleLoopPlayModeIntegration(t *testing.T, manager PlaylistManager, songs []structs.Song) {
	err := manager.SetPlayMode(types.PmSingleLoop)
	if err != nil {
		t.Fatalf("Failed to set single loop play mode: %v", err)
	}

	// 初始化到中间的歌曲
	midIndex := len(songs) / 2
	err = manager.Initialize(midIndex, songs)
	if err != nil {
		t.Fatalf("Failed to reinitialize: %v", err)
	}

	expectedSong := songs[midIndex]

	// 验证自动播放时重复当前歌曲
	for i := 0; i < 5; i++ {
		nextSong, err := manager.NextSong(false) // 自动播放
		if err != nil {
			t.Fatalf("Failed to get next song in single loop (auto) at iteration %d: %v", i, err)
		}
		if nextSong.Id != expectedSong.Id {
			t.Errorf("Expected to repeat current song (ID %d), got %d at iteration %d", expectedSong.Id, nextSong.Id, i)
		}
	}

	// 验证手动切换时正常前进
	nextSong, err := manager.NextSong(true) // 手动切换
	if err != nil {
		t.Fatalf("Failed to get next song in single loop (manual): %v", err)
	}
	expectedNextId := songs[(midIndex+1)%len(songs)].Id
	if nextSong.Id != expectedNextId {
		t.Errorf("Expected manual next to advance to song ID %d, got %d", expectedNextId, nextSong.Id)
	}

	// 验证手动反向切换
	prevSong, err := manager.PreviousSong(true) // 手动切换
	if err != nil {
		t.Fatalf("Failed to get previous song in single loop (manual): %v", err)
	}
	if prevSong.Id != expectedSong.Id {
		t.Errorf("Expected manual previous to return to song ID %d, got %d", expectedSong.Id, prevSong.Id)
	}
}

// testListRandomPlayModeIntegration 测试列表随机播放模式的完整流程
func testListRandomPlayModeIntegration(t *testing.T, manager PlaylistManager, songs []structs.Song) {
	err := manager.SetPlayMode(types.PmListRandom)
	if err != nil {
		t.Fatalf("Failed to set list random play mode: %v", err)
	}

	err = manager.Initialize(0, songs)
	if err != nil {
		t.Fatalf("Failed to reinitialize: %v", err)
	}

	// 验证随机播放不会超出播放列表范围
	// 列表随机播放模式会播放完所有歌曲后停止
	playedSongs := make(map[int64]bool)
	
	// 记录当前歌曲（初始化时的第0首）
	currentSong, err := manager.GetCurrentSong()
	if err != nil {
		t.Fatalf("Failed to get current song: %v", err)
	}
	playedSongs[currentSong.Id] = true
	
	// 尝试播放剩余的歌曲，直到到达列表末尾
	for i := 0; i < len(songs); i++ { // 最多尝试播放所有歌曲
		nextSong, err := manager.NextSong(false)
		if err != nil {
			// 如果遇到"no next song"错误，说明已经播放完所有歌曲
			if err.Error() == "playlist next song: no next song" {
				break
			}
			t.Fatalf("Unexpected error getting next song at iteration %d: %v", i, err)
		}

		// 验证歌曲在原播放列表中
		found := false
		for _, song := range songs {
			if song.Id == nextSong.Id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Random mode returned song ID %d which is not in original playlist", nextSong.Id)
		}

		playedSongs[nextSong.Id] = true
	}

	// 验证播放完所有歌曲后会返回错误
	_, err = manager.NextSong(false)
	if err == nil {
		t.Error("Expected error when trying to get next song after playlist ends in list random mode")
	}
}

// testInfiniteRandomPlayModeIntegration 测试无限随机播放模式的完整流程
func testInfiniteRandomPlayModeIntegration(t *testing.T, manager PlaylistManager, songs []structs.Song) {
	err := manager.SetPlayMode(types.PmInfRandom)
	if err != nil {
		t.Fatalf("Failed to set infinite random play mode: %v", err)
	}

	err = manager.Initialize(0, songs)
	if err != nil {
		t.Fatalf("Failed to reinitialize: %v", err)
	}

	// 验证无限随机播放可以持续进行
	for i := 0; i < len(songs)*3; i++ {
		nextSong, err := manager.NextSong(false)
		if err != nil {
			t.Fatalf("Failed to get next song in infinite random mode at iteration %d: %v", i, err)
		}

		// 验证歌曲在原播放列表中
		found := false
		for _, song := range songs {
			if song.Id == nextSong.Id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Infinite random mode returned song ID %d which is not in original playlist", nextSong.Id)
		}
	}
}

// TestPlaylistManagerWithUIInteraction 测试播放列表管理器与UI层的交互
func TestPlaylistManagerWithUIInteraction(t *testing.T) {
	// 模拟UI层的使用场景
	// 模拟用户选择播放列表
	songs := []structs.Song{
		{Id: 101, Name: "Popular Song 1", Duration: time.Minute * 3},
		{Id: 102, Name: "Popular Song 2", Duration: time.Minute * 4},
		{Id: 103, Name: "Popular Song 3", Duration: time.Minute * 2},
	}

	// 模拟用户切换播放模式
	modes := []types.Mode{
		types.PmOrdered,
		types.PmListLoop,
		types.PmSingleLoop,
		types.PmListRandom,
		types.PmInfRandom,
	}

	for _, mode := range modes {
		// 为每个模式创建新的管理器实例，避免状态污染
		manager := NewPlaylistManager()
		
		// 初始化播放列表
		err := manager.Initialize(0, songs)
		if err != nil {
			t.Fatalf("Failed to initialize playlist: %v", err)
		}
		
		err = manager.SetPlayMode(mode)
		if err != nil {
			t.Errorf("Failed to set play mode %v: %v", mode, err)
			continue
		}

		// 验证模式设置成功
		if manager.GetPlayMode() != mode {
			t.Errorf("Expected play mode %v, got %v", mode, manager.GetPlayMode())
		}

		// 验证模式名称
		modeName := manager.GetPlayModeName()
		if modeName == "" {
			t.Errorf("Play mode name should not be empty for mode %v", mode)
		}

		// 测试在该模式下的基本操作
		if mode == types.PmListRandom {
			// 列表随机播放模式需要特殊处理，因为它在播放完所有歌曲后会停止
			// 只测试一次NextSong
			_, err = manager.NextSong(true)
			if err != nil {
				t.Errorf("Failed to get next song in mode %v: %v", mode, err)
			}
		} else {
			_, err = manager.NextSong(true)
			if err != nil {
				t.Errorf("Failed to get next song in mode %v: %v", mode, err)
			}

			_, err = manager.PreviousSong(true)
			if err != nil {
				t.Errorf("Failed to get previous song in mode %v: %v", mode, err)
			}
		}
	}
	
	// 测试播放列表操作（使用最后一个管理器实例）
	manager := NewPlaylistManager()
	err := manager.Initialize(0, songs)
	if err != nil {
		t.Fatalf("Failed to initialize playlist for operations test: %v", err)
	}
	
	// 模拟用户删除歌曲
	originalLength := len(manager.GetPlaylist())
	_, err = manager.RemoveSong(1)
	if err != nil {
		t.Fatalf("Failed to remove song: %v", err)
	}

	newLength := len(manager.GetPlaylist())
	if newLength != originalLength-1 {
		t.Errorf("Expected playlist length to be %d after removal, got %d", originalLength-1, newLength)
	}
}

// TestPlaylistManagerErrorHandling 测试错误处理和边界条件
func TestPlaylistManagerErrorHandling(t *testing.T) {
	manager := NewPlaylistManager()

	// 测试空播放列表的错误处理
	t.Run("EmptyPlaylistErrors", func(t *testing.T) {
		_, err := manager.GetCurrentSong()
		if err == nil {
			t.Error("Expected error when getting current song from empty playlist")
		}

		_, err = manager.NextSong(false)
		if err == nil {
			t.Error("Expected error when getting next song from empty playlist")
		}

		_, err = manager.PreviousSong(false)
		if err == nil {
			t.Error("Expected error when getting previous song from empty playlist")
		}

		_, err = manager.RemoveSong(0)
		if err == nil {
			t.Error("Expected error when removing song from empty playlist")
		}
	})

	// 测试无效索引的错误处理
	t.Run("InvalidIndexErrors", func(t *testing.T) {
		songs := []structs.Song{
			{Id: 1, Name: "Song 1"},
			{Id: 2, Name: "Song 2"},
		}

		// 测试初始化时的无效索引
		err := manager.Initialize(-1, songs)
		if err == nil {
			t.Error("Expected error when initializing with negative index")
		}

		err = manager.Initialize(10, songs)
		if err == nil {
			t.Error("Expected error when initializing with out-of-range index")
		}

		// 正确初始化后测试删除时的无效索引
		err = manager.Initialize(0, songs)
		if err != nil {
			t.Fatalf("Failed to initialize: %v", err)
		}

		_, err = manager.RemoveSong(-1)
		if err == nil {
			t.Error("Expected error when removing song with negative index")
		}

		_, err = manager.RemoveSong(10)
		if err == nil {
			t.Error("Expected error when removing song with out-of-range index")
		}
	})

	// 测试无效播放模式的错误处理
	t.Run("InvalidPlayModeErrors", func(t *testing.T) {
		err := manager.SetPlayMode(types.PmUnknown) // 无效的播放模式
		if err == nil {
			t.Error("Expected error when setting invalid play mode")
		}
	})
}

// TestPlaylistManagerStateConsistency 测试状态一致性
func TestPlaylistManagerStateConsistency(t *testing.T) {
	manager := NewPlaylistManager()
	songs := []structs.Song{
		{Id: 1, Name: "Song 1"},
		{Id: 2, Name: "Song 2"},
		{Id: 3, Name: "Song 3"},
	}

	err := manager.Initialize(1, songs)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// 验证初始状态一致性
	currentIndex := manager.GetCurrentIndex()
	currentSong, err := manager.GetCurrentSong()
	if err != nil {
		t.Fatalf("Failed to get current song: %v", err)
	}

	playlist := manager.GetPlaylist()
	if currentIndex < 0 || currentIndex >= len(playlist) {
		t.Errorf("Current index %d is out of range for playlist length %d", currentIndex, len(playlist))
	}

	if playlist[currentIndex].Id != currentSong.Id {
		t.Errorf("Current song ID mismatch: index points to %d, but GetCurrentSong returned %d", playlist[currentIndex].Id, currentSong.Id)
	}

	// 测试切换歌曲后的状态一致性
	nextSong, err := manager.NextSong(true)
	if err != nil {
		t.Fatalf("Failed to get next song: %v", err)
	}

	newCurrentSong, err := manager.GetCurrentSong()
	if err != nil {
		t.Fatalf("Failed to get current song after next: %v", err)
	}

	if nextSong.Id != newCurrentSong.Id {
		t.Errorf("State inconsistency: NextSong returned %d, but GetCurrentSong returned %d", nextSong.Id, newCurrentSong.Id)
	}

	// 测试播放模式切换后的状态一致性
	originalMode := manager.GetPlayMode()
	err = manager.SetPlayMode(types.PmListLoop)
	if err != nil {
		t.Fatalf("Failed to set play mode: %v", err)
	}

	if manager.GetPlayMode() == originalMode {
		t.Error("Play mode did not change after SetPlayMode")
	}

	// 验证播放模式切换后当前歌曲保持不变
	currentAfterModeChange, err := manager.GetCurrentSong()
	if err != nil {
		t.Fatalf("Failed to get current song after mode change: %v", err)
	}

	if currentAfterModeChange.Id != newCurrentSong.Id {
		t.Errorf("Current song changed after mode switch: was %d, now %d", newCurrentSong.Id, currentAfterModeChange.Id)
	}
}