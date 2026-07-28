package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

func TestLastfmCustomAPIPageUsesLocalizedSubmitAndBackButton(t *testing.T) {
	previousLocale := model.DefaultCatalog().Locale()
	SetupI18n("zh")
	t.Cleanup(func() { model.SetLocale(previousLocale) })

	app, netease, page := newLastfmCustomAPIPageTest(t)
	view := page.View(app)
	plain := ansi.Strip(view)
	if !strings.Contains(plain, pageSubmitText()) || strings.Contains(plain, "Submit") {
		t.Fatalf("submit text = %q, want localized %q without Submit", plain, pageSubmitText())
	}

	backRow := visibleRowContaining(t, view, page.menuTitle.Title, app.WindowHeight())
	if line := ansi.Strip(strings.Split(view, "\n")[backRow]); !strings.Contains(line, "←") {
		t.Fatal("Lastfm API account menu title does not contain a back button")
	}
	backX := max(0, netease.MustMain().MenuStartColumn()-pageBackButtonWidth)
	_, _ = page.Update(tea.MouseMotionMsg(tea.Mouse{X: backX, Y: backRow}), app)
	if page.mousePointer != "pointer" {
		t.Fatalf("back button mouse pointer = %q, want pointer", page.mousePointer)
	}
	next, _ := page.Update(tea.MouseClickMsg(tea.Mouse{X: backX, Y: backRow, Button: tea.MouseLeft}), app)
	if next != netease.MustMain() {
		t.Fatalf("back button returned %T, want Main", next)
	}
}

func TestLastfmCustomAPIPageMouseControlsAndActiveStyles(t *testing.T) {
	previousLocale := model.DefaultCatalog().Locale()
	SetupI18n("zh")
	t.Cleanup(func() { model.SetLocale(previousLocale) })

	app, netease, page := newLastfmCustomAPIPageTest(t)
	page.keyInput.SetValue("abcdef")
	page.secretInput.SetValue("uvwxyz")
	page.keyInput.SetCursor(0)
	view := page.View(app)

	keyRow := visibleRowContaining(t, view, "abcdef", app.WindowHeight())
	keyX := netease.MustMain().MenuStartColumn() + runewidth.StringWidth(page.keyInput.Prompt) + 2
	_, _ = page.Update(tea.MouseMotionMsg(tea.Mouse{X: keyX, Y: keyRow}), app)
	if page.mousePointer != "text" {
		t.Fatalf("key input mouse pointer = %q, want text", page.mousePointer)
	}
	_, _ = page.Update(tea.MouseClickMsg(tea.Mouse{X: keyX, Y: keyRow, Button: tea.MouseLeft}), app)
	if got := page.keyInput.Position(); got != 2 {
		t.Fatalf("key input cursor position = %d, want 2", got)
	}

	view = page.View(app)
	buttonRow := visibleRowContaining(t, view, pageSubmitText(), app.WindowHeight())
	for _, label := range []string{pageSubmitText(), page.reloadText, page.clearText} {
		x := visibleColumnContaining(t, view, buttonRow, label)
		_, _ = page.Update(tea.MouseMotionMsg(tea.Mouse{X: x, Y: buttonRow}), app)
		if page.mousePointer != "pointer" {
			t.Fatalf("%s mouse pointer = %q, want pointer", label, page.mousePointer)
		}
		cell := formStyledScreen(page.View(app)).CellAt(x, buttonRow)
		if cell == nil || cell.Style.Underline != uv.UnderlineSingle {
			t.Fatalf("%s hover underline = %v, want single", label, cell)
		}
	}

	activePage := newLastfmCustomAPIPageForStyles(t, netease)
	assertLastfmInputUsesActivePageStyles(t, focusedInputStylesOf(activePage.keyInput))
	_, _ = activePage.Update(tea.KeyPressMsg{Code: tea.KeyTab}, app)
	assertLastfmInputUsesActivePageStyles(t, focusedInputStylesOf(activePage.secretInput))
	_, _ = activePage.Update(tea.KeyPressMsg{Code: tea.KeyTab}, app)
	activeView := activePage.View(app)
	activeButtonRow := visibleRowContaining(t, activeView, pageSubmitText(), app.WindowHeight())
	activeSubmitX := visibleColumnContaining(t, activeView, activeButtonRow, pageSubmitText())
	assertCellForeground(t, formStyledScreen(activeView).CellAt(activeSubmitX, activeButtonRow), util.GetPrimaryColor())

	_, _ = page.Update(tea.MouseClickMsg(tea.Mouse{X: visibleColumnContaining(t, view, buttonRow, pageSubmitText()), Y: buttonRow, Button: tea.MouseLeft}), app)
	if !strings.Contains(ansi.Strip(page.tips), "请输入正确的 API 账号或密码") {
		t.Fatalf("invalid Submit tip = %q", page.tips)
	}
}

func newLastfmCustomAPIPageForStyles(t *testing.T, netease *Netease) *LastfmCustomApiPage {
	t.Helper()
	return newLastfmCustomApiPage(netease)
}

func assertLastfmInputUsesActivePageStyles(t *testing.T, styles focusedInputStyles) {
	t.Helper()
	want := util.GetPrimaryColor()
	if styles.text != want || styles.prompt != want || styles.cursor != want {
		t.Fatalf("active input styles = %#v, want text, prompt, and cursor in primary %v", styles, want)
	}
	if styles.placeholder == want {
		t.Fatalf("active placeholder color = primary %v, want non-highlighted", want)
	}
}

func newLastfmCustomAPIPageTest(t *testing.T) (*model.App, *Netease, *LastfmCustomApiPage) {
	t.Helper()
	app, netease := newFormPageTestApp(t)
	return app, netease, newLastfmCustomApiPage(netease)
}

func visibleColumnContaining(t *testing.T, view string, row int, marker string) int {
	t.Helper()
	line := ansi.Strip(strings.Split(view, "\n")[row])
	byteIndex := strings.Index(line, marker)
	if byteIndex < 0 {
		t.Fatalf("row %d does not contain %q: %q", row, marker, line)
	}
	return runewidth.StringWidth(line[:byteIndex])
}
