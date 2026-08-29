package ui

import (
	"context"
	"io"
	"slices"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/charmbracelet/x/ansi"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// newCommandViewFlowTestApp builds a bare *model.App wired with the S1-2
// UnknownMsgHandler — the same dispatch the tuiFrontend assembly installs
// (frontend.go), minus the cover-renderer clear (the bare test shell has no
// cover renderer). The tea event loop is not run: messages are fed
// synchronously via app.Update, mirroring the foxful_integration_test infra.
func newCommandViewFlowTestApp(t *testing.T) *model.App {
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
	// Mirror the frontend.go wiring: a commandViewMsg builds the command_view
	// page through the provider registry; any other message falls through.
	app.With(model.WithUnknownMsgHandler(func(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
		vm, ok := msg.(commandViewMsg)
		if !ok {
			return nil, nil
		}
		if page := buildPageOrToast("command_view", CommandViewOpts{Title: vm.Title, Lines: vm.Lines}); page != nil {
			return page, nil
		}
		return nil, nil
	}))
	_ = app.Run()
	if app.Main() == nil {
		t.Fatal("main page was not initialized")
	}
	_, _ = app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return app
}

// TestCommandViewFlow is the S1 end-to-end message-flow test (no real WASM): a
// track-B command whose Run returns Action=="view" goes through the full chain —
// ① CommandMenu.Action returns a tea.Cmd; ② executing it yields a commandViewMsg
// (title + \r\n-normalized, blank-preserving lines); ③ the UnknownMsgHandler
// builds the page through the provider registry; ④ App.Update switches the
// active page to *CommandViewPage; ⑤ the page scrolls and returns to Main.
func TestCommandViewFlow(t *testing.T) {
	app := newCommandViewFlowTestApp(t)
	key := cmdMenuTestKey(t) + "_flow"
	// 30 lines: more than the ~15-row body viewport, so scrolling is actually
	// exercisable; \r\n and a blank line also run splitLines inside the flow.
	wantLines := []string{"line1", "line2", "", "line3"}
	for i := 4; i <= 28; i++ {
		wantLines = append(wantLines, "line"+strconv.Itoa(i))
	}
	wantLines = append(wantLines, "lineLast")
	message := "line1\r\nline2\n\nline3\n"
	for i := 4; i <= 28; i++ {
		message += "line" + strconv.Itoa(i) + "\n"
	}
	message += "lineLast"
	RegisterCommand(frontend.Command{
		Key:   key,
		Title: "View Flow",
		Run: func(frontend.CommandContext) frontend.CommandResult {
			return frontend.CommandResult{Action: "view", Title: "FlowTitle", Message: message}
		},
	})
	registerCommandMenus()

	menu, err := BuildMenu(key, baseMenu{}, NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(%q): %v", key, err)
	}
	cm, ok := menu.(*CommandMenu)
	if !ok {
		t.Fatalf("BuildMenu(%q) = %T, want *CommandMenu", key, menu)
	}

	// ① The command menu action returns the current main page plus the command
	// to execute (a bubbletea cmd, not yet run).
	page, cmd := cm.Action(app, 0)
	if page != app.MustMain() {
		t.Fatalf("Action() page = %T, want main", page)
	}
	if cmd == nil {
		t.Fatal("Action() returned a nil command")
	}

	// ② Executing the command produces the view message (not a nil msg).
	msg := cmd()
	vm, ok := msg.(commandViewMsg)
	if !ok {
		t.Fatalf("command msg = %T (%v), want commandViewMsg", msg, msg)
	}
	if vm.Title != "FlowTitle" {
		t.Fatalf("commandViewMsg.Title = %q, want %q", vm.Title, "FlowTitle")
	}
	if !slices.Equal(vm.Lines, wantLines) {
		t.Fatalf("commandViewMsg.Lines = %q, want %q", vm.Lines, wantLines)
	}

	// A non-view message is ignored by the handler (the current page stays put).
	_, _ = app.Update(tickLoginMsg{})
	if app.CurPage() != app.Main() {
		t.Fatalf("current page after a non-view message = %T, want main", app.CurPage())
	}

	// ③④ Feeding the view message through the app message loop dispatches it
	// via Main.Update → UnknownMsgHandler → command_view page.
	_, _ = app.Update(msg)
	cvp, ok := app.CurPage().(*CommandViewPage)
	if !ok {
		t.Fatalf("current page after view msg = %T, want *CommandViewPage", app.CurPage())
	}
	if cvp.opts.Title != "FlowTitle" || !slices.Equal(cvp.opts.Lines, wantLines) {
		t.Fatalf("page opts = {Title: %q, Lines: %q}", cvp.opts.Title, cvp.opts.Lines)
	}
	view := ansi.Strip(cvp.View(app))
	if !strings.Contains(view, "line1") || !strings.Contains(view, "line3") {
		t.Fatalf("page view does not render the body lines:\n%s", view)
	}

	// ⑤ The page scrolls (down moves the first visible line) and back keys
	// return to the main page.
	_, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if got := cvp.scroll; got != 1 {
		t.Fatalf("scroll after down = %d, want 1", got)
	}
	_, _ = app.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if app.CurPage() != app.Main() {
		t.Fatalf("current page after esc = %T, want main", app.CurPage())
	}
}
