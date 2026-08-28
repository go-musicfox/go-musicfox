package webui

import (
	"net/http"

	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// commandExecAllowed is the WebUI side-effect policy: a command whose Run
// returns Action "exec" is always rejected, because spawning a local process
// from a browser-exposed endpoint would cross the trust boundary that the
// session cookie establishes (a browser tab is not a shell).
const commandExecAllowed = false

// commandItem is the wire form of one command in the GET /api/commands list.
type commandItem struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	After    string `json:"after,omitempty"`
	PluginID string `json:"pluginId,omitempty"`
}

// handleCommandsList answers the list of runnable commands. Commands of
// disabled plugins and commands whose Show gate reports false are filtered
// out, so the browser only sees what it could actually run right now.
func (s *Server) handleCommandsList(w http.ResponseWriter, r *http.Request) {
	if !s.backend.Ready() {
		writeJSONError(w, http.StatusInternalServerError, "engine unavailable")
		return
	}
	ctx := s.backend.CommandContext()
	items := make([]commandItem, 0, 8)
	for _, cmd := range frontend.Commands() {
		if !configs.IsPluginEnabled(cmd.PluginID) {
			continue
		}
		if cmd.Show != nil && !cmd.Show(ctx) {
			continue
		}
		items = append(items, commandItem{
			Key:      cmd.Key,
			Title:    cmd.Title,
			After:    cmd.After,
			PluginID: cmd.PluginID,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "data": items})
}

// handleCommandExec runs the command registered under the URL key.
//
// NotFound is returned both for an unknown key and for a key belonging to a
// disabled plugin, so the endpoint does not leak the existence of disabled
// commands. The Show gate (if any) is re-checked at execution time against the
// current context, and "exec" results are rejected by commandExecAllowed.
func (s *Server) handleCommandExec(w http.ResponseWriter, r *http.Request) {
	if !s.backend.Ready() {
		writeJSONError(w, http.StatusInternalServerError, "engine unavailable")
		return
	}
	key := r.PathValue("key")
	cmd, ok := frontend.CommandByKey(key)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "命令不存在")
		return
	}
	if !configs.IsPluginEnabled(cmd.PluginID) {
		writeJSONError(w, http.StatusNotFound, "命令不存在")
		return
	}
	ctx := s.backend.CommandContext()
	if cmd.Show != nil && !cmd.Show(ctx) {
		writeJSONError(w, http.StatusForbidden, "命令当前不可用")
		return
	}
	res := cmd.Run(ctx)

	if res.Action == "exec" {
		writeJSONError(w, http.StatusForbidden, "exec 已在 WebUI 中禁用")
		return
	}
	if res.Action == "open_url" && res.URL != "" {
		_ = open.Start(res.URL)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "action": "open_url"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"action":  res.Action,
		"title":   res.Title,
		"message": res.Message,
		"level":   res.Level,
		"data":    res.Data,
	})
}
