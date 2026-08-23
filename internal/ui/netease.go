// Package ui implements the foxful-cli TUI layer: the thin-shell Netease
// coordinator (navigation/assembly/event dispatch), provider-registered menus
// and pages, and the lyric/cover/spectrum renderers composed by the shell.
// Business capabilities are resolved by name through internal/framework.
package ui

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/buger/jsonparser"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/go-musicfox/netease-music/service"
	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/pkg/errors"

	"github.com/go-musicfox/go-musicfox/internal/automator"
	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/desktop_lyrics"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/track"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/filex"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	"github.com/go-musicfox/go-musicfox/utils/version"
)

var appCookieJar *cookiejar.Jar

type Netease struct {
	user   *structs.User
	lastfm *lastfm.Client

	*model.App
	search *SearchPage

	lyricService *lyric.Service

	lyricRenderer       *LyricRenderer
	songInfoRenderer    *SongInfoRenderer
	progressRenderer    *ProgressRenderer
	coverRenderer       *CoverRenderer
	spectrumRenderer    *SpectrumRenderer
	spectrogramRenderer *SpectrogramRenderer

	player       *Player
	shareSvc     *composer.ShareService
	trackManager *track.Manager

	// framework context + scope own the app-wide service registry and its
	// lifecycle (see services_scope.go / services.go).
	ctx   *framework.Context
	scope *framework.Scope

	playbarHoveredElement PlaybarElement

	// Theme switch notification: update in-place when visible, recreate when expired.
	themeNotifID    model.NotificationID
	themeNotifTimer *time.Timer

	desktopLyrics desktop_lyrics.Controller
}

func NewNetease(app *model.App) *Netease {
	n := new(Netease)
	n.lastfm = lastfm.NewClient()

	quality := configs.AppConfig.Player.SongLevel
	maxSizeMB := configs.AppConfig.Storage.Cache.Limit
	nameGen := composer.NewFileNameGenerator()
	_ = nameGen.RegisterSongTemplate(configs.AppConfig.Storage.FileNameTpl)
	_ = nameGen.RegisterLyricTemplate(configs.AppConfig.Storage.FileNameTpl)
	n.trackManager = track.NewManager(
		track.WithNameGenerator(nameGen),
		track.WithCacher(track.NewCacher(maxSizeMB)),
		track.WithSongQuality(quality))

	showTranslation := configs.AppConfig.Main.Lyric.ShowTranslation
	offset := time.Duration(configs.AppConfig.Main.Lyric.Offset) * time.Millisecond
	showLyric := configs.AppConfig.Main.Lyric.Show
	skipParseErr := configs.AppConfig.Main.Lyric.SkipParseErr

	n.lyricService = lyric.NewService(n.trackManager, showTranslation, offset, skipParseErr)
	n.lyricService.EnableYRC(true) // Enable word-by-word lyrics

	// Initialize desktop lyrics
	n.desktopLyrics = desktop_lyrics.NewController(configs.AppConfig.Main.Lyric.DesktopLyrics)

	n.player = NewPlayer(n, n.lyricService)

	n.lyricRenderer = NewLyricRenderer(newMenuServices(n), n.lyricService, showLyric)
	n.songInfoRenderer = NewSongInfoRenderer(newMenuServices(n), n.player)
	n.progressRenderer = NewProgressRenderer(newMenuServices(n), n.player)
	n.coverRenderer = NewCoverRenderer(newMenuServices(n), n.player)
	n.spectrumRenderer = NewSpectrumRenderer(n.player)
	n.spectrogramRenderer = NewSpectrogramRenderer(n.player)

	// The search page is a shell-owned singleton: its wordsInput/result/
	// searchType state is shared with the SearchResultMenu flow (and operate.go
	// searchSong), so the shell keeps one instance built through the registry.
	// The login page is built per-navigation in ToLoginPage (no cross-component
	// state to preserve), so the shell holds no login field.
	searchPage, err := BuildPage("search", SearchPageOpts{Netease: n})
	if err != nil {
		return nil
	}
	n.search = searchPage.(*SearchPage) // BuildPage returns model.Page; concrete type asserted back
	n.App = app

	n.shareSvc = composer.NewShareService()
	_ = n.shareSvc.RegisterTemplates(configs.AppConfig.Share)

	// Wire the framework scope: shareSvc/lastfm are registered into the
	// app-wide context via their scope plugins (Phase 3.1.1 slice), then the
	// remaining startup services are registered (Phase 3.1.2).
	n.ctx = &framework.Context{}
	n.scope = newAppScope(n)
	if err := n.scope.Start(n.ctx); err != nil {
		slog.Error("framework scope start failed", slogx.Error(err))
		return nil
	}
	if err := registerServices(n.ctx, n); err != nil {
		slog.Error("framework service registration failed", slogx.Error(err))
		return nil
	}

	// Startup completeness: a provider set missing any canonical key is a
	// programmer error; fail loudly instead of surfacing it at navigation time.
	AssertMenuRegistryComplete(expectedMenuKeys...)
	AssertPageRegistryComplete(expectedPageKeys...)

	return n
}

