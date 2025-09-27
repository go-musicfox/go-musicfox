package model

import (
	"encoding/json"
	"testing"
	"time"
)

// TestPlayStatus tests the PlayStatus enum
func TestPlayStatus(t *testing.T) {
	tests := []struct {
		status   PlayStatus
		expected string
	}{
		{PlayStatusStopped, "stopped"},
		{PlayStatusPlaying, "playing"},
		{PlayStatusPaused, "paused"},
		{PlayStatusBuffering, "buffering"},
		{PlayStatusError, "error"},
		{PlayStatus(999), "unknown"},
	}

	for _, test := range tests {
		if got := test.status.String(); got != test.expected {
			t.Errorf("PlayStatus.String() = %v, want %v", got, test.expected)
		}
	}
}

// TestPlayMode tests the PlayMode enum
func TestPlayMode(t *testing.T) {
	tests := []struct {
		mode     PlayMode
		expected string
	}{
		{PlayModeSequential, "sequential"},
		{PlayModeRepeatOne, "repeat_one"},
		{PlayModeRepeatAll, "repeat_all"},
		{PlayModeShuffle, "shuffle"},
		{PlayMode(999), "unknown"},
	}

	for _, test := range tests {
		if got := test.mode.String(); got != test.expected {
			t.Errorf("PlayMode.String() = %v, want %v", got, test.expected)
		}
	}
}

// TestQuality tests the Quality enum
func TestQuality(t *testing.T) {
	tests := []struct {
		quality  Quality
		expected string
	}{
		{QualityLow, "low"},
		{QualityMedium, "medium"},
		{QualityHigh, "high"},
		{QualityLossless, "lossless"},
		{Quality(999), "unknown"},
	}

	for _, test := range tests {
		if got := test.quality.String(); got != test.expected {
			t.Errorf("Quality.String() = %v, want %v", got, test.expected)
		}
	}
}

// TestNewSong tests the NewSong constructor
func TestNewSong(t *testing.T) {
	song := NewSong("123", "Test Song", "Test Artist", "test_source")

	if song.ID != "123" {
		t.Errorf("Expected ID to be '123', got '%s'", song.ID)
	}
	if song.Title != "Test Song" {
		t.Errorf("Expected Title to be 'Test Song', got '%s'", song.Title)
	}
	if song.Artist != "Test Artist" {
		t.Errorf("Expected Artist to be 'Test Artist', got '%s'", song.Artist)
	}
	if song.Source != "test_source" {
		t.Errorf("Expected Source to be 'test_source', got '%s'", song.Source)
	}
	if song.Quality != QualityMedium {
		t.Errorf("Expected Quality to be QualityMedium, got %v", song.Quality)
	}
	if song.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}
	if song.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if song.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

// TestSongSerialization tests Song JSON serialization and deserialization
func TestSongSerialization(t *testing.T) {
	original := NewSong("123", "Test Song", "Test Artist", "test_source")
	original.Album = "Test Album"
	original.Duration = 3 * time.Minute
	original.URL = "http://example.com/song.mp3"
	original.CoverURL = "http://example.com/cover.jpg"
	original.Quality = QualityHigh
	original.Metadata["genre"] = "rock"

	// Serialize to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Song: %v", err)
	}

	// Deserialize from JSON
	var deserialized Song
	err = json.Unmarshal(jsonData, &deserialized)
	if err != nil {
		t.Fatalf("Failed to unmarshal Song: %v", err)
	}

	// Compare fields
	if deserialized.ID != original.ID {
		t.Errorf("ID mismatch: got %s, want %s", deserialized.ID, original.ID)
	}
	if deserialized.Title != original.Title {
		t.Errorf("Title mismatch: got %s, want %s", deserialized.Title, original.Title)
	}
	if deserialized.Artist != original.Artist {
		t.Errorf("Artist mismatch: got %s, want %s", deserialized.Artist, original.Artist)
	}
	if deserialized.Quality != original.Quality {
		t.Errorf("Quality mismatch: got %v, want %v", deserialized.Quality, original.Quality)
	}
}

