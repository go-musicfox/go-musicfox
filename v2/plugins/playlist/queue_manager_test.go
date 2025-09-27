package playlist

import (
	"context"
	"testing"

	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
)

// TestSetCurrentQueue 测试设置当前播放队列
func TestSetCurrentQueue(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 创建测试歌曲
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
		createTestSong("song3", "Song 3", "Artist 3"),
	}
	
	// 设置队列
	err := plugin.SetCurrentQueue(ctx, songs)
	if err != nil {
		t.Errorf("Failed to set current queue: %v", err)
	}
	
	// 检查队列是否设置成功
	queue, err := plugin.GetCurrentQueue(ctx)
	if err != nil {
		t.Errorf("Failed to get current queue: %v", err)
	}
	
	if len(queue) != 3 {
		t.Errorf("Expected queue length 3, got %d", len(queue))
	}
	
	for i, song := range queue {
		if song.ID != songs[i].ID {
			t.Errorf("Expected song %d ID '%s', got '%s'", i, songs[i].ID, song.ID)
		}
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "queue.updated" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected queue.updated event to be published")
	}
	
	// 测试设置nil队列
	err = plugin.SetCurrentQueue(ctx, nil)
	if err == nil {
		t.Error("Expected error when setting nil queue")
	}
	
	// 测试设置空队列
	err = plugin.SetCurrentQueue(ctx, []*model.Song{})
	if err != nil {
		t.Errorf("Failed to set empty queue: %v", err)
	}
	
	emptyQueue, err := plugin.GetCurrentQueue(ctx)
	if err != nil {
		t.Errorf("Failed to get empty queue: %v", err)
	}
	
	if len(emptyQueue) != 0 {
		t.Errorf("Expected empty queue, got length %d", len(emptyQueue))
	}
}

// TestAddToQueue 测试添加歌曲到队列
func TestAddToQueue(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 创建测试歌曲
	song1 := createTestSong("song1", "Song 1", "Artist 1")
	song2 := createTestSong("song2", "Song 2", "Artist 2")
	
	// 添加第一首歌曲
	err := plugin.AddToQueue(ctx, song1)
	if err != nil {
		t.Errorf("Failed to add song to queue: %v", err)
	}
	
	// 检查队列
	queue, err := plugin.GetCurrentQueue(ctx)
	if err != nil {
		t.Errorf("Failed to get queue: %v", err)
	}
	
	if len(queue) != 1 {
		t.Errorf("Expected queue length 1, got %d", len(queue))
	}
	
	if queue[0].ID != "song1" {
		t.Errorf("Expected song ID 'song1', got '%s'", queue[0].ID)
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "queue.song_added" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected queue.song_added event to be published")
	}
	
	mockEventBus.ClearEvents()
	
	// 添加第二首歌曲
	err = plugin.AddToQueue(ctx, song2)
	if err != nil {
		t.Errorf("Failed to add second song to queue: %v", err)
	}
	
	// 检查队列长度
	queue, _ = plugin.GetCurrentQueue(ctx)
	if len(queue) != 2 {
		t.Errorf("Expected queue length 2, got %d", len(queue))
	}
	
	// 测试添加重复歌曲
	err = plugin.AddToQueue(ctx, song1)
	if err == nil {
		t.Error("Expected error when adding duplicate song to queue")
	}
	
	// 测试添加nil歌曲
	err = plugin.AddToQueue(ctx, nil)
	if err == nil {
		t.Error("Expected error when adding nil song to queue")
	}
}

// TestRemoveFromQueue 测试从队列移除歌曲
func TestRemoveFromQueue(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 创建测试队列
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
		createTestSong("song3", "Song 3", "Artist 3"),
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	mockEventBus.ClearEvents()
	
	// 移除中间的歌曲
	err := plugin.RemoveFromQueue(ctx, "song2")
	if err != nil {
		t.Errorf("Failed to remove song from queue: %v", err)
	}
	
	// 检查队列
	queue, err := plugin.GetCurrentQueue(ctx)
	if err != nil {
		t.Errorf("Failed to get queue: %v", err)
	}
	
	if len(queue) != 2 {
		t.Errorf("Expected queue length 2, got %d", len(queue))
	}
	
	expectedIDs := []string{"song1", "song3"}
	for i, expectedID := range expectedIDs {
		if queue[i].ID != expectedID {
			t.Errorf("Expected song %d ID '%s', got '%s'", i, expectedID, queue[i].ID)
		}
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "queue.song_removed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected queue.song_removed event to be published")
	}
	
	// 测试移除不存在的歌曲
	err = plugin.RemoveFromQueue(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when removing nonexistent song from queue")
	}
	
	// 测试空ID
	err = plugin.RemoveFromQueue(ctx, "")
	if err == nil {
		t.Error("Expected error when removing song with empty ID")
	}
}

