package core

import (
	"errors"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/desktop_lyrics"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/track"
	apputils "github.com/go-musicfox/go-musicfox/utils/app"
)

// Service constructor plugins (docs/plugin_ecosystem.md §3.2, P3): each plugin
// owns the construction, registration and cleanup of one engine service.
// NewEngine is a pure assembler — it registers the plugins in dependency order
// and starts the root scope; each plugin's Deps resolves its prerequisites
// (missing dependency = explicit Start failure, enforcing the construction
// order that used to be a comment in NewEngine) and its Start constructs the
// instance, records it on the engine (accessors stay intact) and registers it
// into the shared context via ProvideIfAbsent.
//
// Ownership transition: plugin Dispose holds the real per-service cleanup
// (formerly inlined in Engine.Close: player/lastfm/desktopLyrics); Engine.Close
// delegates to rootScope.Dispose which finalizes plugins in reverse
// registration order.

// newServicePlugins returns the engine service constructor plugins in
// dependency order (registration order = Start order = Deps availability).
func newServicePlugins(e *Engine, opts EngineOptions) []framework.Plugin {
	return []framework.Plugin{
		newLastfmPlugin(e),
		newTrackManagerPlugin(e),
		newLyricServicePlugin(e),
		newDesktopLyricsPlugin(e, opts.DesktopLyrics),
		newLoginServicePlugin(e),
		newUserServicePlugin(e),
		newShareSvcPlugin(e),
		newEventBusPlugin(e),
		newPlayerPlugin(e),
		newDispatcherPlugin(e),
	}
}

// lastfmPlugin constructs and owns the Last.fm client (cleanup: Client.Close).
type lastfmPlugin struct {
	framework.NoopPlugin
	e      *Engine
	build  func() *lastfm.Client
	client *lastfm.Client
}

func newLastfmPlugin(e *Engine) *lastfmPlugin {
	return &lastfmPlugin{e: e, build: lastfm.NewClient}
}

func (p *lastfmPlugin) Deps(*framework.Context) error { return nil }

func (p *lastfmPlugin) Start(ctx *framework.Context) error {
	p.client = p.build()
	p.e.lastfm = p.client
	provideIfAbsent(ctx, ServiceLastfm, p.client)
	return nil
}

func (p *lastfmPlugin) Dispose() error {
	if p.client != nil {
		p.client.Close()
	}
	return nil
}

// trackManagerPlugin constructs and owns the track manager (no cleanup).
type trackManagerPlugin struct {
	framework.NoopPlugin
	e     *Engine
	build func() *track.Manager
	svc   *track.Manager
}

func newTrackManagerPlugin(e *Engine) *trackManagerPlugin {
	return &trackManagerPlugin{e: e, build: func() *track.Manager {
		quality := configs.AppConfig.Player.SongLevel
		maxSizeMB := configs.AppConfig.Storage.Cache.Limit
		nameGen := composer.NewFileNameGenerator()
		_ = nameGen.RegisterSongTemplate(configs.AppConfig.Storage.FileNameTpl)
		_ = nameGen.RegisterLyricTemplate(configs.AppConfig.Storage.FileNameTpl)
		return track.NewManager(
			track.WithNameGenerator(nameGen),
			track.WithCacher(track.NewCacher(maxSizeMB)),
			track.WithSongQuality(quality))
	}}
}

func (p *trackManagerPlugin) Deps(*framework.Context) error { return nil }

func (p *trackManagerPlugin) Start(ctx *framework.Context) error {
	p.svc = p.build()
	p.e.trackManager = p.svc
	provideIfAbsent(ctx, ServiceTrackManager, p.svc)
	return nil
}

func (p *trackManagerPlugin) Dispose() error { return nil }

// lyricServicePlugin constructs and owns the lyric service (no cleanup). It
// depends on the track manager (the lyric Fetcher).
type lyricServicePlugin struct {
	framework.NoopPlugin
	e     *Engine
	build func(tm *track.Manager) *lyric.Service
	tm    *track.Manager
	svc   *lyric.Service
}

func newLyricServicePlugin(e *Engine) *lyricServicePlugin {
	return &lyricServicePlugin{e: e, build: func(tm *track.Manager) *lyric.Service {
		showTranslation := configs.AppConfig.Main.Lyric.ShowTranslation
		offset := time.Duration(configs.AppConfig.Main.Lyric.Offset) * time.Millisecond
		skipParseErr := configs.AppConfig.Main.Lyric.SkipParseErr
		svc := lyric.NewService(tm, showTranslation, offset, skipParseErr)
		svc.EnableYRC(true) // Enable word-by-word lyrics
		return svc
	}}
}

