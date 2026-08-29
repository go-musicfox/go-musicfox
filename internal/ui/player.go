// Package ui provides the TUI views, menus and player coordinator for
// go-musicfox.
package ui

import (
	"errors"
	"strconv"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// PlayDirection mirrors the core playback direction; kept so ui call sites
// (operate.go) compile unchanged against the promoted core methods.
type PlayDirection = core.PlayDirection

const (
	DurationNext = core.DurationNext
	DurationPrev = core.DurationPrev
)

// playerRendererState 提高 UI 渲染所需的歌曲信息
type playerRendererState interface {
	CurSong() structs.Song
	CurSongIndex() int
	PassedTime() time.Duration
	State() types.State
	Volume() int
	Mode() types.Mode
	Playlist() []structs.Song
}

// Player 网易云音乐播放器（TUI 前端薄壳）
//
// The playback coordinator lives in internal/core; this wrapper embeds it and
// keeps the UI-facing state (render ticker, playing menu reference, shell) and
// the methods whose signatures mention foxful/bubbletea types.
type Player struct {
	*core.Player
	netease      *Netease
	svc          *menuServices
	renderTicker *tickerByPlayer
	playingMenu  Menu

	// remote is the TUI-connect data surface (D-TC-1 方案 B): when non-nil the
	// shell runs in connect mode and the shadowing methods below forward to it
	// instead of the embedded *core.Player (which is never constructed in
	// connect mode — B9). The embedded field stays nil in connect mode, so any
	// method the shadows miss would nil-dereference as a sentinel panic
	// (roadmap TC-3 risk 3).
	remote *RemotePlayer
}

// remoteVolumeStep mirrors the native playback engines' volume step for the
// remote shell's UpVolume/DownVolume (the dispatcher "volume" command sets an
// absolute value; native engines step by 5).
const remoteVolumeStep = 5

// remoteUnsupported toasts a TUI-connect degradation (D-TC-3): the daemon
// command set has no equivalent for this local-player action.
func (p *Player) remoteUnsupported(action string) {
	if p.netease == nil || p.netease.App == nil {
		return
	}
	p.netease.Notify(model.NotificationSpec{
		Level:   model.NotificationWarning,
		Title:   "遥控模式",
		Message: action + "：daemon 不支持该操作",
	})
}

// --- TUI-connect shadowing methods (S6 TC-3) ---
//
// Each method below explicitly shadows the promoted *core.Player method so
// menu/operate/hotkey call sites compile and run unchanged in connect mode:
// with p.remote != nil they forward to the RemotePlayer data surface (state
// cache + Call forwarding + render ticker); the standalone path keeps calling
// the embedded *core.Player with zero behavior change. Local-playlist-only
// actions (选歌播放/播放队列编辑/智能模式) have no daemon command and degrade
// to a toast (D-TC-3).

// CurSong returns the current song (remote cache in connect mode).
func (p *Player) CurSong() structs.Song {
	if p.remote != nil {
		return p.remote.CurSong()
	}
	return p.Player.CurSong()
}

// CurSongIndex returns the current song index.
func (p *Player) CurSongIndex() int {
	if p.remote != nil {
		return p.remote.CurSongIndex()
	}
	return p.Player.CurSongIndex()
}

// PassedTime returns the current playback position.
func (p *Player) PassedTime() time.Duration {
	if p.remote != nil {
		return p.remote.PassedTime()
	}
	return p.Player.PassedTime()
}

// State returns the current playback state.
func (p *Player) State() types.State {
	if p.remote != nil {
		return p.remote.State()
	}
	return p.Player.State()
}

// Volume returns the current volume.
func (p *Player) Volume() int {
	if p.remote != nil {
		return p.remote.Volume()
	}
	return p.Player.Volume()
}

// Mode returns the current play mode.
func (p *Player) Mode() types.Mode {
	if p.remote != nil {
		return p.remote.Mode()
	}
	return p.Player.Mode()
}

// Playlist returns the current playlist (trimmed snapshot in connect mode).
func (p *Player) Playlist() []structs.Song {
	if p.remote != nil {
		return p.remote.Playlist()
	}
	return p.Player.Playlist()
}

// User returns the current user (daemon nickname only in connect mode, B8).
func (p *Player) User() *structs.User {
	if p.remote != nil {
		return p.remote.User()
	}
	return p.Player.User()
}

// PlaylistUpdateAt returns the playlist update timestamp. Connect mode has no
// local playlist bookkeeping: a zero value hides the "更新于" subtitle.
func (p *Player) PlaylistUpdateAt() time.Time {
	if p.remote != nil {
		return time.Time{}
	}
	return p.Player.PlaylistUpdateAt()
}

// PlayingMenuKey returns the key of the menu that owns the playing playlist.
// Connect mode keeps no local record (the daemon owns the queue), so it always
// reports empty — InPlayingMenu then only matches the current-playlist menu.
func (p *Player) PlayingMenuKey() string {
	if p.remote != nil {
		return ""
	}
	return p.Player.PlayingMenuKey()
}

// CompareWithCurPlaylist reports whether the given songs match the current
// playlist (compared by id against the remote trimmed playlist in connect mode,
// mirroring core.Player's first-20-ids semantics).
func (p *Player) CompareWithCurPlaylist(playlist []structs.Song) bool {
	if p.remote != nil {
		cur := p.remote.Playlist()
		if len(playlist) != len(cur) {
			return false
		}
		for i := 0; i < 20 && i < len(playlist); i++ {
			if playlist[i].Id != cur[i].Id {
				return false
			}
		}
		return true
	}
	return p.Player.CompareWithCurPlaylist(playlist)
}

// CommandContext returns the UI-agnostic command context snapshot.
func (p *Player) CommandContext() frontend.CommandContext {
	if p.remote != nil {
		return p.remote.CommandContext()
	}
	return p.Player.CommandContext()
}

// Seek forwards to the daemon seek command.
func (p *Player) Seek(d time.Duration) {
	if p.remote != nil {
		p.remote.CtrlSeek(d)
		return
	}
	p.Player.Seek(d)
}

// Toggle forwards to the daemon toggle command.
func (p *Player) Toggle() {
	if p.remote != nil {
		p.remote.CtrlToggle()
		return
	}
	p.Player.Toggle()
}

// Pause forwards to the daemon pause command.
func (p *Player) Pause() {
	if p.remote != nil {
		p.remote.CtrlPause()
		return
	}
	p.Player.Pause()
}

// Resume forwards to the daemon resume command (a Stopped daemon with a
// current song starts playback on the daemon side).
func (p *Player) Resume() {
	if p.remote != nil {
		p.remote.CtrlResume()
		return
	}
	p.Player.Resume()
}

// Stop forwards to the daemon stop command.
func (p *Player) Stop() {
	if p.remote != nil {
		p.remote.CtrlStop()
		return
	}
	p.Player.Stop()
}

// NextSong forwards to the daemon next command.
func (p *Player) NextSong(manual bool) {
	if p.remote != nil {
		p.remote.CtrlNext()
		return
	}
	p.Player.NextSong(manual)
}

// PreviousSong forwards to the daemon previous command.
func (p *Player) PreviousSong(manual bool) {
	if p.remote != nil {
		p.remote.CtrlPrevious()
		return
	}
	p.Player.PreviousSong(manual)
}

// SwitchMode cycles the play mode on the daemon (repeat cycle mirroring
// core.Player.cycleRepeat; intelligent is skipped, D-TC-3).
func (p *Player) SwitchMode() {
	if p.remote != nil {
		switch p.remote.Mode() {
		case types.PmOrdered:
			p.remote.CtrlSetRepeat(2) // list loop
		case types.PmListLoop:
			p.remote.CtrlSetRepeat(1) // single loop
		case types.PmSingleLoop:
			p.remote.CtrlSetRepeat(0) // ordered
		default:
			p.remote.CtrlSetRepeat(2) // random/intelligent → list loop
		}
		return
	}
	p.Player.SwitchMode()
}

// SetMode maps the local mode API onto the daemon repeat/shuffle commands.
// Intelligent mode has no daemon command and is disabled in connect mode.
func (p *Player) SetMode(playMode types.Mode) {
	if p.remote != nil {
		switch playMode {
		case types.PmOrdered:
			p.remote.CtrlSetRepeat(0)
		case types.PmSingleLoop:
			p.remote.CtrlSetRepeat(1)
		case types.PmListLoop, types.PmInfRandom:
			p.remote.CtrlSetRepeat(2)
		case types.PmListRandom:
			p.remote.CtrlSetShuffle(1)
		case types.PmIntelligent:
			p.remoteUnsupported("心动/智能模式")
		}
		return
	}
	p.Player.SetMode(playMode)
}

// SetVolume forwards to the daemon volume command.
func (p *Player) SetVolume(v int) {
	if p.remote != nil {
		p.remote.CtrlSetVolume(v)
		return
	}
	p.Player.SetVolume(v)
}

// UpVolume steps the volume up on the daemon (absolute value re-read from the
// remote cache, mirroring the native engines' step).
func (p *Player) UpVolume() {
	if p.remote != nil {
		p.remote.CtrlSetVolume(min(p.remote.Volume()+remoteVolumeStep, 100))
		return
	}
	p.Player.UpVolume()
}

// DownVolume steps the volume down on the daemon.
func (p *Player) DownVolume() {
	if p.remote != nil {
		p.remote.CtrlSetVolume(max(p.remote.Volume()-remoteVolumeStep, 0))
		return
	}
	p.Player.DownVolume()
}

// PlaySong forwards a single-song play to the daemon in connect mode (D-TC-9):
// a one-song play_list (index 0, play=true) is the daemon-side equivalent of
// the local PlaySong — the daemon rebuilds its queue with just this song and
// starts playback, keeping next/prev consistent.
func (p *Player) PlaySong(song structs.Song, dir PlayDirection) {
	if p.remote != nil {
		p.remote.CallPlayList([]structs.Song{song}, 0, true)
		return
	}
	p.Player.PlaySong(song, dir)
}

// ReinitializePlaylist forwards to the daemon in connect mode (D-TC-9): the
// daemon rebuilds its queue with the given songs at index WITHOUT starting
// playback — the caller (playOrToggle) follows with StartPlay, exactly like
// the standalone path.
func (p *Player) ReinitializePlaylist(index int, songs []structs.Song) {
	if p.remote != nil {
		p.remote.CallPlayList(songs, index, false)
		return
	}
	p.Player.ReinitializePlaylist(index, songs)
}

// InitSongManager forwards to the daemon in connect mode (wraps
// ReinitializePlaylist; only reachable from the standalone automator, kept for
// shadow completeness).
func (p *Player) InitSongManager(index int, songs []structs.Song) {
	if p.remote != nil {
		p.remote.CallPlayList(songs, index, false)
		return
	}
	p.Player.InitSongManager(index, songs)
}

// StartPlay forwards as a resume in connect mode (D-TC-9): the daemon's
// resumeOrStart starts the current queue song when Stopped with nothing
// loaded — the exact post-ReinitializePlaylist flow playOrToggle relies on.
func (p *Player) StartPlay() {
	if p.remote != nil {
		p.remote.CtrlResume()
		return
	}
	p.Player.StartPlay()
}

// RemoveSong degrades in connect mode (no daemon playlist mutation command).
func (p *Player) RemoveSong(index int) (structs.Song, error) {
	if p.remote != nil {
		p.remoteUnsupported("编辑播放队列")
		return structs.Song{}, errors.New("tui-connect: playlist editing is not supported")
	}
	return p.Player.RemoveSong(index)
}

// NextPlaylistSong degrades in connect mode (only reached by the disabled
// Intelligence flow; kept for shadow completeness).
func (p *Player) NextPlaylistSong(manual bool) (structs.Song, error) {
	if p.remote != nil {
		return structs.Song{}, errors.New("tui-connect: playlist advance is not supported")
	}
	return p.Player.NextPlaylistSong(manual)
}

// MarkPlaylistUpdated is a no-op in connect mode (the daemon owns the
// playlist timestamp).
func (p *Player) MarkPlaylistUpdated() {
	if p.remote != nil {
		return
	}
	p.Player.MarkPlaylistUpdated()
}

// NewPlayer builds the TUI player wrapper around the core playback
// coordinator and wires the observer/locator/loading seams to this frontend.
func NewPlayer(n *Netease, corePlayer *core.Player) *Player {
	p := &Player{Player: corePlayer, netease: n, svc: newMenuServices(n)}
	p.renderTicker = newTickerByPlayer(p)
	corePlayer.SetObserver(p)
	corePlayer.SetLocator(p)
	corePlayer.SetLoading(&playerLoading{p: p})
	return p
}

// playerLoading adapts the core LoadingIndicator seam to the foxful model
// loading tip, mirroring the original PlaySong loading behavior.
type playerLoading struct {
	p       *Player
	loading *model.Loading
}

func (l *playerLoading) Start() {
	l.loading = model.NewLoading(l.p.netease.MustMain())
	l.loading.Start()
}

func (l *playerLoading) Complete() {
	if l.loading != nil {
		l.loading.Complete()
	}
}

// InPlayingMenu 是否处于正在播放的菜单中
func (p *Player) InPlayingMenu() bool {
	key := p.netease.MustMain().CurMenu().GetMenuKey()
	return key == p.PlayingMenuKey() || key == CurPlaylistKey
}

// SetPlayingMenu records the menu that owns the currently playing playlist.
// A non-ui.Menu resets the playing menu to nil while keeping the key. In
// connect mode the daemon owns the queue, so only the local menu reference is
// kept (no core playing-menu-key record).
func (p *Player) SetPlayingMenu(key string, menu model.Menu) {
	p.playingMenu, _ = menu.(Menu)
	if p.remote != nil {
		return
	}
	p.Player.SetPlayingMenu(key)
}

// PlayingMenu returns the currently playing menu.
func (p *Player) PlayingMenu() Menu {
	return p.playingMenu
}

// MarkPlaylistModified invalidates the playing-menu association after the
// playlist is mutated externally, avoiding stale menu data. In connect mode
// the daemon queue is never mutated by the shell, so only the local reference
// is cleared.
func (p *Player) MarkPlaylistModified() {
	p.playingMenu = nil
	if p.remote != nil {
		return
	}
	p.Player.MarkPlaylistModified()
}

// LocatePlayingSong 定位到正在播放的音乐
func (p *Player) LocatePlayingSong() {
	var (
		main        = p.netease.MustMain()
		curMenu, ok = main.CurMenu().(Menu)
	)
	if !ok {
		return
	}

	if !curMenu.IsLocatable() {
		return
	}

	menu, ok := curMenu.(SongsMenu)
	if !ok {
		return
	}
	if !p.InPlayingMenu() || !p.CompareWithCurPlaylist(menu.Songs()) {
		return
	}

	pageDelta := p.CurSongIndex()/main.PageSize() - (main.CurPage() - 1)
	if pageDelta > 0 {
		for i := 0; i < pageDelta; i++ {
			p.netease.MustMain().NextPage()
		}
	} else if pageDelta < 0 {
		for i := 0; i > pageDelta; i-- {
			p.netease.MustMain().PrePage()
		}
	}
	main.SetSelectedIndex(p.CurSongIndex())
}

// Intelligence 智能/心动模式
func (p *Player) Intelligence(appendMode bool) model.Page {
	if p.remote != nil {
		// D-TC-3: 智能/心动模式 is disabled in connect mode (it builds a local
		// playlist and starts playback — both daemon-unreachable).
		p.remoteUnsupported("心动/智能模式")
		return nil
	}
	var (
		main    = p.netease.MustMain()
		curMenu = main.CurMenu()
	)
	// The concrete *PlaylistDetailMenu now lives in the playlist plugin; reach
	// the open playlist's songs and ID through the shared interface instead.
	playlist, ok := curMenu.(PlaylistDetailGetter)
	if !ok {
		return nil
	}
	playlistSongs := playlist.Songs()

	selectedIndex := curMenu.RealDataIndex(main.SelectedIndex())
	if selectedIndex >= len(playlistSongs) {
		return nil
	}

	if _struct.CheckUserInfo(p.svc.User()) == _struct.NeedLogin {
		page, _ := p.svc.ToLoginPage(nil)
		return page
	}

	// 获取智能推荐歌曲
	intelligenceService := service.PlaymodeIntelligenceListService{
		SongId:       strconv.FormatInt(playlistSongs[selectedIndex].Id, 10),
		PlaylistId:   strconv.FormatInt(playlist.PlaylistID(), 10),
		StartMusicId: strconv.FormatInt(playlistSongs[selectedIndex].Id, 10),
	}
	code, response := intelligenceService.PlaymodeIntelligenceList()
	codeType := _struct.CheckCode(code)
	if codeType == _struct.NeedLogin {
		page, _ := p.svc.ToLoginPage(func() model.Page {
			p.Intelligence(appendMode)
			return nil
		})
		return page
	} else if codeType != _struct.Success {
		return nil
	}
	songs := _struct.GetIntelligenceSongs(response)

	var song structs.Song
	if appendMode {
		p.ReinitializePlaylist(p.CurSongIndex(), append(p.Playlist(), songs...))
		p.MarkPlaylistUpdated()
		var err error
		song, err = p.NextPlaylistSong(true)
		if err != nil {
			return nil
		}
	} else {
		p.SetMode(types.PmIntelligent)
		p.Player.SetPlayingMenu("Intelligent")
		p.ReinitializePlaylist(0, append([]structs.Song{playlistSongs[selectedIndex]}, songs...))
		p.MarkPlaylistUpdated()
		song = p.Playlist()[0]
	}

	p.PlaySong(song, DurationNext)
	return nil
}

// RenderTicker returns the render ticker fed by the core position observer
// (the remote player's ticker in connect mode).
func (p *Player) RenderTicker() model.Ticker {
	if p.remote != nil {
		return p.remote.RenderTicker()
	}
	return p.renderTicker
}

// --- core.Observer implementation ---

// OnSongChanged reports the song that started playing: rerender the shell.
func (p *Player) OnSongChanged(_ structs.Song) {
	p.netease.Rerender(false)
}

// OnStateChanged reports a playback state transition: rerender the shell.
func (p *Player) OnStateChanged(_ types.State) {
	p.netease.Rerender(false)
}

// OnPosition feeds the per-tick render ticker, mirroring the original time
// listener's renderTicker push.
func (p *Player) OnPosition(_ time.Duration) {
	if p.renderTicker != nil {
		select {
		case p.renderTicker.c <- time.Now():
		default:
		}
	}
}

// RequestLogin forwards login gating to the shell login page. In connect mode
// login is owned by the daemon — toast and skip (B8).
func (p *Player) RequestLogin(afterLogin func()) {
	if p.remote != nil {
		p.remoteUnsupported("登录")
		return
	}
	_, _ = p.svc.ToLoginPage(func() model.Page {
		if afterLogin != nil {
			afterLogin()
		}
		return nil
	})
}

// OnPlaylistExhausted restores the original NextSong/PreviousSong bottom-out /
// top-out navigation: dual-column moves and playing-menu hooks.
func (p *Player) OnPlaylistExhausted(dir core.PlayDirection) {
	main := p.netease.MustMain()
	index := p.CurSongIndex()
	if dir == core.DurationNext {
		if p.InPlayingMenu() {
			if main.IsDualColumn() && index%2 == 0 {
				p.netease.MustMain().MoveRight()
			} else {
				p.netease.MustMain().MoveDown()
			}
		} else if p.playingMenu != nil {
			if bottomHook := p.playingMenu.BottomOutHook(); bottomHook != nil {
				bottomHook(main)
			}
		}
		return
	}

	if p.InPlayingMenu() {
		if main.IsDualColumn() && index%2 == 0 {
			p.netease.MustMain().MoveUp()
		} else {
			p.netease.MustMain().MoveLeft()
		}
	} else if p.playingMenu != nil {
		if topHook := p.playingMenu.TopOutHook(); topHook != nil {
			topHook(main)
		}
	}
}

// OnRerender handles CtrlRerender.
func (p *Player) OnRerender() {
	p.netease.Rerender(false)
}

// OnStartupPhase forwards startup phases to the shell for title refresh.
func (p *Player) OnStartupPhase(phase core.StartupPhase) {
	switch phase {
	case core.StartupPhaseUserRestored:
		p.netease.MustMain().RefreshMenuTitle()
		p.netease.Rerender(false)
	case core.StartupPhasePlaylistLoaded:
		p.netease.Rerender(false)
	case core.StartupPhaseBeforeAutoplay:
		p.netease.Rerender(false)
	}
}
