package ui

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/netease-music/service"
	"github.com/go-musicfox/netease-music/util"
	neteaseutil "github.com/go-musicfox/netease-music/util"
	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/pkg/errors"

	"github.com/go-musicfox/go-musicfox/internal/automator"
	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/track"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	apputils "github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
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
	login  *LoginPage
	search *SearchPage

	lyricService *lyric.Service

	lyricRenderer    *LyricRenderer
	songInfoRenderer *SongInfoRenderer
	progressRenderer *ProgressRenderer
	coverRenderer    *CoverRenderer

	player       *Player
	shareSvc     *composer.ShareService
	trackManager *track.Manager
}

func NewNetease(app *model.App) *Netease {
	n := new(Netease)
	n.lastfm = lastfm.NewClient()

	quality := configs.AppConfig.Player.SongLevel
	maxSizeMB := configs.AppConfig.Storage.Cache.Limit
	nameGen := composer.NewFileNameGenerator()
	nameGen.RegisterSongTemplate(configs.AppConfig.Storage.FileNameTpl)
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
	n.player = NewPlayer(n, n.lyricService)

	n.lyricRenderer = NewLyricRenderer(n, n.lyricService, showLyric)
	n.songInfoRenderer = NewSongInfoRenderer(n, n.player)
	n.progressRenderer = NewProgressRenderer(n, n.player)
	n.coverRenderer = NewCoverRenderer(n, n.player)

	n.login = NewLoginPage(n)
	n.search = NewSearchPage(n)
	n.App = app

	n.shareSvc = composer.NewShareService()
	n.shareSvc.RegisterTemplates(configs.AppConfig.Share)

	return n
}

func (n *Netease) Components() []model.Component {
	// CoverRenderer uses absolute positioning and returns 0 lines,
	// so it doesn't affect the layout of other components.
	// It should be rendered LAST to overlay the cover image after normal layout.
	if n.coverRenderer.IsEnabled() {
		return []model.Component{
			n.lyricRenderer,
			n.songInfoRenderer,
			n.progressRenderer,
			n.coverRenderer,
		}
	}

	// Default: just lyrics without cover
	return []model.Component{
		n.lyricRenderer,
		n.songInfoRenderer,
		n.progressRenderer,
	}
}

// ToLoginPage 需要登录的处理
func (n *Netease) ToLoginPage(callback func() model.Page) (model.Page, tea.Cmd) {
	n.login.AfterLogin = callback
	return n.login, tickLogin(time.Nanosecond)
}

// ToSearchPage 搜索
func (n *Netease) ToSearchPage(searchType SearchType) (model.Page, tea.Cmd) {
	n.search.searchType = searchType
	return n.search, tickSearch(time.Nanosecond)
}

func (n *Netease) InitHook(_ *model.App) {
	config := configs.AppConfig
	dataDir := app.DataDir()

	// 全局文件Jar
	cookiePath := filepath.Join(dataDir, "cookie")
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: cookiePath,
	})
	if err != nil {
		slog.Warn("检测到旧版或损坏的 Cookie 文件，开始重置", slogx.Error(err))
		if removeErr := os.Remove(cookiePath); removeErr != nil && !os.IsNotExist(removeErr) {
			slog.Error("无法删除损坏的 Cookie 文件", slogx.Error(removeErr))
			panic("failed to remove broken cookie file")
		}

		jar, err = cookiejar.New(&cookiejar.Options{
			Filename: cookiePath,
		})
		if err != nil {
			slog.Error("重置后仍旧无法创建 cookie jar", slogx.Error(err))
			panic("failed to create persistent cookie jar after reset")
		}

		slog.Info("Cookie 文件已重置，请重新登陆")
	}

	appCookieJar = jar
	util.SetGlobalCookieJar(appCookieJar)

	// 获取用户信息
	errorx.Go(func() {
		table := storage.NewTable()

		// 获取用户信息
		if jsonStr, err := table.GetByKVModel(storage.User{}); err == nil {
			if user, err := structs.NewUserFromLocalJson(jsonStr); err == nil {
				n.user = &user
			}
		}

		cookieStr := os.Getenv("MUSICFOX_COOKIE")
		if cookieStr == "" {
			cookieStr = config.Main.Account.NeteaseCookie
		}
		if n.user == nil && cookieStr != "" {
			// 使用cookie登录

			err := apputils.ParseCookieFromStr(cookieStr, appCookieJar)
			if err != nil {
				slog.Error("网易云 cookies 格式错误", "error", err)
			} else {
				neteaseutil.SetGlobalCookieJar(appCookieJar)
				newJar, err := apputils.RefreshCookieJar()
				if err != nil {
					slog.Error("使用配置项的cookie登录/刷新失败，将以游客模式启动", slogx.Error(err))
					n.user = nil
				} else {
					appCookieJar = newJar
					neteaseutil.SetGlobalCookieJar(appCookieJar)
					if err := n.LoginCallback(); err != nil {
						slog.Warn("使用配置项的cookie获取用户信息失败", slogx.Error(err))
						n.user = nil
					}
				}
			}
		}

		if n.user != nil {
			newJar, err := apputils.RefreshCookieJar()
			if err != nil {
				slog.Error("Token 刷新失败，Cookie已彻底失效，降级为游客模式", slogx.Error(err))
				n.user = nil
				_ = table.DeleteByKVModel(storage.User{})
				_ = os.Remove(cookiePath)
			} else {
				appCookieJar = newJar
				neteaseutil.SetGlobalCookieJar(appCookieJar)
				slog.Info("Token 刷新成功~")

				if err := n.LoginCallback(); err != nil {
					slog.Warn("触发登录回调失败", slogx.Error(err))
				}
			}
		}

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
		if err := n.player.playlistManager.LoadState(); err != nil {
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
				})
			}
		}

		// 检查更新
		if config.Startup.CheckUpdate {
			if ok, newVersion := version.CheckUpdate(); ok {
				if runtime.GOOS == "windows" {
					n.MustMain().EnterMenu(
						NewCheckUpdateMenu(newBaseMenu(n)),
						&model.MenuItem{Title: "新版本: " + newVersion, Subtitle: "当前版本: " + types.AppVersion},
					)
					n.Rerender(false)
				}

				notify.Notify(notify.NotifyContent{
					Title: "发现新版本: " + newVersion,
					Text:  "去看看呗",
					Url:   types.AppLatestReleases,
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
				})
			}
		}
	})
}

func (n *Netease) CloseHook(_ *model.App) {
	_ = n.player.Close()
	n.lastfm.Close()

	if n.coverRenderer != nil {
		n.coverRenderer.Close()
	}
}

func (n *Netease) Player() *Player {
	return n.player
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
