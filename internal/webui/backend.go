package webui

import (
	"context"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// Backend abstracts the WebUI data source: a local engine (standalone / GUI)
// or a remote headless daemon client (connect, S5-3). The Server only depends
// on this interface and never holds a *core.Engine. Every method is a pure
// query/forward — no playback state is maintained here.
type Backend interface {
	// Ready reports backend availability (engine non-nil / daemon connected).
	// When unavailable the auxiliary endpoints degrade (404/503/empty data)
	// and the HTTP layer never panics.
	Ready() bool
	// Dispatch executes a control command (local Dispatcher or remote
	// CtrlClient.Call).
	Dispatch(ctx context.Context, cmd string, args map[string]any) (any, error)
	// SubscribeEvents registers an event listener and returns the
	// unsubscribe function. handler's name is the WebUI frame name
	// (song_changed etc.); payload is the already-serialized event frame
	// ({"type":"event",...}, isomorphic to frontend.EventFrame).
	SubscribeEvents(handler func(name string, payload []byte)) func()
	// Playlist returns the trimmed playlist (id/name/artist/album) for
	// snapshots.
	Playlist() []map[string]any
	// PlayingInfo returns the current song's cover URL; ok=false means
	// unavailable (connect mode always false — the daemon status snapshot
	// carries no PicUrl).
	PlayingInfo() (picURL string, ok bool)
	// LyricState returns the lyric state; connect mode returns the empty
	// structure.
	LyricState() (fragments []lyricFragment, translated map[int64]string, currentIndex int, offsetMs int64)
	// CommandContext returns the command-context snapshot (command endpoints).
	CommandContext() frontend.CommandContext
}

// Compile-time assertion that localBackend satisfies Backend.
var _ Backend = (*localBackend)(nil)

// localBackend is the *core.Engine-bound Backend implementation. Its behavior
// is word-for-word identical to the pre-refactor Server, which consumed the
// engine directly.
type localBackend struct {
	engine *core.Engine
}

// Ready reports whether the engine is available.
func (b *localBackend) Ready() bool {
	return b.engine != nil
}

// Dispatch forwards the control command to the engine's Dispatcher.
func (b *localBackend) Dispatch(ctx context.Context, cmd string, args map[string]any) (any, error) {
	return core.NewDispatcher(b.engine).Dispatch(ctx, cmd, args)
}

// SubscribeEvents registers the core event-bus listeners that forward the
// frontend-relevant events to handler (reusing subscribeEmitter from
// events.go). A nil engine or a missing event bus yields a no-op unsubscribe.
func (b *localBackend) SubscribeEvents(handler func(name string, payload []byte)) func() {
	if b.engine == nil {
		return func() {}
	}
	emitter, ok := framework.ServiceOf[*framework.EventEmitter](b.engine.Ctx(), core.ServiceEventBus)
	if !ok {
		return func() {}
	}
	return subscribeEmitter(emitter, handler)
}

// Playlist maps the engine playlist to the trimmed snapshot shape
// (id/name/artist/album). A nil engine yields an empty list.
func (b *localBackend) Playlist() []map[string]any {
	playlist := make([]map[string]any, 0)
	if b.engine == nil {
		return playlist
	}
	for _, song := range b.engine.Player().Playlist() {
		playlist = append(playlist, map[string]any{
			"id":     song.Id,
			"name":   song.Name,
			"artist": song.ArtistName(),
			"album":  song.Album.Name,
		})
	}
	return playlist
}

// PlayingInfo returns the current song's cover URL.
func (b *localBackend) PlayingInfo() (string, bool) {
	if b.engine == nil {
		return "", false
	}
	return b.engine.Player().PlayingInfo().PicUrl, true
}

// LyricState maps the lyric service state to the wire shape (mirroring the
// handleLyrics mapping). A nil engine yields the empty structure.
func (b *localBackend) LyricState() (fragments []lyricFragment, translated map[int64]string, currentIndex int, offsetMs int64) {
	if b.engine == nil {
		return make([]lyricFragment, 0), map[int64]string{}, 0, 0
	}
	st := b.engine.LyricService().State()
	fragments = make([]lyricFragment, 0, len(st.Fragments))
	for _, f := range st.Fragments {
		fragments = append(fragments, lyricFragment{StartTimeMs: f.StartTimeMs, Content: f.Content})
	}
	translated = st.TranslatedFragments
	if translated == nil {
		translated = map[int64]string{}
	}
	return fragments, translated, st.CurrentIndex, st.OffsetMs
}

// CommandContext returns the player command-context snapshot.
func (b *localBackend) CommandContext() frontend.CommandContext {
	if b.engine == nil {
		return frontend.CommandContext{}
	}
	return b.engine.Player().CommandContext()
}

// ServerOptions configures the Server at construction.
type ServerOptions struct {
	// Auth controls whether the auth layer (token exchange + cookie + Origin
	// validation) is enabled. true = current behavior (standalone/connect);
	// false = disabled (GUI AssetServer scheme: pages load from a built-in
	// scheme where the cookie exchange is impossible).
	Auth bool
}
