package ui

import (
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

	// view renders the same shape (MVP: content through the toast message).
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
