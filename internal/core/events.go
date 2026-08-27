package core

import (
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

// Event names on the app-wide event bus (P4, docs/plugin_ecosystem.md §四).
// The wire names align with the WebUI frame names so a subscriber can map an
// event to its frame 1:1. Playback events are double-written: the existing
// observer call first (TUI rendering unchanged), then emitter.Emit.
const (
	EvSongChanged  = "player.song_changed"
	EvStateChanged = "player.state_changed"
	EvPosition     = "player.position"
	EvPlaylistEnd  = "player.playlist_exhausted"
	EvRerender     = "player.rerender"
	EvLogin        = "auth.login_succeeded"
	EvStartupPhase = "startup.phase"
)

// emit forwards a playback event to the app-wide event bus. The emitter is
// optional (nil in empty players and frontends without a bus); Emit is
// synchronous but fast and listeners must be enqueue-only — they run on the
// emitting player goroutine (docs/plugin_ecosystem.md §四 并发契约).
func (p *Player) emit(name string, payload any) {
	if p.emitter == nil {
		return
	}
	_ = p.emitter.Emit(nil, name, payload)
}

// emit forwards a core-level event (startup/login, which live on the Engine)
// to the app-wide event bus. The bus is registered by the eventBusPlugin; a
// missing bus is a no-op (nil-safe).
func (e *Engine) emit(name string, payload any) {
	emitter, ok := framework.ServiceOf[*framework.EventEmitter](e.ctx, ServiceEventBus)
	if !ok {
		return
	}
	_ = emitter.Emit(e.ctx, name, payload)
}

// songEventPayload trims a structs.Song to the frontend-relevant event fields
// (never the whole song), mirroring the WebUI song_changed frame data shape.
func songEventPayload(song structs.Song) map[string]any {
	return map[string]any{
		"id":              song.Id,
		"name":            song.Name,
		"artist":          song.ArtistName(),
		"album":           song.Album.Name,
		"picUrl":          song.PicUrl,
		"durationSeconds": song.Duration.Seconds(),
	}
}

// loginEventPayload builds the auth.login_succeeded event data, mirroring the
// WebUI login frame data shape.
func loginEventPayload(user *structs.User) map[string]any {
	if user == nil {
		return map[string]any{"user": map[string]any{}}
	}
	return map[string]any{
		"user": map[string]any{
			"userId":    user.UserId,
			"nickname":  user.Nickname,
			"avatarUrl": user.AvatarUrl,
		},
	}
}
