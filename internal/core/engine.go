package core

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
	pkgerrors "github.com/pkg/errors"

	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/desktop_lyrics"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/track"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// EngineOptions carries the frontend-supplied shared state for the engine.
type EngineOptions struct {
	// User is the shared user slot owned by the shell (TUI). When nil the
	// engine allocates its own slot (headless).
	User **structs.User
	// DesktopLyrics creates the desktop-lyrics controller when true (the TUI
	// passes true). When false the engine leaves the controller nil, which the
	// player already tolerates (nil-safe updates) — headless mode has no
	// desktop-lyrics window.
	DesktopLyrics bool
}

// Engine owns the UI-free playback and startup assembly: the service context
// and scope, the cookie jar and user slots, the concrete service instances and
// the startup sequence (including plugin startup hooks). Frontends build it
// once and then wrap the player with their own observer/loading/locator.
type Engine struct {
	ctx      *framework.Context
	scope    *framework.Scope
	user     *structs.User // internal slot when opts.User == nil
	userSlot **structs.User

	player        *Player
	lyricService  *lyric.Service
	trackManager  *track.Manager
	desktopLyrics desktop_lyrics.Controller
	lastfm        *lastfm.Client
	shareSvc      *composer.ShareService
}

// NewEngine assembles the engine services, scope and player. Assembly order is
// preserved from the former TUI shell: context → lastfm → trackManager →
// lyricService → desktopLyrics → shareSvc → user slot → player → scope start →
// service registration.
func NewEngine(opts EngineOptions) *Engine {
	e := &Engine{ctx: &framework.Context{}}

	e.lastfm = lastfm.NewClient()

	quality := configs.AppConfig.Player.SongLevel
	maxSizeMB := configs.AppConfig.Storage.Cache.Limit
	nameGen := composer.NewFileNameGenerator()
	_ = nameGen.RegisterSongTemplate(configs.AppConfig.Storage.FileNameTpl)
	_ = nameGen.RegisterLyricTemplate(configs.AppConfig.Storage.FileNameTpl)
	e.trackManager = track.NewManager(
		track.WithNameGenerator(nameGen),
		track.WithCacher(track.NewCacher(maxSizeMB)),
		track.WithSongQuality(quality))

	showTranslation := configs.AppConfig.Main.Lyric.ShowTranslation
	offset := time.Duration(configs.AppConfig.Main.Lyric.Offset) * time.Millisecond
	skipParseErr := configs.AppConfig.Main.Lyric.SkipParseErr

	e.lyricService = lyric.NewService(e.trackManager, showTranslation, offset, skipParseErr)
	e.lyricService.EnableYRC(true) // Enable word-by-word lyrics

	// Initialize desktop lyrics (only when the frontend requests it: the TUI
	// shows the desktop-lyrics window, headless has no window at all).
	if opts.DesktopLyrics {
		e.desktopLyrics = desktop_lyrics.NewController(configs.AppConfig.Main.Lyric.DesktopLyrics)
	}

	e.shareSvc = composer.NewShareService()
	_ = e.shareSvc.RegisterTemplates(configs.AppConfig.Share)

	// The user slot is shared with the shell when supplied (the TUI passes
	// &n.user so the shell sees live login state); headless frontends let the
	// engine own the slot.
	if opts.User != nil {
		e.userSlot = opts.User
	} else {
		e.userSlot = &e.user
	}

	e.player = NewPlayer(PlayerOptions{
		LyricService:  e.lyricService,
		TrackManager:  e.trackManager,
		DesktopLyrics: e.desktopLyrics,
		User:          e.userSlot,
		LastfmTracker: e.lastfm.Tracker,
	})

	// Wire the framework scope: shareSvc/lastfm are registered into the
	// app-wide context via their scope plugins, then the remaining startup
	// services are registered.
	e.scope = newAppScope(e)
	if err := e.scope.Start(e.ctx); err != nil {
		// Unreachable for a fresh scope; log and continue so the shell keeps
		// a non-nil engine (a nil engine would panic at the first accessor).
		slog.Error("framework scope start failed", slogx.Error(err))
		return e
	}
	if err := registerServices(e.ctx, e); err != nil {
		// Unreachable with non-nil ctx/engine; log and continue.
		slog.Error("framework service registration failed", slogx.Error(err))
		return e
	}

	return e
}

