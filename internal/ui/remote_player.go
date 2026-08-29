package ui

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/headless"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// Compile-time assertion: RemotePlayer satisfies the renderer read surface.
var _ playerRendererState = (*RemotePlayer)(nil)

// RemotePlayer 是 TUI-connect 的播放器数据面（D-TC-1 方案 B）：状态 =
// SubscribeClient 快照 + 事件流增量缓存；控制 = Call 转发 daemon；渲染事件经
// renderTicker/Rerender 投递——与 ui.Player 的 Observer 回调同构（core 播放
// goroutine 调 Rerender 是既有线程安全前提，事件消费 goroutine 沿用同一
// 模式，B1）。daemon 无对应命令的降级（播放/收藏等）见 TC-3。
type RemotePlayer struct {
	client  *headless.SubscribeClient
	netease *Netease

	renderTicker *tickerByRemotePlayer
	// posThrottle bounds the position event rate the TUI consumes (mirrors
	// internal/webui/events.go positionThrottle): the daemon already throttles,
	// this second layer keeps a stable tick rate for the renderer.
	posThrottle *positionThrottle

	// running gates render pokes: the consumer goroutine starts at
	// construction (the snapshot is already buffered when DialSubscribe
	// returns), but the shell's App.program is only assigned by App.Run.
	// Calling App.Rerender before that races foxful's program field, so
	// pre-run frames update the cache and skip the render poke. The connect
	// InitHook marks the shell running — it fires after the program starts
	// (foxful App.Init), so post-run events render normally.
	running atomic.Bool

	// mu guards the state cache below: written by the event consumer goroutine
	// (snapshot + event frames), read by the renderer/query accessors.
	mu sync.Mutex
	// The following fields are maintained from the snapshot (Dispatcher status
	// + trimmed playlist) and the event stream.
	ready    bool
	song     structs.Song // snapshot song trimmed mapping (Id/Name/Artist/Album)
	state    types.State
	passed   time.Duration
	volume   int
	mode     types.Mode
	playlist []structs.Song // snapshot playlist trimmed mapping
	user     *structs.User  // status.user nickname only (B8)
}

// NewRemotePlayer 构造遥控播放器并启动事件消费 goroutine。A nil client (a
// detached cache used by unit tests) skips the consumer goroutine.
func NewRemotePlayer(n *Netease, client *headless.SubscribeClient) *RemotePlayer {
	p := &RemotePlayer{
		client:      client,
		netease:     n,
		posThrottle: new(positionThrottle),
		playlist:    make([]structs.Song, 0),
	}
	p.renderTicker = newTickerByRemotePlayer(p)
	if client != nil {
		go p.consumeEvents()
	}
	return p
}

// --- playerRendererState 接口（B2）：renderer 组读取面 ---

// CurSong returns the cached current song (zero value before any snapshot).
func (p *RemotePlayer) CurSong() structs.Song {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.song
}

// CurSongIndex returns the index of the current song within the cached
// playlist, derived by id match; 0 when no song is loaded or it is absent.
func (p *RemotePlayer) CurSongIndex() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.song.Id == 0 {
		return 0
	}
	for i, s := range p.playlist {
		if s.Id == p.song.Id {
			return i
		}
	}
	return 0
}

// PassedTime returns the cached position of the current song.
func (p *RemotePlayer) PassedTime() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.passed
}

// State returns the cached playback state.
func (p *RemotePlayer) State() types.State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// Volume returns the cached volume.
func (p *RemotePlayer) Volume() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.volume
}

// Mode returns the cached play mode.
func (p *RemotePlayer) Mode() types.Mode {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.mode
}

// Playlist returns a copy of the cached playlist.
func (p *RemotePlayer) Playlist() []structs.Song {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]structs.Song, len(p.playlist))
	copy(out, p.playlist)
	return out
}

// --- 控制面（B3）：Call 转发；daemon 无对应命令的降级见 TC-3 ---

// call forwards a control command to the daemon. A failure (disconnect or
// timeout) is logged and surfaced as an in-app notification when the shell is
// attached; the daemon disconnect itself is also surfaced by consumeEvents via
// ready=false (D-TC-4). SubscribeClient.Call owns the 3s deadline.
func (p *RemotePlayer) call(cmd string, args map[string]any) {
	if _, err := p.client.Call(context.Background(), cmd, args); err != nil {
		slog.Warn("remote player: control forward failed", slog.String("cmd", cmd), slog.Any("err", err))
		if p.netease != nil {
			p.netease.Notify(model.NotificationSpec{
				Level:   model.NotificationWarning,
				Title:   "遥控失败",
				Message: err.Error(),
			})
		}
	}
}

