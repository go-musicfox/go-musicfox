package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/charmbracelet/x/ansi"
)

// newCommandViewPage builds a "command_view" page through the provider registry
// (mirroring the S1-2 buildPageOrToast jump path).
func newCommandViewPage(t *testing.T, opts CommandViewOpts) *CommandViewPage {
	t.Helper()
	page, err := BuildPage("command_view", opts)
	if err != nil {
		t.Fatalf("BuildPage(command_view) failed: %v", err)
	}
	cvp, ok := page.(*CommandViewPage)
	if !ok {
		t.Fatalf("command_view provider returned %T, want *CommandViewPage", page)
	}
	return cvp
}

// TestCommandViewPageEmptyLinesNoPanic verifies that a page with empty lines
// renders without panicking and still fills the window height.
func TestCommandViewPageEmptyLinesNoPanic(t *testing.T) {
	app, _ := newFormPageTestApp(t)
	page := newCommandViewPage(t, CommandViewOpts{Title: "Empty"})

	view := page.View(app)
	if got := lipgloss.Height(view); got != app.WindowHeight() {
		t.Fatalf("view height = %d, want window height %d", got, app.WindowHeight())
	}
	if !strings.Contains(ansi.Strip(view), "Empty") {
		t.Fatalf("view is missing the page title: %q", ansi.Strip(view))
	}
}

// TestCommandViewPageRendersAtWindowHeightAndTruncatesWideLines verifies that
// View always produces exactly the window height (fillPageHeight contract) and
// that over-wide lines are truncated so no rendered row exceeds the terminal
// width.
func TestCommandViewPageRendersAtWindowHeightAndTruncatesWideLines(t *testing.T) {
	const width, height = 80, 24
	app, _ := newFormPageTestApp(t)
	lines := []string{
		"short line",
		strings.Repeat("x", 500),
		"",
		"tail line",
	}
	page := newCommandViewPage(t, CommandViewOpts{Title: "View", Lines: lines})

	view := page.View(app)
	if got := lipgloss.Height(view); got != height {
		t.Fatalf("view height = %d, want window height %d", got, height)
	}
	for i, line := range strings.Split(ansi.Strip(view), "\n") {
		if w := lipgloss.Width(line); w > width {
			t.Fatalf("rendered line %d width = %d > terminal width %d: %q", i, w, width, line)
		}
	}
}

