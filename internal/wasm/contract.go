// Package wasm implements the runtime side of the go-musicfox WASM plugin
// system: plugin discovery from a directory, manifest parsing and validation,
// SHA-256 verification, and executing guest plugins compiled from standard Go
// (//go:wasmexport, GOOS=wasip1) with the wazero runtime.
package wasm

// ProtocolVersion is the version of the host<->guest JSON contract. A plugin
// may inspect request.Version and refuse to serve unsupported versions.
const ProtocolVersion = 1

// Request is what the host sends to a WASM plugin's exported function.
type Request struct {
	Version int            `json:"version"`        // = ProtocolVersion
	Action  string         `json:"action"`         // menu key that triggered the call
	Args    map[string]any `json:"args,omitempty"` // static args from the manifest menu entry
	Context RequestContext `json:"context,omitempty"`
}

// RequestContext carries current player state to the plugin.
type RequestContext struct {
	UserID   int64     `json:"userId,omitempty"`
	UserName string    `json:"userName,omitempty"`
	Playing  bool      `json:"playing"`
	Song     *SongInfo `json:"song,omitempty"`
}

// SongInfo describes the currently playing song.
type SongInfo struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
}

// Response is what the plugin returns. Action is one of:
// "toast" (Title+Message+Level), "view" (Title+Message rendered as text),
// "open_url" (URL), "exec" (Command+Args). Unknown actions are ignored.
type Response struct {
	Action  string   `json:"action"` // "toast" | "view" | "open_url" | "exec"
	Title   string   `json:"title,omitempty"`
	Message string   `json:"message,omitempty"`
	Level   string   `json:"level,omitempty"` // "info" | "success" | "warning" | "error"
	URL     string   `json:"url,omitempty"`
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
}

// LevelValid reports whether the response Level is one of the documented
// values. An empty level is allowed.
func (r *Response) LevelValid() bool {
	switch r.Level {
	case "", "info", "success", "warning", "error":
		return true
	default:
		return false
	}
}
