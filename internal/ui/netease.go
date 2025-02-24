package ui

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/netease-music/service"
	"github.com/go-musicfox/netease-music/util"
	"github.com/telanflow/cookiejar"

	"github.com/go-musicfox/go-musicfox/internal/automator"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	"github.com/go-musicfox/go-musicfox/utils/version"
)

type Netease struct {
	user       *structs.User
	lastfm     *lastfm.Client
	lastfmUser *storage.LastfmUser

	*model.App
	login  *LoginPage
	search *SearchPage

	player *Player
}

func NewNetease(app *model.App) *Netease {
	n := new(Netease)
	n.lastfm = lastfm.NewClient()
	n.player = NewPlayer(n)
	n.login = NewLoginPage(n)
	n.search = NewSearchPage(n)
	n.App = app

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
	projectDir := app.DataRootDir()

	// 全局文件Jar
	cookieJar, _ := cookiejar.NewFileJar(filepath.Join(projectDir, "cookie"), nil)
	util.SetGlobalCookieJar(cookieJar)

	// DBManager初始化
	storage.DBManager = new(storage.LocalDBManager)

	// 获取用户信息
	errorx.Go(func() {
		table := storage.NewTable()

		// 获取用户信息
		if jsonStr, err := table.GetByKVModel(storage.User{}); err == nil {
			if user, err := structs.NewUserFromLocalJson(jsonStr); err == nil {
				n.user = &user
			}
		}
		// 刷新界面用户名
		n.MustMain().RefreshMenuTitle()

		// 获取lastfm用户信息
		var lastfmUser storage.LastfmUser
		if jsonStr, err := table.GetByKVModel(&lastfmUser); err == nil {
			if err = json.Unmarshal(jsonStr, &lastfmUser); err == nil {
				if lastfmUser.ApiKey == config.Lastfm.Key {
					n.lastfmUser = &lastfmUser
					n.lastfm.SetSession(lastfmUser.SessionKey)
				}
			}
		}
		n.MustMain().RefreshMenuList()

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

		// 获取播放歌曲信息
		if jsonStr, err := table.GetByKVModel(storage.PlayerSnapshot{}); err == nil && len(jsonStr) > 0 {
			var snapshot storage.PlayerSnapshot
			if err = json.Unmarshal(jsonStr, &snapshot); err == nil {
				p := n.player
                p.songManager.init(snapshot.CurSongIndex, snapshot.Playlist)
				p.playlistUpdateAt = snapshot.PlaylistUpdateAt
				p.playingMenuKey = "from_local_db" // 启动后，重置菜单Key，避免很多问题
			}
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
				_ = os.RemoveAll(filepath.Join(projectDir, "musicfox-notifier.app"))

				// 删除旧logo
				_ = os.Remove(filepath.Join(projectDir, types.DefaultNotifyIcon))

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
}

func (n *Netease) Player() *Player {
	return n.player
}
