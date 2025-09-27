package layouts

import (
	"strings"
	"testing"
	"time"

	ui "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/ui"
	"github.com/go-musicfox/go-musicfox/v2/plugins/tui/themes"
)

func TestNewMainLayout(t *testing.T) {
	width, height := 80, 24
	theme := themes.DefaultTheme
	
	layout := NewMainLayout(width, height, theme)
	
	if layout == nil {
		t.Fatal("NewMainLayout returned nil")
	}
	
	if layout.width != width {
		t.Errorf("Expected width %d, got %d", width, layout.width)
	}
	
	if layout.height != height {
		t.Errorf("Expected height %d, got %d", height, layout.height)
	}
	
	if layout.theme != theme {
		t.Error("Theme not set correctly")
	}
}

func TestMainLayoutRender(t *testing.T) {
	layout := NewMainLayout(80, 24, themes.DefaultTheme)
	
	// Create test app state
	appState := &ui.AppState{
		CurrentView: "main",
		Player: &ui.PlayerState{
			Status: ui.PlayStatusStopped,
			Volume: 0.8,
		},
		Config: map[string]string{
			"selected_index": "0",
			"theme": "default",
			"is_logged_in": "false",
		},
	}
	
	lines, err := layout.Render(appState)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	
	if len(lines) != layout.height {
		t.Errorf("Expected %d lines, got %d", layout.height, len(lines))
	}
	
	// Check that all lines have correct width
	for i, line := range lines {
		if len(line) != layout.width {
			t.Errorf("Line %d has incorrect width: expected %d, got %d", i, layout.width, len(line))
		}
	}
}

func TestMainLayoutRenderWithNilState(t *testing.T) {
	layout := NewMainLayout(80, 24, themes.DefaultTheme)
	
	_, err := layout.Render(nil)
	if err == nil {
		t.Error("Expected error when rendering with nil state")
	}
}

func TestMainLayoutSetSize(t *testing.T) {
	layout := NewMainLayout(80, 24, themes.DefaultTheme)
	
	newWidth, newHeight := 120, 30
	layout.SetSize(newWidth, newHeight)
	
	width, height := layout.GetSize()
	if width != newWidth || height != newHeight {
		t.Errorf("SetSize failed: expected %dx%d, got %dx%d", newWidth, newHeight, width, height)
	}
}

func TestMainLayoutSetTheme(t *testing.T) {
	layout := NewMainLayout(80, 24, themes.DefaultTheme)
	
	newTheme := themes.DarkTheme
	layout.SetTheme(newTheme)
	
	if layout.GetTheme() != newTheme {
		t.Error("SetTheme failed")
	}
}

func TestNewPlayerLayout(t *testing.T) {
	width, height := 80, 24
	theme := themes.DefaultTheme
	
	layout := NewPlayerLayout(width, height, theme)
	
	if layout == nil {
		t.Fatal("NewPlayerLayout returned nil")
	}
	
	if layout.width != width {
		t.Errorf("Expected width %d, got %d", width, layout.width)
	}
	
	if layout.height != height {
		t.Errorf("Expected height %d, got %d", height, layout.height)
	}
}

func TestPlayerLayoutRender(t *testing.T) {
	layout := NewPlayerLayout(80, 24, themes.DefaultTheme)
	
	// Create test app state with current song
	appState := &ui.AppState{
		CurrentView: "player",
		Player: &ui.PlayerState{
			Status: ui.PlayStatusPlaying,
			Volume: 0.75,
			Position: time.Duration(120500000000), // 120.5 seconds in nanoseconds
			Duration: time.Duration(240000000000), // 240 seconds in nanoseconds
			CurrentSong: &ui.Song{
				Title:  "Test Song",
				Artist: "Test Artist",
				Album:  "Test Album",
			},
			IsMuted:     false,
			PlayMode:    ui.PlayModeSequential,
		},
		Config: map[string]string{
			"selected_index": "0",
			"theme": "default",
		},
	}
	
	lines, err := layout.Render(appState)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	
	if len(lines) != layout.height {
		t.Errorf("Expected %d lines, got %d", layout.height, len(lines))
	}
	
	// Check that all lines have correct width
	for i, line := range lines {
		if len(line) != layout.width {
			t.Errorf("Line %d has incorrect width: expected %d, got %d", i, layout.width, len(line))
		}
	}
}

func TestPlayerLayoutRenderWithoutSong(t *testing.T) {
	layout := NewPlayerLayout(80, 24, themes.DefaultTheme)
	
	// Create test app state without current song
	appState := &ui.AppState{
		CurrentView: "player",
		Player: &ui.PlayerState{
			Status:      ui.PlayStatusStopped,
			Volume:      0.5,
			CurrentSong: nil,
		},
		Config: map[string]string{
			"selected_index": "0",
			"theme": "default",
		},
	}
	
	lines, err := layout.Render(appState)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	
	if len(lines) != layout.height {
		t.Errorf("Expected %d lines, got %d", layout.height, len(lines))
	}
}

