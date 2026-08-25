// Package ui provides the TUI views, menus and player coordinator for
// go-musicfox.
package ui

import (
	"strconv"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/core"
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
// A non-ui.Menu resets the playing menu to nil while keeping the key.
func (p *Player) SetPlayingMenu(key string, menu model.Menu) {
	p.Player.SetPlayingMenu(key)
	p.playingMenu, _ = menu.(Menu)
}

// PlayingMenu returns the currently playing menu.
func (p *Player) PlayingMenu() Menu {
	return p.playingMenu
}

// MarkPlaylistModified invalidates the playing-menu association after the
// playlist is mutated externally, avoiding stale menu data.
func (p *Player) MarkPlaylistModified() {
	p.playingMenu = nil
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

// RenderTicker returns the render ticker fed by the core position observer.
func (p *Player) RenderTicker() model.Ticker {
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

// RequestLogin forwards login gating to the shell login page.
func (p *Player) RequestLogin(afterLogin func()) {
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