func (n *Netease) Components() []model.Component {
	var components []model.Component
	if n.spectrumRenderer.IsEnabled() {
		components = append(components, n.spectrumRenderer)
	}
	if n.spectrogramRenderer.IsEnabled() {
		components = append(components, n.spectrogramRenderer)
	}
	components = append(components, n.lyricRenderer)
	components = append(components, n.songInfoRenderer, n.progressRenderer)
	// CoverRenderer uses absolute positioning and returns 0 lines, so it must
	// be rendered last to overlay the normal layout.
	if n.coverRenderer.IsEnabled() {
		components = append(components, n.coverRenderer)
	}
	return components
}

func (n *Netease) SpectrumLines(main *model.Main) int {
	windowHeight := n.EffectiveWindowHeight()
	menuBottomRow := main.MenuBottomRow()
	if n.spectrumRenderer != nil && n.spectrumRenderer.IsEnabled() {
		return n.spectrumRenderer.LineCount(windowHeight, menuBottomRow)
	}
	if n.spectrogramRenderer != nil && n.spectrogramRenderer.IsEnabled() {
		return n.spectrogramRenderer.LineCount(windowHeight, menuBottomRow)
	}
	return 0
}

// EffectiveWindowHeight returns the available content height, excluding the
// status bar if present. Components should use this instead of WindowHeight()
// for bottom-anchored layout to avoid overlapping the status bar.
func (n *Netease) EffectiveWindowHeight() int {
	return n.MustMain().EffectiveWindowHeight(n.App)
}

// ToLoginPage 需要登录的处理
func (n *Netease) ToLoginPage(callback func() model.Page) (model.Page, tea.Cmd) {
	page := buildPageOrToast("login", LoginPageOpts{Netease: n})
	if page == nil {
		return nil, nil
	}
	page.(*LoginPage).AfterLogin = callback
	n.coverRenderer.ClearDisplayed()
	return page, tickLogin(time.Nanosecond)
}

// ToSearchPage 搜索
func (n *Netease) ToSearchPage(searchType SearchType) (model.Page, tea.Cmd) {
	n.search.searchType = searchType
	n.coverRenderer.ClearDisplayed()
	return n.search, tickSearch(time.Nanosecond)
}