// TestNewPlaylist tests the NewPlaylist constructor
func TestNewPlaylist(t *testing.T) {
	playlist := NewPlaylist("pl123", "Test Playlist", "test_source", "user123")

	if playlist.ID != "pl123" {
		t.Errorf("Expected ID to be 'pl123', got '%s'", playlist.ID)
	}
	if playlist.Name != "Test Playlist" {
		t.Errorf("Expected Name to be 'Test Playlist', got '%s'", playlist.Name)
	}
	if playlist.Source != "test_source" {
		t.Errorf("Expected Source to be 'test_source', got '%s'", playlist.Source)
	}
	if playlist.CreatedBy != "user123" {
		t.Errorf("Expected CreatedBy to be 'user123', got '%s'", playlist.CreatedBy)
	}
	if playlist.IsPublic {
		t.Error("Expected IsPublic to be false")
	}
	if playlist.Songs == nil {
		t.Error("Expected Songs to be initialized")
	}
	if len(playlist.Songs) != 0 {
		t.Errorf("Expected Songs to be empty, got %d songs", len(playlist.Songs))
	}
}

// TestPlaylistOperations tests playlist song operations
func TestPlaylistOperations(t *testing.T) {
	playlist := NewPlaylist("pl123", "Test Playlist", "test_source", "user123")
	song1 := NewSong("s1", "Song 1", "Artist 1", "source1")
	song2 := NewSong("s2", "Song 2", "Artist 2", "source2")

	// Test AddSong
	playlist.AddSong(song1)
	playlist.AddSong(song2)

	if playlist.GetSongCount() != 2 {
		t.Errorf("Expected 2 songs, got %d", playlist.GetSongCount())
	}

	// Test RemoveSong
	removed := playlist.RemoveSong("s1")
	if !removed {
		t.Error("Expected RemoveSong to return true")
	}
	if playlist.GetSongCount() != 1 {
		t.Errorf("Expected 1 song after removal, got %d", playlist.GetSongCount())
	}

	// Test removing non-existent song
	removed = playlist.RemoveSong("nonexistent")
	if removed {
		t.Error("Expected RemoveSong to return false for non-existent song")
	}
}

// TestPlaylistSerialization tests Playlist JSON serialization
func TestPlaylistSerialization(t *testing.T) {
	original := NewPlaylist("pl123", "Test Playlist", "test_source", "user123")
	original.Description = "A test playlist"
	original.IsPublic = true
	original.AddSong(NewSong("s1", "Song 1", "Artist 1", "source1"))

	// Serialize to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Playlist: %v", err)
	}

	// Deserialize from JSON
	var deserialized Playlist
	err = json.Unmarshal(jsonData, &deserialized)
	if err != nil {
		t.Fatalf("Failed to unmarshal Playlist: %v", err)
	}

	// Compare fields
	if deserialized.ID != original.ID {
		t.Errorf("ID mismatch: got %s, want %s", deserialized.ID, original.ID)
	}
	if deserialized.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", deserialized.Name, original.Name)
	}
	if deserialized.IsPublic != original.IsPublic {
		t.Errorf("IsPublic mismatch: got %v, want %v", deserialized.IsPublic, original.IsPublic)
	}
	if len(deserialized.Songs) != len(original.Songs) {
		t.Errorf("Songs count mismatch: got %d, want %d", len(deserialized.Songs), len(original.Songs))
	}
}

// TestNewPlayerState tests the NewPlayerState constructor
func TestNewPlayerState(t *testing.T) {
	ps := NewPlayerState()

	if ps.Status != PlayStatusStopped {
		t.Errorf("Expected Status to be PlayStatusStopped, got %v", ps.Status)
	}
	if ps.Volume != 0.8 {
		t.Errorf("Expected Volume to be 0.8, got %f", ps.Volume)
	}
	if ps.IsMuted {
		t.Error("Expected IsMuted to be false")
	}
	if ps.PlayMode != PlayModeSequential {
		t.Errorf("Expected PlayMode to be PlayModeSequential, got %v", ps.PlayMode)
	}
	if ps.Queue == nil {
		t.Error("Expected Queue to be initialized")
	}
	if ps.History == nil {
		t.Error("Expected History to be initialized")
	}
}

// TestPlayerStateMethods tests PlayerState methods
func TestPlayerStateMethods(t *testing.T) {
	ps := NewPlayerState()

	// Test IsPlaying
	if ps.IsPlaying() {
		t.Error("Expected IsPlaying to be false for stopped state")
	}

	ps.Status = PlayStatusPlaying
	if !ps.IsPlaying() {
		t.Error("Expected IsPlaying to be true for playing state")
	}

	// Test IsPaused
	ps.Status = PlayStatusPaused
	if !ps.IsPaused() {
		t.Error("Expected IsPaused to be true for paused state")
	}

	ps.Status = PlayStatusPlaying
	if ps.IsPaused() {
		t.Error("Expected IsPaused to be false for playing state")
	}
}