func (p *RemotePlayer) CtrlPlay() {
	// The dispatcher "play" needs a search query; a bare play is a daemon-side
	// error surfaced as a toast. The connect shell's UX degrade (hide/remap the
	// play action) is TC-3's responsibility.
	p.call("play", nil)
}

// callResult executes a control command and parses the daemon's Dispatcher
// response data. A transport failure or a non-ok response yields an error.
// Unlike call it does not toast — callers surface their own message (e.g. the
// QR login page shows the error inline).
func (p *RemotePlayer) callResult(cmd string, args map[string]any) (map[string]any, error) {
	resp, err := p.client.Call(context.Background(), cmd, args)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, errors.New(resp.Error)
	}
	data, _ := resp.Data.(map[string]any)
	return data, nil
}

// CallQRKey requests a fresh QR-login unikey and its scan URL from the daemon
// (D-TC-7: the daemon is the sole login-state host, the TUI only renders and
// polls). The url feeds the local QR-code renderer.
func (p *RemotePlayer) CallQRKey() (uniKey, qrcodeUrl string, err error) {
	data, err := p.callResult("login_qr_key", nil)
	if err != nil {
		return "", "", err
	}
	return strOf(data["uniKey"]), strOf(data["qrcodeUrl"]), nil
}

// CallQRStatus polls the QR scan status for uniKey on the daemon (D-TC-7).
// Code semantics (mirror core/qrlogin): 800 = expired, 801 = awaiting scan,
// 802 = scanned awaiting confirm, 803 = confirmed — on 803 the daemon has
// already completed the login and broadcast EvLogin (the TUI observes the
// user-state change through the subscription, D-TC-8).
func (p *RemotePlayer) CallQRStatus(uniKey string) (code float64, err error) {
	data, err := p.callResult("login_qr_status", map[string]any{"key": uniKey})
	if err != nil {
		return 0, err
	}
	code, ok := numOf(data["code"])
	if !ok {
		return 0, errors.New("login_qr_status 响应缺少 code 字段")
	}
	return code, nil
}

// playListArgs maps a local song list to the play_list wire shape (D-TC-9):
// each song is trimmed to {id,name,artist,album} — artist joined by comma via
// ArtistName — mirroring the daemon's songsFromWire/trimmedPlaylist round-trip
// so the response can refresh the local queue cache in place.
func playListArgs(songs []structs.Song, index int, play bool) map[string]any {
	w := make([]map[string]any, 0, len(songs))
	for _, s := range songs {
		w = append(w, map[string]any{
			"id":     s.Id,
			"name":   s.Name,
			"artist": s.ArtistName(),
			"album":  s.Album.Name,
		})
	}
	return map[string]any{"songs": w, "index": index, "play": play}
}

// CallPlayList delivers a whole song list to the daemon queue (D-TC-9): the
// daemon rebuilds its playlist at index and optionally starts playback. The
// response carries the rebuilt (trimmed) playlist, which is written back into
// the local queue cache so the playlist view and next/prev navigation stay in
// sync (EvPlaylistChanged events are P2). Failures are logged and toasted like
// the other control forwards.
func (p *RemotePlayer) CallPlayList(songs []structs.Song, index int, play bool) error {
	data, err := p.callResult("play_list", playListArgs(songs, index, play))
	if err != nil {
		slog.Warn("remote player: play_list forward failed", slog.Int("songs", len(songs)), slog.Any("err", err))
		if p.netease != nil {
			p.netease.Notify(model.NotificationSpec{
				Level:   model.NotificationWarning,
				Title:   "遥控失败",
				Message: err.Error(),
			})
		}
		return err
	}
	if pl, ok := data["playlist"].([]any); ok {
		p.mu.Lock()
		p.playlist = playlistFromWire(pl)
		p.mu.Unlock()
	}
	return nil
}

func (p *RemotePlayer) CtrlPause()      { p.call("pause", nil) }
func (p *RemotePlayer) CtrlResume()     { p.call("resume", nil) }
func (p *RemotePlayer) CtrlToggle()     { p.call("toggle", nil) }
func (p *RemotePlayer) CtrlNext()       { p.call("next", nil) }
func (p *RemotePlayer) CtrlPrevious()   { p.call("prev", nil) }
func (p *RemotePlayer) CtrlStop()       { p.call("stop", nil) }
func (p *RemotePlayer) CtrlLikeNowPlaying()    { p.call("like", nil) }
func (p *RemotePlayer) CtrlDislikeNowPlaying() { p.call("dislike", nil) }

