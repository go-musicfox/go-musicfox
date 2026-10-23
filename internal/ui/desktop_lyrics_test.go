package ui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/desktop_lyrics"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/player"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func TestToggleDesktopLyricsPersistsAndReusesWindow(t *testing.T) {
	n, window := newDesktopLyricsTestPlayer(t, types.Playing)
	n.desktopLyrics = nil
	path := filepath.Join(t.TempDir(), "config.toml")
	const original = "# User settings\n[main.lyric.desktopLyrics]\nenable = false # Keep this comment\nfontSize = 30\n\n[theme]\nactiveTheme = \"Nord\"\n"
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	created := 0
	create := func(cfg configs.DesktopLyricsConfig) desktop_lyrics.Controller {
		created++
		if !cfg.Enable {
			t.Fatal("window creation received disabled configuration")
		}
		return window
	}
	for _, want := range []bool{true, false, true} {
		enabled, err := n.toggleDesktopLyrics(path, create)
		if err != nil || enabled != want {
			t.Fatalf("toggle = (%v, %v), want (%v, nil)", enabled, err, want)
		}
		if !want && window.visible {
			t.Fatal("disabling did not immediately hide the window")
		}
		n.player.updateDesktopLyrics()
		if window.visible != want || configs.AppConfig.Main.Lyric.DesktopLyrics.Enable != want {
			t.Fatalf("window and runtime enable state did not follow toggle to %v", want)
		}
		reloaded, err := configs.NewConfigFromTomlFile(path)
		if err != nil || reloaded.Main.Lyric.DesktopLyrics.Enable != want {
			t.Fatalf("next startup does not preserve enabled=%v: %v", want, err)
		}
	}
	if created != 1 {
		t.Fatalf("created %d windows, want one reused window", created)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(content), strings.Replace(original, "enable = false", "enable = true", 1); got != want {
		t.Fatalf("toggle changed unrelated configuration: %s", got)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("config permissions changed: %v", err)
	}
}

func TestToggleDesktopLyricsFailureKeepsCurrentState(t *testing.T) {
	for _, tt := range []struct {
		name          string
		enabled       bool
		createFailure bool
	}{
		{name: "creation failure", createFailure: true},
		{name: "save failure while enabling"},
		{name: "save failure while disabling", enabled: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			n, window := newDesktopLyricsTestPlayer(t, types.Playing)
			configs.AppConfig.Main.Lyric.DesktopLyrics.Enable = tt.enabled
			window.visible = tt.enabled
			if !tt.enabled {
				n.desktopLyrics = nil
			}
			path := filepath.Join(t.TempDir(), "missing.toml")
			if tt.createFailure {
				if err := os.WriteFile(path, []byte("[main.lyric.desktopLyrics]\nenable = false\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			enabled, err := n.toggleDesktopLyrics(path, func(configs.DesktopLyricsConfig) desktop_lyrics.Controller {
				if tt.createFailure {
					return nil
				}
				return window
			})
			if err == nil || enabled != tt.enabled || configs.AppConfig.Main.Lyric.DesktopLyrics.Enable != tt.enabled || window.visible != tt.enabled {
				t.Fatalf("failed toggle changed state: enabled=%v visible=%v err=%v", enabled, window.visible, err)
			}
			if tt.createFailure {
				reloaded, err := configs.NewConfigFromTomlFile(path)
				if err != nil || reloaded.Main.Lyric.DesktopLyrics.Enable {
					t.Fatalf("failed creation persisted an enabled flag: %v", err)
				}
			}
		})
	}
}

func TestToggleDesktopLyricsDuringPlaybackUpdates(t *testing.T) {
	n, window := newDesktopLyricsTestPlayer(t, types.Playing)
	path := filepath.Join(t.TempDir(), "config.toml")
	// Older configurations may not contain the desktopLyrics table yet.
	if err := os.WriteFile(path, []byte("[main]\nframeRate = 30\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var updates sync.WaitGroup
	updates.Go(func() {
		for range 100 {
			n.player.updateDesktopLyrics()
		}
	})
	for range 10 {
		if _, err := n.toggleDesktopLyrics(path, nil); err != nil {
			t.Error(err)
			break
		}
	}
	updates.Wait()
	n.player.updateDesktopLyrics()
	if window.visible || configs.AppConfig.Main.Lyric.DesktopLyrics.Enable {
		t.Fatal("playback updates reopened disabled desktop lyrics")
	}
	reloaded, err := configs.NewConfigFromTomlFile(path)
	if err != nil || reloaded.Main.Lyric.DesktopLyrics.Enable {
		t.Fatalf("failed to append the setting to an older configuration: %v", err)
	}
}

func TestUpdateDesktopLyricsVisibility(t *testing.T) {
	for _, tt := range []struct {
		name        string
		state       types.State
		enabled     bool
		hideOnPause bool
		visible     bool
	}{
		{name: "playing", state: types.Playing, enabled: true, visible: true},
		{name: "disabled while playing", state: types.Playing},
		{name: "enabled while paused", state: types.Paused, enabled: true, visible: true},
		{name: "hide on pause", state: types.Paused, enabled: true, hideOnPause: true},
		{name: "stopped", state: types.Stopped, enabled: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			n, window := newDesktopLyricsTestPlayer(t, tt.state)
			configs.AppConfig.Main.Lyric.DesktopLyrics.Enable = tt.enabled
			configs.AppConfig.Main.Lyric.DesktopLyrics.HideOnPause = tt.hideOnPause
			n.player.updateDesktopLyrics()
			if window.visible != tt.visible {
				t.Errorf("visible=%v, want %v", window.visible, tt.visible)
			}
			if !tt.enabled && window.updates != 0 {
				t.Fatal("disabled desktop lyrics still received playback updates")
			}
			if tt.visible && window.line != "current lyric" {
				t.Errorf("displayed lyric=%q, want current lyric", window.line)
			}
		})
	}
}

// Replace only the native window and audio device; lyric parsing remains real.
func newDesktopLyricsTestPlayer(t *testing.T, state types.State) (*Netease, *desktopLyricsTestWindow) {
	t.Helper()
	previous := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previous })
	service := lyric.NewService(desktopLyricsTestFetcher{}, false, 0, false)
	if err := service.SetSong(context.Background(), structs.Song{Id: 1}); err != nil {
		t.Fatal(err)
	}
	service.UpdatePosition(time.Second)
	window := &desktopLyricsTestWindow{}
	n := &Netease{desktopLyrics: window, lyricService: service}
	n.player = &Player{netease: n, Player: desktopLyricsTestPlayer{state: state}}
	return n, window
}

type desktopLyricsTestFetcher struct{}

func (desktopLyricsTestFetcher) GetLyric(context.Context, structs.Song) (structs.LRCData, error) {
	return structs.LRCData{Original: "[00:00.00]current lyric\n[00:10.00]next lyric"}, nil
}

type desktopLyricsTestPlayer struct {
	player.Player
	state types.State
}

func (p desktopLyricsTestPlayer) State() types.State      { return p.state }
func (desktopLyricsTestPlayer) PassedTime() time.Duration { return time.Second }

type desktopLyricsTestWindow struct {
	desktop_lyrics.Controller
	visible bool
	updates int
	line    string
}

func (w *desktopLyricsTestWindow) Show()           { w.visible = true }
func (w *desktopLyricsTestWindow) Hide()           { w.visible = false }
func (w *desktopLyricsTestWindow) IsVisible() bool { return w.visible }
func (w *desktopLyricsTestWindow) Update(line, _ desktop_lyrics.LyricLine, _ int, _ int64, _ bool) {
	w.updates++
	w.line = line.Text
}
func (*desktopLyricsTestWindow) SetSpectrumAvailable(bool) {}
