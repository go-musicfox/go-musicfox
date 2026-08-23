package ui

import (
	"image/color"
	"testing"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

func TestCustomPagesFillExplicitAppBackground(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	appBackground := lipgloss.Color("#FAF0C8")
	withAppBackgroundTheme(t, appBackground)

	login := NewLoginPage(netease)
	pages := map[string]model.Page{
		"search":             NewSearchPage(netease),
		"login":              login,
		"qr-login":           NewQRLoginPage(netease, login, nil),
		"lastfm-auth":        NewLastfmAuthPage(newMenuServices(netease)),
		"lastfm-qr-auth":     NewLastfmQRAuthPage(newMenuServices(netease), login, nil),
		"lastfm-api-account": newLastfmCustomAPIPageForBackgroundTest(netease),
	}
	for name, page := range pages {
		t.Run(name, func(t *testing.T) {
			assertPageBackground(t, page.View(app), app.WindowWidth(), app.WindowHeight(), appBackground)
		})
	}
}

func TestCustomPagesLeaveTransparentAppBackgroundUnpainted(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	login := NewLoginPage(netease)
	pages := map[string]model.Page{
		"search":             NewSearchPage(netease),
		"login":              login,
		"qr-login":           NewQRLoginPage(netease, login, nil),
		"lastfm-auth":        NewLastfmAuthPage(newMenuServices(netease)),
		"lastfm-qr-auth":     NewLastfmQRAuthPage(newMenuServices(netease), login, nil),
		"lastfm-api-account": newLastfmCustomAPIPageForBackgroundTest(netease),
	}
	for name, page := range pages {
		t.Run(name, func(t *testing.T) {
			screen := customPageStyledScreen(page.View(app), app.WindowWidth(), app.WindowHeight())
			cell := screen.CellAt(app.WindowWidth()-1, app.WindowHeight()-1)
			if cell == nil {
				t.Fatal("bottom-right cell is missing")
			}
			if cell.Style.Bg != nil {
				t.Fatalf("bottom-right background = %#v, want unset", cell.Style.Bg)
			}
		})
	}
}

func newLastfmCustomAPIPageForBackgroundTest(netease *Netease) *LastfmCustomAPIPage {
	keyInput := textinput.New()
	keyInput.Placeholder = " Key"
	secretInput := textinput.New()
	secretInput.Placeholder = " Secret"

	return &LastfmCustomAPIPage{
		svc:          newMenuServices(netease),
		menuTitle:    &model.MenuItem{Title: "Lastfm API account"},
		keyInput:     keyInput,
		secretInput:  secretInput,
		submitButton: util.GetBlurredSubmitButton(),
		reloadButton: util.GetBlurredButton("重载"),
		clearButton:  util.GetBlurredButton("清空"),
		reloadText:   "重载",
		clearText:    "清空",
	}
}

func assertPageBackground(t *testing.T, content string, width, height int, want color.Color) {
	t.Helper()
	screen := customPageStyledScreen(content, width, height)
	wantR, wantG, wantB, wantA := want.RGBA()
	for y := range height {
		for x := range width {
			cell := screen.CellAt(x, y)
			if cell == nil || cell.Width == 0 {
				continue
			}
			if cell.Style.Bg == nil {
				t.Fatalf("cell (%d,%d) content=%q has no app background", x, y, cell.Content)
			}
			gotR, gotG, gotB, gotA := cell.Style.Bg.RGBA()
			if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
				t.Fatalf("cell (%d,%d) background = %v, want app background %v", x, y, cell.Style.Bg, want)
			}
		}
	}
}

func customPageStyledScreen(content string, width, height int) uv.ScreenBuffer {
	screen := uv.NewScreenBuffer(width, height)
	screen.Method = ansi.GraphemeWidth
	uv.NewStyledString(content).Draw(screen, screen.Bounds())
	return screen
}