func (p *RemotePlayer) CtrlSeek(d time.Duration) { p.call("seek", remoteSeekArgs(d)) }
func (p *RemotePlayer) CtrlSetVolume(v int)      { p.call("volume", remoteVolumeArgs(v)) }
func (p *RemotePlayer) CtrlSetRepeat(m int)      { p.call("repeat", remoteRepeatArgs(m)) }
func (p *RemotePlayer) CtrlSetShuffle(on int)    { p.call("shuffle", remoteShuffleArgs(on)) }

// remoteSeekArgs maps a duration to the dispatcher "seek" arg (seconds).
func remoteSeekArgs(d time.Duration) map[string]any {
	return map[string]any{"seconds": d.Seconds()}
}

// remoteVolumeArgs maps a volume to the dispatcher "volume" arg.
func remoteVolumeArgs(v int) map[string]any {
	return map[string]any{"value": v}
}

// repeatModeName maps the repeat int (0=off, 1=one, 2=all — the core CtrlSetRepeat
// semantics, see dispatcher.cmdRepeat) to the dispatcher wire mode.
func repeatModeName(m int) string {
	switch m {
	case 1:
		return "one"
	case 2:
		return "all"
	default:
		return "off"
	}
}

// remoteRepeatArgs maps a repeat int to the dispatcher "repeat" arg.
func remoteRepeatArgs(m int) map[string]any {
	return map[string]any{"mode": repeatModeName(m)}
}

// remoteShuffleArgs maps a shuffle switch (1=on, 0=off) to the dispatcher arg.
func remoteShuffleArgs(on int) map[string]any {
	return map[string]any{"on": on != 0}
}

// --- 查询面 ---

// User returns a copy of the cached user with the UserId stripped (D-TC-8:
// login gating stays nickname-based — CheckUserInfo treats a zero UserId as
// not logged in, which is what the connect shell wants for the still-degraded
// local login-gated menus); nil before any snapshot.
func (p *RemotePlayer) User() *structs.User {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.user == nil {
		return nil
	}
	u := *p.user
	u.UserId = 0
	return &u
}

// UserID returns the cached daemon user id (0 before login). It is written by
// the status snapshot (idempotent, D-TC-8) and the EvLogin event (live
// update). Display and the P2 command surface read it here instead of User(),
// which strips it for gating.
func (p *RemotePlayer) UserID() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.user == nil {
		return 0
	}
	return p.user.UserId
}

// UserLoggedIn reports whether the daemon has a logged-in user (D-TC-8).
func (p *RemotePlayer) UserLoggedIn() bool {
	return p.UserID() != 0
}

// CommandContext maps the cached state to the UI-agnostic command context
// (B10: the MVP connect command surface stays empty; kept for P2).
func (p *RemotePlayer) CommandContext() frontend.CommandContext {
	p.mu.Lock()
	defer p.mu.Unlock()
	ctx := frontend.CommandContext{}
	if p.user != nil {
		ctx.UserID = p.user.UserId
		ctx.UserName = p.user.Nickname
	}
	ctx.Playing = p.state == types.Playing
	if p.song.Id != 0 {
		ctx.Song = &frontend.SongInfo{
			ID:     p.song.Id,
			Name:   p.song.Name,
			Artist: p.song.ArtistName(),
			Album:  p.song.Album.Name,
		}
	}
	return ctx
}

// RenderTicker returns the render ticker fed by the position event consumer.
func (p *RemotePlayer) RenderTicker() model.Ticker {
	return p.renderTicker
}

// Ready reports whether the subscription is live (snapshot received and the
// daemon connection still open). It is false before the snapshot arrives and
// after a disconnect (D-TC-4).
func (p *RemotePlayer) Ready() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.ready
}

// rerender refreshes the shell from the event consumer goroutine, mirroring
// ui.Player.OnSongChanged/OnStateChanged (B1). A detached shell (unit tests)
// or a not-yet-running App (the consumer starts before App.Run assigns the
// program) is a no-op — pre-run frames stay cache-only, which is what the
// initial render shows anyway.
func (p *RemotePlayer) rerender() {
	if p.netease == nil || !p.running.Load() {
		return
	}
	p.netease.Rerender(false)
}