// Ctx returns the app-wide framework context holding the service registry.
func (e *Engine) Ctx() *framework.Context {
	return e.ctx
}

// Player returns the core playback coordinator.
func (e *Engine) Player() *Player {
	return e.player
}

// User returns the current user (nil-safe deref of the user slot).
func (e *Engine) User() *structs.User {
	if e.userSlot == nil || *e.userSlot == nil {
		return nil
	}
	return *e.userSlot
}

// UserSlot returns the shared user slot itself (the engine writes login
// results through it; the TUI shares its own slot).
func (e *Engine) UserSlot() **structs.User {
	return e.userSlot
}

// TrackManager returns the track manager service.
func (e *Engine) TrackManager() *track.Manager {
	return e.trackManager
}

// LyricService returns the lyric service.
func (e *Engine) LyricService() *lyric.Service {
	return e.lyricService
}

// DesktopLyrics returns the desktop lyrics controller (nil when disabled).
func (e *Engine) DesktopLyrics() desktop_lyrics.Controller {
	return e.desktopLyrics
}

// Lastfm returns the Last.fm client.
func (e *Engine) Lastfm() *lastfm.Client {
	return e.lastfm
}

// ShareSvc returns the share service.
func (e *Engine) ShareSvc() *composer.ShareService {
	return e.shareSvc
}

// LoginCallback refreshes the user profile after a successful login: account
// fetch, slot update, like-playlist id resolution, storage persist, cookie jar
// save and like-list refresh.
func (e *Engine) LoginCallback() error {
	code, resp := (&service.UserAccountService{}).AccountInfo()
	if code != 200 {
		return pkgerrors.Errorf("accountInfo code: %f, resp: %s", code, string(resp))
	}

	user, err := structs.NewUserFromJsonForLogin(resp)
	if err != nil {
		return pkgerrors.WithMessagef(err, "parse user err, code: %f, resp: %s", code, string(resp))
	}
	*e.userSlot = &user
	e.trackManager.SetCloudUserID(user.UserId)

	// 获取我喜欢的歌单
	userPlaylists := service.UserPlaylistService{
		Uid:    strconv.FormatInt((*e.userSlot).UserId, 10),
		Limit:  strconv.Itoa(1),
		Offset: strconv.Itoa(0),
	}
	_, response := userPlaylists.UserPlaylist()
	(*e.userSlot).MyLikePlaylistID, err = jsonparser.GetInt(response, "playlist", "[0]", "id")
	if err != nil {
		slog.Warn("获取歌单ID失败", slogx.Error(err), slog.String("response", string(response)))
	}

	// 写入本地数据库
	table := storage.NewTable()
	_ = table.SetByKVModel(storage.User{}, user)

	// 持久化存储
	if err := appCookieJar.Save(); err != nil {
		slog.Error("登录成功，但持久化 cookie 到文件失败", slogx.Error(err))
	} else {
		slog.Info("登录成功，会话Cookie成功保存")
	}

	// 更新like list
	go likelist.RefreshLikeList(user.UserId)

	return nil
}

// Close shuts down the engine-owned services: player, lastfm, desktop lyrics
// and the framework scope (stop + dispose).
func (e *Engine) Close() error {
	var errs []error
	if e.player != nil {
		errs = append(errs, e.player.Close())
	}
	if e.lastfm != nil {
		e.lastfm.Close()
	}
	if e.desktopLyrics != nil {
		e.desktopLyrics.Close()
	}
	if e.scope != nil {
		errs = append(errs, e.scope.Stop())
		errs = append(errs, e.scope.Dispose())
	}
	return errors.Join(errs...)
}
