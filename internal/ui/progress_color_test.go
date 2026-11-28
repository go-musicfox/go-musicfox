package ui

import (
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func TestProgressSmoothRampUsesThemeColors(t *testing.T) {
	startColor := lipgloss.Color("#000000")
	endColor := lipgloss.Color("#FFFFFF")
	theme := style.DefaultDarkTheme()
	theme.Custom = configs.AppCustomStyles{
		PlaybarProgressStart: style.Highlight{Fg: startColor},
		PlaybarProgressEnd:   style.Highlight{Fg: endColor},
	}
	style.SetStyleSet(style.NewStyleSet(theme))
	t.Cleanup(func() { style.SetStyleSet(style.DefaultStyleSet()) })

	ramp := progressSmoothRamp(4)
	if len(ramp) != 4 {
		t.Fatalf("ramp length = %d, want 4", len(ramp))
	}
	if !colorsEqual(ramp[0], startColor) {
		t.Fatalf("ramp start = %v, want theme playbarProgressStart %v", ramp[0], startColor)
	}
	if colorsEqual(ramp[0], ramp[3]) {
		t.Fatal("ramp is flat; theme progress colors were not applied")
	}
}

func TestProgressSmoothRampFallsBackWithoutThemeColors(t *testing.T) {
	style.SetStyleSet(style.DefaultStyleSet())

	if ramp := progressSmoothRamp(4); ramp != nil {
		t.Fatalf("ramp = %v, want nil fallback when theme has no progress colors", ramp)
	}
}

func TestProgressRendererRampInvalidatesOnThemeSwitch(t *testing.T) {
	app, netease := newFormPageTestApp(t)

	setProgressColors := func(start, end string) {
		theme := style.DefaultDarkTheme()
		theme.Custom = configs.AppCustomStyles{
			PlaybarProgressStart: style.Highlight{Fg: lipgloss.Color(start)},
			PlaybarProgressEnd:   style.Highlight{Fg: lipgloss.Color(end)},
		}
		style.SetStyleSet(style.NewStyleSet(theme))
	}
	setProgressColors("#FF0000", "#FF0000")
	t.Cleanup(func() { style.SetStyleSet(style.DefaultStyleSet()) })

	song := structs.Song{Name: "缓存歌曲", Duration: time.Minute}
	renderer := NewProgressRenderer(netease, progressTestState{song: song, passed: time.Minute})
	main := app.MustMain()

	view1, _ := renderer.View(app, main)
	cell1 := formStyledScreen(view1).CellAt(0, 0)
	if cell1 == nil || !colorsEqual(cell1.Style.Fg, lipgloss.Color("#FF0000")) {
		t.Fatalf("initial progress cell = %+v, want red ramp", cell1)
	}

	setProgressColors("#00FF00", "#00FF00")
	view2, _ := renderer.View(app, main)
	cell2 := formStyledScreen(view2).CellAt(0, 0)
	if cell2 == nil || !colorsEqual(cell2.Style.Fg, lipgloss.Color("#00FF00")) {
		t.Fatalf("progress cell after theme switch = %+v, want green ramp (stale ramp cache)", cell2)
	}
}
