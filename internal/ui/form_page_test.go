package ui

import (
	"context"
	"image/color"
	"io"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

func TestStandaloneFormsUseLocalizedSubmitText(t *testing.T) {
	previousLocale := model.DefaultCatalog().Locale()
	SetupI18n("zh")
	t.Cleanup(func() { model.SetLocale(previousLocale) })

	app, netease := newFormPageTestApp(t)
	views := map[string]string{
		"search": NewSearchPage(netease).View(app),
		"login":  NewLoginPage(netease).View(app),
	}
	for name, view := range views {
		t.Run(name, func(t *testing.T) {
			plain := ansi.Strip(view)
			if !strings.Contains(plain, "确认") {
				t.Fatalf("submit button is not localized:\n%s", plain)
			}
			if strings.Contains(plain, "Submit") {
				t.Fatalf("submit button still contains English text:\n%s", plain)
			}
		})
	}
}

func TestStandaloneFormFocusedStylesUsePrimaryColor(t *testing.T) {
	_, netease := newFormPageTestApp(t)
	search := NewSearchPage(netease)
	login := NewLoginPage(netease)
	want := util.GetPrimaryColor()

	inputs := map[string]struct {
		styles focusedInputStyles
	}{
		"search":   {focusedInputStylesOf(search.wordsInput)},
		"account":  {focusedInputStylesOf(login.accountInput)},
		"password": {focusedInputStylesOf(login.passwordInput)},
		"cookie":   {focusedInputStylesOf(login.cookieInput)},
	}
	for name, input := range inputs {
		t.Run(name, func(t *testing.T) {
			if input.styles.text != want {
				t.Errorf("focused text color = %v, want primary %v", input.styles.text, want)
			}
			if input.styles.prompt != want {
				t.Errorf("focused prompt color = %v, want primary %v", input.styles.prompt, want)
			}
			if input.styles.placeholder == want {
				t.Errorf("focused placeholder color = primary %v, want non-highlighted", want)
			}
			if input.styles.cursor != want {
				t.Errorf("cursor color = %v, want primary %v", input.styles.cursor, want)
			}
		})
	}
}

func TestStandaloneFormMouseClickPositionsCursor(t *testing.T) {
	app, netease := newFormPageTestApp(t)

	t.Run("search", func(t *testing.T) {
		page := NewSearchPage(netease)
		page.searchType = StSingleSong
		page.wordsInput.SetValue("abcdef")
		page.View(app)
		_, _ = page.Update(tea.MouseMotionMsg(tea.Mouse{X: page.inputStartX, Y: page.inputRowY}), app)
		if !page.hoveredInput || page.mousePointer != "text" {
			t.Fatalf("input hover = %v, pointer = %q; want hovered text pointer", page.hoveredInput, page.mousePointer)
		}
		_, _ = page.Update(tea.MouseMotionMsg(tea.Mouse{X: page.submitStartX, Y: page.submitRowY}), app)
		if !page.hoveredSubmit || page.mousePointer != "pointer" {
			t.Fatalf("submit hover = %v, pointer = %q; want hovered pointer", page.hoveredSubmit, page.mousePointer)
		}

		x := netease.MustMain().MenuStartColumn() + runewidth.StringWidth(page.wordsInput.Prompt) + 2
		_, _ = page.Update(tea.MouseClickMsg(tea.Mouse{X: x, Y: page.inputRowY, Button: tea.MouseLeft}), app)
		if got := page.wordsInput.Position(); got != 2 {
			t.Fatalf("cursor position = %d, want 2", got)
		}
	})

	t.Run("login", func(t *testing.T) {
		page := NewLoginPage(netease)
		page.accountInput.SetValue("abcdef")
		page.View(app)
		_, _ = page.Update(tea.MouseMotionMsg(tea.Mouse{X: page.inputStartX, Y: page.accountRowY - 1}), app)
		if page.hoveredInputBox != 0 || page.mousePointer != "text" {
			t.Fatalf("account hover = %d, pointer = %q; want account text pointer", page.hoveredInputBox, page.mousePointer)
		}

		x := netease.MustMain().MenuStartColumn() + runewidth.StringWidth(page.accountInput.Prompt) + 2
		_, _ = page.Update(tea.MouseClickMsg(tea.Mouse{X: x, Y: page.accountRowY - 1, Button: tea.MouseLeft}), app)
		if got := page.accountInput.Position(); got != 2 {
			t.Fatalf("cursor position = %d, want 2", got)
		}
	})
}

func TestLoginPageUsesFoxfulTabsWithMouseAndKeyboard(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	page := NewLoginPage(netease)
	page.View(app)
	if page.tabs == nil {
		t.Fatal("foxful-cli Tabs was not initialized")
	}

	accountWidth := lipgloss.Width(model.T(MsgLoginAccountTab)) + 4
	cookieX := page.tabStartX + accountWidth + 1
	tabY := page.tabsStartRowY - 1
	_, _ = page.Update(tea.MouseMotionMsg(tea.Mouse{X: cookieX, Y: tabY}), app)
	if page.hoveredTab != tabCookie || page.mousePointer != "pointer" {
		t.Fatalf("tab hover = %d, pointer = %q; want cookie pointer", page.hoveredTab, page.mousePointer)
	}
	_, _ = page.Update(tea.MouseClickMsg(tea.Mouse{X: cookieX, Y: tabY, Button: tea.MouseLeft}), app)
	if page.tabs.Active() != tabCookie || !page.tabs.Focused() {
		t.Fatalf("after click active = %d, focused = %v; want cookie tab focused", page.tabs.Active(), page.tabs.Focused())
	}
	if page.menuTitle.Subtitle != model.T(MsgLoginCookieTab) {
		t.Fatalf("menu subtitle = %q, want cookie login", page.menuTitle.Subtitle)
	}

	page.tabs.SetActive(tabAccount)
	page.index = idxTabAccount
	page.updateTabStyle()
	page.applyFocus()
	_, _ = page.Update(tea.KeyPressMsg{Code: tea.KeyRight}, app)
	if page.tabs.Active() != tabCookie {
		t.Fatalf("right key active tab = %d, want cookie", page.tabs.Active())
	}
}

func TestSearchPageMouseRowsMatchRenderedControls(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	page := NewSearchPage(netease)
	page.wordsInput.SetValue("SEARCH_INPUT_MARKER")
	view := page.View(app)

	inputRow := visibleRowContaining(t, view, "SEARCH_INPUT_MARKER", app.WindowHeight())
	submitRow := visibleRowContaining(t, view, pageSubmitText(), app.WindowHeight())
	if page.inputRowY != inputRow {
		t.Errorf("input mouse row = %d, rendered row = %d", page.inputRowY, inputRow)
	}
	if page.submitRowY != submitRow {
		t.Errorf("submit mouse row = %d, rendered row = %d", page.submitRowY, submitRow)
	}

	_, _ = page.Update(tea.MouseMotionMsg(tea.Mouse{X: page.inputStartX, Y: inputRow - 1}), app)
	if page.hoveredInput || page.hoveredSubmit {
		t.Fatal("hover activated one row above the rendered input")
	}
	_, _ = page.Update(tea.MouseMotionMsg(tea.Mouse{X: page.inputStartX, Y: inputRow}), app)
	if !page.hoveredInput {
		t.Fatal("hover did not activate on the rendered input row")
	}
}

func TestFocusedFormPromptsContainNoANSI(t *testing.T) {
	app, netease := newFormPageTestApp(t)

	search := NewSearchPage(netease)
	search.View(app)
	_, _ = search.Update(tea.MouseMotionMsg(tea.Mouse{X: search.inputStartX, Y: search.inputRowY}), app)

	login := NewLoginPage(netease)
	login.View(app)
	_, _ = login.Update(tea.MouseMotionMsg(tea.Mouse{X: login.inputStartX, Y: login.accountRowY - 1}), app)

	prompts := map[string]string{
		"search": search.wordsInput.Prompt,
		"login":  login.accountInput.Prompt,
	}
	for name, prompt := range prompts {
		t.Run(name, func(t *testing.T) {
			if stripped := ansi.Strip(prompt); prompt != stripped {
				t.Fatalf("focused prompt contains ANSI control sequences: %q", prompt)
			}
			if prompt != "> " {
				t.Fatalf("focused prompt = %q, want plain > prompt", prompt)
			}
		})
	}
}

func TestLoginTabsKeepEveryRenderedRowAligned(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	page := NewLoginPage(netease)
	view := ansi.Strip(page.View(app))
	lines := strings.Split(view, "\n")
	wantPrefix := strings.Repeat(" ", page.tabStartX)

	for row := page.tabsStartRowY - 1; row < page.tabsEndRowY; row++ {
		if row >= len(lines) {
			t.Fatalf("Tabs row %d is outside rendered view", row)
		}
		if !strings.HasPrefix(lines[row], wantPrefix) {
			t.Errorf("Tabs row %d is not aligned to column %d: %q", row, page.tabStartX, lines[row])
		}
	}
}

func TestHoveredInputsHaveNoUnderline(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	page := NewSearchPage(netease)
	page.wordsInput.SetValue("hover")
	page.View(app)
	_, _ = page.Update(tea.MouseMotionMsg(tea.Mouse{X: page.inputStartX, Y: page.inputRowY}), app)
	screen := formStyledScreen(page.View(app))
	cell := screen.CellAt(page.inputStartX, page.inputRowY)
	if cell == nil {
		t.Fatal("hovered input prompt cell is missing")
	}
	if cell.Style.Underline != uv.UnderlineNone {
		t.Fatalf("hovered input underline = %v, want none", cell.Style.Underline)
	}
}

func TestActiveButtonsUsePrimaryColor(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	want := util.GetPrimaryColor()

	search := NewSearchPage(netease)
	search.submitButton = pageSubmitButton(true)
	search.View(app)
	assertCellForeground(t, formStyledScreen(search.View(app)).CellAt(search.submitStartX+2, search.submitRowY), want)

	login := NewLoginPage(netease)
	login.index = submitIndex
	login.applyFocus()
	login.View(app)
	assertCellForeground(t, formStyledScreen(login.View(app)).CellAt(login.submitStartX+2, login.buttonsRowY-1), want)

	login.index = qrLoginIndex
	login.applyFocus()
	login.View(app)
	assertCellForeground(t, formStyledScreen(login.View(app)).CellAt(login.qrStartX+2, login.buttonsRowY-1), want)
}

func TestPlaceholderDoesNotHighlightOnFocusOrHover(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	page := NewSearchPage(netease)
	view := page.View(app)
	placeholderX := page.inputStartX + lipgloss.Width(page.wordsInput.Prompt) + 1
	activeCell := formStyledScreen(view).CellAt(placeholderX, page.inputRowY)
	if activeCell == nil || activeCell.Style.Fg == nil {
		t.Fatal("active placeholder cell or foreground is missing")
	}
	if colorsEqual(activeCell.Style.Fg, util.GetPrimaryColor()) {
		t.Fatalf("active placeholder uses primary color %v", activeCell.Style.Fg)
	}

	_, _ = page.Update(tea.MouseMotionMsg(tea.Mouse{X: page.inputStartX, Y: page.inputRowY}), app)
	hoverCell := formStyledScreen(page.View(app)).CellAt(placeholderX, page.inputRowY)
	if hoverCell == nil || !colorsEqual(hoverCell.Style.Fg, activeCell.Style.Fg) {
		t.Fatalf("hover placeholder color = %v, want unchanged %v", hoverCell, activeCell.Style.Fg)
	}
}

func TestLoginTabsOnlyFocusedTabUsesPrimaryWithoutBackground(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	page := NewLoginPage(netease)
	accountWidth := lipgloss.Width(model.T(MsgLoginAccountTab)) + 4

	page.index = idxTabAccount
	page.applyFocus()
	view := page.View(app)
	screen := formStyledScreen(view)
	topRow := page.tabsStartRowY - 1
	assertCellForeground(t, screen.CellAt(page.tabStartX, topRow), style.CurrentStyleSet().SelectedItem.GetForeground())
	if cell := screen.CellAt(page.tabStartX+accountWidth, topRow); cell != nil && cell.Style.Fg != nil {
		t.Fatalf("inactive tab border color = %v, want unspecified", cell.Style.Fg)
	}
	if cell := screen.CellAt(page.tabStartX+2, topRow+1); cell == nil || cell.Style.Bg != nil {
		if cell == nil {
			t.Fatal("focused tab title cell is missing")
		}
		t.Fatalf("focused tab title background = %v, want unspecified", cell.Style.Bg)
	}
	bottomRow := page.tabsEndRowY - 1
	assertCellForeground(t, screen.CellAt(page.tabStartX, bottomRow), style.CurrentStyleSet().SelectedItem.GetForeground())

	page.tabs.SetActive(tabCookie)
	page.index = idxTabCookie
	page.updateTabStyle()
	page.applyFocus()
	view = page.View(app)
	screen = formStyledScreen(view)
	bottomRow = page.tabsEndRowY - 1
	bottomLine := strings.TrimRight(strings.Split(ansi.Strip(view), "\n")[bottomRow], " ")
	farRight := lipgloss.Width(bottomLine) - 1
	assertCellForeground(t, screen.CellAt(farRight, bottomRow), style.CurrentStyleSet().SelectedItem.GetForeground())
	if cell := screen.CellAt(page.tabStartX, bottomRow); cell != nil && cell.Style.Fg != nil {
		t.Fatalf("left edge color = %v while rightmost tab is focused, want unspecified", cell.Style.Fg)
	}

	page.index = 0
	page.applyFocus()
	view = page.View(app)
	screen = formStyledScreen(view)
	if cell := screen.CellAt(page.tabStartX, page.tabsStartRowY-1); cell != nil && cell.Style.Fg != nil {
		t.Fatalf("blurred active tab border color = %v, want unspecified", cell.Style.Fg)
	}
}

func TestStandaloneFormPlaceholdersRenderCompletely(t *testing.T) {
	previousLocale := model.DefaultCatalog().Locale()
	SetupI18n("zh")
	t.Cleanup(func() { model.SetLocale(previousLocale) })
	app, netease := newFormPageTestApp(t)

	search := NewSearchPage(netease)
	searchLines := strings.Split(ansi.Strip(search.View(app)), "\n")
	if line := searchLines[search.inputRowY]; !strings.Contains(line, "输入关键词") {
		t.Fatalf("search placeholder is incomplete: %q", line)
	}

	login := NewLoginPage(netease)
	loginLines := strings.Split(ansi.Strip(login.View(app)), "\n")
	if line := loginLines[login.accountRowY-1]; !strings.Contains(line, "手机号或邮箱") {
		t.Fatalf("account placeholder is incomplete: %q", line)
	}
	if line := loginLines[login.passwordRowY-1]; !strings.Contains(line, "密码") {
		t.Fatalf("password placeholder is incomplete: %q", line)
	}

	login.tabs.SetActive(tabCookie)
	login.updateTabStyle()
	login.index = 0
	login.applyFocus()
	loginLines = strings.Split(ansi.Strip(login.View(app)), "\n")
	if line := loginLines[login.cookieRowY-1]; !strings.Contains(line, "请输入 Cookie") {
		t.Fatalf("cookie placeholder is incomplete: %q", line)
	}
}

func TestQRLoginPageMenuTitleHasClickableBackButton(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	from := NewLoginPage(netease)
	page := NewQRLoginPage(netease, from, nil)
	view := page.View(app)
	if row := visibleRowContaining(t, view, model.T(MsgQRLoginPageTitle), app.WindowHeight()); !strings.Contains(ansi.Strip(strings.Split(view, "\n")[row]), "←") {
		t.Fatal("QR login menu title does not contain a back button")
	}

	backRow := visibleRowContaining(t, view, model.T(MsgQRLoginPageTitle), app.WindowHeight())
	backX := max(0, netease.MustMain().MenuStartColumn()-pageBackButtonWidth)
	next, _ := page.Update(tea.MouseClickMsg(tea.Mouse{X: backX, Y: backRow, Button: tea.MouseLeft}), app)
	if next != from {
		t.Fatalf("back button returned %T, want originating login page", next)
	}
}

func formStyledScreen(content string) uv.ScreenBuffer {
	width := max(lipgloss.Width(content), 1)
	height := strings.Count(content, "\n") + 1
	screen := uv.NewScreenBuffer(width, height)
	screen.Method = ansi.GraphemeWidth
	uv.NewStyledString(content).Draw(screen, screen.Bounds())
	return screen
}

func assertCellForeground(t *testing.T, cell *uv.Cell, want color.Color) {
	t.Helper()
	if cell == nil || cell.Style.Fg == nil {
		t.Fatalf("cell foreground = nil, want %v", want)
	}
	gotR, gotG, gotB, gotA := cell.Style.Fg.RGBA()
	wantR, wantG, wantB, wantA := want.RGBA()
	if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
		t.Fatalf("cell foreground = %#x %#x %#x %#x, want %#x %#x %#x %#x", gotR, gotG, gotB, gotA, wantR, wantG, wantB, wantA)
	}
}
func colorsEqual(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	aR, aG, aB, aA := a.RGBA()
	bR, bG, bB, bA := b.RGBA()
	return aR == bR && aG == bG && aB == bB && aA == bA
}

type focusedInputStyles struct {
	text        color.Color
	prompt      color.Color
	placeholder color.Color
	cursor      color.Color
}

func focusedInputStylesOf(input textinput.Model) focusedInputStyles {
	styles := input.Styles()
	return focusedInputStyles{
		text:        styles.Focused.Text.GetForeground(),
		prompt:      styles.Focused.Prompt.GetForeground(),
		placeholder: styles.Focused.Placeholder.GetForeground(),
		cursor:      styles.Cursor.Color,
	}
}

func newFormPageTestApp(t *testing.T) (*model.App, *Netease) {
	t.Helper()

	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Theme.ShowTitle = true
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	opts := model.DefaultOptions()
	opts.EnableStartup = false
	opts.WhetherDisplayTitle = true
	opts.MainMenu = &pageLayoutTestMenu{}
	opts.MainMenuTitle = &model.MenuItem{Title: "Main"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	opts.TeaOptions = []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithInput(nil),
		tea.WithOutput(io.Discard),
	}

	app := model.NewApp(opts)
	_ = app.Run()
	if app.Main() == nil {
		t.Fatal("main page was not initialized")
	}
	_, _ = app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return app, &Netease{App: app}
}