// TestCommandViewPageScrollKeysClamp verifies the keyboard scroll steps and
// that the scroll offset is clamped to [0, maxScroll] from either side.
func TestCommandViewPageScrollKeysClamp(t *testing.T) {
	app, _ := newFormPageTestApp(t)
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}
	page := newCommandViewPage(t, CommandViewOpts{Title: "Scroll", Lines: lines})
	maxScroll := page.maxScroll(app)
	if maxScroll <= 0 {
		t.Fatalf("maxScroll = %d, want > 0 for 100 lines", maxScroll)
	}

	// End jumps to the clamped bottom; further down presses stay clamped.
	if next, _ := page.Update(tea.KeyPressMsg{Code: tea.KeyEnd}, app); next != page {
		t.Fatalf("end key returned %T, want the page itself", next)
	}
	if page.scroll != maxScroll {
		t.Fatalf("scroll after end = %d, want maxScroll %d", page.scroll, maxScroll)
	}
	if next, _ := page.Update(tea.KeyPressMsg{Code: tea.KeyDown}, app); next != page {
		t.Fatalf("down key returned %T, want the page itself", next)
	}
	if page.scroll != maxScroll {
		t.Fatalf("scroll after clamped down = %d, want %d", page.scroll, maxScroll)
	}

	// A scroll already out of range above is clamped back on the next move.
	page.scroll = maxScroll + 50
	if next, _ := page.Update(tea.KeyPressMsg{Code: tea.KeyUp}, app); next != page {
		t.Fatalf("up key returned %T, want the page itself", next)
	}
	if page.scroll != maxScroll {
		t.Fatalf("scroll after up from overflow = %d, want %d", page.scroll, maxScroll)
	}

	// Home returns to the top; further up presses stay clamped.
	if next, _ := page.Update(tea.KeyPressMsg{Code: tea.KeyHome}, app); next != page {
		t.Fatalf("home key returned %T, want the page itself", next)
	}
	if page.scroll != 0 {
		t.Fatalf("scroll after home = %d, want 0", page.scroll)
	}
	if next, _ := page.Update(tea.KeyPressMsg{Code: tea.KeyUp}, app); next != page {
		t.Fatalf("up key returned %T, want the page itself", next)
	}
	if page.scroll != 0 {
		t.Fatalf("scroll after clamped up = %d, want 0", page.scroll)
	}

	// A negative scroll (out of range below) is clamped back on the next move.
	page.scroll = -50
	if next, _ := page.Update(tea.KeyPressMsg{Code: tea.KeyDown}, app); next != page {
		t.Fatalf("down key returned %T, want the page itself", next)
	}
	if page.scroll != 0 {
		t.Fatalf("scroll after down from underflow = %d, want 0", page.scroll)
	}

	// PgUp/PgDn step by half a page and clamp.
	half := page.pageStep(app)
	if next, _ := page.Update(tea.KeyPressMsg{Code: tea.KeyPgDown}, app); next != page {
		t.Fatalf("pgdn key returned %T, want the page itself", next)
	}
	if page.scroll != half {
		t.Fatalf("scroll after pgdn = %d, want %d", page.scroll, half)
	}
	if next, _ := page.Update(tea.KeyPressMsg{Code: tea.KeyPgUp}, app); next != page {
		t.Fatalf("pgup key returned %T, want the page itself", next)
	}
	if page.scroll != 0 {
		t.Fatalf("scroll after pgup = %d, want 0", page.scroll)
	}
}

// TestCommandViewPageMouseWheelScrolls verifies the wheel scroll steps (±3
// lines, mirroring the menu wheel habit) and bottom clamping.
func TestCommandViewPageMouseWheelScrolls(t *testing.T) {
	app, _ := newFormPageTestApp(t)
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "wheel line"
	}
	page := newCommandViewPage(t, CommandViewOpts{Title: "Wheel", Lines: lines})
	maxScroll := page.maxScroll(app)

	if next, _ := page.Update(tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelDown}), app); next != page {
		t.Fatalf("wheel down returned %T, want the page itself", next)
	}
	if page.scroll != 3 {
		t.Fatalf("scroll after wheel down = %d, want 3", page.scroll)
	}
	if next, _ := page.Update(tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelUp}), app); next != page {
		t.Fatalf("wheel up returned %T, want the page itself", next)
	}
	if page.scroll != 0 {
		t.Fatalf("scroll after wheel up = %d, want 0", page.scroll)
	}

	// Wheel at the bottom clamps instead of overflowing.
	page.scroll = maxScroll
	if next, _ := page.Update(tea.MouseWheelMsg(tea.Mouse{Button: tea.MouseWheelDown}), app); next != page {
		t.Fatalf("wheel down at bottom returned %T, want the page itself", next)
	}
	if page.scroll != maxScroll {
		t.Fatalf("scroll after clamped wheel down = %d, want %d", page.scroll, maxScroll)
	}
}

// TestCommandViewPageBackKeysReturnToMain verifies that left/esc/backspace
// navigate back to the main page.
func TestCommandViewPageBackKeysReturnToMain(t *testing.T) {
	app, _ := newFormPageTestApp(t)
	page := newCommandViewPage(t, CommandViewOpts{Title: "Back"})

	for _, key := range []rune{tea.KeyEsc, tea.KeyLeft, tea.KeyBackspace} {
		next, cmd := page.Update(tea.KeyPressMsg{Code: key}, app)
		if next != app.Main() {
			t.Fatalf("key %q returned %T, want *model.Main", key, next)
		}
		if cmd != nil {
			t.Fatalf("key %q returned a non-nil cmd %v, want nil", key, cmd)
		}
	}
}

