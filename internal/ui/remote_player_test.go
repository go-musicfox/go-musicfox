package ui

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/headless"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	apputils "github.com/go-musicfox/go-musicfox/utils/app"
)

// newRemotePlayerForTest builds a detached RemotePlayer (no consumer goroutine,
// no shell) for direct frame-processing assertions.
func newRemotePlayerForTest() *RemotePlayer {
	p := &RemotePlayer{
		posThrottle: new(positionThrottle),
	}
	p.renderTicker = newTickerByRemotePlayer(p)
	return p
}

// TestRemotePlayerSnapshotMapping feeds a snapshot frame and asserts the cache
// mapping: song (id/name/artist/album), state, passed, volume, mode, playlist
// and user all come from the Dispatcher-status + trimmed-playlist JSON keys.
func TestRemotePlayerSnapshotMapping(t *testing.T) {
	p := newRemotePlayerForTest()
	p.handleFrame([]byte(`{
		"type":"snapshot",
		"data":{
			"playing": true,
			"state": "playing",
			"song": {"id": 42, "name": "测试歌曲", "artist": "歌手A", "album": "专辑X"},
			"positionSeconds": 12.5,
			"durationSeconds": 180,
			"volume": 66,
			"mode": "列表循环",
			"playlistLen": 2,
			"user": "tester",
			"playlist": [
				{"id": 42, "name": "测试歌曲", "artist": "歌手A", "album": "专辑X"},
				{"id": 7, "name": "第二首", "artist": "歌手B", "album": "专辑Y"}
			]
		}
	}`))

	if !p.Ready() {
		t.Fatal("Ready() = false after snapshot, want true")
	}
	song := p.CurSong()
	if song.Id != 42 || song.Name != "测试歌曲" {
		t.Fatalf("song = %+v, want id 42 / 测试歌曲", song)
	}
	if len(song.Artists) != 1 || song.Artists[0].Name != "歌手A" {
		t.Fatalf("song.Artists = %+v, want [歌手A]", song.Artists)
	}
	if song.Album.Name != "专辑X" {
		t.Fatalf("song.Album.Name = %q, want 专辑X", song.Album.Name)
	}
	if got := p.State(); got != types.Playing {
		t.Fatalf("State() = %v, want Playing", got)
	}
	if got := p.PassedTime(); got != 12500*time.Millisecond {
		t.Fatalf("PassedTime() = %v, want 12.5s", got)
	}
	if got := p.Volume(); got != 66 {
		t.Fatalf("Volume() = %d, want 66", got)
	}
	if got := p.Mode(); got != types.PmListLoop {
		t.Fatalf("Mode() = %v, want PmListLoop", got)
	}
	if got := p.CurSongIndex(); got != 0 {
		t.Fatalf("CurSongIndex() = %d, want 0 (first playlist entry)", got)
	}
	if got := len(p.Playlist()); got != 2 {
		t.Fatalf("Playlist() len = %d, want 2", got)
	}
	u := p.User()
	if u == nil || u.Nickname != "tester" {
		t.Fatalf("User() = %+v, want nickname tester", u)
	}
}

