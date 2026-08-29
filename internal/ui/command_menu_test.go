package ui

import (
	"slices"
	"strings"
	"testing"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// cmdMenuTestKey derives a unique command/menu key from the test name so the
// package-global registries stay pollution-free across a single test run.
func cmdMenuTestKey(t *testing.T) string {
	t.Helper()
	return "cmdmenu_" + strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, t.Name())
}

// --- Menu surface ---

// TestCommandMenuSurface checks the pure menu surface of CommandMenu: it is an
// action menu (not playable / not locatable), renders the command's title,
// produces no submenu and reports the command's key. Run is not exercised.
func TestCommandMenuSurface(t *testing.T) {
	menu := &CommandMenu{
		BaseMenu: BaseMenu{},
		cmd:      frontend.Command{Key: "cmdmenu_surface", Title: "X"},
	}

	if got := menu.GetMenuKey(); got != "cmdmenu_surface" {
		t.Fatalf("GetMenuKey() = %q, want %q", got, "cmdmenu_surface")
	}
	if menu.IsPlayable() {
		t.Fatal("IsPlayable() = true, want false")
	}
	if menu.IsLocatable() {
		t.Fatal("IsLocatable() = true, want false")
	}

	views := menu.MenuViews()
	if len(views) != 1 {
		t.Fatalf("MenuViews() returned %d items, want 1", len(views))
	}
	if views[0].Title != "X" {
		t.Fatalf("MenuViews()[0].Title = %q, want %q", views[0].Title, "X")
	}

	if sub := menu.SubMenu(nil, 0); sub != nil {
		t.Fatalf("SubMenu() = %v, want nil", sub)
	}
}

// --- Registration wiring ---

// TestCommandMenuRegistration proves registerCommandMenus adapts a registered
// track-B command into a *CommandMenu provider and a main-menu entry. The
// registry is package-global, so the key is derived from the test name and
// registerCommandMenus is called exactly once in the whole test binary.
func TestCommandMenuRegistration(t *testing.T) {
	key := cmdMenuTestKey(t)
	cmd := testCommand(key)
	cmd.Title = "CmdMenu Test"
	RegisterCommand(cmd)

	registerCommandMenus()

	// The menu provider is buildable through the registry.
	menu, err := BuildMenu(key, baseMenu{}, NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(%q): %v", key, err)
	}
	cm, ok := menu.(*CommandMenu)
	if !ok {
		t.Fatalf("BuildMenu(%q) = %T, want *CommandMenu", key, menu)
	}
	if got := cm.GetMenuKey(); got != key {
		t.Fatalf("GetMenuKey() = %q, want %q", got, key)
	}

	// The main-menu item is present.
	var found bool
	for _, item := range MainMenuPluginItems() {
		if item.Key == key {
			found = true
			if item.Title != "CmdMenu Test" {
				t.Fatalf("main-menu item title = %q, want %q", item.Title, "CmdMenu Test")
			}
		}
	}
	if !found {
		t.Fatalf("main-menu item %q not registered", key)
	}
}

// --- Action gating ---

// TestCommandMenuActionResolvesCurrentCommand proves the P8 hot-reload refresh
// at the action point: commandActionCmd resolves the command from the frontend
// registry by key at action time, so a menu instance built from a pre-reload
// snapshot executes the CURRENT (replaced) definition — never the stale one.
func TestCommandMenuActionResolvesCurrentCommand(t *testing.T) {
	key := cmdMenuTestKey(t) + "_reload"
	v1Ran, v2Ran := false, false

	RegisterCommand(frontend.Command{
		Key:   key,
		Title: "v1",
		Run: func(frontend.CommandContext) frontend.CommandResult {
			v1Ran = true
			return frontend.CommandResult{}
		},
	})

	// The menu instance is a pre-reload snapshot (v1).
	oldCmd, ok := frontend.CommandByKey(key)
	if !ok {
		t.Fatalf("command %q not registered", key)
	}
	m := &CommandMenu{BaseMenu: BaseMenu{}, cmd: oldCmd}
	a := &model.App{} // zero app: Notify is a no-op without a program

	// Hot reload replaces the definition (v2).
	frontend.ReplaceCommand(frontend.Command{
		Key:   key,
		Title: "v2",
		Run: func(frontend.CommandContext) frontend.CommandResult {
			v2Ran = true
			return frontend.CommandResult{}
		},
	})

	if msg := commandActionCmd(a, m)(); msg != nil {
		t.Fatalf("commandActionCmd() msg = %v, want nil", msg)
	}
	if v1Ran {
		t.Fatal("v1 closure ran: the menu executed the stale pre-reload command")
	}
	if !v2Ran {
		t.Fatal("v2 closure did not run: commandActionCmd must resolve the current command by key")
	}
}