// TestCommandViewPageBackButtonHoverAndClick verifies the recorded back-button
// hit coordinates: motion toggles hover (with a rerender cmd), and a left click
// returns to Main while a click elsewhere is ignored.
func TestCommandViewPageBackButtonHoverAndClick(t *testing.T) {
	app, _ := newFormPageTestApp(t)
	page := newCommandViewPage(t, CommandViewOpts{Title: "BackBtn"})
	page.View(app) // records backBtnRowY / backBtnStartX

	backX, backY := page.backBtnStartX, page.backBtnRowY

	// Motion over the back button toggles hover and requests a rerender.
	if next, cmd := page.Update(tea.MouseMotionMsg(tea.Mouse{X: backX, Y: backY}), app); next != page || cmd == nil {
		t.Fatalf("motion over back button returned page=%T cmd=%v, want page itself + rerender cmd", next, cmd)
	}
	if !page.hovered {
		t.Fatalf("hovered = false after motion over the back button")
	}
	// Unchanged hover returns no cmd.
	if next, cmd := page.Update(tea.MouseMotionMsg(tea.Mouse{X: backX, Y: backY}), app); next != page || cmd != nil {
		t.Fatalf("unchanged hover returned page=%T cmd=%v, want page itself + nil cmd", next, cmd)
	}
	// Motion away clears hover and requests a rerender.
	if next, cmd := page.Update(tea.MouseMotionMsg(tea.Mouse{X: 60, Y: backY + 3}), app); next != page || cmd == nil {
		t.Fatalf("motion away returned page=%T cmd=%v, want page itself + rerender cmd", next, cmd)
	}
	if page.hovered {
		t.Fatalf("hovered = true after motion away from the back button")
	}

	// Left click on the back button returns to Main.
	if next, _ := page.Update(tea.MouseClickMsg(tea.Mouse{X: backX, Y: backY, Button: tea.MouseLeft}), app); next != app.Main() {
		t.Fatalf("back button click returned %T, want *model.Main", next)
	}

	// A click outside the back button is ignored.
	if next, cmd := page.Update(tea.MouseClickMsg(tea.Mouse{X: backX + pageBackButtonWidth + 10, Y: backY, Button: tea.MouseLeft}), app); next != page || cmd != nil {
		t.Fatalf("click outside back button returned page=%T cmd=%v, want page itself + nil cmd", next, cmd)
	}
}

// TestCommandViewPageIgnoresUnknownMessages verifies the page contract basics:
// non-input messages are ignored, the type identifier is correct, and the
// content is available without async loading.
func TestCommandViewPageIgnoresUnknownMessages(t *testing.T) {
	app, _ := newFormPageTestApp(t)
	page := newCommandViewPage(t, CommandViewOpts{Title: "Misc"})

	if next, cmd := page.Update(tea.WindowSizeMsg{Width: 80, Height: 24}, app); next != page || cmd != nil {
		t.Fatalf("window size msg returned page=%T cmd=%v, want page itself + nil cmd", next, cmd)
	}
	if next, cmd := page.Update(model.ShowNotificationMsg{}, app); next != page || cmd != nil {
		t.Fatalf("notification msg returned page=%T cmd=%v, want page itself + nil cmd", next, cmd)
	}
	if page.Type() != PtCommandView {
		t.Fatalf("Type() = %q, want %q", page.Type(), PtCommandView)
	}
	if page.Msg() != nil {
		t.Fatalf("Msg() = %v, want nil", page.Msg())
	}
	if page.IgnoreQuitKeyMsg(tea.KeyPressMsg{Text: "q"}) {
		t.Fatalf("IgnoreQuitKeyMsg(q) = true, want false (global quit stays active)")
	}
}