// TestClearQueue 测试清空队列
func TestClearQueue(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 创建测试队列
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	mockEventBus.ClearEvents()
	
	// 清空队列
	err := plugin.ClearQueue(ctx)
	if err != nil {
		t.Errorf("Failed to clear queue: %v", err)
	}
	
	// 检查队列是否为空
	queue, err := plugin.GetCurrentQueue(ctx)
	if err != nil {
		t.Errorf("Failed to get queue: %v", err)
	}
	
	if len(queue) != 0 {
		t.Errorf("Expected empty queue, got length %d", len(queue))
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "queue.cleared" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected queue.cleared event to be published")
	}
}

// TestShuffleQueue 测试打乱队列
func TestShuffleQueue(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 创建测试队列（足够大以便观察打乱效果）
	songs := make([]*model.Song, 10)
	for i := 0; i < 10; i++ {
		songs[i] = createTestSong(string(rune('a'+i)), "Song "+string(rune('A'+i)), "Artist")
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	// 记录原始顺序
	originalQueue, _ := plugin.GetCurrentQueue(ctx)
	originalOrder := make([]string, len(originalQueue))
	for i, song := range originalQueue {
		originalOrder[i] = song.ID
	}
	
	mockEventBus.ClearEvents()
	
	// 打乱队列
	err := plugin.ShuffleQueue(ctx)
	if err != nil {
		t.Errorf("Failed to shuffle queue: %v", err)
	}
	
	// 检查队列长度是否保持不变
	shuffledQueue, err := plugin.GetCurrentQueue(ctx)
	if err != nil {
		t.Errorf("Failed to get shuffled queue: %v", err)
	}
	
	if len(shuffledQueue) != len(originalQueue) {
		t.Errorf("Expected queue length %d, got %d", len(originalQueue), len(shuffledQueue))
	}
	
	// 检查所有歌曲是否仍然存在
	shuffledIDs := make(map[string]bool)
	for _, song := range shuffledQueue {
		shuffledIDs[song.ID] = true
	}
	
	for _, originalID := range originalOrder {
		if !shuffledIDs[originalID] {
			t.Errorf("Song '%s' missing after shuffle", originalID)
		}
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "queue.shuffled" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected queue.shuffled event to be published")
	}
	
	// 测试打乱空队列
	plugin.ClearQueue(ctx)
	err = plugin.ShuffleQueue(ctx)
	if err != nil {
		t.Errorf("Failed to shuffle empty queue: %v", err)
	}
	
	// 测试打乱单首歌曲的队列
	plugin.AddToQueue(ctx, songs[0])
	err = plugin.ShuffleQueue(ctx)
	if err != nil {
		t.Errorf("Failed to shuffle single song queue: %v", err)
	}
}

// TestQueueWithShuffleMode 测试随机播放模式下的队列操作
func TestQueueWithShuffleMode(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 设置随机播放模式
	plugin.SetPlayMode(ctx, model.PlayModeShuffle)
	
	// 创建测试队列
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
		createTestSong("song3", "Song 3", "Artist 3"),
	}
	
	// 设置队列（应该自动生成随机索引）
	err := plugin.SetCurrentQueue(ctx, songs)
	if err != nil {
		t.Errorf("Failed to set queue in shuffle mode: %v", err)
	}
	
	// 检查随机索引是否生成
	plugin.mu.RLock()
	shuffleIndexExists := plugin.shuffleIndex != nil
	shuffleIndexLength := 0
	if plugin.shuffleIndex != nil {
		shuffleIndexLength = len(plugin.shuffleIndex)
	}
	plugin.mu.RUnlock()
	
	if !shuffleIndexExists {
		t.Error("Expected shuffle index to be generated in shuffle mode")
	}
	
	if shuffleIndexLength != len(songs) {
		t.Errorf("Expected shuffle index length %d, got %d", len(songs), shuffleIndexLength)
	}
	
	// 添加歌曲到队列（应该重新生成随机索引）
	newSong := createTestSong("song4", "Song 4", "Artist 4")
	err = plugin.AddToQueue(ctx, newSong)
	if err != nil {
		t.Errorf("Failed to add song to queue in shuffle mode: %v", err)
	}
	
	// 检查随机索引是否更新
	plugin.mu.RLock()
	newShuffleIndexLength := 0
	if plugin.shuffleIndex != nil {
		newShuffleIndexLength = len(plugin.shuffleIndex)
	}
	plugin.mu.RUnlock()
	
	if newShuffleIndexLength != 4 {
		t.Errorf("Expected updated shuffle index length 4, got %d", newShuffleIndexLength)
	}
}