// TestCommandMenuProviderResolvesCurrentCommand proves the P8 hot-reload refresh
// at the build point: the menu provider registered by registerCommandMenus
// resolves the command from the registry by key, so a menu built after a reload
// carries the new definition (title) without re-adapting the provider.
func TestCommandMenuProviderResolvesCurrentCommand(t *testing.T) {
	key := cmdMenuTestKey(t) + "_provider"
	RegisterCommand(frontend.Command{
		Key:   key,
		Title: "v1",
		Run:   func(frontend.CommandContext) frontend.CommandResult { return frontend.CommandResult{} },
	})
	registerCommandMenus()

	menu, err := BuildMenu(key, baseMenu{}, NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(%q): %v", key, err)
	}
	if cm := menu.(*CommandMenu); cm.cmd.Title != "v1" {
		t.Fatalf("pre-reload Title = %q, want %q", cm.cmd.Title, "v1")
	}

	// Reload replaces the definition; the SAME provider key must serve it.
	frontend.ReplaceCommand(frontend.Command{
		Key:   key,
		Title: "v2",
		Run:   func(frontend.CommandContext) frontend.CommandResult { return frontend.CommandResult{} },
	})
	menu, err = BuildMenu(key, baseMenu{}, NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(%q) after reload: %v", key, err)
	}
	if cm := menu.(*CommandMenu); cm.cmd.Title != "v2" {
		t.Fatalf("post-reload Title = %q, want %q (provider must resolve the current command)", cm.cmd.Title, "v2")
	}
}

// TestCommandMenuActionShowGate proves a command whose Show gate returns false
// is rejected before Run: commandActionCmd notifies (no-op on a zero App whose
// program is nil) and Run never executes.
func TestCommandMenuActionShowGate(t *testing.T) {
	ran := false
	cmd := frontend.Command{
		Key:   "cmdmenu_show_gate",
		Title: "ShowGate",
		Show:  func(frontend.CommandContext) bool { return false },
		Run: func(frontend.CommandContext) frontend.CommandResult {
			ran = true
			return frontend.CommandResult{}
		},
	}
	m := &CommandMenu{BaseMenu: BaseMenu{}, cmd: cmd}
	a := &model.App{} // zero app: Notify is a no-op without a program

	if msg := commandActionCmd(a, m)(); msg != nil {
		t.Fatalf("commandActionCmd() msg = %v, want nil", msg)
	}
	if ran {
		t.Fatal("Run executed despite Show returning false")
	}
}

// TestCommandMenuActionDisabledPlugin proves the [plugins] disabled gate: a
// command whose PluginID is disabled in config is rejected before Run.
func TestCommandMenuActionDisabledPlugin(t *testing.T) {
	withPluginConfig(t, []string{"cmdmenu_disabled_owner"})

	ran := false
	cmd := frontend.Command{
		Key:      "cmdmenu_disabled",
		Title:    "Disabled",
		PluginID: "cmdmenu_disabled_owner",
		Run: func(frontend.CommandContext) frontend.CommandResult {
			ran = true
			return frontend.CommandResult{}
		},
	}
	m := &CommandMenu{BaseMenu: BaseMenu{}, cmd: cmd}
	a := &model.App{}

	if msg := commandActionCmd(a, m)(); msg != nil {
		t.Fatalf("commandActionCmd() msg = %v, want nil", msg)
	}
	if ran {
		t.Fatal("Run executed despite plugin being disabled")
	}
}

