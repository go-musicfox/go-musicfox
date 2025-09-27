package handlers

import (
	"strings"
	"testing"
)

func TestNewKeyMappingManager(t *testing.T) {
	manager := NewKeyMappingManager()
	
	if manager == nil {
		t.Fatal("NewKeyMappingManager returned nil")
	}
	
	if len(manager.mappings) == 0 {
		t.Error("Manager should have default mappings")
	}
	
	// Check that default contexts exist
	expectedContexts := []string{"global", "main", "player", "search", "playlist"}
	for _, context := range expectedContexts {
		if _, exists := manager.mappings[context]; !exists {
			t.Errorf("Missing default context: %s", context)
		}
	}
}

func TestKeyMappingManagerFindAction(t *testing.T) {
	manager := NewKeyMappingManager()
	
	tests := []struct {
		key       string
		modifiers []string
		context   string
		expected  string
		found     bool
	}{
		{"q", []string{}, "global", "quit", true},
		{"h", []string{}, "global", "help", true},
		{"k", []string{}, "global", "navigate_up", true},
		{"j", []string{}, "global", "navigate_down", true},
		{" ", []string{}, "global", "toggle_play", true},
		{"c", []string{"ctrl"}, "global", "quit", true},
		{"unknown", []string{}, "global", "", false},
		{"/", []string{}, "main", "search", true},
		{"L", []string{}, "player", "toggle_lyrics", true},
		{"tab", []string{}, "search", "search_type_next", true},
		{"d", []string{}, "playlist", "remove_song", true},
	}
	
	for _, test := range tests {
		action, found := manager.FindAction(test.key, test.modifiers, test.context)
		if found != test.found {
			t.Errorf("FindAction(%s, %v, %s) found = %v, expected %v", test.key, test.modifiers, test.context, found, test.found)
		}
		if found && action != test.expected {
			t.Errorf("FindAction(%s, %v, %s) = %s, expected %s", test.key, test.modifiers, test.context, action, test.expected)
		}
	}
}

func TestKeyMappingManagerFindActionGlobalFallback(t *testing.T) {
	manager := NewKeyMappingManager()
	
	// Test that global mappings work in specific contexts
	action, found := manager.FindAction("q", []string{}, "main")
	if !found {
		t.Error("Global mapping should be found in specific context")
	}
	if action != "quit" {
		t.Errorf("Expected action 'quit', got '%s'", action)
	}
}