func (n *Netease) InitHook(_ *model.App) {
	config := configs.AppConfig
	dataDir := app.DataDir()

	// 注册 TUI 内 toast 回调（此时 App.Run 已启动，program 就绪）
	n.registerToastHook()

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
	//  10. 检查更新（受 config.Startup.CheckUpdate 控制）。
	//  11. 自动播放（受 config.Autoplay.Enable 控制）。
	//  12. changelog 弹窗。
	// ---------------------------------------------------------------------

	// 1. 全局文件Jar（loginService 拥有整个 cookie-jar 生命周期）
	loginSvc, ok := framework.ServiceOf[*LoginService](n.ctx, ServiceLoginService)
	var jar *cookiejar.Jar
	if !ok || loginSvc == nil {
		slog.Error("loginService 未注册，跳过 cookie jar 初始化", slog.String("hook", "InitHook"))
	} else {
		var err error
		jar, err = loginSvc.InitJar(filepath.Join(dataDir, "cookie"))
		if err != nil {
			slog.Error("cookie jar 初始化失败，已降级为临时会话", slogx.Error(err))
		}
	}

	// 获取用户信息
	errorx.Go(func() {
		table := storage.NewTable()

		// 2-3. 用户恢复 + cookie 登录（userService 拥有用户与登录流程）
		if userSvc, ok := framework.ServiceOf[*UserService](n.ctx, ServiceUserService); ok && userSvc != nil {
			if jar != nil {
				userSvc.LoadFromStorage(jar)
				userSvc.LoginWithCookie(jar, filepath.Join(dataDir, "cookie"))
			} else {
				slog.Error("cookie jar 不可用，跳过用户恢复，以游客模式启动")
			}
		} else {
			slog.Error("userService 未注册，跳过用户恢复", slog.String("hook", "InitHook"))
		}

		cloudUserID := int64(0)
		if n.user != nil {
			cloudUserID = n.user.UserId
		}
		n.trackManager.SetCloudUserID(cloudUserID)

		// 刷新界面用户名
		n.MustMain().RefreshMenuTitle()

		// 获取播放模式
		if jsonStr, err := table.GetByKVModel(storage.PlayMode{}); err == nil && len(jsonStr) > 0 {
			var playMode types.Mode
			if err = json.Unmarshal(jsonStr, &playMode); err == nil {
				n.player.SetMode(playMode)
			}
		}

		// 获取音量
		if jsonStr, err := table.GetByKVModel(storage.Volume{}); err == nil && len(jsonStr) > 0 {
			var volume int
			if err = json.Unmarshal(jsonStr, &volume); err == nil {
				v, ok := n.player.Player.(storage.VolumeStorable)
				if ok {
					v.SetVolume(volume)
				}
			}
		}

		// 加载播放列表状态
		if err := n.player.LoadPlaylistState(); err != nil {
			// 如果加载失败，记录错误但不影响启动
			slog.Warn("Failed to load playlist state", slogx.Error(err))
		}
		n.Rerender(false)

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
		if n.user != nil {
			likelist.RefreshLikeList(n.user.UserId)
			n.Rerender(false)
		}

		// 签到
		if config.Startup.SignIn {
			var lastSignIn int
			if jsonStr, err := table.GetByKVModel(storage.LastSignIn{}); err == nil && len(jsonStr) > 0 {
				_ = json.Unmarshal(jsonStr, &lastSignIn)
			}
			today, err := strconv.Atoi(time.Now().Format("20060102"))
			if n.user != nil && err == nil && lastSignIn != today {
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

		// 检查更新：启动自动检查保留在 shell（无启动钩子机制；菜单触发的检查
		// 由 internal/plugins/checkupdate 插件承载，ui 不得反向导入插件包）。
		if config.Startup.CheckUpdate {
			if ok, newVersion := version.CheckUpdate(); ok {
				notify.Notify(notify.NotifyContent{
					Title:       "发现新版本: " + newVersion,
					Text:        "去看看呗",
					Url:         types.AppGithubUrl + "/releases/tag/" + newVersion,
					ActionLabel: "前往 GitHub",
				})
			}
		}

		// 自动播放
		if config.Autoplay.Enable {
			autoPlayer := automator.NewAutoPlayer(n.user, n.player, config.Autoplay)
			if err := autoPlayer.Start(); err != nil {
				slog.Error("自动播放失败", slogx.Error(err))
				notify.Notify(notify.NotifyContent{
					Title: "自动播放失败",
					Text:  err.Error(),
					Level: notify.ToastError,
				})
			}
		}

		// changelog: 首次启动新版本或 debug 模式 → 弹更新日志
		// 使用 AfterFunc 延迟弹窗，确保 startup 页完成、主页面已进入
		{
			slog.Debug("changelog: entering check",
				"debug", configs.AppConfig.Main.Debug,
				"appVersion", types.AppVersion,
				"app", n.App != nil,
			)
			jsonStr, _ := table.GetByKVModel(storage.ChangelogSeen{})
			var seen storage.ChangelogSeen
			if len(jsonStr) > 0 {
				_ = json.Unmarshal(jsonStr, &seen)
			}
			shouldShow := configs.AppConfig.Main.Debug || seen.Version == "" || version.CompareVersion(types.AppVersion, seen.Version, false)
			slog.Debug("changelog: shouldShow",
				"shouldShow", shouldShow,
				"debug", configs.AppConfig.Main.Debug,
				"seenVersion", seen.Version,
			)
			if shouldShow {
				if !configs.AppConfig.Main.Debug {
					if err := table.SetByKVModel(storage.ChangelogSeen{}, storage.ChangelogSeen{Version: types.AppVersion}); err != nil {
						slog.Error("changelog: failed to persist seen version", slogx.Error(err))
					} else {
						slog.Debug("changelog: persisted seen version", "version", types.AppVersion)
					}
				}
				app := n.App
				slog.Debug("changelog: scheduling AfterFunc", "hasApp", app != nil)
				time.AfterFunc(max(configs.AppConfig.Startup.ToModel().LoadingDuration, time.Second)-750*time.Millisecond, func() {
					slog.Debug("changelog: AfterFunc triggered, showing popup")
					showChangelogPopup(app)
				})
			}
		}
	})
}

func (n *Netease) CloseHook(_ *model.App) {
	_ = n.player.Close()
	n.lastfm.Close()

	if n.desktopLyrics != nil {
		n.desktopLyrics.Close()
	}

	if n.coverRenderer != nil {
		n.coverRenderer.Close()
	}

	if n.scope != nil {
		_ = n.scope.Stop()
		_ = n.scope.Dispose()
	}

	CloseGohookLogger()
}

// Ctx returns the app-wide framework context holding the service registry.
func (n *Netease) Ctx() *framework.Context {
	return n.ctx
}

func (n *Netease) Player() *Player {
	return n.player
}

// DesktopLyrics returns the desktop lyrics controller.
func (n *Netease) DesktopLyrics() desktop_lyrics.Controller {
	return n.desktopLyrics
}

// GetDesktopLyricsLines returns the current lyrics lines for desktop display.
// Returns the current line, the next line (with word-level data when YRC is available),
// and the current index.
func (n *Netease) GetDesktopLyricsLines() (curLine, nextLine desktop_lyrics.LyricLine, currentIndex int) {
	if n.desktopLyrics == nil || n.lyricService == nil {
		return desktop_lyrics.LyricLine{}, desktop_lyrics.LyricLine{}, -1
	}

	state := n.lyricService.State()
	if !state.IsRunning {
		return desktop_lyrics.LyricLine{}, desktop_lyrics.LyricLine{}, -1
	}

	// Helper to build display text including translation
	buildText := func(content, translation string) string {
		if translation != "" {
			return content + "\n" + translation
		}
		return content
	}

	if state.YRCEnabled && len(state.YRCLines) > 0 {
		idx := state.YRCLineIndex
		if idx < 0 {
			idx = 0
		}
		if idx >= len(state.YRCLines) {
			idx = len(state.YRCLines) - 1
		}

		// Current line — with word data
		if idx < len(state.YRCLines) {
			line := state.YRCLines[idx]
			var sb strings.Builder
			words := make([]desktop_lyrics.LyricWord, len(line.Words))
			for i, w := range line.Words {
				sb.WriteString(w.Word)
				words[i] = desktop_lyrics.LyricWord{
					Word:      w.Word,
					StartTime: w.StartTime,
					EndTime:   w.EndTime,
				}
			}
			curLine = desktop_lyrics.LyricLine{Text: buildText(sb.String(), line.TranslatedLyric), Words: words}
		}

		// Next line
		nextIdx := idx + 1
		if nextIdx < len(state.YRCLines) {
			line := state.YRCLines[nextIdx]
			var sb strings.Builder
			nextWords := make([]desktop_lyrics.LyricWord, len(line.Words))
			for i, w := range line.Words {
				sb.WriteString(w.Word)
				nextWords[i] = desktop_lyrics.LyricWord{
					Word:      w.Word,
					StartTime: w.StartTime,
					EndTime:   w.EndTime,
				}
			}
			nextLine = desktop_lyrics.LyricLine{Text: buildText(sb.String(), line.TranslatedLyric), Words: nextWords}
		}

		return curLine, nextLine, idx

	} else if len(state.Fragments) > 0 {
		idx := state.CurrentIndex
		if idx < 0 {
			idx = 0
		}
		if idx >= len(state.Fragments) {
			idx = len(state.Fragments) - 1
		}

		// Current line — plain text (no word data for LRC)
		if idx < len(state.Fragments) {
			f := state.Fragments[idx]
			trans := ""
			if state.ShowTranslation {
				trans = state.TranslatedFragments[f.StartTimeMs]
			}
			curLine = desktop_lyrics.LyricLine{Text: buildText(f.Content, trans)}
		}

		// Next line
		if idx+1 < len(state.Fragments) {
			f := state.Fragments[idx+1]
			trans := ""
			if state.ShowTranslation {
				trans = state.TranslatedFragments[f.StartTimeMs]
			}
			nextLine = desktop_lyrics.LyricLine{Text: buildText(f.Content, trans)}
		}

		return curLine, nextLine, idx
	}

	return desktop_lyrics.LyricLine{}, desktop_lyrics.LyricLine{}, -1
}

// GetCoverWidth returns the cover image width in columns, or 0 if cover is disabled.
func (n *Netease) GetCoverWidth() int {
	if n.coverRenderer == nil {
		return 0
	}
	return n.coverRenderer.GetCoverWidth()
}

// GetCoverEndColumn returns the column where the cover ends, or 0 if cover is disabled.
func (n *Netease) GetCoverEndColumn() int {
	if n.coverRenderer == nil {
		return 0
	}
	return n.coverRenderer.GetCoverEndColumn()
}

// GetLyricPosition returns the current lyric display position.
// Returns (startRow, lineCount). If lyrics are not visible, returns (0, 0).
func (n *Netease) GetLyricPosition() (startRow int, lineCount int) {
	if n.lyricRenderer == nil {
		return 0, 0
	}
	return n.lyricRenderer.GetLyricPosition()
}

func (n *Netease) LoginCallback() error {
	code, resp := (&service.UserAccountService{}).AccountInfo()
	if code != 200 {
		return errors.Errorf("accountInfo code: %f, resp: %s", code, string(resp))
	}

	user, err := structs.NewUserFromJsonForLogin(resp)
	if err != nil {
		return errors.WithMessagef(err, "parse user err, code: %f, resp: %s", code, string(resp))
	}
	n.user = &user
	n.trackManager.SetCloudUserID(user.UserId)

	// 获取我喜欢的歌单
	userPlaylists := service.UserPlaylistService{
		Uid:    strconv.FormatInt(n.user.UserId, 10),
		Limit:  strconv.Itoa(1),
		Offset: strconv.Itoa(0),
	}
	_, response := userPlaylists.UserPlaylist()
	n.user.MyLikePlaylistID, err = jsonparser.GetInt(response, "playlist", "[0]", "id")
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

// showChangelogPopup reads the embedded CHANGELOG.md and displays it as a markdown popup.
func showChangelogPopup(app *model.App) {
	data, err := filex.ReadFileFromEmbed("embed/changelog.md")
	if err != nil {
		slog.Error("changelog load failure", slogx.Error(err))
		return
	}

	popupWidth := min(max(app.WindowWidth()*70/100, 40), 120)

	maxHeight := max(app.WindowHeight()*80/100, 10)

	popup, err := model.NewMarkdownPopup(model.MarkdownPopupSpec{
		Title:           types.AppVersion + " 更新日志",
		MarkdownContent: string(data),
		MaxWidth:        popupWidth,
		MaxHeight:       maxHeight,
		Anchor:          model.AnchorCenter,
		DisableResize:   false,
		CloseKeys:       []string{"esc", "q"},
	})
	if err != nil {
		slog.Error("changelog popup creation failure", slogx.Error(err))
		return
	}
	app.ShowPopup(popup)
	// Trigger a rerender so the popup appears immediately without waiting for a keypress
	app.Rerender(true)
}

// Update intercepts system background-change messages to rebuild the StyleSet
// from go-musicfox's ThemeRegistry with the correct dark/light variant.
// Foxful-cli's built-in onBackgroundChanged handles SetDarkBackground and popup
// cache invalidation; we override the StyleSet afterwards to use our theme files.
func (n *Netease) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.BackgroundColorMsg, uv.LightColorSchemeEvent, uv.DarkColorSchemeEvent:
		registry := configs.CurrentThemeRegistry()
		syncActiveThemePair(n.App, registry)
		// Let foxful-cli process the message first (SetDarkBackground, popup cache, etc.)
		_, cmd := n.App.Update(msg)
		// Rebuild StyleSet from our ThemeRegistry with correct dark/light variant
		isDark := style.HasDarkBackground()
		ss := registry.CurrentStyleSet(isDark)
		if ss != nil {
			style.SetStyleSet(*ss)
			n.SetStyleSet(*ss)
		}
		n.notifyThemeSwitch(n.App, "外观模式已切换", configs.CurrentThemeRegistry().CurrentName(isDark))
		return n, tea.Sequence(cmd, n.RerenderCmd(true))
	default:
		_, cmd := n.App.Update(msg)
		return n, cmd
	}
}

func syncActiveThemePair(app *model.App, registry *configs.ThemeRegistry) {
	dark, light, ok := registry.CurrentThemePair()
	if ok {
		app.SetThemePair(dark, light)
	}
}

// notifyThemeSwitch shows or updates a theme switch notification.
// If a previous notification is still visible, updates it in place.
// If it has expired, creates a new one.
func (n *Netease) notifyThemeSwitch(app *model.App, title, name string) {
	const timeout = 4 * time.Second
	spec := model.NotificationSpec{
		Level:   model.NotificationInfo,
		Title:   title,
		Message: name,
		Timeout: timeout,
	}
	if n.themeNotifID == 0 {
		n.themeNotifID = app.Notify(spec)
	} else {
		app.UpdateNotification(n.themeNotifID, spec)
	}
	if n.themeNotifTimer != nil {
		n.themeNotifTimer.Stop()
	}
	n.themeNotifTimer = time.AfterFunc(timeout, func() {
		n.themeNotifID = 0
	})
}