// TestCommandMenuActionRuns proves the enabled + Show-pass path executes Run
// and presents the result without panicking on a nil player (the context
// degrades to the zero-value snapshot).
func TestCommandMenuActionRuns(t *testing.T) {
	ran := false
	cmd := frontend.Command{
		Key:   "cmdmenu_run",
		Title: "Run",
		Run: func(frontend.CommandContext) frontend.CommandResult {
			ran = true
			return frontend.CommandResult{Action: "toast", Title: "T", Message: "M"}
		},
	}
	m := &CommandMenu{BaseMenu: BaseMenu{}, cmd: cmd}
	a := &model.App{}

	if msg := commandActionCmd(a, m)(); msg != nil {
		t.Fatalf("commandActionCmd() msg = %v, want nil", msg)
	}
	if !ran {
		t.Fatal("Run not executed")
	}
}

// --- Result presentation helpers ---

// TestCommandMenuLevelToModel proves the CommandResult Level string maps to the
// model notification levels; unknown and empty levels default to info.
func TestCommandMenuLevelToModel(t *testing.T) {
	cases := []struct {
		level string
		want  model.NotificationLevel
	}{
		{"", model.NotificationInfo},
		{"info", model.NotificationInfo},
		{"success", model.NotificationSuccess},
		{"warning", model.NotificationWarning},
		{"error", model.NotificationError},
		{"bogus", model.NotificationInfo},
	}
	for _, tc := range cases {
		if got := commandLevelToModel(tc.level); got != tc.want {
			t.Fatalf("commandLevelToModel(%q) = %v, want %v", tc.level, got, tc.want)
		}
	}
}

// TestCommandMenuResultSpec proves toast/view results map to a multi-line
// in-app notification spec with the declared title, message and mapped level.
func TestCommandMenuResultSpec(t *testing.T) {
	spec := commandResultSpec(frontend.CommandResult{Action: "toast", Title: "T", Message: "M", Level: "warning"})
	if spec.Title != "T" || spec.Message != "M" || spec.Level != model.NotificationWarning {
		t.Fatalf("toast spec = %+v, want title T message M level warning", spec)
	}

	// view keeps the same toast shape (the notification fires alongside the
	// command_view page navigation — see TestCommandMenuActionViewReturnsCommandViewMsg).
	spec = commandResultSpec(frontend.CommandResult{Action: "view", Title: "V", Message: "body", Level: ""})
	if spec.Title != "V" || spec.Message != "body" || spec.Level != model.NotificationInfo {
		t.Fatalf("view spec = %+v, want title V message body level info", spec)
	}
}

// TestCommandMenuSideEffects proves the exec side effect reports success on a
// clean run and an error spec (with the error message) on failure. open_url is
// intentionally not unit-tested: open.Start would launch a real browser.
func TestCommandMenuSideEffects(t *testing.T) {
	// exec success
	spec := runCommandSideEffects(frontend.CommandResult{Action: "exec", Command: "true"})
	if spec.Level != model.NotificationSuccess {
		t.Fatalf("exec success spec = %+v, want level success", spec)
	}
	if spec.Message != "true" {
		t.Fatalf("exec success spec.Message = %q, want %q", spec.Message, "true")
	}

	// exec failure (the binary does not exist)
	spec = runCommandSideEffects(frontend.CommandResult{Action: "exec", Command: "definitely-not-a-real-binary-xyz"})
	if spec.Level != model.NotificationError {
		t.Fatalf("exec failure spec = %+v, want level error", spec)
	}
	if spec.Message == "" {
		t.Fatal("exec failure spec.Message is empty, want the exec error")
	}

	// Unknown/empty action: ignored (empty spec; the caller skips Notify).
	spec = runCommandSideEffects(frontend.CommandResult{Action: "bogus"})
	if spec.Level != model.NotificationInfo || spec.Title != "" || spec.Message != "" {
		t.Fatalf("unknown action spec = %+v, want zero spec", spec)
	}
}

// --- S1 view navigation ---

