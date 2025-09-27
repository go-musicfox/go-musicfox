package playlist

import (
	"context"
	"testing"
)

// TestCreatePlaylist 测试创建播放列表
func TestCreatePlaylist(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 测试正常创建
	playlist, err := plugin.CreatePlaylist(ctx, "My Playlist", "My favorite songs")
	if err != nil {
		t.Errorf("Failed to create playlist: %v", err)
	}
	
	if playlist == nil {
		t.Fatal("Playlist should not be nil")
	}
	
	if playlist.Name != "My Playlist" {
		t.Errorf("Expected playlist name 'My Playlist', got '%s'", playlist.Name)
	}
	
	if playlist.Description != "My favorite songs" {
		t.Errorf("Expected description 'My favorite songs', got '%s'", playlist.Description)
	}
	
	if playlist.Source != "local" {
		t.Errorf("Expected source 'local', got '%s'", playlist.Source)
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "playlist.created" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected playlist.created event to be published")
	}
	
	// 测试空名称
	_, err = plugin.CreatePlaylist(ctx, "", "Description")
	if err == nil {
		t.Error("Expected error when creating playlist with empty name")
	}
}

// TestDeletePlaylist 测试删除播放列表
func TestDeletePlaylist(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 创建播放列表
	playlist, err := plugin.CreatePlaylist(ctx, "Test Playlist", "Test")
	if err != nil {
		t.Fatalf("Failed to create playlist: %v", err)
	}
	
	mockEventBus.ClearEvents()
	
	// 删除播放列表
	err = plugin.DeletePlaylist(ctx, playlist.ID)
	if err != nil {
		t.Errorf("Failed to delete playlist: %v", err)
	}
	
	// 检查播放列表是否被删除
	_, err = plugin.GetPlaylist(ctx, playlist.ID)
	if err == nil {
		t.Error("Expected error when getting deleted playlist")
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "playlist.deleted" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected playlist.deleted event to be published")
	}
	
	// 测试删除不存在的播放列表
	err = plugin.DeletePlaylist(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when deleting nonexistent playlist")
	}
	
	// 测试空ID
	err = plugin.DeletePlaylist(ctx, "")
	if err == nil {
		t.Error("Expected error when deleting playlist with empty ID")
	}
}

// TestUpdatePlaylist 测试更新播放列表
func TestUpdatePlaylist(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 创建播放列表
	playlist, err := plugin.CreatePlaylist(ctx, "Original Name", "Original Description")
	if err != nil {
		t.Fatalf("Failed to create playlist: %v", err)
	}
	
	mockEventBus.ClearEvents()
	
	// 更新播放列表
	updates := map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated Description",
		"is_public":   true,
	}
	
	err = plugin.UpdatePlaylist(ctx, playlist.ID, updates)
	if err != nil {
		t.Errorf("Failed to update playlist: %v", err)
	}
	
	// 检查更新是否生效
	updatedPlaylist, err := plugin.GetPlaylist(ctx, playlist.ID)
	if err != nil {
		t.Errorf("Failed to get updated playlist: %v", err)
	}
	
	if updatedPlaylist.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updatedPlaylist.Name)
	}
	
	if updatedPlaylist.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got '%s'", updatedPlaylist.Description)
	}
	
	if !updatedPlaylist.IsPublic {
		t.Error("Expected playlist to be public")
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "playlist.updated" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected playlist.updated event to be published")
	}
	
	// 测试更新不存在的播放列表
	err = plugin.UpdatePlaylist(ctx, "nonexistent", updates)
	if err == nil {
		t.Error("Expected error when updating nonexistent playlist")
	}
}

// TestGetPlaylist 测试获取播放列表
func TestGetPlaylist(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 创建播放列表
	originalPlaylist, err := plugin.CreatePlaylist(ctx, "Test Playlist", "Test Description")
	if err != nil {
		t.Fatalf("Failed to create playlist: %v", err)
	}
	
	// 获取播放列表
	playlist, err := plugin.GetPlaylist(ctx, originalPlaylist.ID)
	if err != nil {
		t.Errorf("Failed to get playlist: %v", err)
	}
	
	if playlist.ID != originalPlaylist.ID {
		t.Errorf("Expected playlist ID '%s', got '%s'", originalPlaylist.ID, playlist.ID)
	}
	
	if playlist.Name != "Test Playlist" {
		t.Errorf("Expected playlist name 'Test Playlist', got '%s'", playlist.Name)
	}
	
	// 测试获取不存在的播放列表
	_, err = plugin.GetPlaylist(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when getting nonexistent playlist")
	}
	
	// 测试空ID
	_, err = plugin.GetPlaylist(ctx, "")
	if err == nil {
		t.Error("Expected error when getting playlist with empty ID")
	}
}