func TestKeyMappingManagerNormalizeKey(t *testing.T) {
	manager := NewKeyMappingManager()
	
	tests := []struct {
		input    string
		expected string
	}{
		{"Q", "q"},
		{"SPACE", " "},
		{"space", " "},
		{"RETURN", "enter"},
		{"return", "enter"},
		{"ESCAPE", "esc"},
		{"escape", "esc"},
		{"UP", "↑"},
		{"up", "↑"},
		{"DOWN", "↓"},
		{"LEFT", "←"},
		{"RIGHT", "→"},
		{"PAGE_UP", "pgup"},
		{"PAGE_DOWN", "pgdn"},
		{"a", "a"},
		{"1", "1"},
	}
	
	for _, test := range tests {
		result := manager.normalizeKey(test.input)
		if result != test.expected {
			t.Errorf("normalizeKey(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestKeyMappingManagerNormalizeModifiers(t *testing.T) {
	manager := NewKeyMappingManager()
	
	input := []string{"CTRL", "SHIFT", "Alt"}
	expected := []string{"ctrl", "shift", "alt"}
	
	result := manager.normalizeModifiers(input)
	
	if len(result) != len(expected) {
		t.Errorf("Expected %d modifiers, got %d", len(expected), len(result))
	}
	
	for i, mod := range result {
		if mod != expected[i] {
			t.Errorf("normalizeModifiers[%d] = %s, expected %s", i, mod, expected[i])
		}
	}
}

func TestKeyMappingManagerGetMappings(t *testing.T) {
	manager := NewKeyMappingManager()
	
	// Test existing context
	globalMappings := manager.GetMappings("global")
	if len(globalMappings) == 0 {
		t.Error("Global mappings should not be empty")
	}
	
	// Test non-existing context
	unknownMappings := manager.GetMappings("unknown")
	if len(unknownMappings) != 0 {
		t.Error("Unknown context should return empty mappings")
	}
}

func TestKeyMappingManagerGetAllMappings(t *testing.T) {
	manager := NewKeyMappingManager()
	
	allMappings := manager.GetAllMappings()
	
	if len(allMappings) == 0 {
		t.Error("Should have mappings")
	}
	
	expectedContexts := []string{"global", "main", "player", "search", "playlist"}
	for _, context := range expectedContexts {
		if _, exists := allMappings[context]; !exists {
			t.Errorf("Missing context in all mappings: %s", context)
		}
	}
}

func TestKeyMappingManagerAddMapping(t *testing.T) {
	manager := NewKeyMappingManager()
	
	newMapping := KeyMapping{
		Key:         "x",
		Action:      "test_action",
		Context:     "test_context",
		Description: "Test mapping",
	}
	
	manager.AddMapping(newMapping)
	
	mappings := manager.GetMappings("test_context")
	if len(mappings) != 1 {
		t.Errorf("Expected 1 mapping in test_context, got %d", len(mappings))
	}
	
	if mappings[0].Action != "test_action" {
		t.Errorf("Expected action 'test_action', got '%s'", mappings[0].Action)
	}
}

func TestKeyMappingManagerRemoveMapping(t *testing.T) {
	manager := NewKeyMappingManager()
	
	// Remove existing mapping
	removed := manager.RemoveMapping("global", "q", []string{})
	if !removed {
		t.Error("Should have removed existing mapping")
	}
	
	// Verify it's removed
	_, found := manager.FindAction("q", []string{}, "global")
	if found {
		t.Error("Mapping should have been removed")
	}
	
	// Try to remove non-existing mapping
	removed = manager.RemoveMapping("global", "nonexistent", []string{})
	if removed {
		t.Error("Should not have removed non-existing mapping")
	}
}

func TestKeyMappingManagerUpdateMapping(t *testing.T) {
	manager := NewKeyMappingManager()
	
	// Update existing mapping
	updated := manager.UpdateMapping("global", "q", []string{}, "new_quit", "New quit action")
	if !updated {
		t.Error("Should have updated existing mapping")
	}
	
	// Verify it's updated
	action, found := manager.FindAction("q", []string{}, "global")
	if !found {
		t.Error("Updated mapping should still be found")
	}
	if action != "new_quit" {
		t.Errorf("Expected action 'new_quit', got '%s'", action)
	}
	
	// Try to update non-existing mapping
	updated = manager.UpdateMapping("global", "nonexistent", []string{}, "action", "desc")
	if updated {
		t.Error("Should not have updated non-existing mapping")
	}
}

func TestKeyMappingManagerLoadMappingsFromConfig(t *testing.T) {
	manager := NewKeyMappingManager()
	
	config := map[string]interface{}{
		"test_context": []interface{}{
			map[string]interface{}{
				"key":         "t",
				"action":      "test_action",
				"description": "Test action",
				"modifiers":   []interface{}{"ctrl"},
			},
		},
	}
	
	err := manager.LoadMappingsFromConfig(config)
	if err != nil {
		t.Fatalf("LoadMappingsFromConfig failed: %v", err)
	}
	
	action, found := manager.FindAction("t", []string{"ctrl"}, "test_context")
	if !found {
		t.Error("Loaded mapping not found")
	}
	if action != "test_action" {
		t.Errorf("Expected action 'test_action', got '%s'", action)
	}
}

func TestKeyMappingManagerExportMappingsToConfig(t *testing.T) {
	manager := NewKeyMappingManager()
	
	// Add a test mapping
	testMapping := KeyMapping{
		Key:         "t",
		Action:      "test_action",
		Context:     "test_context",
		Description: "Test action",
		Modifiers:   []string{"ctrl"},
	}
	manager.AddMapping(testMapping)
	
	config := manager.ExportMappingsToConfig()
	
	if len(config) == 0 {
		t.Error("Exported config should not be empty")
	}
	
	testContextConfig, exists := config["test_context"]
	if !exists {
		t.Error("test_context should exist in exported config")
	}
	
	mappingsConfig, ok := testContextConfig.([]interface{})
	if !ok {
		t.Error("test_context should be a slice of mappings")
	}
	
	if len(mappingsConfig) != 1 {
		t.Errorf("Expected 1 mapping in test_context, got %d", len(mappingsConfig))
	}
}

func TestKeyMappingManagerGetHelpText(t *testing.T) {
	manager := NewKeyMappingManager()
	
	helpText := manager.GetHelpText("global")
	
	if len(helpText) == 0 {
		t.Error("Help text should not be empty")
	}
	
	// Check that it contains a title
	if len(helpText) < 2 || !strings.Contains(helpText[0], "全局快捷键") {
		t.Error("Help text should contain context title")
	}
}

func TestKeyMappingManagerGetAllHelpText(t *testing.T) {
	manager := NewKeyMappingManager()
	
	allHelpText := manager.GetAllHelpText()
	
	if len(allHelpText) == 0 {
		t.Error("All help text should not be empty")
	}
	
	// Check that it contains help for all contexts
	contexts := []string{"全局快捷键", "主菜单快捷键", "播放器快捷键", "搜索快捷键", "播放列表快捷键"}
	helpTextStr := strings.Join(allHelpText, "\n")
	
	for _, context := range contexts {
		if !strings.Contains(helpTextStr, context) {
			t.Errorf("All help text should contain %s", context)
		}
	}
}

func TestKeyMappingManagerValidateMapping(t *testing.T) {
	manager := NewKeyMappingManager()
	
	// Valid mapping
	validMapping := KeyMapping{
		Key:         "t",
		Action:      "test_action",
		Context:     "test_context",
		Description: "Test action",
	}
	
	err := manager.ValidateMapping(validMapping)
	if err != nil {
		t.Errorf("Valid mapping should not return error: %v", err)
	}
	
	// Invalid mappings
	invalidMappings := []KeyMapping{
		{Key: "", Action: "action", Context: "context"}, // empty key
		{Key: "k", Action: "", Context: "context"},      // empty action
		{Key: "k", Action: "action", Context: ""},        // empty context
	}
	
	for i, mapping := range invalidMappings {
		err := manager.ValidateMapping(mapping)
		if err == nil {
			t.Errorf("Invalid mapping %d should return error", i)
		}
	}
}

func TestKeyMappingManagerValidateMappingConflict(t *testing.T) {
	manager := NewKeyMappingManager()
	
	// Try to add a mapping that conflicts with existing one
	conflictMapping := KeyMapping{
		Key:         "q", // conflicts with existing quit mapping
		Action:      "different_action",
		Context:     "global",
		Description: "Different action",
	}
	
	err := manager.ValidateMapping(conflictMapping)
	if err == nil {
		t.Error("Conflicting mapping should return error")
	}
}

func TestKeyMappingManagerMatchesMapping(t *testing.T) {
	manager := NewKeyMappingManager()
	
	mapping := KeyMapping{
		Key:       "t",
		Modifiers: []string{"ctrl", "shift"},
	}
	
	tests := []struct {
		key       string
		modifiers []string
		expected  bool
	}{
		{"t", []string{"ctrl", "shift"}, true},
		{"t", []string{"shift", "ctrl"}, true}, // order doesn't matter
		{"t", []string{"ctrl"}, false},          // missing modifier
		{"t", []string{"ctrl", "shift", "alt"}, false}, // extra modifier
		{"T", []string{"ctrl", "shift"}, true},  // case insensitive
		{"s", []string{"ctrl", "shift"}, false}, // different key
	}
	
	for _, test := range tests {
		normalizedKey := manager.normalizeKey(test.key)
		normalizedModifiers := manager.normalizeModifiers(test.modifiers)
		result := manager.matchesMapping(normalizedKey, normalizedModifiers, mapping)
		if result != test.expected {
			t.Errorf("matchesMapping(%s, %v) = %v, expected %v", test.key, test.modifiers, result, test.expected)
		}
	}
}

// Benchmark tests
func BenchmarkKeyMappingManagerFindAction(b *testing.B) {
	manager := NewKeyMappingManager()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.FindAction("q", []string{}, "global")
	}
}

func BenchmarkKeyMappingManagerFindActionWithModifiers(b *testing.B) {
	manager := NewKeyMappingManager()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.FindAction("c", []string{"ctrl"}, "global")
	}
}

func BenchmarkKeyMappingManagerNormalizeKey(b *testing.B) {
	manager := NewKeyMappingManager()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.normalizeKey("SPACE")
	}
}