// TestPlayerStateSerialization tests PlayerState JSON serialization
func TestPlayerStateSerialization(t *testing.T) {
	original := NewPlayerState()
	original.Status = PlayStatusPlaying
	original.CurrentSong = NewSong("s1", "Current Song", "Artist", "source")
	original.Position = 30 * time.Second
	original.Duration = 3 * time.Minute
	original.Volume = 0.7
	original.PlayMode = PlayModeShuffle

	// Serialize to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal PlayerState: %v", err)
	}

	// Deserialize from JSON
	var deserialized PlayerState
	err = json.Unmarshal(jsonData, &deserialized)
	if err != nil {
		t.Fatalf("Failed to unmarshal PlayerState: %v", err)
	}

	// Compare fields
	if deserialized.Status != original.Status {
		t.Errorf("Status mismatch: got %v, want %v", deserialized.Status, original.Status)
	}
	if deserialized.Volume != original.Volume {
		t.Errorf("Volume mismatch: got %f, want %f", deserialized.Volume, original.Volume)
	}
	if deserialized.PlayMode != original.PlayMode {
		t.Errorf("PlayMode mismatch: got %v, want %v", deserialized.PlayMode, original.PlayMode)
	}
}

// TestNewAppState tests the NewAppState constructor
func TestNewAppState(t *testing.T) {
	appState := NewAppState()

	if appState.Player == nil {
		t.Error("Expected Player to be initialized")
	}
	if appState.CurrentView != "main" {
		t.Errorf("Expected CurrentView to be 'main', got '%s'", appState.CurrentView)
	}
	if appState.Config == nil {
		t.Error("Expected Config to be initialized")
	}
	if appState.Plugins == nil {
		t.Error("Expected Plugins to be initialized")
	}
	if appState.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

// TestAppStateOperations tests AppState operations
func TestAppStateOperations(t *testing.T) {
	appState := NewAppState()

	// Test UpdateConfig and GetConfig
	appState.UpdateConfig("theme", "dark")
	value, exists := appState.GetConfig("theme")
	if !exists {
		t.Error("Expected config key 'theme' to exist")
	}
	if value != "dark" {
		t.Errorf("Expected config value to be 'dark', got '%s'", value)
	}

	// Test non-existent config
	_, exists = appState.GetConfig("nonexistent")
	if exists {
		t.Error("Expected config key 'nonexistent' to not exist")
	}

	// Test AddPlugin
	appState.AddPlugin("audio_plugin")
	appState.AddPlugin("ui_plugin")
	if len(appState.Plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(appState.Plugins))
	}

	// Test adding duplicate plugin
	appState.AddPlugin("audio_plugin")
	if len(appState.Plugins) != 2 {
		t.Errorf("Expected 2 plugins after adding duplicate, got %d", len(appState.Plugins))
	}

	// Test RemovePlugin
	removed := appState.RemovePlugin("audio_plugin")
	if !removed {
		t.Error("Expected RemovePlugin to return true")
	}
	if len(appState.Plugins) != 1 {
		t.Errorf("Expected 1 plugin after removal, got %d", len(appState.Plugins))
	}

	// Test removing non-existent plugin
	removed = appState.RemovePlugin("nonexistent")
	if removed {
		t.Error("Expected RemovePlugin to return false for non-existent plugin")
	}
}

// TestAppStateSerialization tests AppState JSON serialization
func TestAppStateSerialization(t *testing.T) {
	original := NewAppState()
	original.CurrentView = "playlist"
	original.User = &User{
		ID:       "u123",
		Username: "testuser",
		Email:    "test@example.com",
		Avatar:   "http://example.com/avatar.jpg",
	}
	original.UpdateConfig("theme", "dark")
	original.AddPlugin("test_plugin")

	// Serialize to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal AppState: %v", err)
	}

	// Deserialize from JSON
	var deserialized AppState
	err = json.Unmarshal(jsonData, &deserialized)
	if err != nil {
		t.Fatalf("Failed to unmarshal AppState: %v", err)
	}

	// Compare fields
	if deserialized.CurrentView != original.CurrentView {
		t.Errorf("CurrentView mismatch: got %s, want %s", deserialized.CurrentView, original.CurrentView)
	}
	if deserialized.User.Username != original.User.Username {
		t.Errorf("User.Username mismatch: got %s, want %s", deserialized.User.Username, original.User.Username)
	}
	if len(deserialized.Plugins) != len(original.Plugins) {
		t.Errorf("Plugins count mismatch: got %d, want %d", len(deserialized.Plugins), len(original.Plugins))
	}
}