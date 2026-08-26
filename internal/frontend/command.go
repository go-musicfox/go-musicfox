// Track B (UI-agnostic command contribution) contract and registry.
//
// This file, together with frontend.go, makes up the frontend package, which
// keeps zero business dependencies (no internal/* imports and no third-party
// business libraries; standard library only). The command contract is pure
// data plus func fields; its JSON shapes align with
// internal/wasm/contract.go so a later track can map/alias between them.
package frontend

// SongInfo is a minimal snapshot of the currently playing song
// (UI-agnostic; does not reference structs.Song).
type SongInfo struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
}

// CommandContext is the context snapshot passed to a command (pure data).
type CommandContext struct {
	UserID   int64     `json:"userId"`
	UserName string    `json:"userName"`
	Playing  bool      `json:"playing"`
	Song     *SongInfo `json:"song,omitempty"`
}

// CommandResult is what a command returns. Action decides host behavior:
// "" / "toast" / "view" / "open_url" / "exec" / "data".
type CommandResult struct {
	Action  string   `json:"action"`
	Title   string   `json:"title,omitempty"`
	Message string   `json:"message,omitempty"`
	Level   string   `json:"level,omitempty"` // info|success|warning|error
	URL     string   `json:"url,omitempty"`
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	Data    any      `json:"data,omitempty"`
}

// Command is a track-B command contribution. Key/Title/After are used for TUI
// main-menu placement; PluginID is stamped by the registering side (any
// registration inside a ui.WithPlugin scope gets it automatically).
type Command struct {
	Key      string
	Title    string
	After    string // main-menu after-anchor predecessor key; empty = chain tail
	PluginID string
	Show     func(CommandContext) bool          // may be nil (always true): whether the command may run now
	Run      func(CommandContext) CommandResult // required
}

var (
	cmdRegistry = map[string]Command{}
	cmdOrder    []string
)

// RegisterCommand registers a command. It panics on an empty Key, a nil Run,
// or a duplicate Key.
func RegisterCommand(cmd Command) {
	if cmd.Key == "" {
		panic("frontend: RegisterCommand with empty Key")
	}
	if cmd.Run == nil {
		panic("frontend: RegisterCommand with nil Run for key " + cmd.Key)
	}
	mu.Lock()
	defer mu.Unlock()
	if _, ok := cmdRegistry[cmd.Key]; ok {
		panic("frontend: duplicate command registration for key " + cmd.Key)
	}
	cmdRegistry[cmd.Key] = cmd
	cmdOrder = append(cmdOrder, cmd.Key)
}

// Commands returns a snapshot of all registered commands in registration
// order. The returned slice is a copy; mutating it does not affect the
// registry.
func Commands() []Command {
	mu.RLock()
	defer mu.RUnlock()
	cmds := make([]Command, len(cmdOrder))
	for i, key := range cmdOrder {
		cmds[i] = cmdRegistry[key]
	}
	return cmds
}

// CommandByKey returns the command registered under key.
func CommandByKey(key string) (Command, bool) {
	mu.RLock()
	defer mu.RUnlock()
	cmd, ok := cmdRegistry[key]
	return cmd, ok
}
