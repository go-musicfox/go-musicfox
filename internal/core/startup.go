package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-musicfox/netease-music/service"
	cookiejar "github.com/juju/persistent-cookiejar"

	"github.com/go-musicfox/go-musicfox/internal/automator"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	"github.com/go-musicfox/go-musicfox/utils/version"
)

// StartupPhase names a milestone of the engine startup sequence. Frontends
// react per phase (the TUI refreshes titles/rerenders; headless ignores them).
type StartupPhase string

const (
	// StartupPhaseUserRestored fires after login/user restore (TUI:
	// RefreshMenuTitle + Rerender).
	StartupPhaseUserRestored = StartupPhase("user_restored")
	// StartupPhasePlaylistLoaded fires after playlist state load.
	StartupPhasePlaylistLoaded = StartupPhase("playlist_loaded")
	// StartupPhaseBeforeAutoplay fires right before autoplay.
	StartupPhaseBeforeAutoplay = StartupPhase("before_autoplay")
)

// emitStartupPhase forwards a phase milestone to the observer (nil-safe).
func emitStartupPhase(observer Observer, phase StartupPhase) {
	if o, ok := observer.(StartupPhaseObserver); ok {
		o.OnStartupPhase(phase)
	}
}

// Startup runs the full startup sequence synchronously. It must be called
// after the engine is built; frontends run it in their own goroutine. The
// sequence order is preserved from the former shell InitHook (jar before
// userService callback, login before like-list/sign-in/autoplay, hooks before
// autoplay).
func (e *Engine) Startup(ctx context.Context, observer Observer) error {
	config := configs.AppConfig
	dataDir := app.DataDir()

	// ---------------------------------------------------------------------
	// Startup sequence — order constraints (逐条枚举，拆分后必须保持):
	//
	//  1. loginService.InitJar — cookie jar 初始化（损坏备份/重置、appCookieJar
	//     全局赋值、util.SetGlobalCookieJar 同步）。必须先于 userService 回调：
	//     LoginCallback 内部调用 appCookieJar.Save()（jar 先于 userService 回调）。
	//  2. userService.LoadFromStorage — 从 BoltDB 恢复持久化用户。
	//  3. userService.LoginWithCookie — cookie 登录 / token 刷新（ParseCookieFromStr
	//     → RefreshCookieJar → jar 保存 → LoginCallback）。
	//  4. 播放模式恢复（storage.PlayMode）。
	//  5. 音量恢复（storage.Volume）。
	//  6. 播放列表状态加载（playlistManager.LoadState）。
	//  7. extInfo / notifier / logo 清理。
	//  8. like list 刷新（仅登录态）。
	//  9. 每日签到（仅登录态，受 config.Startup.SignIn 控制）。
	//  10. 插件启动钩子（插件注册的启动任务，如 checkupdate 的自动检查）。
	//  11. 自动播放（受 config.Autoplay.Enable 控制）。
	//  12. changelog 弹窗（TUI-only，由前端在 Startup 返回后执行）。
	// ---------------------------------------------------------------------

	// 1. 全局文件Jar（loginService 拥有整个 cookie-jar 生命周期）
	loginSvc, ok := framework.ServiceOf[*LoginService](e.ctx, ServiceLoginService)
	var jar *cookiejar.Jar
	if !ok || loginSvc == nil {
		slog.Error("loginService 未注册，跳过 cookie jar 初始化", slog.String("hook", "Startup"))
	} else {
		var err error
		jar, err = loginSvc.InitJar(filepath.Join(dataDir, "cookie"))
		if err != nil {
			slog.Error("cookie jar 初始化失败，已降级为临时会话", slogx.Error(err))
		}
	}

	// 获取用户信息
	table := storage.NewTable()

	// 2-3. 用户恢复 + cookie 登录（userService 拥有用户与登录流程）
	if userSvc, ok := framework.ServiceOf[*UserService](e.ctx, ServiceUserService); ok && userSvc != nil {
		if jar != nil {
			userSvc.LoadFromStorage(jar)
			userSvc.LoginWithCookie(jar, filepath.Join(dataDir, "cookie"))
		} else {
			slog.Error("cookie jar 不可用，跳过用户恢复，以游客模式启动")
		}
	} else {
		slog.Error("userService 未注册，跳过用户恢复", slog.String("hook", "Startup"))
	}

	cloudUserID := int64(0)
	if user := e.User(); user != nil {
		cloudUserID = user.UserId
	}
	e.trackManager.SetCloudUserID(cloudUserID)

	// 刷新界面用户名（前端在 OnStartupPhase 处理）
	emitStartupPhase(observer, StartupPhaseUserRestored)

	// 获取播放模式
	if jsonStr, err := table.GetByKVModel(storage.PlayMode{}); err == nil && len(jsonStr) > 0 {
		var playMode types.Mode
		if err = json.Unmarshal(jsonStr, &playMode); err == nil {
			e.player.SetMode(playMode)
		}
	}

	// 获取音量
	if jsonStr, err := table.GetByKVModel(storage.Volume{}); err == nil && len(jsonStr) > 0 {
		var volume int
		if err = json.Unmarshal(jsonStr, &volume); err == nil {
			v, ok := e.player.Engine().(storage.VolumeStorable)
			if ok {
				v.SetVolume(volume)
			}
		}
	}

	// 加载播放列表状态
	if err := e.player.LoadPlaylistState(); err != nil {
		// 如果加载失败，记录错误但不影响启动
		slog.Warn("Failed to load playlist state", slogx.Error(err))
	}
	emitStartupPhase(observer, StartupPhasePlaylistLoaded)

	// 获取扩展信息
	{
		var (
			extInfo    storage.ExtInfo
			needUpdate = true
		)
		jsonStr, _ := table.GetByKVModel(extInfo)
		if len(jsonStr) != 0 {
			if err := json.Unmarshal(jsonStr, &extInfo); err == nil && version.CompareVersion(extInfo.StorageVersion, types.AppVersion, true) {
				needUpdate = false
			}
		}
		if needUpdate {
			// 删除旧notifier
			_ = os.RemoveAll(filepath.Join(dataDir, "musicfox-notifier.app"))

			// 删除旧logo
			_ = os.Remove(filepath.Join(dataDir, types.DefaultNotifyIcon))

			extInfo.StorageVersion = types.AppVersion
			_ = table.SetByKVModel(extInfo, extInfo)
		}
	}

	// 刷新like list
	if user := e.User(); user != nil {
		likelist.RefreshLikeList(user.UserId)
	}

	// 签到
	if config.Startup.SignIn {
		var lastSignIn int
		if jsonStr, err := table.GetByKVModel(storage.LastSignIn{}); err == nil && len(jsonStr) > 0 {
			_ = json.Unmarshal(jsonStr, &lastSignIn)
		}
		today, err := strconv.Atoi(time.Now().Format("20060102"))
		if e.User() != nil && err == nil && lastSignIn != today {
			var notifyMsg string
			// 手机签到
			signInService := service.DailySigninService{}
			signInService.Type = "0"
			signInService.DailySignin()
			notifyMsg += "手机✅ "
			// PC签到
			signInService.Type = "1"
			signInService.DailySignin()
			notifyMsg += "PC✅ "
			// 云贝签到
			yunbeiService := service.YunbeiService{}
			result, err := yunbeiService.Sign()

			var yunbeiResult string
			if err != nil {
				slog.Error("云贝签到网络/接口错误", slogx.Error(err))
				yunbeiResult = "云贝:异常❌"
			} else if result.Code != 200 {
				slog.Warn("云贝签到返回非200", "code", result.Code, "msg", result.Message)
				yunbeiResult = "云贝:失败❌"
			} else {
				if result.Data.YunbeiNum > 0 {
					yunbeiResult = fmt.Sprintf("云贝:+%d✅", result.Data.YunbeiNum)
					slog.Info("云贝签到成功", "数量", result.Data.YunbeiNum)
				} else {
					yunbeiResult = "云贝:无奖励✅"
				}
			}
			notifyMsg += yunbeiResult

			_ = table.SetByKVModel(storage.LastSignIn{}, today)
			notify.Notify(notify.NotifyContent{
				Title:   "自动签到完成",
				Text:    notifyMsg,
				Url:     types.AppGithubUrl,
				GroupId: types.GroupID,
				Level:   notify.ToastSuccess,
			})
		}
	}

	// 插件启动钩子（Phase 3.9）：运行插件注册的启动任务。此位置即原
	// shell 级「检查更新」自动检查（config.Startup.CheckUpdate 门控）的
	// 位置——该逻辑已移入 internal/plugins/checkupdate 并注册为启动钩子，
	// 引擎层不再反向导入插件包。此时用户/登录已就绪、services 已注册；
	// 每个 hook 带 panic 隔离（recover + 日志，不阻断启动）。
	framework.RunStartupHooks(configs.IsPluginEnabled)

	// 自动播放
	emitStartupPhase(observer, StartupPhaseBeforeAutoplay)
	if config.Autoplay.Enable {
		autoPlayer := automator.NewAutoPlayer(e.User(), e.player, config.Autoplay)
		if err := autoPlayer.Start(); err != nil {
			slog.Error("自动播放失败", slogx.Error(err))
			notify.Notify(notify.NotifyContent{
				Title: "自动播放失败",
				Text:  err.Error(),
				Level: notify.ToastError,
			})
		}
	}

	return nil
}