func (p *lyricServicePlugin) Deps(ctx *framework.Context) error {
	tm, ok := framework.ServiceOf[*track.Manager](ctx, ServiceTrackManager)
	if !ok {
		return errors.New("lyricServicePlugin: trackManager not resolved")
	}
	p.tm = tm
	return nil
}

func (p *lyricServicePlugin) Start(ctx *framework.Context) error {
	p.svc = p.build(p.tm)
	p.e.lyricService = p.svc
	provideIfAbsent(ctx, ServiceLyricService, p.svc)
	return nil
}

func (p *lyricServicePlugin) Dispose() error { return nil }

// desktopLyricsPlugin constructs and owns the desktop-lyrics controller
// (cleanup: Controller.Close). The controller is only created when the
// frontend requests it (TUI); headless mode provides a nil controller so the
// service name stays registered with the same service set.
type desktopLyricsPlugin struct {
	framework.NoopPlugin
	e       *Engine
	enabled bool
	build   func() desktop_lyrics.Controller
	ctrl    desktop_lyrics.Controller
}

func newDesktopLyricsPlugin(e *Engine, enabled bool) *desktopLyricsPlugin {
	return &desktopLyricsPlugin{
		e:       e,
		enabled: enabled,
		build: func() desktop_lyrics.Controller {
			if !enabled {
				return nil
			}
			return desktop_lyrics.NewController(configs.AppConfig.Main.Lyric.DesktopLyrics)
		},
	}
}

func (p *desktopLyricsPlugin) Deps(*framework.Context) error { return nil }

func (p *desktopLyricsPlugin) Start(ctx *framework.Context) error {
	p.ctrl = p.build()
	p.e.desktopLyrics = p.ctrl
	provideIfAbsent(ctx, ServiceDesktopLyrics, p.ctrl)
	return nil
}

func (p *desktopLyricsPlugin) Dispose() error {
	if p.ctrl != nil {
		p.ctrl.Close()
	}
	return nil
}

// loginServicePlugin registers the login service (cookie-jar owner; the jar
// slot is the package-level appCookieJar). No cleanup.
type loginServicePlugin struct {
	framework.NoopPlugin
	e   *Engine
	svc *LoginService
}

func newLoginServicePlugin(e *Engine) *loginServicePlugin {
	return &loginServicePlugin{e: e}
}

func (p *loginServicePlugin) Deps(*framework.Context) error { return nil }

func (p *loginServicePlugin) Start(ctx *framework.Context) error {
	p.svc = &LoginService{CookieJar: &appCookieJar, User: p.e.UserSlot()}
	provideIfAbsent(ctx, ServiceLoginService, p.svc)
	return nil
}

func (p *loginServicePlugin) Dispose() error { return nil }

// userServicePlugin registers the user service. It depends on loginService to
// document the cross-service ordering constraint (jar 先于 userService 回调:
// InitJar must have run before the user callback persists the jar). No
// cleanup.
type userServicePlugin struct {
	framework.NoopPlugin
	e   *Engine
	svc *UserService
}

func newUserServicePlugin(e *Engine) *userServicePlugin {
	return &userServicePlugin{e: e}
}

func (p *userServicePlugin) Deps(ctx *framework.Context) error {
	if _, ok := framework.ServiceOf[*LoginService](ctx, ServiceLoginService); !ok {
		return errors.New("userServicePlugin: loginService not resolved")
	}
	return nil
}

func (p *userServicePlugin) Start(ctx *framework.Context) error {
	p.svc = &UserService{
		User:           p.e.UserSlot(),
		Login:          p.e.LoginCallback,
		Jar:            &appCookieJar,
		loadStoredUser: loadStoredUserFromStorage,
		refreshJar:     apputils.RefreshCookieJar,
	}
	provideIfAbsent(ctx, ServiceUserService, p.svc)
	return nil
}

func (p *userServicePlugin) Dispose() error { return nil }

// shareSvcPlugin constructs and owns the share service (no cleanup).
type shareSvcPlugin struct {
	framework.NoopPlugin
	e     *Engine
	build func() *composer.ShareService
	svc   *composer.ShareService
}

func newShareSvcPlugin(e *Engine) *shareSvcPlugin {
	return &shareSvcPlugin{e: e, build: func() *composer.ShareService {
		svc := composer.NewShareService()
		_ = svc.RegisterTemplates(configs.AppConfig.Share)
		return svc
	}}
}

