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

func TestAddSongsToNextPrioritizesNewestBatchWithoutAdvancingMode(t *testing.T) {
	manager := NewPlaylistManager()
	base := []structs.Song{{Id: 1}, {Id: 2}, {Id: 3}}
	if err := manager.Initialize(0, base); err != nil {
		t.Fatal(err)
	}
	if err := manager.SetPlayMode(types.PmSingleLoop); err != nil {
		t.Fatal(err)
	}
	manager.AddSongsToNext([]structs.Song{{Id: 10}, {Id: 11}})
	manager.AddSongsToNext([]structs.Song{{Id: 20}, {Id: 21}})
	for _, id := range []int64{20, 21, 10, 11, 1} {
		song, err := manager.NextSong(false)
		if err != nil || song.Id != id {
			t.Fatalf("NextSong() = %d, %v; want %d", song.Id, err, id)
		}
	}
	if got := manager.GetPlaylist(); len(got) != len(base) {
		t.Fatalf("temporary songs changed main playlist: %d songs", len(got))
	}
}

func TestTemporarySongsPrecedeEveryPlayMode(t *testing.T) {
	modes := []types.Mode{
		types.PmOrdered,
		types.PmListLoop,
		types.PmSingleLoop,
		types.PmListRandom,
		types.PmInfRandom,
		types.PmIntelligent,
	}
	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			manager := NewPlaylistManager().(*playlistManager)
			if err := manager.Initialize(0, []structs.Song{{Id: 1}, {Id: 2}}); err != nil {
				t.Fatal(err)
			}
			if err := manager.SetPlayMode(mode); err != nil {
				t.Fatal(err)
			}
			manager.AddSongsToNext([]structs.Song{{Id: 9}})
			song, err := manager.NextSong(false)
			if err != nil || song.Id != 9 {
				t.Fatalf("NextSong() = %d, %v; want inserted song 9", song.Id, err)
			}
			switch modeState := manager.playMode.(type) {
			case *ListRandomPlayMode:
				wantIndex := modeState.randomOrder[modeState.currentPos+1]
				song, err = manager.NextSong(false)
				if err != nil || song.Id != manager.playlist[wantIndex].Id {
					t.Fatalf("random continuation = %d, %v; want playlist index %d", song.Id, err, wantIndex)
				}
			case *InfiniteRandomPlayMode:
				if len(modeState.history) != 1 || modeState.currentPos != 0 {
					t.Fatalf("inserted song advanced infinite-random history: %#v", modeState.history)
				}
				if _, err = manager.NextSong(false); err != nil {
					t.Fatalf("infinite-random continuation: %v", err)
				}
				if len(modeState.history) != 2 || modeState.currentPos != 1 {
					t.Fatalf("normal continuation did not advance history once: %#v", modeState.history)
				}
			default:
				song, err = manager.NextSong(false)
				wantID := int64(2)
				if mode == types.PmSingleLoop {
					wantID = 1
				}
				if err != nil || song.Id != wantID {
					t.Fatalf("mode continuation = %d, %v; want %d", song.Id, err, wantID)
				}
			}
		})
	}
}

func TestInsertedSongWithSameIDIsConsumedBeforeModeAdvance(t *testing.T) {
	manager := NewPlaylistManager()
	if err := manager.Initialize(0, []structs.Song{{Id: 1}, {Id: 2}}); err != nil {
		t.Fatal(err)
	}
	manager.AddSongsToNext([]structs.Song{{Id: 1}})
	if song, err := manager.NextSong(false); err != nil || song.Id != 1 {
		t.Fatalf("inserted NextSong() = %d, %v; want same-ID insert 1", song.Id, err)
	}
	if _, playing := manager.InsertedSongsState(); !playing {
		t.Fatal("same-ID inserted song was not marked as playing")
	}
	if song, err := manager.NextSong(false); err != nil || song.Id != 2 {
		t.Fatalf("mode NextSong() = %d, %v; want main-list successor 2", song.Id, err)
	}
	if _, playing := manager.InsertedSongsState(); playing {
		t.Fatal("inserted playback state remained after returning to main list")
	}
}