// TestListPlaylists 测试列出播放列表
func TestListPlaylists(t *testing.T) {
	plugin, _ := setupTestPlugin()
	ctx := context.Background()
	
	// 初始应该没有播放列表
	playlists, err := plugin.ListPlaylists(ctx)
	if err != nil {
		t.Errorf("Failed to list playlists: %v", err)
	}
	
	if len(playlists) != 0 {
		t.Errorf("Expected 0 playlists, got %d", len(playlists))
	}
	
	// 创建几个播放列表
	playlist1, _ := plugin.CreatePlaylist(ctx, "Playlist 1", "Description 1")
	playlist2, _ := plugin.CreatePlaylist(ctx, "Playlist 2", "Description 2")
	playlist3, _ := plugin.CreatePlaylist(ctx, "Playlist 3", "Description 3")
	
	// 列出播放列表
	playlists, err = plugin.ListPlaylists(ctx)
	if err != nil {
		t.Errorf("Failed to list playlists: %v", err)
	}
	
	if len(playlists) != 3 {
		t.Errorf("Expected 3 playlists, got %d", len(playlists))
	}
	
	// 检查播放列表是否都存在
	playlistIDs := make(map[string]bool)
	for _, playlist := range playlists {
		playlistIDs[playlist.ID] = true
	}
	
	if !playlistIDs[playlist1.ID] {
		t.Error("Playlist 1 not found in list")
	}
	if !playlistIDs[playlist2.ID] {
		t.Error("Playlist 2 not found in list")
	}
	if !playlistIDs[playlist3.ID] {
		t.Error("Playlist 3 not found in list")
	}
}

// TestAddSong 测试添加歌曲到播放列表
func TestAddSong(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 创建播放列表
	playlist, err := plugin.CreatePlaylist(ctx, "Test Playlist", "Test")
	if err != nil {
		t.Fatalf("Failed to create playlist: %v", err)
	}
	
	mockEventBus.ClearEvents()
	
	// 创建测试歌曲
	song := createTestSong("song1", "Test Song", "Test Artist")
	
	// 添加歌曲到播放列表
	err = plugin.AddSong(ctx, playlist.ID, song)
	if err != nil {
		t.Errorf("Failed to add song to playlist: %v", err)
	}
	
	// 检查歌曲是否被添加
	updatedPlaylist, err := plugin.GetPlaylist(ctx, playlist.ID)
	if err != nil {
		t.Errorf("Failed to get updated playlist: %v", err)
	}
	
	if len(updatedPlaylist.Songs) != 1 {
		t.Errorf("Expected 1 song in playlist, got %d", len(updatedPlaylist.Songs))
	}
	
	if updatedPlaylist.Songs[0].ID != "song1" {
		t.Errorf("Expected song ID 'song1', got '%s'", updatedPlaylist.Songs[0].ID)
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "playlist.song_added" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected playlist.song_added event to be published")
	}
	
	// 测试添加重复歌曲
	err = plugin.AddSong(ctx, playlist.ID, song)
	if err == nil {
		t.Error("Expected error when adding duplicate song")
	}
	
	// 测试添加到不存在的播放列表
	err = plugin.AddSong(ctx, "nonexistent", song)
	if err == nil {
		t.Error("Expected error when adding song to nonexistent playlist")
	}
	
	// 测试添加nil歌曲
	err = plugin.AddSong(ctx, playlist.ID, nil)
	if err == nil {
		t.Error("Expected error when adding nil song")
	}
}