func (p *shareSvcPlugin) Deps(*framework.Context) error { return nil }

func (p *shareSvcPlugin) Start(ctx *framework.Context) error {
	p.svc = p.build()
	p.e.shareSvc = p.svc
	provideIfAbsent(ctx, ServiceShareSvc, p.svc)
	return nil
}

func (p *shareSvcPlugin) Dispose() error { return nil }

// playerPlugin constructs and owns the playback coordinator (cleanup:
// Player.Close). It depends on lyricService, trackManager, desktopLyrics
// (optional: nil in headless mode, the player is nil-safe), lastfm and the
// event bus (P4: the player double-writes playback events to it).
type playerPlugin struct {
	framework.NoopPlugin
	e      *Engine
	build  func(opts PlayerOptions) *Player
	opts   PlayerOptions
	player *Player
}

func newPlayerPlugin(e *Engine) *playerPlugin {
	return &playerPlugin{e: e, build: NewPlayer}
}

func (p *playerPlugin) Deps(ctx *framework.Context) error {
	ls, ok := framework.ServiceOf[*lyric.Service](ctx, ServiceLyricService)
	if !ok {
		return errors.New("playerPlugin: lyricService not resolved")
	}
	tm, ok := framework.ServiceOf[*track.Manager](ctx, ServiceTrackManager)
	if !ok {
		return errors.New("playerPlugin: trackManager not resolved")
	}
	lf, ok := framework.ServiceOf[*lastfm.Client](ctx, ServiceLastfm)
	if !ok {
		return errors.New("playerPlugin: lastfm not resolved")
	}
	eb, ok := framework.ServiceOf[*framework.EventEmitter](ctx, ServiceEventBus)
	if !ok {
		return errors.New("playerPlugin: eventBus not resolved")
	}
	// desktopLyrics is optional: headless provides a nil controller, so a
	// missing/nil resolution must not fail Deps (the player tolerates nil).
	dl, _ := framework.ServiceOf[desktop_lyrics.Controller](ctx, ServiceDesktopLyrics)
	p.opts = PlayerOptions{
		LyricService:  ls,
		TrackManager:  tm,
		DesktopLyrics: dl,
		User:          p.e.UserSlot(),
		LastfmTracker: lf.Tracker,
		EventEmitter:  eb,
	}
	return nil
}

func (p *playerPlugin) Start(ctx *framework.Context) error {
	p.player = p.build(p.opts)
	p.e.player = p.player
	provideIfAbsent(ctx, ServicePlayer, p.player)
	return nil
}

func (p *playerPlugin) Dispose() error {
	if p.player != nil {
		return p.player.Close()
	}
	return nil
}

// dispatcherPlugin registers the control dispatcher as a core service (P3
// lift: headless/webui still construct their own transport dispatchers, but
// ServiceDispatcher becomes the framework-resolvable mount point for P4+). It
// depends on the player. No cleanup.
type dispatcherPlugin struct {
	framework.NoopPlugin
	e *Engine
	d *Dispatcher
}

func newDispatcherPlugin(e *Engine) *dispatcherPlugin {
	return &dispatcherPlugin{e: e}
}

func (p *dispatcherPlugin) Deps(ctx *framework.Context) error {
	if _, ok := framework.ServiceOf[*Player](ctx, ServicePlayer); !ok {
		return errors.New("dispatcherPlugin: player not resolved")
	}
	return nil
}

func (p *dispatcherPlugin) Start(ctx *framework.Context) error {
	p.d = NewDispatcher(p.e)
	provideIfAbsent(ctx, ServiceDispatcher, p.d)
	return nil
}

func (p *dispatcherPlugin) Dispose() error { return nil }

// eventBusPlugin registers the app-wide event emitter as a service (P4: the
// player double-writes playback events and the engine startup/login flows emit
// through it). No cleanup.
type eventBusPlugin struct {
	framework.NoopPlugin
	e       *Engine
	emitter *framework.EventEmitter
}

func newEventBusPlugin(e *Engine) *eventBusPlugin {
	return &eventBusPlugin{e: e}
}

func (p *eventBusPlugin) Deps(*framework.Context) error { return nil }

func (p *eventBusPlugin) Start(ctx *framework.Context) error {
	p.emitter = framework.NewEventEmitter()
	provideIfAbsent(ctx, ServiceEventBus, p.emitter)
	return nil
}

func (p *eventBusPlugin) Dispose() error { return nil }