func TestAddingBatchDuringInsertedPlaybackKeepsCurrentAndPrioritizesNewBatch(t *testing.T) {
	manager := NewPlaylistManager()
	if err := manager.Initialize(0, []structs.Song{{Id: 1}, {Id: 2}}); err != nil {
		t.Fatal(err)
	}
	manager.AddSongsToNext([]structs.Song{{Id: 10}, {Id: 11}})
	if song, err := manager.NextSong(false); err != nil || song.Id != 10 {
		t.Fatalf("NextSong() = %d, %v; want 10", song.Id, err)
	}
	manager.AddSongsToNext([]structs.Song{{Id: 20}, {Id: 21}})
	if song, err := manager.GetCurrentSong(); err != nil || song.Id != 10 {
		t.Fatalf("current song after enqueue = %d, %v; want still-playing 10", song.Id, err)
	}
	for _, want := range []int64{20, 21, 11} {
		song, err := manager.NextSong(false)
		if err != nil || song.Id != want {
			t.Fatalf("NextSong() = %d, %v; want %d", song.Id, err, want)
		}
	}
}

func TestFailedModeAdvanceKeepsInsertedCurrentSong(t *testing.T) {
	manager := NewPlaylistManager()
	if err := manager.Initialize(0, []structs.Song{{Id: 1}}); err != nil {
		t.Fatal(err)
	}
	manager.AddSongsToNext([]structs.Song{{Id: 9}})
	if song, err := manager.NextSong(false); err != nil || song.Id != 9 {
		t.Fatalf("NextSong() = %d, %v; want inserted song 9", song.Id, err)
	}
	if _, err := manager.NextSong(false); err == nil {
		t.Fatal("expected ordered mode to have no next song")
	}
	if song, err := manager.GetCurrentSong(); err != nil || song.Id != 9 {
		t.Fatalf("GetCurrentSong() = %d, %v; want still-playing inserted song 9", song.Id, err)
	}
}

func TestPreviousSongAndRemoveSongClearInsertedCurrentSong(t *testing.T) {
	manager := NewPlaylistManager()
	if err := manager.Initialize(1, []structs.Song{{Id: 1}, {Id: 2}, {Id: 3}}); err != nil {
		t.Fatal(err)
	}
	manager.AddSongsToNext([]structs.Song{{Id: 9}})
	_, _ = manager.NextSong(false)
	if song, err := manager.PreviousSong(true); err != nil || song.Id != 1 {
		t.Fatalf("PreviousSong() = %d, %v; want 1", song.Id, err)
	}
	if song, err := manager.GetCurrentSong(); err != nil || song.Id != 1 {
		t.Fatalf("GetCurrentSong() = %d, %v; want 1", song.Id, err)
	}

	manager.AddSongsToNext([]structs.Song{{Id: 9}})
	_, _ = manager.NextSong(false)
	if song, err := manager.RemoveSong(2); err != nil || song.Id != 1 {
		t.Fatalf("RemoveSong() = %d, %v; want 1", song.Id, err)
	}
	if song, err := manager.GetCurrentSong(); err != nil || song.Id != 1 {
		t.Fatalf("GetCurrentSong() after removal = %d, %v; want 1", song.Id, err)
	}
}

func TestRemovingLastMainSongClearsTemporarySongs(t *testing.T) {
	manager := NewPlaylistManager()
	if err := manager.Initialize(0, []structs.Song{{Id: 1}}); err != nil {
		t.Fatal(err)
	}
	manager.AddSongsToNext([]structs.Song{{Id: 9}})
	if _, err := manager.RemoveSong(0); err == nil {
		t.Fatal("expected removing the last main song to return an error")
	}
	queued, playing := manager.InsertedSongsState()
	if queued || playing {
		t.Fatalf("temporary state remained after removing the playlist: queued=%v playing=%v", queued, playing)
	}
}

func TestInitializeClearsTemporarySongs(t *testing.T) {
	manager := NewPlaylistManager()
	base := []structs.Song{{Id: 1}, {Id: 2}}
	if err := manager.Initialize(0, base); err != nil {
		t.Fatal(err)
	}
	manager.AddSongsToNext([]structs.Song{{Id: 10}})
	if err := manager.Initialize(0, base); err != nil {
		t.Fatal(err)
	}
	song, err := manager.NextSong(false)
	if err != nil || song.Id != 2 {
		t.Fatalf("NextSong() after reset = %d, %v; want 2", song.Id, err)
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