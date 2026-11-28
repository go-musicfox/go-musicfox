package configs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTmuxCoverSyncDefaultAndOverride(t *testing.T) {
	previous := EffectiveKeybindings
	defer func() { EffectiveKeybindings = previous }()
	for _, tc := range []struct {
		name, toml string
		want       bool
	}{
		{"old config", "[main.lyric.cover]\nshow = true\ntmuxPassthrough = true\n", false},
		{"explicit enable", "[main.lyric.cover]\ntmuxSyncOutput = true\n", true},
		{"explicit disable", "[main.lyric.cover]\ntmuxSyncOutput = false\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.toml")
			if err := os.WriteFile(path, []byte(tc.toml), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := NewConfigFromTomlFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Main.Lyric.Cover.TmuxSyncOutput != tc.want {
				t.Fatalf("sync output = %v, want %v", cfg.Main.Lyric.Cover.TmuxSyncOutput, tc.want)
			}
		})
	}
}
