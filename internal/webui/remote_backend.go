package webui

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/headless"
)

// Compile-time assertion that remoteBackend satisfies Backend.
var _ Backend = (*remoteBackend)(nil)

// remoteBackend is the Backend implementation for connect mode: the data
// source is a local headless daemon (musicfox --headless) reached through a
// SubscribeClient. The control plane goes through SubscribeClient.Call; the
// event plane comes over the subscription connection (wire names mapped to
// WebUI frame names); status/playlist come from the client's snapshot cache.
type remoteBackend struct {
	client *headless.SubscribeClient

	// unsub is the teardown of the last SubscribeEvents call (stops the
	// consume loop). The Server holds the returned func and calls it on Close.
	unsub func()
	// posThrottle keeps the WebUI-side 4Hz position rate on top of the
	// daemon-side throttle (same value as the standalone emitter).
	posThrottle *positionThrottle
}

// newRemoteBackend wraps a live subscription client. The event consume loop is
// started by SubscribeEvents (the Server calls it at construction).
func newRemoteBackend(client *headless.SubscribeClient) *remoteBackend {
	return &remoteBackend{
		client:      client,
		posThrottle: new(positionThrottle),
	}
}

// Ready reports whether the daemon connection is still alive (not closed).
func (b *remoteBackend) Ready() bool {
	return !b.client.Closed()
}

// Dispatch forwards a control command to the daemon over the subscription
// connection. "quit" is rejected here: connect mode never forwards the
// transport-layer shutdown to the daemon (D-S5-2) — the WebUI server already
// intercepts WS "quit" before reaching the backend.
func (b *remoteBackend) Dispatch(ctx context.Context, cmd string, args map[string]any) (any, error) {
	if cmd == "quit" {
		return nil, errors.New("quit 不转发到 daemon（connect 模式）")
	}
	resp, err := b.client.Call(ctx, cmd, args)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, errors.New(resp.Error)
	}
	return resp.Data, nil
}

// SubscribeEvents starts the consume loop that forwards daemon event frames to
// handler. handler's name is the WebUI frame name (song_changed etc.) and its
// payload is the serialized event frame ({"type":"event",...}), so the Server
// can broadcast it verbatim. The returned func stops the consume loop.
func (b *remoteBackend) SubscribeEvents(handler func(name string, payload []byte)) func() {
	stop := make(chan struct{})
	b.unsub = func() { close(stop) }
	go b.consumeEvents(handler, stop)
	return b.unsub
}

// consumeEvents reads side frames from the subscription connection until it is
// stopped or the connection dies, forwarding mapped event frames to handler.
func (b *remoteBackend) consumeEvents(handler func(name string, payload []byte), stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case frame, ok := <-b.client.Events():
			if !ok {
				// Connection closed (Close or a daemon disconnect): the event
				// stream is gone; auxiliary endpoints degrade via Ready().
				return
			}
			b.forwardFrame(handler, frame)
		}
	}
}

// forwardFrame maps one daemon side-frame to a WebUI event frame. The daemon
// snapshot frame is skipped: the WebUI rebuilds the WS snapshot independently
// (Dispatch status + Playlist), so the daemon's snapshot is only cached by the
// client, never re-broadcast.
func (b *remoteBackend) forwardFrame(handler func(name string, payload []byte), frame []byte) {
	var f struct {
		Type  string         `json:"type"`
		Event string         `json:"event"`
		Data  map[string]any `json:"data"`
	}
	if err := json.Unmarshal(frame, &f); err != nil {
		return
	}
	if f.Type != "event" {
		return
	}
	name, ok := eventWireToFrame[f.Event]
	if !ok {
		return // not consumed by the WebUI frontend
	}
	if name == "position" && !b.posThrottle.shouldEmit(time.Now()) {
		return
	}
	if payload := eventFrame(name, f.Data); payload != nil {
		handler(name, payload)
	}
}

// Playlist returns the trimmed playlist from the daemon snapshot cache. Before
// the snapshot frame arrives the list is empty (the frontend is idempotent and
// must not crash on a mid-handshake connect).
func (b *remoteBackend) Playlist() []map[string]any {
	playlist := make([]map[string]any, 0)
	snap := b.client.Snapshot()
	if snap == nil {
		return playlist
	}
	raw, _ := snap["playlist"].([]any)
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok {
			playlist = append(playlist, m)
		}
	}
	return playlist
}

// PlayingInfo is always unavailable in connect mode: the daemon status
// snapshot carries no PicUrl (D-S5-2).
func (b *remoteBackend) PlayingInfo() (string, bool) {
	return "", false
}

// LyricState returns the empty lyric structure: the daemon exposes no lyric
// service surface on the control channel (D-S5-2).
func (b *remoteBackend) LyricState() (fragments []lyricFragment, translated map[int64]string, currentIndex int, offsetMs int64) {
	return make([]lyricFragment, 0), map[int64]string{}, 0, 0
}

// CommandContext maps the daemon status snapshot fields onto the command
// context. The snapshot carries the user nickname but no user id, so UserID
// stays 0.
func (b *remoteBackend) CommandContext() frontend.CommandContext {
	ctx := frontend.CommandContext{}
	snap := b.client.Snapshot()
	if snap == nil {
		return ctx
	}
	if user, ok := snap["user"].(string); ok {
		ctx.UserName = user
	}
	if playing, ok := snap["playing"].(bool); ok {
		ctx.Playing = playing
	}
	if song, ok := snap["song"].(map[string]any); ok {
		info := &frontend.SongInfo{}
		if id, ok := song["id"].(float64); ok {
			info.ID = int64(id)
		}
		if name, ok := song["name"].(string); ok {
			info.Name = name
		}
		if artist, ok := song["artist"].(string); ok {
			info.Artist = artist
		}
		if album, ok := song["album"].(string); ok {
			info.Album = album
		}
		ctx.Song = info
	}
	return ctx
}