// TestCommandMenuActionViewReturnsCommandViewMsg proves the S1 upgrade: a
// "view" result keeps firing the in-app notification (toast 照发) and the
// command's tea.Cmd returns a commandViewMsg carrying the title and the message
// split into lines (\r\n normalized, empty lines preserved). Main.Update routes
// that msg to the command_view page via the foxful UnknownMsgHandler (wired in
// frontend.go).
func TestCommandMenuActionViewReturnsCommandViewMsg(t *testing.T) {
	ran := false
	cmd := frontend.Command{
		Key:   "cmdmenu_view",
		Title: "View",
		Run: func(frontend.CommandContext) frontend.CommandResult {
			ran = true
			return frontend.CommandResult{Action: "view", Title: "ViewTitle", Message: "line1\r\nline2\n\nline3"}
		},
	}
	m := &CommandMenu{BaseMenu: BaseMenu{}, cmd: cmd}
	a := &model.App{} // zero app: Notify is a no-op without a program

	msg := commandActionCmd(a, m)()
	if !ran {
		t.Fatal("Run not executed")
	}
	vm, ok := msg.(commandViewMsg)
	if !ok {
		t.Fatalf("commandActionCmd() msg = %T (%v), want commandViewMsg", msg, msg)
	}
	if vm.Title != "ViewTitle" {
		t.Fatalf("commandViewMsg.Title = %q, want %q", vm.Title, "ViewTitle")
	}
	want := []string{"line1", "line2", "", "line3"}
	if !slices.Equal(vm.Lines, want) {
		t.Fatalf("commandViewMsg.Lines = %q, want %q", vm.Lines, want)
	}
}

// TestCommandMenuActionToastReturnsNilAndNotifies proves a "toast" result
// keeps the command's tea.Cmd returning nil (no page navigation) and delivers
// the result as an in-app notification (a.Notify(commandResultSpec(res))). The
// zero-app run pins the no-navigation contract; the real-app run pins that the
// toast spec renders through the app's notification path (mirrors
// TestCheckUpdateResultRendersDirectlyInTUI; the async a.Notify call itself is
// exercised in the zero-app run — it is a no-op without a program).
func TestCommandMenuActionToastReturnsNilAndNotifies(t *testing.T) {
	ran := false
	cmd := frontend.Command{
		Key:   "cmdmenu_toast",
		Title: "Toast",
		Run: func(frontend.CommandContext) frontend.CommandResult {
			ran = true
			return frontend.CommandResult{Action: "toast", Title: "ToastTitle", Message: "ToastBody"}
		},
	}
	m := &CommandMenu{BaseMenu: BaseMenu{}, cmd: cmd}

	a := &model.App{} // zero app: Notify is a no-op without a program
	if msg := commandActionCmd(a, m)(); msg != nil {
		t.Fatalf("toast commandActionCmd() msg = %v, want nil (no page navigation)", msg)
	}
	if !ran {
		t.Fatal("Run not executed")
	}

	// The toast result is delivered through app.Notify: process the equivalent
	// ShowNotificationMsg the app would receive and assert it renders.
	app, _ := newFormPageTestApp(t)
	_, _ = app.Update(model.ShowNotificationMsg{Spec: commandResultSpec(frontend.CommandResult{
		Action: "toast", Title: "ToastTitle", Message: "ToastBody",
	})})
	if view := app.View().Content; !strings.Contains(view, "ToastTitle") {
		t.Fatalf("toast not rendered through the app notification path:\n%s", view)
	}
}

// TestSplitLinesNormalizesAndPreservesEmptyLines proves splitLines normalizes
// \r\n to \n and keeps intentional blank lines (including a trailing one).
func TestSplitLinesNormalizesAndPreservesEmptyLines(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty message", "", []string{""}},
		{"single line", "hello", []string{"hello"}},
		{"crlf", "a\r\nb", []string{"a", "b"}},
		{"crlf blank line", "a\r\n\r\nb", []string{"a", "", "b"}},
		{"blank line", "a\n\nb", []string{"a", "", "b"}},
		{"trailing newline", "a\n", []string{"a", ""}},
		{"mixed", "a\r\nb\n\nc\r\n", []string{"a", "b", "", "c", ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := splitLines(tc.input); !slices.Equal(got, tc.want) {
				t.Fatalf("splitLines(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
