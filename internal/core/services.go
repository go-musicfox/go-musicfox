package core

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	neteaseutil "github.com/go-musicfox/netease-music/util"
	cookiejar "github.com/juju/persistent-cookiejar"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	apputils "github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// Core service names (engine-owned services). The TUI-only services
// (coverRenderer / menuRegistry / pageRegistry) stay registered by the ui
// layer into the same context.
const (
	ServicePlayer        = "player"
	ServiceLyricService  = "lyricService"
	ServiceTrackManager  = "trackManager"
	ServiceDesktopLyrics = "desktopLyrics"
	ServiceUserService   = "userService"
	ServiceLoginService  = "loginService"
	ServiceShareSvc      = "shareSvc"
	ServiceLastfm        = "lastfm"
	// ServiceEventBus is the app-wide framework event emitter (P4 mount point;
	// this ticket only registers the service).
	ServiceEventBus = "eventBus"
	// ServiceDispatcher is the core control dispatcher (lifted from the
	// headless/webui transport-created instances as a framework-resolvable
	// mount point; the transports still construct their own dispatchers).
	ServiceDispatcher = "dispatcher"
)

// appCookieJar is the app-wide persistent cookie jar slot. LoginService.InitJar
// assigns it at startup; the TUI login flows may replace it after a successful
// token refresh so the engine-side LoginCallback persists the refreshed jar.
var appCookieJar *cookiejar.Jar

// AppCookieJar returns the app-wide cookie jar slot (nil until InitJar assigns
// it or a TUI login flow replaces it).
func AppCookieJar() *cookiejar.Jar {
	return appCookieJar
}

// SetAppCookieJar replaces the app-wide cookie jar slot. The TUI login flows
// call it after a successful token refresh so LoginCallback persists the
// refreshed jar.
func SetAppCookieJar(jar *cookiejar.Jar) {
	appCookieJar = jar
}

// registeredServiceNames is the exact engine startup service set (no missing,
// no extras). The registration-completeness test compares it against the
// context populated by the service constructor plugins after the root scope
// starts.
var registeredServiceNames = []string{
	ServicePlayer,
	ServiceLyricService,
	ServiceTrackManager,
	ServiceDesktopLyrics,
	ServiceUserService,
	ServiceLoginService,
	ServiceShareSvc,
	ServiceLastfm,
	ServiceEventBus,
	ServiceDispatcher,
}

// UserService carries the current login state (user 状态迁移).
type UserService struct {
	// User points at the live engine-owned user slot; dereferencing always
	// yields the current user (nil until LoadFromStorage/LoginWithCookie or a
	// LoginCallback populate it).
	User **structs.User
	// Login refreshes the user profile after a successful login.
	Login func() error
	// Jar points at the app-wide appCookieJar slot (the same slot
	// LoginService.CookieJar points at). LoginWithCookie replaces it after a
	// successful token refresh so LoginCallback persists the refreshed jar.
	Jar **cookiejar.Jar

	// loadStoredUser restores the persisted user from BoltDB; overridable in
	// tests to isolate from the real data dir.
	loadStoredUser func() (*structs.User, bool)
	// refreshJar refreshes the global cookie jar; overridable in tests.
	refreshJar func() (*cookiejar.Jar, error)
}

// LoadFromStorage restores the persisted user from BoltDB into the User slot.
// jar must have been initialized by LoginService.InitJar first; it is not read
// here but documents the cross-service ordering constraint (jar 先于 userService).
func (s *UserService) LoadFromStorage(jar *cookiejar.Jar) {
	if s.User == nil || s.loadStoredUser == nil {
		return
	}
	if user, ok := s.loadStoredUser(); ok {
		*s.User = user
	}
}