// markRunning opens the render-poke path once the shell's App is running
// (called from the connect InitHook, which fires after App.Run started the
// program). Safe for concurrent use.
func (p *RemotePlayer) markRunning() {
	p.running.Store(true)
}

// --- 事件消费 goroutine（TC-2 核心） ---

// consumeEvents 逐帧读取 client.Events()：
//   - {"type":"snapshot"} → 全量刷新缓存（song/state/volume/mode/playlist/user）
//   - {"type":"event","event":"song_changed"|"state_changed"|"position"|...}
//     → 增量更新缓存（position 更新 passed + 喂 renderTicker；song/state
//     更新后 Rerender）
// 通道关闭（断线）→ ready=false → Rerender + 状态降级（D-TC-4）。
func (p *RemotePlayer) consumeEvents() {
	for frame := range p.client.Events() {
		p.handleFrame(frame)
	}
	// The Events channel closed: the daemon is gone or the client was closed.
	p.mu.Lock()
	p.ready = false
	p.mu.Unlock()
	p.rerender()
}

// handleFrame parses one raw side frame and applies it to the cache.
func (p *RemotePlayer) handleFrame(raw []byte) {
	var f struct {
		Type  string         `json:"type"`
		Event string         `json:"event"`
		Data  map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		return
	}
	switch f.Type {
	case "snapshot":
		p.applySnapshot(f.Data)
		p.rerender()
	case "event":
		p.applyEvent(f.Event, f.Data)
	}
}

// applySnapshot fully refreshes the cache from the snapshot frame's data
// (Dispatcher status fields + trimmed playlist).
func (p *RemotePlayer) applySnapshot(data map[string]any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ready = true
	p.song = songFromWire(mapOf(data["song"]))
	p.state = stateFromWire(strOf(data["state"]))
	p.passed = secondsOf(data["positionSeconds"])
	p.volume = intOf(data["volume"])
	p.mode = modeFromWire(strOf(data["mode"]))
	p.playlist = playlistFromWire(data["playlist"])
	if nick := strOf(data["user"]); nick != "" {
		if p.user == nil {
			p.user = &structs.User{}
		}
		p.user.Nickname = nick
	}
	// userId is the snapshot-side login state (D-TC-8): reconnect restores it
	// idempotently; the EvLogin event covers live in-session updates.
	if id, ok := numOf(data["userId"]); ok {
		if p.user == nil {
			p.user = &structs.User{}
		}
		p.user.UserId = int64(id)
	}
}

// applyEvent applies one event frame incrementally.
func (p *RemotePlayer) applyEvent(event string, data map[string]any) {
	switch event {
	case core.EvSongChanged:
		p.mu.Lock()
		p.song = songFromWire(data)
		// The song_changed event carries picUrl/durationSeconds beyond the
		// snapshot song shape.
		if secs, ok := numOf(data["durationSeconds"]); ok {
			p.song.Duration = time.Duration(secs * float64(time.Second))
		}
		if url := strOf(data["picUrl"]); url != "" {
			p.song.Album.PicUrl = url
		}
		p.mu.Unlock()
		p.rerender()
	case core.EvStateChanged:
		p.mu.Lock()
		p.state = stateFromWire(strOf(data["state"]))
		p.mu.Unlock()
		p.rerender()
	case core.EvPosition:
		if !p.posThrottle.shouldEmit(time.Now()) {
			return
		}
		p.mu.Lock()
		p.passed = secondsOf(data["positionSeconds"])
		p.mu.Unlock()
		// Feed the render ticker with a non-blocking push (mirrors
		// ui.Player.OnPosition): a slow renderer must never stall the event
		// consumer goroutine.
		if p.renderTicker != nil {
			select {
			case p.renderTicker.c <- time.Now():
			default:
			}
		}
	case core.EvStartupPhase:
		// The daemon already ran its startup; a late phase only refreshes the
		// shell title/state.
		p.rerender()
	case core.EvLogin:
		p.applyLogin(data)
		p.rerender()
	}
}

// applyLogin updates the cached user from the auth.login_succeeded event
// (D-TC-8: EvLogin carries the fresh nickname AND userId — the live
// in-session update; the snapshot covers reconnect).
func (p *RemotePlayer) applyLogin(data map[string]any) {
	user := mapOf(data["user"])
	nick := strOf(user["nickname"])
	userId, _ := numOf(user["userId"])
	p.mu.Lock()
	defer p.mu.Unlock()
	if nick == "" && userId == 0 {
		return
	}
	if p.user == nil {
		p.user = &structs.User{}
	}
	if nick != "" {
		p.user.Nickname = nick
	}
	if userId != 0 {
		p.user.UserId = int64(userId)
	}
}

