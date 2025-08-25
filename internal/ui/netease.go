package ui

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
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
	"github.com/pkg/errors"
	"github.com/telanflow/cookiejar"

	"github.com/go-musicfox/go-musicfox/internal/automator"
	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/track"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	"github.com/go-musicfox/go-musicfox/utils/version"
)

type Netease struct {
	user   *structs.User
	lastfm *lastfm.Client

	*model.App
	login  *LoginPage
	search *SearchPage

	player       *Player
	shareSvc     *composer.ShareService
	trackManager *track.Manager
}

func NewNetease(app *model.App) *Netease {
	n := new(Netease)
	n.lastfm = lastfm.NewClient()
	n.player = NewPlayer(n)
	n.login = NewLoginPage(n)
	n.search = NewSearchPage(n)
	n.App = app

	n.shareSvc = composer.NewShareService()
	n.shareSvc.RegisterTemplates(configs.ConfigRegistry.Share)

	quality := configs.ConfigRegistry.Main.PlayerSongLevel
	maxSizeMB := configs.ConfigRegistry.Main.CacheLimit
	nameGen := composer.NewFileNameGenerator()
	nameGen.RegisterSongTemplate(configs.ConfigRegistry.Main.DownloadFileNameTpl)
	n.trackManager = track.NewManager(
		track.WithNameGenerator(nameGen),
		track.WithCacher(track.NewCacher(maxSizeMB)),
		track.WithSongQuality(quality))

	return n
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
	config := configs.ConfigRegistry
	dataDir := app.DataDir()

	// 全局文件Jar
	cookieJar, _ := cookiejar.NewFileJar(filepath.Join(dataDir, "cookie"), nil)
	util.SetGlobalCookieJar(cookieJar)

	// 获取用户信息
	errorx.Go(func() {
		table := storage.NewTable()

		// 获取用户信息
		if jsonStr, err := table.GetByKVModel(storage.User{}); err == nil {
			if user, err := structs.NewUserFromLocalJson(jsonStr); err == nil {
				n.user = &user
			}
		}

		if n.user == nil && config.Main.NeteaseCookie != "" {
			// 使用cookie登录
			cookieJar.SetCookies(
				errorx.Must1(url.Parse("https://music.163.com")),
				errorx.Must1(http.ParseCookie(config.Main.NeteaseCookie)),
			)
			if err := n.LoginCallback(); err != nil {
				slog.Warn("使用cookie登录失败", slogx.Error(err))
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
				// 手机签到
				signInService := service.DailySigninService{}
				signInService.Type = "0"
				signInService.DailySignin()
				// PC签到
				signInService.Type = "1"
				signInService.DailySignin()
				// 云贝签到
				ybSignService := service.YunbeiSigninService{}
				ybSignService.Signin()

				_ = table.SetByKVModel(storage.LastSignIn{}, today)

				notify.Notify(notify.NotifyContent{
					Title:   "签到成功",
					Text:    "今日手机、PC端签到成功",
					Url:     types.AppGithubUrl,
					GroupId: types.GroupID,
				})
			}
		}

		// 刷新登录状态
		if n.user != nil {
			refreshLoginService := service.LoginRefreshService{}
			refreshLoginService.LoginRefresh()
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
		if config.AutoPlayer.Enable {
			autoPlayer := automator.NewAutoPlayer(n.user, n.player, config.AutoPlayer)
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
}

func (n *Netease) Player() *Player {
	return n.player
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

	// 更新like list
	go likelist.RefreshLikeList(user.UserId)

	return nil
}
