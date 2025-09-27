package handlers

import (
	"context"
	"testing"
	"time"

	ui "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/ui"
)

func TestNewPlaybackEventHandler(t *testing.T) {
	handler := NewPlaybackEventHandler()
	
	if handler == nil {
		t.Fatal("NewPlaybackEventHandler returned nil")
	}
	
	if handler.GetPriority() != 100 {
		t.Errorf("Expected priority 100, got %d", handler.GetPriority())
	}
}

func TestPlaybackEventHandlerCanHandle(t *testing.T) {
	handler := NewPlaybackEventHandler()
	
	tests := []struct {
		eventType string
		expected  bool
	}{
		{"play", true},
		{"pause", true},
		{"stop", true},
		{"next", true},
		{"previous", true},
		{"shuffle", true},
		{"repeat", true},
		{"seek", true},
		{"volume_change", true},
		{"unknown", false},
		{"navigate_up", false},
	}
	
	for _, test := range tests {
		event := &ui.UIEvent{Type: test.eventType}
		result := handler.CanHandle(event)
		if result != test.expected {
			t.Errorf("CanHandle(%s) = %v, expected %v", test.eventType, result, test.expected)
		}
	}
}

func TestPlaybackEventHandlerHandlePlay(t *testing.T) {
	handler := NewPlaybackEventHandler()
	ctx := context.Background()
	
	state := &ui.AppState{
		Player: &ui.PlayerState{
			Status: ui.PlayStatusStopped,
		},
	}
	
	event := &ui.UIEvent{
		Type: "play",
		Data: map[string]interface{}{},
	}
	
	err := handler.Handle(ctx, event, state)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	
	if state.Player.Status != ui.PlayStatusPlaying {
		t.Error("Expected Status to be Playing after play event")
	}
}

func TestPlaybackEventHandlerHandlePause(t *testing.T) {
	handler := NewPlaybackEventHandler()
	ctx := context.Background()
	
	state := &ui.AppState{
		Player: &ui.PlayerState{
			Status: ui.PlayStatusPlaying,
		},
	}
	
	event := &ui.UIEvent{
		Type: "pause",
		Data: map[string]interface{}{},
	}
	
	err := handler.Handle(ctx, event, state)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	
	if state.Player.Status != ui.PlayStatusPaused {
		t.Error("Expected Status to be Paused after pause event")
	}
}

func TestPlaybackEventHandlerHandleVolumeChange(t *testing.T) {
	handler := NewPlaybackEventHandler()
	ctx := context.Background()
	
	state := &ui.AppState{
		Player: &ui.PlayerState{
			Volume: 0.5,
		},
	}
	
	event := &ui.UIEvent{
		Type: "volume_change",
		Data: map[string]interface{}{
			"volume": 0.75,
		},
	}
	
	err := handler.Handle(ctx, event, state)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	
	if state.Player.Volume != 0.75 {
		t.Errorf("Expected volume 0.75, got %f", state.Player.Volume)
	}
}

func TestPlaybackEventHandlerHandleNext(t *testing.T) {
	handler := NewPlaybackEventHandler()
	ctx := context.Background()
	
	state := &ui.AppState{
		Player: &ui.PlayerState{
			PlayMode: ui.PlayModeSequential,
		},
		Config: map[string]string{
			"playlist_songs": "Song 1,Song 2,Song 3",
			"current_index": "0",
		},
	}
	
	event := &ui.UIEvent{
		Type: "next",
		Data: map[string]interface{}{},
	}
	
	err := handler.Handle(ctx, event, state)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	
	if currentIndexStr, ok := state.Config["current_index"]; !ok || currentIndexStr != "1" {
		t.Errorf("Expected current index '1', got %v", state.Config["current_index"])
	}
}