// --- wire helpers：daemon 帧 JSON key → 缓存字段 ---

// mapOf returns v as a map[string]any (nil when not one).
func mapOf(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

// strOf returns v as a string ("" when not a string).
func strOf(v any) string {
	s, _ := v.(string)
	return s
}

// numOf converts any JSON-number-ish value to float64.
func numOf(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

// intOf converts a JSON number to int (0 when absent).
func intOf(v any) int {
	if n, ok := numOf(v); ok {
		return int(n)
	}
	return 0
}

// secondsOf converts a JSON seconds value to a time.Duration.
func secondsOf(v any) time.Duration {
	if n, ok := numOf(v); ok {
		return time.Duration(n * float64(time.Second))
	}
	return 0
}

// songFromWire maps the trimmed song record (id/name/artist/album) to a
// structs.Song, filling only the fields the wire carries (the rest stay zero).
func songFromWire(m map[string]any) structs.Song {
	song := structs.Song{}
	if m == nil {
		return song
	}
	if id, ok := numOf(m["id"]); ok {
		song.Id = int64(id)
	}
	song.Name = strOf(m["name"])
	if artist := strOf(m["artist"]); artist != "" {
		song.Artists = []structs.Artist{{Name: artist}}
	}
	if album := strOf(m["album"]); album != "" {
		song.Album = structs.Album{Name: album}
	}
	return song
}

// playlistFromWire maps the trimmed playlist array to []structs.Song.
func playlistFromWire(raw any) []structs.Song {
	items, _ := raw.([]any)
	out := make([]structs.Song, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			out = append(out, songFromWire(m))
		}
	}
	return out
}

// stateFromWire maps the dispatcher stateName wire string to types.State.
func stateFromWire(s string) types.State {
	switch s {
	case "playing":
		return types.Playing
	case "paused":
		return types.Paused
	case "stopped":
		return types.Stopped
	case "interrupted":
		return types.Interrupted
	default:
		return types.Unknown
	}
}

// modeFromWire maps the dispatcher "mode" wire value (types.Mode.Name(), a
// Chinese display name) back to types.Mode.
func modeFromWire(s string) types.Mode {
	switch s {
	case types.PmListLoop.String():
		return types.PmListLoop
	case types.PmOrdered.String():
		return types.PmOrdered
	case types.PmSingleLoop.String():
		return types.PmSingleLoop
	case types.PmListRandom.String():
		return types.PmListRandom
	case types.PmInfRandom.String():
		return types.PmInfRandom
	case types.PmIntelligent.String():
		return types.PmIntelligent
	default:
		return types.PmUnknown
	}
}

// remotePositionMinInterval bounds the position event rate (mirrors
// internal/webui/events.go positionMinInterval).
const remotePositionMinInterval = 250 * time.Millisecond

// positionThrottle drops position events closer than remotePositionMinInterval
// to the previous emit. Not safe for concurrent use — the event consumer
// goroutine is the only caller.
type positionThrottle struct {
	lastAt time.Time
}

// shouldEmit reports whether now is far enough from the last emit, updating
// the state on a hit.
func (t *positionThrottle) shouldEmit(now time.Time) bool {
	if now.Sub(t.lastAt) >= remotePositionMinInterval {
		t.lastAt = now
		return true
	}
	return false
}

// tickerByRemotePlayer mirrors tickerByPlayer (internal/ui/ticker_by_player.go)
// for the remote shell: the model.Ticker surface fed by the position event
// consumer, reporting the cached remote passed time. It exists because
// tickerByPlayer is concretely bound to *Player; the two share the same
// channel semantics (unbuffered, non-blocking push).
type tickerByRemotePlayer struct {
	c chan time.Time
	p *RemotePlayer
}

func newTickerByRemotePlayer(p *RemotePlayer) *tickerByRemotePlayer {
	return &tickerByRemotePlayer{
		c: make(chan time.Time),
		p: p,
	}
}

func (*tickerByRemotePlayer) Start() error {
	return nil
}

func (t *tickerByRemotePlayer) Ticker() <-chan time.Time {
	return t.c
}

func (t *tickerByRemotePlayer) PassedTime() time.Duration {
	return t.p.PassedTime()
}

func (tickerByRemotePlayer) Close() error {
	return nil
}
