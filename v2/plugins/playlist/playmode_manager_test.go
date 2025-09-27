package playlist

import (
	"context"
	"testing"

	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
)

// TestSetPlayMode 测试设置播放模式
func TestSetPlayMode(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 测试设置顺序播放模式
	err := plugin.SetPlayMode(ctx, model.PlayModeSequential)
	if err != nil {
		t.Errorf("Failed to set sequential play mode: %v", err)
	}
	
	mode := plugin.GetPlayMode(ctx)
	if mode != model.PlayModeSequential {
		t.Errorf("Expected sequential mode, got %v", mode)
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "player.mode_changed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected player.mode_changed event to be published")
	}
	
	mockEventBus.ClearEvents()
	
	// 测试设置随机播放模式
	err = plugin.SetPlayMode(ctx, model.PlayModeShuffle)
	if err != nil {
		t.Errorf("Failed to set shuffle play mode: %v", err)
	}
	
	mode = plugin.GetPlayMode(ctx)
	if mode != model.PlayModeShuffle {
		t.Errorf("Expected shuffle mode, got %v", mode)
	}
	
	// 检查随机索引是否生成（需要先设置队列）
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	// 重新设置随机模式以触发随机索引生成
	err = plugin.SetPlayMode(ctx, model.PlayModeShuffle)
	if err != nil {
		t.Errorf("Failed to set shuffle mode with queue: %v", err)
	}
	
	plugin.mu.RLock()
	shuffleIndexExists := plugin.shuffleIndex != nil
	plugin.mu.RUnlock()
	
	if !shuffleIndexExists {
		t.Error("Expected shuffle index to be generated when setting shuffle mode")
	}
}

// TestGetPlayMode 测试获取播放模式
func TestGetPlayMode(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 默认应该是顺序播放
	mode := plugin.GetPlayMode(ctx)
	if mode != model.PlayModeSequential {
		t.Errorf("Expected default mode to be sequential, got %v", mode)
	}
	
	// 设置不同模式并验证
	modes := []model.PlayMode{
		model.PlayModeRepeatOne,
		model.PlayModeRepeatAll,
		model.PlayModeShuffle,
		model.PlayModeSequential,
	}
	
	for _, expectedMode := range modes {
		plugin.SetPlayMode(ctx, expectedMode)
		actualMode := plugin.GetPlayMode(ctx)
		if actualMode != expectedMode {
			t.Errorf("Expected mode %v, got %v", expectedMode, actualMode)
		}
	}
}

// TestGetNextSongSequential 测试顺序播放模式下获取下一首歌曲
func TestGetNextSongSequential(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 设置顺序播放模式
	plugin.SetPlayMode(ctx, model.PlayModeSequential)
	
	// 创建测试队列
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
		createTestSong("song3", "Song 3", "Artist 3"),
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	// 测试从第一首获取下一首
	nextSong, err := plugin.GetNextSong(ctx, songs[0])
	if err != nil {
		t.Errorf("Failed to get next song: %v", err)
	}
	if nextSong.ID != "song2" {
		t.Errorf("Expected next song 'song2', got '%s'", nextSong.ID)
	}
	
	// 测试从第二首获取下一首
	nextSong, err = plugin.GetNextSong(ctx, songs[1])
	if err != nil {
		t.Errorf("Failed to get next song: %v", err)
	}
	if nextSong.ID != "song3" {
		t.Errorf("Expected next song 'song3', got '%s'", nextSong.ID)
	}
	
	// 测试从最后一首获取下一首（应该失败）
	_, err = plugin.GetNextSong(ctx, songs[2])
	if err == nil {
		t.Error("Expected error when getting next song from last song in sequential mode")
	}
	
	// 测试空队列
	plugin.ClearQueue(ctx)
	_, err = plugin.GetNextSong(ctx, songs[0])
	if err == nil {
		t.Error("Expected error when getting next song from empty queue")
	}
	
	// 测试nil当前歌曲
	plugin.SetCurrentQueue(ctx, songs)
	nextSong, err = plugin.GetNextSong(ctx, nil)
	if err != nil {
		t.Errorf("Failed to get first song when current is nil: %v", err)
	}
	if nextSong.ID != "song1" {
		t.Errorf("Expected first song 'song1', got '%s'", nextSong.ID)
	}
}