func TestNewNavigationEventHandler(t *testing.T) {
	handler := NewNavigationEventHandler()
	
	if handler == nil {
		t.Fatal("NewNavigationEventHandler returned nil")
	}
	
	if handler.GetPriority() != 90 {
		t.Errorf("Expected priority 90, got %d", handler.GetPriority())
	}
}

func TestNavigationEventHandlerCanHandle(t *testing.T) {
	handler := NewNavigationEventHandler()
	
	tests := []struct {
		eventType string
		expected  bool
	}{
		{"navigate_up", true},
		{"navigate_down", true},
		{"navigate_left", true},
		{"navigate_right", true},
		{"page_up", true},
		{"page_down", true},
		{"home", true},
		{"end", true},
		{"select", true},
		{"back", true},
		{"view_change", true},
		{"play", false},
		{"unknown", false},
	}
	
	for _, test := range tests {
		event := &ui.UIEvent{Type: test.eventType}
		result := handler.CanHandle(event)
		if result != test.expected {
			t.Errorf("CanHandle(%s) = %v, expected %v", test.eventType, result, test.expected)
		}
	}
}

func TestNavigationEventHandlerHandleNavigateUp(t *testing.T) {
	handler := NewNavigationEventHandler()
	ctx := context.Background()
	
	state := &ui.AppState{
		Config: map[string]string{
			"selected_index": "5",
		},
	}
	
	event := &ui.UIEvent{
		Type: "navigate_up",
		Data: map[string]interface{}{},
	}
	
	err := handler.Handle(ctx, event, state)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	
	if selectedIndexStr, ok := state.Config["selected_index"]; !ok || selectedIndexStr != "4" {
		t.Errorf("Expected selected index '4', got %v", state.Config["selected_index"])
	}
}

func TestNavigationEventHandlerHandleNavigateDown(t *testing.T) {
	handler := NewNavigationEventHandler()
	ctx := context.Background()
	
	state := &ui.AppState{
		Config: map[string]string{
			"selected_index": "2",
		},
	}
	
	event := &ui.UIEvent{
		Type: "navigate_down",
		Data: map[string]interface{}{
			"max_index": 10,
		},
	}
	
	err := handler.Handle(ctx, event, state)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	
	if selectedIndexStr, ok := state.Config["selected_index"]; !ok || selectedIndexStr != "3" {
		t.Errorf("Expected selected index '3', got %v", state.Config["selected_index"])
	}
}

func TestNavigationEventHandlerHandleHome(t *testing.T) {
	handler := NewNavigationEventHandler()
	ctx := context.Background()
	
	state := &ui.AppState{
		Config: map[string]string{
			"selected_index": "10",
		},
	}
	
	event := &ui.UIEvent{
		Type: "home",
		Data: map[string]interface{}{},
	}
	
	err := handler.Handle(ctx, event, state)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	
	if selectedIndexStr, ok := state.Config["selected_index"]; !ok || selectedIndexStr != "0" {
		t.Errorf("Expected selected index '0', got %v", state.Config["selected_index"])
	}
}

func TestNewSearchEventHandler(t *testing.T) {
	handler := NewSearchEventHandler()
	
	if handler == nil {
		t.Fatal("NewSearchEventHandler returned nil")
	}
	
	if handler.GetPriority() != 80 {
		t.Errorf("Expected priority 80, got %d", handler.GetPriority())
	}
}

func TestSearchEventHandlerHandleSearch(t *testing.T) {
	handler := NewSearchEventHandler()
	ctx := context.Background()
	
	state := &ui.AppState{
		Config: map[string]string{
			"search_history": "",
			"search_query": "",
			"is_searching": "false",
		},
	}
	
	event := &ui.UIEvent{
		Type: "search",
		Data: map[string]interface{}{
			"query": "test query",
		},
	}
	
	err := handler.Handle(ctx, event, state)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	
	if query, ok := state.Config["search_query"]; !ok || query != "test query" {
		t.Errorf("Expected query 'test query', got '%v'", state.Config["search_query"])
	}
	
	if isSearching, ok := state.Config["is_searching"]; !ok || isSearching != "true" {
		t.Error("Expected IsSearching to be 'true'")
	}
	
	if history, ok := state.Config["search_history"]; !ok || history != "test query" {
		t.Error("Query not added to history")
	}
}

