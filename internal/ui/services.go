package ui

import (
	"errors"

	"github.com/juju/persistent-cookiejar"

	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

// Canonical service names (spec API Contracts table, Phase 3.1).
// menuRegistry / pageRegistry are reserved for Phase 3.2 and are NOT
// registered here yet.
const (
	ServicePlayer        = "player"
	ServiceLyricService  = "lyricService"
	ServiceTrackManager  = "trackManager"
	ServiceDesktopLyrics = "desktopLyrics"
	ServiceCoverRenderer = "coverRenderer"
	ServiceUserService   = "userService"
	ServiceLoginService  = "loginService"
	ServiceShareSvc      = "shareSvc"
	ServiceLastfm        = "lastfm"

	ServiceMenuRegistry = "menuRegistry" // reserved, Phase 3.2
	ServicePageRegistry = "pageRegistry" // reserved, Phase 3.2
)

// registeredServiceNames is the exact startup service set (no missing, no
// extras). The registration-completeness test compares it against the context.
var registeredServiceNames = []string{
	ServicePlayer,
	ServiceLyricService,
	ServiceTrackManager,
	ServiceDesktopLyrics,
	ServiceCoverRenderer,
	ServiceUserService,
	ServiceLoginService,
	ServiceShareSvc,
	ServiceLastfm,
}

// UserService carries the current login state (user 状态迁移).
type UserService struct {
	// User points at the live Netease-owned user slot; dereferencing always
	// yields the current user (nil until InitHook/LoginCallback populate it).
	User **structs.User
	// Login refreshes the user profile after a successful login.
	Login func() error
}

// LoginService owns the login-flow state. For this slice it is a placeholder
// wrapping the app-wide cookie jar (appCookieJar 归属收敛).
type LoginService struct {
	// CookieJar points at the package-level appCookieJar slot; dereferencing
	// always yields the current jar (nil until InitHook initializes it).
	CookieJar **cookiejar.Jar
}

// provideIfAbsent registers svc under name unless it is already present, so the
// scope-provided services (shareSvc/lastfm) and the startup registration can
// coexist without a duplicate-Provide panic.
func provideIfAbsent(ctx *framework.Context, name string, svc any) {
	if ctx.Service(name) == nil {
		ctx.Provide(name, svc)
	}
}

// registerServices registers the nine startup services into ctx. It is a plain
// callable so tests can run it against a fresh context without a full app boot.
func registerServices(ctx *framework.Context, n *Netease) error {
	if ctx == nil || n == nil {
		return errors.New("registerServices: nil ctx or Netease")
	}

	provideIfAbsent(ctx, ServicePlayer, n.player)
	provideIfAbsent(ctx, ServiceLyricService, n.lyricService)
	provideIfAbsent(ctx, ServiceTrackManager, n.trackManager)
	provideIfAbsent(ctx, ServiceDesktopLyrics, n.desktopLyrics)
	provideIfAbsent(ctx, ServiceCoverRenderer, n.coverRenderer)
	provideIfAbsent(ctx, ServiceUserService, &UserService{User: &n.user, Login: n.LoginCallback})
	provideIfAbsent(ctx, ServiceLoginService, &LoginService{CookieJar: &appCookieJar})
	provideIfAbsent(ctx, ServiceShareSvc, n.shareSvc)
	provideIfAbsent(ctx, ServiceLastfm, n.lastfm)
	return nil
}