// TestRemotePlayerEventIncrements feeds event frames and asserts the
// incremental cache updates: song_changed replaces the song (incl. the event's
// picUrl/durationSeconds), state_changed updates the state, and position
// advances passed and feeds the render ticker (throttled to 250ms).
func TestRemotePlayerEventIncrements(t *testing.T) {
	p := newRemotePlayerForTest()

	// song_changed carries the full event shape (incl. picUrl/durationSeconds).
	p.handleFrame([]byte(`{"type":"event","event":"player.song_changed","data":{
		"id": 42, "name": "新歌", "artist": "歌手A", "album": "专辑X",
		"picUrl": "https://p1.music.126.net/x.jpg", "durationSeconds": 210
	}}`))
	song := p.CurSong()
	if song.Id != 42 || song.Name != "新歌" {
		t.Fatalf("song = %+v after song_changed", song)
	}
	if song.Duration != 210*time.Second {
		t.Fatalf("song.Duration = %v, want 210s", song.Duration)
	}
	if song.Album.PicUrl != "https://p1.music.126.net/x.jpg" {
		t.Fatalf("song.Album.PicUrl = %q", song.Album.PicUrl)
	}

	// state_changed updates the state.
	p.handleFrame([]byte(`{"type":"event","event":"player.state_changed","data":{"state":"paused"}}`))
	if got := p.State(); got != types.Paused {
		t.Fatalf("State() = %v after state_changed, want Paused", got)
	}

	// position (first after construction passes the throttle) updates passed
	// and feeds the render ticker. The MAIN goroutine parks on the ticker
	// channel while a helper fires the event, so the non-blocking push (mirror
	// of ui.Player.OnPosition) is deterministically observed.
	go func() {
		time.Sleep(20 * time.Millisecond)
		p.handleFrame([]byte(`{"type":"event","event":"player.position","data":{"positionSeconds": 30.0}}`))
	}()
	select {
	case <-p.renderTicker.Ticker():
	case <-time.After(2 * time.Second):
		t.Fatal("render ticker was not fed by the position event")
	}
	if got := p.PassedTime(); got != 30*time.Second {
		t.Fatalf("PassedTime() = %v, want 30s", got)
	}

	// A second position within the 250ms throttle window is dropped (the
	// window is forced to guarantee determinism regardless of machine speed).
	p.posThrottle.lastAt = time.Now().Add(-100 * time.Millisecond)
	p.handleFrame([]byte(`{"type":"event","event":"player.position","data":{"positionSeconds": 31.0}}`))
	if got := p.PassedTime(); got != 30*time.Second {
		t.Fatalf("PassedTime() = %v, want unchanged 30s (throttled)", got)
	}
}

// TestRemotePlayerControlArgs locks the Call arg mapping: seek seconds,
// repeat off/one/all and the shuffle/volume switches.
func TestRemotePlayerControlArgs(t *testing.T) {
	if got := remoteSeekArgs(30 * time.Second); got["seconds"] != 30.0 {
		t.Fatalf("remoteSeekArgs(30s) seconds = %v, want 30", got["seconds"])
	}
	for m, want := range map[int]string{0: "off", 1: "one", 2: "all", 5: "off"} {
		if got := remoteRepeatArgs(m); got["mode"] != want {
			t.Fatalf("remoteRepeatArgs(%d) mode = %v, want %q", m, got["mode"], want)
		}
	}
	if got := remoteShuffleArgs(1); got["on"] != true {
		t.Fatalf("remoteShuffleArgs(1) on = %v, want true", got["on"])
	}
	if got := remoteShuffleArgs(0); got["on"] != false {
		t.Fatalf("remoteShuffleArgs(0) on = %v, want false", got["on"])
	}
	if got := remoteVolumeArgs(66); got["value"] != 66 {
		t.Fatalf("remoteVolumeArgs(66) value = %v, want 66", got["value"])
	}
}

// --- daemon-backed tests (temp unix socket + DialSubscribe) ---

var (
	remoteEnvOnce sync.Once
	// remoteTestRoot is a process-lifetime short temp root: macOS
	// sockaddr_un.sun_path is only 104 bytes, so the daemon socket path must
	// stay short (the headless tests use the same trick). It is kept alive for
	// the whole test process because the shared engine's beep background
	// goroutine reads app paths lazily; the socket file itself is removed by
	// server.Close.
	remoteTestRoot string
	remoteEngine   *core.Engine
)