func TestPlayerLayoutFormatDuration(t *testing.T) {
	layout := NewPlayerLayout(80, 24, themes.DefaultTheme)
	
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0:00"},
		{30, "0:30"},
		{60, "1:00"},
		{90, "1:30"},
		{3600, "1:00:00"},
		{3661, "1:01:01"},
		{-1, "--:--"},
	}
	
	for _, test := range tests {
		result := layout.formatDuration(test.input)
		if result != test.expected {
			t.Errorf("formatDuration(%f) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestNewSearchLayout(t *testing.T) {
	width, height := 80, 24
	theme := themes.DefaultTheme
	
	layout := NewSearchLayout(width, height, theme)
	
	if layout == nil {
		t.Fatal("NewSearchLayout returned nil")
	}
	
	if layout.width != width {
		t.Errorf("Expected width %d, got %d", width, layout.width)
	}
	
	if layout.height != height {
		t.Errorf("Expected height %d, got %d", height, layout.height)
	}
}

func TestSearchLayoutRender(t *testing.T) {
	layout := NewSearchLayout(80, 24, themes.DefaultTheme)
	
	// Create test app state with search results
	appState := &ui.AppState{
		CurrentView: "search",
		Config: map[string]string{
			"search_query":   "test query",
			"search_type":    "song",
			"is_searching":   "false",
			"selected_index": "0",
			"theme":          "default",
		},
	}
	
	lines, err := layout.Render(appState)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	
	if len(lines) != layout.height {
		t.Errorf("Expected %d lines, got %d", layout.height, len(lines))
	}
	
	// Check that all lines have correct width
	for i, line := range lines {
		if len(line) != layout.width {
			t.Errorf("Line %d has incorrect width: expected %d, got %d", i, layout.width, len(line))
		}
	}
}

func TestSearchLayoutRenderEmpty(t *testing.T) {
	layout := NewSearchLayout(80, 24, themes.DefaultTheme)
	
	// Create test app state without search results
	appState := &ui.AppState{
		CurrentView: "search",
		Config: map[string]string{
			"search_query":   "",
			"search_type":    "song",
			"is_searching":   "false",
			"selected_index": "0",
			"theme":          "default",
		},
	}
	
	lines, err := layout.Render(appState)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	
	if len(lines) != layout.height {
		t.Errorf("Expected %d lines, got %d", layout.height, len(lines))
	}
}

func TestSearchLayoutFormatSearchResult(t *testing.T) {
	layout := NewSearchLayout(80, 24, themes.DefaultTheme)
	
	tests := []struct {
		result   interface{}
		expected string
	}{
		{
			"test_song",
			"  ♪ 搜索结果项",
		},
	}
	
	for _, test := range tests {
		result := layout.formatSearchResult(test.result, 0, -1)
		if result != test.expected {
			t.Errorf("formatSearchResult failed: expected %s, got %s", test.expected, result)
		}
	}
}

func TestSearchLayoutFormatSearchResultSelected(t *testing.T) {
	layout := NewSearchLayout(80, 24, themes.DefaultTheme)
	
	song := "test_song"
	result := layout.formatSearchResult(song, 0, 0) // selected
	
	// Should contain cursor
	if !strings.Contains(result, "▶") {
		t.Error("Selected result should contain cursor")
	}
}

// Benchmark tests
func BenchmarkMainLayoutRender(b *testing.B) {
	layout := NewMainLayout(80, 24, themes.DefaultTheme)
	appState := &ui.AppState{
		CurrentView: "main",
		Player: &ui.PlayerState{
			Status: ui.PlayStatusStopped,
			Volume: 0.8,
		},
		User: &ui.User{
			Username: "",
		},
		Config: map[string]string{
			"selected_index": "0",
			"theme":          "default",
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := layout.Render(appState)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

func BenchmarkPlayerLayoutRender(b *testing.B) {
	layout := NewPlayerLayout(80, 24, themes.DefaultTheme)
	appState := &ui.AppState{
		CurrentView: "player",
		Player: &ui.PlayerState{
			Status:   ui.PlayStatusPlaying,
			Volume:   0.75,
			Position: time.Duration(120500000000), // 120.5 seconds
			Duration: time.Duration(240000000000), // 240 seconds
			CurrentSong: &ui.Song{
				Title:  "Test Song",
				Artist: "Test Artist",
				Album:  "Test Album",
			},
		},
		Config: map[string]string{
			"selected_index": "0",
			"theme":          "default",
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := layout.Render(appState)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}

func BenchmarkSearchLayoutRender(b *testing.B) {
	layout := NewSearchLayout(80, 24, themes.DefaultTheme)
	appState := &ui.AppState{
		CurrentView: "search",
		Config: map[string]string{
			"search_query":   "test query",
			"search_type":    "song",
			"is_searching":   "false",
			"selected_index": "0",
			"theme":          "default",
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := layout.Render(appState)
		if err != nil {
			b.Fatalf("Render failed: %v", err)
		}
	}
}