package keybindings

import "testing"

func TestDesktopLyricsBindings(t *testing.T) {
	previous := userOperateToKeys
	t.Cleanup(func() { userOperateToKeys = previous })
	for _, tt := range []struct {
		name string
		user map[string][]string
		key  string
	}{
		{name: "default", key: "ctrl+o"},
		{name: "custom", user: map[string][]string{"toggleDesktopLyrics": {"ctrl+y"}}, key: "ctrl+y"},
		{name: "unbound", user: map[string][]string{"toggleDesktopLyrics": {}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			bindings := BuildEffectiveBindings(ProcessUserBindings(tt.user), true)
			keys := BuildKeyToOperateTypeMap(bindings)
			if tt.key != "" {
				if op, ok := keys[tt.key]; !ok || op.Name() != "toggleDesktopLyrics" {
					t.Fatalf("%s does not toggle desktop lyrics: %v, found=%v", tt.key, op, ok)
				}
			}
			if tt.name != "default" {
				if _, ok := keys["ctrl+o"]; ok {
					t.Fatal("default shortcut is still bound after customization")
				}
			}
			if keys["ctrl+d"] != OpPageDown || keys["ctrl+l"] != OpDownloadPlayingSongLrc {
				t.Fatal("desktop lyrics shortcut replaced an existing operation")
			}
		})
	}
}
