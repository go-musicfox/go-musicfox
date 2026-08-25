package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
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

// fixedStatusBarComponent is the minimal plugin status bar component: its View
// ignores the App/Main arguments (zero values are fine) and returns a fixed
// string.
type fixedStatusBarComponent struct{}

func (fixedStatusBarComponent) View(*model.App, *model.Main) string {
	return "plugin-component"
}

func TestStatusBarComponentsAppendAfterBuiltIn(t *testing.T) {
	previous := statusBarComponents
	t.Cleanup(func() { statusBarComponents = previous })

	RegisterStatusBarComponent(fixedStatusBarComponent{})

	bar := newQueueQualityStatusBar(nil, func(url string) error { return nil })

	// 内置组件在前，插件组件在后，按注册序显示。
	if got, want := len(bar.Components), 2; got != want {
		t.Fatalf("status bar components = %d, want %d", got, want)
	}
	if _, ok := bar.Components[0].(*queueQualityStatusBarComponent); !ok {
		t.Fatalf("components[0] = %T, want *queueQualityStatusBarComponent", bar.Components[0])
	}
	if got := bar.Components[1].View(nil, nil); got != "plugin-component" {
		t.Fatalf("components[1].View(nil, nil) = %q, want %q", got, "plugin-component")
	}

	// 注册序保持：连续注册两个组件，后注册的排在最后。
	RegisterStatusBarComponent(fixedStatusBarComponent{})
	if got, want := len(bar.Components), 2; got != want {
		t.Fatalf("status bar components after re-register = %d, want %d (bar is a snapshot)", got, want)
	}

	// 快照查询函数反映当前注册集合。
	registered := StatusBarComponents()
	if got, want := len(registered), 2; got != want {
		t.Fatalf("StatusBarComponents() = %d entries, want %d", got, want)
	}
}

func TestRegisterStatusBarComponentNilPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("RegisterStatusBarComponent(nil) did not panic")
		}
	}()
	RegisterStatusBarComponent(nil)
}