func TestNewThemeEventHandler(t *testing.T) {
	handler := NewThemeEventHandler()
	
	if handler == nil {
		t.Fatal("NewThemeEventHandler returned nil")
	}
	
	if handler.GetPriority() != 70 {
		t.Errorf("Expected priority 70, got %d", handler.GetPriority())
	}
}

func TestThemeEventHandlerHandleThemeChange(t *testing.T) {
	handler := NewThemeEventHandler()
	ctx := context.Background()
	
	state := &ui.AppState{
		Config: map[string]string{
			"theme": "default",
		},
	}
	
	event := &ui.UIEvent{
		Type: "theme_change",
		Data: map[string]interface{}{
			"theme": "dark",
		},
	}
	
	err := handler.Handle(ctx, event, state)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	
	if theme, ok := state.Config["theme"]; !ok || theme != "dark" {
		t.Errorf("Expected theme 'dark', got '%s'", state.Config["theme"])
	}
}

func TestThemeEventHandlerHandleThemeToggle(t *testing.T) {
	handler := NewThemeEventHandler()
	ctx := context.Background()
	
	state := &ui.AppState{
		Config: map[string]string{
			"theme": "default",
		},
	}
	
	event := &ui.UIEvent{
		Type: "theme_toggle",
		Data: map[string]interface{}{},
	}
	
	err := handler.Handle(ctx, event, state)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	
	if theme, ok := state.Config["theme"]; !ok || theme == "default" {
		t.Error("Theme should have changed from default")
	}
}

func TestNewEventDispatcher(t *testing.T) {
	dispatcher := NewEventDispatcher()
	
	if dispatcher == nil {
		t.Fatal("NewEventDispatcher returned nil")
	}
	
	if len(dispatcher.handlers) != 0 {
		t.Error("New dispatcher should have no handlers")
	}
}

func TestEventDispatcherRegisterHandler(t *testing.T) {
	dispatcher := NewEventDispatcher()
	handler := NewPlaybackEventHandler()
	
	dispatcher.RegisterHandler(handler)
	
	if len(dispatcher.handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(dispatcher.handlers))
	}
	
	if dispatcher.handlers[0] != handler {
		t.Error("Handler not registered correctly")
	}
}

func TestEventDispatcherRegisterMultipleHandlers(t *testing.T) {
	dispatcher := NewEventDispatcher()
	playbackHandler := NewPlaybackEventHandler()   // priority 100
	navigationHandler := NewNavigationEventHandler() // priority 90
	searchHandler := NewSearchEventHandler()         // priority 80
	
	// Register in reverse priority order
	dispatcher.RegisterHandler(searchHandler)
	dispatcher.RegisterHandler(navigationHandler)
	dispatcher.RegisterHandler(playbackHandler)
	
	if len(dispatcher.handlers) != 3 {
		t.Errorf("Expected 3 handlers, got %d", len(dispatcher.handlers))
	}
	
	// Check that handlers are sorted by priority (highest first)
	if dispatcher.handlers[0] != playbackHandler {
		t.Error("Playback handler should be first (highest priority)")
	}
	if dispatcher.handlers[1] != navigationHandler {
		t.Error("Navigation handler should be second")
	}
	if dispatcher.handlers[2] != searchHandler {
		t.Error("Search handler should be third (lowest priority)")
	}
}