// TestGetPreviousSongSequential 测试顺序播放模式下获取上一首歌曲
func TestGetPreviousSongSequential(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 设置顺序播放模式
	plugin.SetPlayMode(ctx, model.PlayModeSequential)
	
	// 创建测试队列
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
		createTestSong("song3", "Song 3", "Artist 3"),
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	// 测试从最后一首获取上一首
	prevSong, err := plugin.GetPreviousSong(ctx, songs[2])
	if err != nil {
		t.Errorf("Failed to get previous song: %v", err)
	}
	if prevSong.ID != "song2" {
		t.Errorf("Expected previous song 'song2', got '%s'", prevSong.ID)
	}
	
	// 测试从第二首获取上一首
	prevSong, err = plugin.GetPreviousSong(ctx, songs[1])
	if err != nil {
		t.Errorf("Failed to get previous song: %v", err)
	}
	if prevSong.ID != "song1" {
		t.Errorf("Expected previous song 'song1', got '%s'", prevSong.ID)
	}
	
	// 测试从第一首获取上一首（应该失败）
	_, err = plugin.GetPreviousSong(ctx, songs[0])
	if err == nil {
		t.Error("Expected error when getting previous song from first song in sequential mode")
	}
	
	// 测试nil当前歌曲
	prevSong, err = plugin.GetPreviousSong(ctx, nil)
	if err != nil {
		t.Errorf("Failed to get last song when current is nil: %v", err)
	}
	if prevSong.ID != "song3" {
		t.Errorf("Expected last song 'song3', got '%s'", prevSong.ID)
	}
}

// TestGetNextSongRepeatOne 测试单曲循环模式
func TestGetNextSongRepeatOne(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 设置单曲循环模式
	plugin.SetPlayMode(ctx, model.PlayModeRepeatOne)
	
	// 创建测试队列
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	// 测试单曲循环（应该返回相同歌曲）
	nextSong, err := plugin.GetNextSong(ctx, songs[0])
	if err != nil {
		t.Errorf("Failed to get next song in repeat one mode: %v", err)
	}
	if nextSong.ID != "song1" {
		t.Errorf("Expected same song 'song1', got '%s'", nextSong.ID)
	}
	
	// 测试上一首也应该返回相同歌曲
	prevSong, err := plugin.GetPreviousSong(ctx, songs[0])
	if err != nil {
		t.Errorf("Failed to get previous song in repeat one mode: %v", err)
	}
	if prevSong.ID != "song1" {
		t.Errorf("Expected same song 'song1', got '%s'", prevSong.ID)
	}
}

// TestGetNextSongRepeatAll 测试列表循环模式
func TestGetNextSongRepeatAll(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 设置列表循环模式
	plugin.SetPlayMode(ctx, model.PlayModeRepeatAll)
	
	// 创建测试队列
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
		createTestSong("song3", "Song 3", "Artist 3"),
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	// 测试正常的下一首
	nextSong, err := plugin.GetNextSong(ctx, songs[0])
	if err != nil {
		t.Errorf("Failed to get next song: %v", err)
	}
	if nextSong.ID != "song2" {
		t.Errorf("Expected next song 'song2', got '%s'", nextSong.ID)
	}
	
	// 测试从最后一首获取下一首（应该循环到第一首）
	nextSong, err = plugin.GetNextSong(ctx, songs[2])
	if err != nil {
		t.Errorf("Failed to get next song from last: %v", err)
	}
	if nextSong.ID != "song1" {
		t.Errorf("Expected first song 'song1', got '%s'", nextSong.ID)
	}
	
	// 测试从第一首获取上一首（应该循环到最后一首）
	prevSong, err := plugin.GetPreviousSong(ctx, songs[0])
	if err != nil {
		t.Errorf("Failed to get previous song from first: %v", err)
	}
	if prevSong.ID != "song3" {
		t.Errorf("Expected last song 'song3', got '%s'", prevSong.ID)
	}
}

