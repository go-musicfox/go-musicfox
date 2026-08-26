package wasm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// RegistrySink is the registration target of WASM plugin commands. TUI and
// WebUI each implement it (TUI: tuiWasmSink in internal/ui); the sink decides
// how a frontend.Command becomes visible/actionable in its frontend.
type RegistrySink interface {
	RegisterCommands(p *Plugin, cmds []frontend.Command) error
}

// CommandsOf maps every manifest-declared menu of a plugin into a track-B
// frontend.Command (PluginID stamped with the plugin id). The command's Run
// invokes the plugin export through callWasm; Show is nil (always available) —
// availability gating is delegated to the frontend's own rules.
func CommandsOf(p *Plugin) []frontend.Command {
	cmds := make([]frontend.Command, 0, len(p.Menus))
	for _, decl := range p.Menus {
		cmds = append(cmds, frontend.Command{
			Key:      decl.Key,
			Title:    decl.Title,
			After:    decl.After,
			PluginID: p.ID,
			Show:     nil,
			Run:      callWasm(p, decl),
		})
	}
	return cmds
}

// callWasm returns the execution closure of one WASM menu command: the
// frontend.CommandContext is copied field-by-field into the wasm.Request
// context (keeping the guest wire protocol unchanged — internal/wasm/
// contract.go is frozen), the request is marshalled, p.Run is invoked (the
// plugin's own 5s watchdog handles hangs) and the parsed wasm.Response is
// mapped into a frontend.CommandResult. An unknown response action maps to an
// empty result (the presenter ignores it), mirroring the pre-track-B WASM
// action-menu behavior.
func callWasm(p *Plugin, decl MenuDecl) func(frontend.CommandContext) frontend.CommandResult {
	return func(cmdCtx frontend.CommandContext) frontend.CommandResult {
		ctx := RequestContext{
			UserID:   cmdCtx.UserID,
			UserName: cmdCtx.UserName,
			Playing:  cmdCtx.Playing,
		}
		if cmdCtx.Song != nil {
			ctx.Song = &SongInfo{
				ID:     cmdCtx.Song.ID,
				Name:   cmdCtx.Song.Name,
				Artist: cmdCtx.Song.Artist,
				Album:  cmdCtx.Song.Album,
			}
		}
		reqJSON, err := json.Marshal(Request{
			Version: ProtocolVersion,
			Action:  decl.Key,
			Args:    decl.Args,
			Context: ctx,
		})
		if err != nil {
			return callError(err)
		}
		out, err := p.Run(context.Background(), decl, reqJSON)
		if err != nil {
			return callError(err)
		}
		var resp Response
		if err := json.Unmarshal(out, &resp); err != nil {
			return callError(err)
		}
		switch resp.Action {
		case "toast", "view", "open_url", "exec":
			return frontend.CommandResult{
				Action:  resp.Action,
				Title:   resp.Title,
				Message: resp.Message,
				Level:   resp.Level,
				URL:     resp.URL,
				Command: resp.Command,
				Args:    resp.Args,
			}
		default:
			return frontend.CommandResult{}
		}
	}
}

// callError maps a plugin call/marshal failure into a toast-able error result
// (the presenter renders Title/Message/Level "error").
func callError(err error) frontend.CommandResult {
	return frontend.CommandResult{
		Action:  "toast",
		Title:   "WASM 插件执行失败",
		Message: err.Error(),
		Level:   "error",
	}
}

// LoadAndRegister is the one-stop loader: NewManager → LoadDir(dir) → per
// plugin sink.RegisterCommands(p, CommandsOf(p)). It returns the manager (nil
// only when the runtime failed to initialize) and the collected errors — a
// single failing plugin or registration does not stop the others.
func LoadAndRegister(ctx context.Context, dir string, sink RegistrySink) (*Manager, []error) {
	mgr, err := NewManager()
	if err != nil {
		return nil, []error{fmt.Errorf("wasm: new manager: %w", err)}
	}
	errs := mgr.LoadDir(ctx, dir)
	for _, p := range mgr.Plugins() {
		if err := sink.RegisterCommands(p, CommandsOf(p)); err != nil {
			errs = append(errs, fmt.Errorf("wasm: register commands for plugin %q: %w", p.ID, err))
		}
	}
	return mgr, errs
}
