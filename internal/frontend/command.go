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
	// cmdPos tracks each registered key's index in cmdOrder so ReplaceCommand
	// can re-insert a hot-reloaded key at its original position.
	cmdPos = map[string]int{}
	// tombPos remembers the last known cmdOrder index of a key removed by
	// UnregisterCommand. A later ReplaceCommand of that key (the WASM hot-reload
	// Stop→Unregister→re-Start→Replace sequence) restores the position instead
	// of drifting to the tail. Entries are cleared once the key is re-registered.
	tombPos = map[string]int{}
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
	delete(tombPos, cmd.Key)
	cmdRegistry[cmd.Key] = cmd
	cmdPos[cmd.Key] = len(cmdOrder)
	cmdOrder = append(cmdOrder, cmd.Key)
}

// UnregisterCommand deletes the command registered under key. It is a no-op
// when key is not registered. The key's last order position is remembered so a
// later ReplaceCommand (hot reload) can restore it.
func UnregisterCommand(key string) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := cmdRegistry[key]; !ok {
		return
	}
	delete(cmdRegistry, key)
	if pos, ok := cmdPos[key]; ok {
		tombPos[key] = pos
		delete(cmdPos, key)
	}
	for i, k := range cmdOrder {
		if k == key {
			cmdOrder = append(cmdOrder[:i], cmdOrder[i+1:]...)
			break
		}
	}
}

// ReplaceCommand replaces the definition of the command registered under
// cmd.Key while keeping its registration-order position. When key is not
// registered it behaves like RegisterCommand (appends at the tail) — EXCEPT
// when the key was previously unregistered: the last known position (recorded
// by UnregisterCommand) is restored, so the WASM hot-reload sequence
// (Stop unregisters → new generation Start replaces) keeps the command at its
// original order slot instead of drifting to the tail. Unlike RegisterCommand
// it is allowed to overwrite an existing key, so it is the only overwrite entry
// point. It shares RegisterCommand's panic semantics (empty Key / nil Run). The
// new cmd's PluginID replaces the old one — the caller is responsible for
// stamping it correctly.
func ReplaceCommand(cmd Command) {
	if cmd.Key == "" {
		panic("frontend: ReplaceCommand with empty Key")
	}
	if cmd.Run == nil {
		panic("frontend: ReplaceCommand with nil Run for key " + cmd.Key)
	}
	mu.Lock()
	defer mu.Unlock()
	if _, ok := cmdRegistry[cmd.Key]; ok {
		cmdRegistry[cmd.Key] = cmd
		return
	}
	// Not registered: restore the original order slot when the key was
	// unregistered earlier (hot reload), else append at the tail.
	if pos, ok := tombPos[cmd.Key]; ok {
		delete(tombPos, cmd.Key)
		if pos > len(cmdOrder) {
			pos = len(cmdOrder) // clamp: other keys may have been removed too
		}
		newOrder := make([]string, 0, len(cmdOrder)+1)
		newOrder = append(newOrder, cmdOrder[:pos]...)
		newOrder = append(newOrder, cmd.Key)
		newOrder = append(newOrder, cmdOrder[pos:]...)
		cmdOrder = newOrder
		for i := pos; i < len(cmdOrder); i++ {
			cmdPos[cmdOrder[i]] = i
		}
		cmdRegistry[cmd.Key] = cmd
		return
	}
	cmdRegistry[cmd.Key] = cmd
	cmdPos[cmd.Key] = len(cmdOrder)
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
