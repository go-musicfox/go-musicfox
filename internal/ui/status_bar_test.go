package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"

	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/playlist"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func TestFormatQueueAndQualityIncludesMusicfox(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Player.SongLevel = service.Lossless
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	player := &Player{playlistManager: playlist.NewPlaylistManager()}
	player.InitSongManager(0, []structs.Song{{Id: 1}, {Id: 2}})

	statusTextStyle := style.CurrentStyleSet().StatusBarText
	want := statusTextStyle.Foreground(util.GetPrimaryColor()).Render("musicfox") +
		statusTextStyle.Render(" · [1/2] · 无损")
	if got := formatQueueAndQuality(player); got != want {
		t.Fatalf("formatQueueAndQuality() = %q, want %q", got, want)
	}
}

func TestQueueQualityStatusBarMusicfoxComponentOpensRepository(t *testing.T) {
	openedURL := ""
	bar := newQueueQualityStatusBar(nil, func(url string) error {
		openedURL = url
		return nil
	})
	if got, want := len(bar.Components), 1; got != want {
		t.Fatalf("status bar components = %d, want %d", got, want)
	}
	component, ok := bar.Components[0].(*queueQualityStatusBarComponent)
	if !ok {
		t.Fatalf("status bar component = %T, want *queueQualityStatusBarComponent", bar.Components[0])
	}
	if !component.IsMouseOver(0, 0) {
		t.Fatal("musicfox prefix is not clickable")
	}
	handled, cmd := component.HandleMouse(tea.Mouse{Button: tea.MouseLeft}, 0, 0)
	if !handled || cmd == nil {
		t.Fatalf("musicfox click handled=%v cmd=%v, want handled command", handled, cmd)
	}
	_ = cmd()
	if got, want := openedURL, types.AppGithubUrl; got != want {
		t.Fatalf("opened URL = %q, want %q", got, want)
	}

	handled, cmd = component.HandleMouse(tea.Mouse{Button: tea.MouseLeft}, len(musicfoxStatusBarLabel), 0)
	if handled || cmd != nil {
		t.Fatalf("adjacent text click handled=%v cmd=%v, want no action", handled, cmd)
	}
}