// setupRemoteEnv performs the ONE-TIME process-level setup shared by the
// daemon tests (mirrors the headless/webui TestMain fixtures): it redirects
// the app paths to a temp root, stubs config/storage and builds the shared
// engine. Everything is bundled in a single Once so SetupPureRoot completes
// BEFORE the engine — and its beep background goroutine, which lazily reads
// app.RuntimeDir — is created (a later mutation would race that goroutine).
// Config/storage are intentionally left set for the process (existing ui
// tests save/restore configs.AppConfig themselves; none touches storage).
func setupRemoteEnv() {
	remoteEnvOnce.Do(func() {
		remoteTestRoot = filepath.Join(os.TempDir(), fmt.Sprintf("mfox-ui-%d", os.Getpid()))
		_ = os.MkdirAll(remoteTestRoot, 0755)
		apputils.SetupPureRoot(remoteTestRoot)

		configs.AppConfig = &configs.Config{Player: configs.PlayerConfig{Engine: types.BeepPlayer}}
		storage.DBManager = new(storage.LocalDBManager)

		remoteEngine = core.NewEngine(core.EngineOptions{})
	})
}

// startRemoteDaemon starts a headless control daemon bound to the
// DialAddr-resolved unix socket (scoped to the temp root via app.SetupPureRoot)
// so headless.DialSubscribe reaches it through the exported API — the same
// trick the connect shell uses. Windows keeps its TCP port-file path as a
// manual smoke item (roadmap S5 §2.4).
func startRemoteDaemon(t *testing.T) *headless.Server {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("daemon tests use a unix socket; the Windows TCP path is covered by the manual smoke checklist")
	}
	setupRemoteEnv()

	server := headless.NewServerWithAddr(remoteEngine, "unix", headless.ListenAddr())
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = server.Serve(ctx) }()
	t.Cleanup(server.Close)
	waitRemoteDaemon(t)
	return server
}