// LoginWithCookie runs the cookie-login flow: when no user is loaded and a
// cookie (MUSICFOX_COOKIE env or config.Main.Account.NeteaseCookie) is present,
// it parses the cookie into jar and refreshes the token; otherwise it refreshes
// the token for the already-loaded user. Both paths persist the refreshed jar
// BEFORE running the Login callback, so LoginService.InitJar must have run
// first (the callback persists the jar again via appCookieJar.Save()).
func (s *UserService) LoginWithCookie(jar *cookiejar.Jar, cookiePath string) {
	if s.refreshJar == nil {
		slog.Error("userService.refreshJar not wired, skipping cookie login")
		return
	}

	cookieStr := os.Getenv("MUSICFOX_COOKIE")
	if cookieStr == "" && configs.AppConfig != nil {
		cookieStr = configs.AppConfig.Main.Account.NeteaseCookie
	}

	if s.User != nil && *s.User == nil && cookieStr != "" {
		// 使用cookie登录
		if err := apputils.ParseCookieFromStr(cookieStr, jar); err != nil {
			slog.Error("网易云 cookies 格式错误", "error", err)
		} else {
			neteaseutil.SetGlobalCookieJar(jar)
			newJar, err := s.refreshJar()
			if err != nil {
				slog.Error("使用配置项的cookie登录/刷新失败，将以游客模式启动", slogx.Error(err))
				*s.User = nil
			} else {
				jar = newJar
				if s.Jar != nil {
					*s.Jar = jar
				}
				neteaseutil.SetGlobalCookieJar(jar)

				// 先保存 cookie，确保 token 刷新成功后 cookie 被持久化
				// 即使后续 LoginCallback 失败，cookie 也已保存
				if err := jar.Save(); err != nil {
					slog.Warn("持久化 Cookie 失败", slogx.Error(err))
				}

				if s.Login != nil {
					if err := s.Login(); err != nil {
						slog.Warn("使用配置项的cookie获取用户信息失败", slogx.Error(err))
						*s.User = nil
					}
				}
			}
		}
	}

	if s.User != nil && *s.User != nil {
		newJar, err := s.refreshJar()
		if err != nil {
			slog.Error("Token 刷新失败，Cookie已彻底失效，降级为游客模式", slogx.Error(err))
			*s.User = nil
			table := storage.NewTable()
			_ = table.DeleteByKVModel(storage.User{})
			_ = os.Remove(cookiePath)
		} else {
			jar = newJar
			if s.Jar != nil {
				*s.Jar = jar
			}
			neteaseutil.SetGlobalCookieJar(jar)
			slog.Info("Token 刷新成功~")

			// 先保存 cookie，确保 token 刷新成功后 cookie 被持久化
			// 即使后续 LoginCallback 失败，cookie 也已保存
			if err := jar.Save(); err != nil {
				slog.Warn("持久化 Cookie 失败", slogx.Error(err))
			}

			if s.Login != nil {
				if err := s.Login(); err != nil {
					slog.Warn("触发登录回调失败", slogx.Error(err))
				}
			}
		}
	}
}

// loadStoredUserFromStorage reads the persisted user from BoltDB.
func loadStoredUserFromStorage() (*structs.User, bool) {
	table := storage.NewTable()
	if jsonStr, err := table.GetByKVModel(storage.User{}); err == nil {
		if user, err := structs.NewUserFromLocalJson(jsonStr); err == nil {
			return &user, true
		}
	}
	return nil, false
}

// LoginService owns the login-flow state (appCookieJar 归属收敛).
type LoginService struct {
	// CookieJar points at the package-level appCookieJar slot; dereferencing
	// always yields the current jar (nil until InitJar initializes it).
	CookieJar **cookiejar.Jar
	// User points at the live engine-owned user slot; it is cleared when the
	// jar cannot be restored so the app degrades to tourist mode.
	User **structs.User
}

// InitJar initializes the persistent cookie jar at cookiePath, owning the whole
// jar lifecycle: corrupt-file backup/reset, appCookieJar slot assignment and
// the neteaseutil.SetGlobalCookieJar sync. It must run before any userService
// callback that persists the jar.
func (s *LoginService) InitJar(cookiePath string) (*cookiejar.Jar, error) {
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: cookiePath,
	})
	if err != nil {
		slog.Warn("检测到旧版或损坏的 Cookie 文件，准备备份并重置", slogx.Error(err))

		// 备份旧文件
		timestamp := time.Now().Format("20060102-150405")
		backupPath := fmt.Sprintf("%s.bak.%s", cookiePath, timestamp)

		if renameErr := os.Rename(cookiePath, backupPath); renameErr != nil && !os.IsNotExist(renameErr) {
			slog.Error("无法备份损坏的 Cookie 文件", slogx.Error(renameErr))
			if s.User != nil {
				*s.User = nil
			}
		} else {
			slog.Info("已将损坏的 Cookie 文件备份", "backup_path", backupPath)
		}

		// 重新初始化
		jar, err = cookiejar.New(&cookiejar.Options{
			Filename: cookiePath,
		})
		if err != nil {
			slog.Error("无法创建持久化 Cookie Jar，将降级为临时会话，重启后将丢失登录状态", slogx.Error(err))
			// 降级为内存模式
			memJar, _ := cookiejar.New(nil)
			jar = memJar

			if s.User != nil {
				*s.User = nil
			}
		} else {
			slog.Info("Cookie 文件已重置，请重新登陆")
		}
	}

	if s.CookieJar != nil {
		*s.CookieJar = jar
	}
	neteaseutil.SetGlobalCookieJar(jar)

	return jar, nil
}

// provideIfAbsent registers svc under name unless it is already present, so the
// scope-provided services (shareSvc/lastfm) and the startup registration can
// coexist without a duplicate-Provide panic.
func provideIfAbsent(ctx *framework.Context, name string, svc any) {
	if ctx.Service(name) == nil {
		ctx.Provide(name, svc)
	}
}

// ProvideIfAbsent is the exported form of provideIfAbsent so the ui layer can
// register its TUI-only services into the same engine context.
func ProvideIfAbsent(ctx *framework.Context, name string, svc any) {
	provideIfAbsent(ctx, name, svc)
}