// TestGetNextSongShuffle 测试随机播放模式
func TestGetNextSongShuffle(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 设置随机播放模式
	plugin.SetPlayMode(ctx, model.PlayModeShuffle)
	
	// 创建测试队列
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
		createTestSong("song3", "Song 3", "Artist 3"),
		createTestSong("song4", "Song 4", "Artist 4"),
		createTestSong("song5", "Song 5", "Artist 5"),
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	// 测试获取下一首歌曲（应该是随机的）
	nextSong, err := plugin.GetNextSong(ctx, songs[0])
	if err != nil {
		t.Errorf("Failed to get next song in shuffle mode: %v", err)
	}
	
	// 验证返回的歌曲在队列中
	found := false
	for _, song := range songs {
		if song.ID == nextSong.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Next song '%s' not found in original queue", nextSong.ID)
	}
	
	// 测试获取上一首歌曲
	prevSong, err := plugin.GetPreviousSong(ctx, songs[0])
	if err != nil {
		t.Errorf("Failed to get previous song in shuffle mode: %v", err)
	}
	
	// 验证返回的歌曲在队列中
	found = false
	for _, song := range songs {
		if song.ID == prevSong.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Previous song '%s' not found in original queue", prevSong.ID)
	}
	
	// 测试多次获取下一首，确保能遍历所有歌曲
	playedSongs := make(map[string]bool)
	currentSong := songs[0]
	for i := 0; i < len(songs)*2; i++ { // 多遍历一轮确保循环
		nextSong, err := plugin.GetNextSong(ctx, currentSong)
		if err != nil {
			t.Errorf("Failed to get next song at iteration %d: %v", i, err)
			break
		}
		playedSongs[nextSong.ID] = true
		currentSong = nextSong
	}
	
	// 验证所有歌曲都被播放过
	for _, song := range songs {
		if !playedSongs[song.ID] {
			t.Errorf("Song '%s' was never played in shuffle mode", song.ID)
		}
	}
}