// waitRemoteDaemon polls until the daemon accepts connections at the address
// headless.DialAddr() resolves to.
func waitRemoteDaemon(t *testing.T) {
	t.Helper()
	network, addr := headless.DialAddr()
	if network == "" || addr == "" {
		t.Fatal("headless.DialAddr() empty; daemon address unavailable")
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		conn, err := net.Dial(network, addr)
		if err == nil {
			_ = conn.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("daemon not ready at %s: %v", addr, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// dialRemote dials the daemon subscription the way RunConnect does.
func dialRemote(t *testing.T) *headless.SubscribeClient {
	t.Helper()
	client, err := headless.DialSubscribe(remoteEventWireNames())
	if err != nil {
		t.Fatalf("DialSubscribe error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// waitRemoteReady polls until the RemotePlayer has processed the snapshot.
func waitRemoteReady(t *testing.T, p *RemotePlayer) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for !p.Ready() {
		if time.Now().After(deadline) {
			t.Fatal("RemotePlayer not ready after snapshot")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestRemotePlayerCallForward verifies the control plane reaches the daemon:
// CtrlSetVolume forwards {"value": 66} over the subscription connection and
// the daemon engine's volume reflects it.
func TestRemotePlayerCallForward(t *testing.T) {
	startRemoteDaemon(t)
	client := dialRemote(t)
	p := NewRemotePlayer(nil, client)

	p.CtrlSetVolume(66)

	deadline := time.Now().Add(3 * time.Second)
	for {
		resp, err := client.Call(context.Background(), "volume", nil)
		if err == nil && resp.Ok {
			if v, _ := resp.Data.(map[string]any)["volume"].(float64); v == 66 {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("daemon volume not updated after CtrlSetVolume; last resp = %+v err = %v", resp, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestRemotePlayerDisconnect verifies the disconnect semantics (D-TC-4):
// closing the client closes the Events channel, so consumeEvents marks the
// shell not ready.
func TestRemotePlayerDisconnect(t *testing.T) {
	startRemoteDaemon(t)
	client := dialRemote(t)
	p := NewRemotePlayer(nil, client)
	waitRemoteReady(t, p)

	// Closing the client closes the Events channel → ready=false.
	_ = client.Close()

	deadline := time.Now().Add(3 * time.Second)
	for p.Ready() {
		if time.Now().After(deadline) {
			t.Fatal("Ready() still true after client.Close")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestRemotePlayerUserIDFromSnapshot locks the D-TC-8 snapshot path: the
// status snapshot's userId field restores the login state idempotently
// (reconnect), while User() keeps stripping the id for the local gating
// semantics.
func TestRemotePlayerUserIDFromSnapshot(t *testing.T) {
	p := newRemotePlayerForTest()
	p.handleFrame([]byte(`{
		"type":"snapshot",
		"data":{
			"playing": false,
			"state": "stopped",
			"song": {"id": 0, "name": "", "artist": "", "album": ""},
			"positionSeconds": 0,
			"durationSeconds": 0,
			"volume": 50,
			"mode": "列表循环",
			"playlistLen": 0,
			"user": "tester",
			"userId": 123,
			"playlist": []
		}
	}`))
	if got := p.UserID(); got != 123 {
		t.Fatalf("UserID() = %d, want 123", got)
	}
	if !p.UserLoggedIn() {
		t.Fatal("UserLoggedIn() = false, want true")
	}
	u := p.User()
	if u == nil || u.Nickname != "tester" {
		t.Fatalf("User() = %+v, want nickname tester", u)
	}
	if u.UserId != 0 {
		t.Fatalf("User().UserId = %d, want 0 (stripped for local gating, D-TC-8)", u.UserId)
	}
}

// TestRemotePlayerUserIDFromLoginEvent locks the D-TC-8 event path: the
// auth.login_succeeded event carries the live user update (in-session login
// completion inside the daemon), and an empty EvLogin frame is a no-op.
func TestRemotePlayerUserIDFromLoginEvent(t *testing.T) {
	p := newRemotePlayerForTest()
	p.handleFrame([]byte(`{"type":"event","event":"auth.login_succeeded","data":{"user":{"userId":456,"nickname":"fox"}}}`))
	if got := p.UserID(); got != 456 {
		t.Fatalf("UserID() = %d, want 456", got)
	}
	if !p.UserLoggedIn() {
		t.Fatal("UserLoggedIn() = false, want true")
	}
	u := p.User()
	if u == nil || u.Nickname != "fox" || u.UserId != 0 {
		t.Fatalf("User() = %+v, want nickname fox with stripped UserId", u)
	}

	// An EvLogin without user data must not clear the cached state.
	p.handleFrame([]byte(`{"type":"event","event":"auth.login_succeeded","data":{}}`))
	if got := p.UserID(); got != 456 {
		t.Fatalf("UserID() = %d after empty EvLogin, want unchanged 456", got)
	}
}

// TestRemotePlayerPlayListArgs locks the play_list wire shape (D-TC-9): songs
// are trimmed to {id,name,artist,album} with artist joined by comma, matching
// the daemon's songsFromWire/trimmedPlaylist round-trip so the response can
// refresh the local queue cache in place.
func TestRemotePlayerPlayListArgs(t *testing.T) {
	songs := []structs.Song{
		{Id: 1, Name: "A", Artists: []structs.Artist{{Name: "X"}, {Name: "Y"}}, Album: structs.Album{Name: "AL"}},
		{Id: 2, Name: "B", Artists: []structs.Artist{{Name: "Z"}}, Album: structs.Album{Name: "BL"}},
	}
	args := playListArgs(songs, 1, false)
	if args["index"] != 1 || args["play"] != false {
		t.Fatalf("args index/play = %v/%v, want 1/false", args["index"], args["play"])
	}
	w, ok := args["songs"].([]map[string]any)
	if !ok {
		t.Fatalf("args songs = %T, want []map[string]any", args["songs"])
	}
	if len(w) != 2 {
		t.Fatalf("songs len = %d, want 2", len(w))
	}
	if w[0]["id"] != int64(1) || w[0]["name"] != "A" || w[0]["artist"] != "X,Y" || w[0]["album"] != "AL" {
		t.Fatalf("song0 = %v, want {id 1, A, X,Y, AL}", w[0])
	}
	if w[1]["artist"] != "Z" {
		t.Fatalf("song1 artist = %v, want Z", w[1]["artist"])
	}
}