// TestGetCurrentSongIndex 测试获取当前歌曲索引
func TestGetCurrentSongIndex(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 创建测试队列
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
		createTestSong("song3", "Song 3", "Artist 3"),
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	// 测试获取存在的歌曲索引
	index := plugin.getCurrentSongIndex(songs[1])
	if index != 1 {
		t.Errorf("Expected index 1, got %d", index)
	}
	
	// 测试获取不存在的歌曲索引
	nonExistentSong := createTestSong("nonexistent", "Non-existent", "Artist")
	index = plugin.getCurrentSongIndex(nonExistentSong)
	if index != -1 {
		t.Errorf("Expected index -1 for non-existent song, got %d", index)
	}
	
	// 测试nil歌曲
	index = plugin.getCurrentSongIndex(nil)
	if index != -1 {
		t.Errorf("Expected index -1 for nil song, got %d", index)
	}
}

// TestShuffleIndexOperations 测试随机索引操作
func TestShuffleIndexOperations(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 创建测试队列
	songs := []*model.Song{
		createTestSong("song1", "Song 1", "Artist 1"),
		createTestSong("song2", "Song 2", "Artist 2"),
		createTestSong("song3", "Song 3", "Artist 3"),
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	// 生成随机索引
	plugin.generateShuffleIndex()
	
	// 测试获取随机索引
	shuffleIndex := plugin.getShuffleIndex(1) // 实际索引1对应的随机索引
	if shuffleIndex < 0 || shuffleIndex >= len(songs) {
		t.Errorf("Invalid shuffle index %d", shuffleIndex)
	}
	
	// 测试从随机索引获取实际索引
	actualIndex := plugin.getActualIndex(shuffleIndex)
	if actualIndex != 1 {
		t.Errorf("Expected actual index 1, got %d", actualIndex)
	}
	
	// 测试无效的随机索引
	invalidShuffleIndex := plugin.getShuffleIndex(10)
	if invalidShuffleIndex != -1 {
		t.Errorf("Expected -1 for invalid actual index, got %d", invalidShuffleIndex)
	}
	
	invalidActualIndex := plugin.getActualIndex(10)
	if invalidActualIndex != -1 {
		t.Errorf("Expected -1 for invalid shuffle index, got %d", invalidActualIndex)
	}
}

// BenchmarkAddToQueueManager 队列添加性能基准测试
func BenchmarkAddToQueueManager(b *testing.B) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		song := createTestSong(string(rune(i)), "Song", "Artist")
		err := plugin.AddToQueue(ctx, song)
		if err != nil {
			b.Errorf("Failed to add song to queue: %v", err)
		}
	}
}

// BenchmarkShuffleQueue 队列打乱性能基准测试
func BenchmarkShuffleQueue(b *testing.B) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 创建大队列
	songs := make([]*model.Song, 1000)
	for i := 0; i < 1000; i++ {
		songs[i] = createTestSong(string(rune(i)), "Song", "Artist")
	}
	plugin.SetCurrentQueue(ctx, songs)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := plugin.ShuffleQueue(ctx)
		if err != nil {
			b.Errorf("Failed to shuffle queue: %v", err)
		}
	}
}