// TestTogglePlayMode 测试切换播放模式
func TestTogglePlayMode(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 创建测试队列（随机模式需要）
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	mockEventBus.ClearEvents()
	
	// 测试模式切换序列：Sequential -> RepeatOne -> RepeatAll -> Shuffle -> Sequential
	expectedModes := []model.PlayMode{
		model.PlayModeRepeatOne,   // 从Sequential切换到RepeatOne
		model.PlayModeRepeatAll,   // 从RepeatOne切换到RepeatAll
		model.PlayModeShuffle,     // 从RepeatAll切换到Shuffle
		model.PlayModeSequential,  // 从Shuffle切换到Sequential
	}
	
	for i, expectedMode := range expectedModes {
		actualMode, err := plugin.TogglePlayMode(ctx)
		if err != nil {
			t.Errorf("Failed to toggle play mode at step %d: %v", i, err)
		}
		
		if actualMode != expectedMode {
			t.Errorf("Expected mode %v at step %d, got %v", expectedMode, i, actualMode)
		}
		
		// 验证GetPlayMode返回相同结果
		getMode := plugin.GetPlayMode(ctx)
		if getMode != expectedMode {
			t.Errorf("GetPlayMode returned %v, expected %v at step %d", getMode, expectedMode, i)
		}
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	if len(events) != len(expectedModes) {
		t.Errorf("Expected %d mode change events, got %d", len(expectedModes), len(events))
	}
}

// TestGetPlayModeDescription 测试获取播放模式描述
func TestGetPlayModeDescription(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	testCases := []struct {
		mode        model.PlayMode
		expectedDesc string
	}{
		{model.PlayModeSequential, "顺序播放"},
		{model.PlayModeRepeatOne, "单曲循环"},
		{model.PlayModeRepeatAll, "列表循环"},
		{model.PlayModeShuffle, "随机播放"},
	}
	
	for _, tc := range testCases {
		plugin.SetPlayMode(ctx, tc.mode)
		desc := plugin.GetPlayModeDescription(ctx)
		if desc != tc.expectedDesc {
			t.Errorf("Expected description '%s' for mode %v, got '%s'", tc.expectedDesc, tc.mode, desc)
		}
	}
}

// TestIsShuffleMode 测试检查是否为随机播放模式
func TestIsShuffleMode(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 默认不是随机模式
	if plugin.IsShuffleMode() {
		t.Error("Expected IsShuffleMode to be false by default")
	}
	
	// 设置随机模式
	plugin.SetPlayMode(ctx, model.PlayModeShuffle)
	if !plugin.IsShuffleMode() {
		t.Error("Expected IsShuffleMode to be true after setting shuffle mode")
	}
	
	// 设置其他模式
	plugin.SetPlayMode(ctx, model.PlayModeSequential)
	if plugin.IsShuffleMode() {
		t.Error("Expected IsShuffleMode to be false after setting sequential mode")
	}
}

// TestIsRepeatMode 测试检查是否为循环播放模式
func TestIsRepeatMode(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 默认不是循环模式
	if plugin.IsRepeatMode() {
		t.Error("Expected IsRepeatMode to be false by default")
	}
	
	// 设置单曲循环模式
	plugin.SetPlayMode(ctx, model.PlayModeRepeatOne)
	if !plugin.IsRepeatMode() {
		t.Error("Expected IsRepeatMode to be true for repeat one mode")
	}
	
	// 设置列表循环模式
	plugin.SetPlayMode(ctx, model.PlayModeRepeatAll)
	if !plugin.IsRepeatMode() {
		t.Error("Expected IsRepeatMode to be true for repeat all mode")
	}
	
	// 设置顺序播放模式
	plugin.SetPlayMode(ctx, model.PlayModeSequential)
	if plugin.IsRepeatMode() {
		t.Error("Expected IsRepeatMode to be false for sequential mode")
	}
	
	// 设置随机播放模式
	plugin.SetPlayMode(ctx, model.PlayModeShuffle)
	if plugin.IsRepeatMode() {
		t.Error("Expected IsRepeatMode to be false for shuffle mode")
	}
}

// TestPlayModeWithEmptyQueue 测试空队列下的播放模式操作
func TestPlayModeWithEmptyQueue(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 测试各种播放模式下空队列的行为
	modes := []model.PlayMode{
		model.PlayModeSequential,
		model.PlayModeRepeatOne,
		model.PlayModeRepeatAll,
		model.PlayModeShuffle,
	}
	
	for _, mode := range modes {
		plugin.SetPlayMode(ctx, mode)
		
		// 空队列下获取下一首应该失败
		_, err := plugin.GetNextSong(ctx, nil)
		if err == nil {
			t.Errorf("Expected error when getting next song from empty queue in mode %v", mode)
		}
		
		// 空队列下获取上一首应该失败
		_, err = plugin.GetPreviousSong(ctx, nil)
		if err == nil {
			t.Errorf("Expected error when getting previous song from empty queue in mode %v", mode)
		}
	}
}

// BenchmarkGetNextSongSequential 顺序播放模式性能基准测试
func BenchmarkGetNextSongSequential(b *testing.B) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	plugin.SetPlayMode(ctx, model.PlayModeSequential)
	
	// 创建大队列
	songs := make([]*model.Song, 1000)
	for i := 0; i < 1000; i++ {
		songs[i] = createTestSong(string(rune(i)), "Song", "Artist")
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		currentSong := songs[i%len(songs)]
		_, err := plugin.GetNextSong(ctx, currentSong)
		if err != nil && i < len(songs)-1 { // 最后一首会出错，这是正常的
			b.Errorf("Failed to get next song: %v", err)
		}
	}
}

// BenchmarkGetNextSongShuffle 随机播放模式性能基准测试
func BenchmarkGetNextSongShuffle(b *testing.B) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	plugin.SetPlayMode(ctx, model.PlayModeShuffle)
	
	// 创建大队列
	songs := make([]*model.Song, 1000)
	for i := 0; i < 1000; i++ {
		songs[i] = createTestSong(string(rune(i)), "Song", "Artist")
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		currentSong := songs[i%len(songs)]
		_, err := plugin.GetNextSong(ctx, currentSong)
		if err != nil {
			b.Errorf("Failed to get next song: %v", err)
		}
	}
}