// Track-B command adaptation: a frontend.Command becomes a TUI CommandMenu
// (the generalized action-menu pattern that WASM plugin menus and track-B
// commands now share). Entering/activating it runs the command's Run with the
// current player snapshot and presents the result (toast/view/open_url/exec)
// via the app's thread-safe notification channel; a "view" result additionally
// returns a commandViewMsg that opens the independent command_view page.
package ui

import (
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// CommandMenu renders a track-B frontend.Command as a TUI menu entry. A command
// menu is an action menu: the main-menu entry triggers Action directly and the
// command runs asynchronously so a slow command (WASM calls may take up to 5s)
// never blocks the UI.
type CommandMenu struct {
	BaseMenu
	cmd frontend.Command
}

// GetMenuKey returns the command's registry key.
func (m *CommandMenu) GetMenuKey() string {
	return m.cmd.Key
}

// IsPlayable reports whether the menu is playable. A command menu is an action
// menu, never a song list.
func (m *CommandMenu) IsPlayable() bool {
	return false
}

// IsLocatable reports whether the menu participates in playback auto-locate.
func (m *CommandMenu) IsLocatable() bool {
	return false
}

// MenuViews returns a single static item; the menu is never rendered as a
// submenu (the main-menu entry triggers Action directly).
func (m *CommandMenu) MenuViews() []model.MenuItem {
	return []model.MenuItem{{Title: m.cmd.Title}}
}

// SubMenu returns nil: a command menu produces no submenu.
func (m *CommandMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

// Action triggers the command and stays on the current page. The non-nil
// page/cmd skip submenu navigation; the command runs in a bubbletea goroutine
// and delivers the result via app.Notify (thread-safe).
func (m *CommandMenu) Action(a *model.App, _ int) (model.Page, tea.Cmd) {
	return a.MustMain(), commandActionCmd(a, m)
}

// commandLevelToModel maps a CommandResult Level string to
// model.NotificationLevel. Unknown or empty levels default to NotificationInfo.
func commandLevelToModel(level string) model.NotificationLevel {
	switch level {
	case "success":
		return model.NotificationSuccess
	case "warning":
		return model.NotificationWarning
	case "error":
		return model.NotificationError
	default:
		return model.NotificationInfo
	}
}

// commandResultSpec builds the notification spec for a result whose action is
// toast/view (rendered as a multi-line in-app notification; a "view" result
// additionally opens the command_view page — see commandActionCmd).
func commandResultSpec(res frontend.CommandResult) model.NotificationSpec {
	return model.NotificationSpec{
		Level:   commandLevelToModel(res.Level),
		Title:   res.Title,
		Message: res.Message,
	}
}

// commandViewMsg is delivered by commandActionCmd for Action=="view" results
// (the only non-nil tea.Msg the command produces). Main.Update dispatches it
// to the TUI assembly point via the foxful Options.UnknownMsgHandler, which
// builds the command_view page.
type commandViewMsg struct {
	Title string
	Lines []string
}

// splitLines splits a command result message into display lines: \r\n is
// normalized to \n and empty lines are preserved (the message body may carry
// intentional blank lines).
func splitLines(message string) []string {
	return strings.Split(strings.ReplaceAll(message, "\r\n", "\n"), "\n")
}

// runCommandSideEffects executes the side effects of open_url/exec results and
// returns the notification spec for the outcome. Unknown/empty actions are
// ignored and return an empty spec (the caller skips Notify for them).
//
// open_url is intentionally not unit-tested: it would open a real browser via
// open.Start. exec side effects are covered by tests.
func runCommandSideEffects(res frontend.CommandResult) model.NotificationSpec {
	switch res.Action {
	case "open_url":
		if err := open.Start(res.URL); err != nil {
			return model.NotificationSpec{Level: model.NotificationError, Title: "打开链接失败", Message: err.Error()}
		}
		return model.NotificationSpec{Level: model.NotificationInfo, Title: "已打开链接", Message: res.URL}
	case "exec":
		if err := exec.Command(res.Command, res.Args...).Run(); err != nil {
			return model.NotificationSpec{Level: model.NotificationError, Title: "命令执行失败", Message: err.Error()}
		}
		return model.NotificationSpec{Level: model.NotificationSuccess, Title: "命令执行成功", Message: res.Command}
	default:
		return model.NotificationSpec{}
	}
}

// commandActionCmd builds the tea.Cmd that runs a track-B command menu action.
// The command is resolved from the frontend registry by key at action time, so
// a hot reload that replaces the registered definition (P8) executes the
// CURRENT command even when the menu instance was built from a pre-reload
// snapshot. Gating runs first: ① [plugins] disabled — a disabled plugin's
// command is rejected with a toast; ② the command's own Show gate — a command
// that is not currently available is rejected with a toast. Then Run executes
// with the current player snapshot and the result is presented by action
// (toast → commandResultSpec notification; view → the same notification plus
// a commandViewMsg opening the command_view page; open_url/exec →
// runCommandSideEffects). Notifications are delivered via app.Notify
// (thread-safe); the view branch is the only path returning a non-nil tea.Msg.
func commandActionCmd(a *model.App, m *CommandMenu) tea.Cmd {
	return func() tea.Msg {
		// Resolve the current definition by key; fall back to the captured one
		// (defensive — directly-constructed menus in tests).
		cmd, ok := frontend.CommandByKey(m.cmd.Key)
		if !ok {
			cmd = m.cmd
		}

		// ① [plugins] disabled gate.
		if !IsPluginEnabled(cmd.PluginID) {
			a.Notify(model.NotificationSpec{
				Level:   model.NotificationWarning,
				Title:   "命令不可用",
				Message: "插件已禁用",
			})
			return nil
		}

		// ② Show gate (nil Show means always available). Defensive: a nil
		// player degrades to the zero-value snapshot.
		var ctx frontend.CommandContext
		if p := m.Player(); p != nil {
			ctx = p.CommandContext()
		}
		if cmd.Show != nil && !cmd.Show(ctx) {
			a.Notify(model.NotificationSpec{
				Level:   model.NotificationInfo,
				Title:   "命令不可用",
				Message: "命令当前不可用",
			})
			return nil
		}

		// ③ Run the command with the current player snapshot.
		res := cmd.Run(ctx)

		// ④ Present the result by action.
		switch res.Action {
		case "toast":
			a.Notify(commandResultSpec(res))
		case "view":
			// The toast still fires (instant visibility); the view msg drives
			// the independent command_view page via the foxful
			// UnknownMsgHandler (wired in frontend.go).
			a.Notify(commandResultSpec(res))
			return commandViewMsg{Title: res.Title, Lines: splitLines(res.Message)}
		case "open_url", "exec":
			a.Notify(runCommandSideEffects(res))
		}
		return nil
	}
}

// registerCommandMenus adapts every registered track-B command into a TUI
// CommandMenu provider and main-menu entry. Called from NewNetease after WASM
// plugins load so command menus and their main-menu items join the after-anchor
// chain from the start (before the main menu is constructed in internal/commands).
//
// The provider resolves the command from the frontend registry BY KEY at build
// time (and commandActionCmd re-resolves at action time), so a hot reload that
// replaces a command definition needs no re-adaptation: the existing menu key's
// provider keeps serving the current command. The dup-guard therefore only skips
// re-registering an already-adapted key (RegisterMenu would panic on a duplicate
// key); it is not a "freeze the first closure" guard. Main-menu item titles are
// captured at adaptation time and are only refreshed by a restart — the TUI hot
// reload boundary (docs/plugin_ecosystem.md §八 P8).
func registerCommandMenus() {
	for _, cmd := range frontend.Commands() {
		if _, dup := menuRegistry[cmd.Key]; dup {
			// Already adapted: the provider resolves the current command by
			// key, so nothing to refresh here.
			continue
		}
		RegisterMenu(cmd.Key, func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
			cur, ok := frontend.CommandByKey(cmd.Key)
			if !ok {
				// Defensive: the key was unregistered after adaptation
				// (plugin removed); fall back to the adaptation-time snapshot.
				cur = cmd
			}
			return &CommandMenu{BaseMenu: base, cmd: cur}, nil
		})
		registerItem := func() {
			if cmd.After == "" {
				RegisterMainMenuItem(cmd.Key, cmd.Title)
			} else {
				RegisterMainMenuItemAfter(cmd.Key, cmd.Title, cmd.After, nil)
			}
		}
		if cmd.PluginID != "" {
			// Declare the main-menu item inside the plugin's scope so the item
			// carries PluginID and [plugins] disabled hides it in NewMainMenu
			// (idempotent merge: the plugin name recorded earlier is kept).
			WithPlugin(cmd.PluginID, "", registerItem)
		} else {
			registerItem()
		}
	}
}