// TestRemoveSong 测试从播放列表移除歌曲
func TestRemoveSong(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 创建播放列表和歌曲
	playlist, _ := plugin.CreatePlaylist(ctx, "Test Playlist", "Test")
	song1 := createTestSong("song1", "Song 1", "Artist 1")
	song2 := createTestSong("song2", "Song 2", "Artist 2")
	
	plugin.AddSong(ctx, playlist.ID, song1)
	plugin.AddSong(ctx, playlist.ID, song2)
	
	mockEventBus.ClearEvents()
	
	// 移除歌曲
	err := plugin.RemoveSong(ctx, playlist.ID, "song1")
	if err != nil {
		t.Errorf("Failed to remove song from playlist: %v", err)
	}
	
	// 检查歌曲是否被移除
	updatedPlaylist, err := plugin.GetPlaylist(ctx, playlist.ID)
	if err != nil {
		t.Errorf("Failed to get updated playlist: %v", err)
	}
	
	if len(updatedPlaylist.Songs) != 1 {
		t.Errorf("Expected 1 song in playlist, got %d", len(updatedPlaylist.Songs))
	}
	
	if updatedPlaylist.Songs[0].ID != "song2" {
		t.Errorf("Expected remaining song ID 'song2', got '%s'", updatedPlaylist.Songs[0].ID)
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "playlist.song_removed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected playlist.song_removed event to be published")
	}
	
	// 测试移除不存在的歌曲
	err = plugin.RemoveSong(ctx, playlist.ID, "nonexistent")
	if err == nil {
		t.Error("Expected error when removing nonexistent song")
	}
}

// TestMoveSong 测试移动播放列表中的歌曲
func TestMoveSong(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 创建播放列表和歌曲
	playlist, _ := plugin.CreatePlaylist(ctx, "Test Playlist", "Test")
	song1 := createTestSong("song1", "Song 1", "Artist 1")
	song2 := createTestSong("song2", "Song 2", "Artist 2")
	song3 := createTestSong("song3", "Song 3", "Artist 3")
	
	plugin.AddSong(ctx, playlist.ID, song1)
	plugin.AddSong(ctx, playlist.ID, song2)
	plugin.AddSong(ctx, playlist.ID, song3)
	
	mockEventBus.ClearEvents()
	
	// 移动歌曲（将第一首歌移到最后）
	err := plugin.MoveSong(ctx, playlist.ID, "song1", 2)
	if err != nil {
		t.Errorf("Failed to move song: %v", err)
	}
	
	// 检查歌曲顺序
	updatedPlaylist, err := plugin.GetPlaylist(ctx, playlist.ID)
	if err != nil {
		t.Errorf("Failed to get updated playlist: %v", err)
	}
	
	expectedOrder := []string{"song2", "song3", "song1"}
	for i, expectedID := range expectedOrder {
		if updatedPlaylist.Songs[i].ID != expectedID {
			t.Errorf("Expected song at position %d to be '%s', got '%s'", i, expectedID, updatedPlaylist.Songs[i].ID)
		}
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "playlist.song_moved" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected playlist.song_moved event to be published")
	}
	
	// 测试移动到无效位置
	err = plugin.MoveSong(ctx, playlist.ID, "song1", 10)
	if err == nil {
		t.Error("Expected error when moving song to invalid position")
	}
	
	// 测试移动不存在的歌曲
	err = plugin.MoveSong(ctx, playlist.ID, "nonexistent", 1)
	if err == nil {
		t.Error("Expected error when moving nonexistent song")
	}
}

// TestClearPlaylist 测试清空播放列表
func TestClearPlaylist(t *testing.T) {
	plugin, mockEventBus := setupTestPlugin()
	ctx := context.Background()
	
	// 创建播放列表和歌曲
	playlist, _ := plugin.CreatePlaylist(ctx, "Test Playlist", "Test")
	song1 := createTestSong("song1", "Song 1", "Artist 1")
	song2 := createTestSong("song2", "Song 2", "Artist 2")
	
	plugin.AddSong(ctx, playlist.ID, song1)
	plugin.AddSong(ctx, playlist.ID, song2)
	
	mockEventBus.ClearEvents()
	
	// 清空播放列表
	err := plugin.ClearPlaylist(ctx, playlist.ID)
	if err != nil {
		t.Errorf("Failed to clear playlist: %v", err)
	}
	
	// 检查播放列表是否被清空
	updatedPlaylist, err := plugin.GetPlaylist(ctx, playlist.ID)
	if err != nil {
		t.Errorf("Failed to get updated playlist: %v", err)
	}
	
	if len(updatedPlaylist.Songs) != 0 {
		t.Errorf("Expected 0 songs in playlist, got %d", len(updatedPlaylist.Songs))
	}
	
	// 检查事件是否发送
	events := mockEventBus.GetPublishedEvents()
	found := false
	for _, event := range events {
		if event.GetType() == "playlist.cleared" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected playlist.cleared event to be published")
	}
	
	// 测试清空不存在的播放列表
	err = plugin.ClearPlaylist(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error when clearing nonexistent playlist")
	}
}