package ui

import (
	"context"
	"encoding/json"
	"os/exec"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/internal/wasm"
)

// WasmPluginMenu renders a WASM plugin menu entry: entering/activating it
// invokes the plugin's exported function with a JSON request and shows the
// response (toast/view/open_url/exec) via the app's thread-safe notification
// channel.
type WasmPluginMenu struct {
	BaseMenu
	plugin *wasm.Plugin
	decl   wasm.MenuDecl
}

// GetMenuKey returns the manifest-declared menu registry key.
func (m *WasmPluginMenu) GetMenuKey() string {
	return m.decl.Key
}

// IsPlayable reports whether the menu is playable. A WASM menu is an action
// menu, never a song list.
func (m *WasmPluginMenu) IsPlayable() bool {
	return false
}

// IsLocatable reports whether the menu participates in playback auto-locate.
func (m *WasmPluginMenu) IsLocatable() bool {
	return false
}

// MenuViews returns a single static item; the menu is never rendered as a
// submenu (the main-menu entry triggers Action directly).
func (m *WasmPluginMenu) MenuViews() []model.MenuItem {
	return []model.MenuItem{{Title: m.decl.Title}}
}

// SubMenu returns nil: a WASM menu produces no submenu.
func (m *WasmPluginMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

// Action triggers the plugin call and stays on the current page. The non-nil
// page/cmd skip submenu navigation; the command runs in a bubbletea goroutine
// and delivers the result via app.Notify (thread-safe).
func (m *WasmPluginMenu) Action(a *model.App, _ int) (model.Page, tea.Cmd) {
	return a.MustMain(), wasmActionCmd(a, m.plugin, m.decl, m)
}

// wasmLevelToModel maps a wasm.Response Level string to model.NotificationLevel.
// Unknown or empty levels default to NotificationInfo.
func wasmLevelToModel(level string) model.NotificationLevel {
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

// wasmResponseSpec builds the notification spec for a response whose action is
// toast/view (rendered as a multi-line in-app notification; MVP: no separate
// page/popup — view content is delivered through the toast message).
func wasmResponseSpec(resp wasm.Response) model.NotificationSpec {
	return model.NotificationSpec{
		Level:   wasmLevelToModel(resp.Level),
		Title:   resp.Title,
		Message: resp.Message,
	}
}

// runWasmSideEffects executes the side effects of open_url/exec responses and
// returns the notification spec for the outcome. Unknown/empty actions are
// ignored and return an empty spec (the caller skips Notify for them).
//
// open_url is intentionally not unit-tested: it would open a real browser via
// open.Start. exec side effects are covered by tests.
func runWasmSideEffects(resp wasm.Response) model.NotificationSpec {
	switch resp.Action {
	case "open_url":
		if err := open.Start(resp.URL); err != nil {
			return model.NotificationSpec{Level: model.NotificationError, Title: "打开链接失败", Message: err.Error()}
		}
		return model.NotificationSpec{Level: model.NotificationInfo, Title: "已打开链接", Message: resp.URL}
	case "exec":
		if err := exec.Command(resp.Command, resp.Args...).Run(); err != nil {
			return model.NotificationSpec{Level: model.NotificationError, Title: "命令执行失败", Message: err.Error()}
		}
		return model.NotificationSpec{Level: model.NotificationSuccess, Title: "命令执行成功", Message: resp.Command}
	default:
		return model.NotificationSpec{}
	}
}

// wasmActionCmd builds the tea.Cmd that runs a WASM plugin menu action: the
// request is marshalled, the plugin's export is invoked (the plugin's own 5s
// watchdog handles hangs), and the parsed response is delivered through the
// app's thread-safe notification channel. Every path returns nil tea.Msg.
func wasmActionCmd(a *model.App, p *wasm.Plugin, decl wasm.MenuDecl, m *WasmPluginMenu) tea.Cmd {
	return func() tea.Msg {
		req := wasm.Request{
			Version: wasm.ProtocolVersion,
			Action:  decl.Key,
			Args:    decl.Args,
			Context: buildWasmContext(m),
		}
		reqJSON, err := json.Marshal(req)
		if err != nil {
			a.Notify(wasmErrSpec(err))
			return nil
		}
		out, err := p.Run(context.Background(), decl, reqJSON)
		if err != nil {
			a.Notify(wasmErrSpec(err))
			return nil
		}
		var resp wasm.Response
		if err := json.Unmarshal(out, &resp); err != nil {
			a.Notify(wasmErrSpec(err))
			return nil
		}
		switch resp.Action {
		case "toast", "view":
			a.Notify(wasmResponseSpec(resp))
		case "open_url", "exec":
			a.Notify(runWasmSideEffects(resp))
		}
		return nil
	}
}

// wasmErrSpec builds the error notification spec for a failed plugin call or an
// unparseable response.
func wasmErrSpec(err error) model.NotificationSpec {
	return model.NotificationSpec{
		Level:   model.NotificationError,
		Title:   "WASM 插件执行失败",
		Message: err.Error(),
	}
}

// buildWasmContext assembles the request context from the current app state:
// the logged-in user (zero values when not logged in), the player's playing
// state and the currently playing song. Defensive: an empty playlist or nil
// services degrade to the zero values without a panic.
func buildWasmContext(m *WasmPluginMenu) wasm.RequestContext {
	var ctx wasm.RequestContext
	if u := m.User(); u != nil {
		ctx.UserID = u.UserId
		ctx.UserName = u.Nickname
	}
	p := m.Player()
	if p == nil {
		return ctx
	}
	ctx.Playing = p.State() == types.Playing
	songs := p.Playlist()
	if len(songs) == 0 {
		return ctx
	}
	idx := p.CurSongIndex()
	song := songs[len(songs)-1]
	if idx >= 0 && idx < len(songs) {
		song = songs[idx]
	}
	if song.Id != 0 || song.Name != "" {
		ctx.Song = &wasm.SongInfo{
			ID:     song.Id,
			Name:   song.Name,
			Artist: song.ArtistName(),
			Album:  song.Album.Name,
		}
	}
	return ctx
}