func TestEventDispatcherDispatchEvent(t *testing.T) {
	dispatcher := NewEventDispatcher()
	handler := NewPlaybackEventHandler()
	dispatcher.RegisterHandler(handler)
	
	ctx := context.Background()
	state := &ui.AppState{
		Player: &ui.PlayerState{
			Status: ui.PlayStatusStopped,
		},
	}
	
	event := &ui.UIEvent{
		Type: "play",
		Data: map[string]interface{}{},
	}
	
	err := dispatcher.DispatchEvent(ctx, event, state)
	if err != nil {
		t.Fatalf("DispatchEvent failed: %v", err)
	}
	
	if state.Player.Status != ui.PlayStatusPlaying {
		t.Error("Event was not handled correctly")
	}
}

func TestEventDispatcherDispatchUnknownEvent(t *testing.T) {
	dispatcher := NewEventDispatcher()
	handler := NewPlaybackEventHandler()
	dispatcher.RegisterHandler(handler)
	
	ctx := context.Background()
	state := &ui.AppState{}
	
	event := &ui.UIEvent{
		Type: "unknown_event",
		Data: map[string]interface{}{},
	}
	
	err := dispatcher.DispatchEvent(ctx, event, state)
	if err == nil {
		t.Error("Expected error for unknown event")
	}
}

func TestEventDispatcherRemoveHandler(t *testing.T) {
	dispatcher := NewEventDispatcher()
	handler := NewPlaybackEventHandler()
	dispatcher.RegisterHandler(handler)
	
	if len(dispatcher.handlers) != 1 {
		t.Error("Handler not registered")
	}
	
	dispatcher.RemoveHandler(handler)
	
	if len(dispatcher.handlers) != 0 {
		t.Error("Handler not removed")
	}
}

func TestCreateDefaultEventDispatcher(t *testing.T) {
	dispatcher := CreateDefaultEventDispatcher()
	
	if dispatcher == nil {
		t.Fatal("CreateDefaultEventDispatcher returned nil")
	}
	
	if len(dispatcher.handlers) == 0 {
		t.Error("Default dispatcher should have handlers")
	}
	
	// Check that all default handlers are registered
	expectedHandlerTypes := []string{"playback", "navigation", "search", "theme"}
	handlerTypes := make(map[string]bool)
	
	for _, handler := range dispatcher.handlers {
		switch handler.(type) {
		case *PlaybackEventHandler:
			handlerTypes["playback"] = true
		case *NavigationEventHandler:
			handlerTypes["navigation"] = true
		case *SearchEventHandler:
			handlerTypes["search"] = true
		case *ThemeEventHandler:
			handlerTypes["theme"] = true
		}
	}
	
	for _, expectedType := range expectedHandlerTypes {
		if !handlerTypes[expectedType] {
			t.Errorf("Missing %s handler in default dispatcher", expectedType)
		}
	}
}

// Benchmark tests
func BenchmarkPlaybackEventHandlerHandle(b *testing.B) {
	handler := NewPlaybackEventHandler()
	ctx := context.Background()
	
	state := &ui.AppState{
		Player: &ui.PlayerState{
			Status: ui.PlayStatusStopped,
		},
	}
	
	event := &ui.UIEvent{
		Type: "play",
		Data: map[string]interface{}{},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := handler.Handle(ctx, event, state)
		if err != nil {
			b.Fatalf("Handle failed: %v", err)
		}
		state.Player.Status = ui.PlayStatusStopped // reset for next iteration
	}
}

func BenchmarkEventDispatcherDispatchEvent(b *testing.B) {
	dispatcher := CreateDefaultEventDispatcher()
	ctx := context.Background()
	
	state := &ui.AppState{
		Player: &ui.PlayerState{
			Status: ui.PlayStatusStopped,
		},
	}
	
	event := &ui.UIEvent{
		Type:      "play",
		Data:      map[string]interface{}{},
		Timestamp: time.Now(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := dispatcher.DispatchEvent(ctx, event, state)
		if err != nil {
			b.Fatalf("DispatchEvent failed: %v", err)
		}
		state.Player.Status = ui.PlayStatusStopped // reset for next iteration
	}
}