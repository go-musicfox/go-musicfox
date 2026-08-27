package core

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
	pkgerrors "github.com/pkg/errors"

	"github.com/go-musicfox/go-musicfox/internal/composer"
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

// NewEngine assembles the engine as a pure plugin assembler: it creates the
// service context and the root scope, registers the service constructor
// plugins in dependency order and starts the scope (each plugin's Deps
// resolves its prerequisites, its Start constructs + registers the instance
// and records it on the engine, keeping the accessors intact). Assembly order
// is preserved from the former TUI shell (lastfm → trackManager → lyricService
// → desktopLyrics → loginService → userService → shareSvc → eventBus → player
// → dispatcher), now enforced by the plugins' Deps instead of a comment; the
// event bus moved ahead of the player so the player can resolve it in Deps.
func NewEngine(opts EngineOptions) *Engine {
	e := &Engine{ctx: &framework.Context{}}

	// The user slot is shared with the shell when supplied (the TUI passes
	// &n.user so the shell sees live login state); headless frontends let the
	// engine own the slot. It must exist before the service plugins start:
	// userService/player capture the slot (double pointer).
	if opts.User != nil {
		e.userSlot = opts.User
	} else {
		e.userSlot = &e.user
	}

	// Pure assembly: root scope + service constructor plugins, then start.
	e.scope = framework.NewScope()
	for _, p := range newServicePlugins(e, opts) {
		if err := e.scope.AddWithEnabled(p, true); err != nil {
			// Unreachable for a fresh scope; log and continue so the shell
			// keeps a non-nil engine (a nil engine would panic at the first
			// accessor).
			slog.Error("framework service plugin registration failed", slogx.Error(err))
			return e
		}
	}
	if err := e.scope.Start(e.ctx); err != nil {
		// Unreachable for the static plugin set with correct ordering; log and
		// continue so the shell keeps a non-nil engine.
		slog.Error("framework scope start failed", slogx.Error(err))
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

	// P4: 登录成功事件（二维码登录经 CompleteQRLogin → LoginCallback 同路径）
	e.emit(EvLogin, loginEventPayload(e.User()))

	return nil
}

// Close shuts down the engine-owned services. The root scope's Dispose runs
// every service constructor plugin's Dispose in reverse registration order
// (eventBus → dispatcher → player → shareSvc → userService → loginService →
// desktopLyrics → lyricService → trackManager → lastfm), which owns the real
// per-service cleanup (player/lastfm/desktopLyrics) formerly inlined here.
func (e *Engine) Close() error {
	var errs []error
	if e.scope != nil {
		errs = append(errs, e.scope.Stop())
		errs = append(errs, e.scope.Dispose())
	}
	return errors.Join(errs...)
}
