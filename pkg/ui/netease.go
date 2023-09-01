package ui

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/pkg/lastfm"
	"github.com/go-musicfox/go-musicfox/pkg/player"
	"github.com/go-musicfox/go-musicfox/pkg/storage"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"
	"github.com/go-musicfox/go-musicfox/utils/like_list"

	"github.com/go-musicfox/netease-music/service"
	"github.com/go-musicfox/netease-music/util"
	"github.com/telanflow/cookiejar"
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
	projectDir := utils.GetLocalDataDir()

	// 全局文件Jar
	cookieJar, _ := cookiejar.NewFileJar(path.Join(projectDir, "cookie"), nil)
	util.SetGlobalCookieJar(cookieJar)

	// DBManager初始化
	storage.DBManager = new(storage.LocalDBManager)

	// 获取用户信息
	go utils.PanicRecoverWrapper(false, func() {
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
				n.lastfmUser = &lastfmUser
				n.lastfm.SetSession(lastfmUser.SessionKey)
			}
		}
		n.MustMain().RefreshMenuList()

		// 获取播放模式
		if jsonStr, err := table.GetByKVModel(storage.PlayMode{}); err == nil && len(jsonStr) > 0 {
			var playMode player.Mode
			if err = json.Unmarshal(jsonStr, &playMode); err == nil {
				n.player.mode = playMode
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
				p.curSongIndex = snapshot.CurSongIndex
				p.playlist = snapshot.Playlist
				p.playlistUpdateAt = snapshot.PlaylistUpdateAt
				p.curSong = p.playlist[p.curSongIndex]
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
				if err := json.Unmarshal(jsonStr, &extInfo); err == nil && utils.CompareVersion(extInfo.StorageVersion, constants.AppVersion, true) {
					needUpdate = false
				}
			}
			if needUpdate {
				localDir := utils.GetLocalDataDir()

				// 删除旧notifier
				_ = os.RemoveAll(path.Join(localDir, "musicfox-notifier.app"))

				// 删除旧logo
				_ = os.Remove(path.Join(localDir, constants.DefaultNotifyIcon))

				extInfo.StorageVersion = constants.AppVersion
				_ = table.SetByKVModel(extInfo, extInfo)
			}
		}

		// 刷新like list
		if n.user != nil {
			like_list.RefreshLikeList(n.user.UserId)
			n.Rerender(false)
		}

		// 签到
		if config.StartupSignIn {
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

				utils.Notify(utils.NotifyContent{
					Title:   "签到成功",
					Text:    "今日手机、PC端签到成功",
					Url:     constants.AppGithubUrl,
					GroupId: constants.GroupID,
				})
			}
		}

		// 刷新登录状态
		if n.user != nil {
			refreshLoginService := service.LoginRefreshService{}
			refreshLoginService.LoginRefresh()
		}

		// 检查更新
		if config.StartupCheckUpdate {
			if ok, newVersion := utils.CheckUpdate(); ok {
				if runtime.GOOS == "windows" {
					n.MustMain().EnterMenu(
						NewCheckUpdateMenu(newBaseMenu(n)),
						&model.MenuItem{Title: "新版本: " + newVersion, Subtitle: "当前版本: " + constants.AppVersion},
					)
				}

				utils.Notify(utils.NotifyContent{
					Title: "发现新版本: " + newVersion,
					Text:  "去看看呗",
					Url:   constants.AppLatestReleases,
				})
			}
		}

		// 自动播放
		// TODO optimize
		if config.AutoPlay {
			var (
				notice   string // 错误通知文本
				index    int    // 歌曲索引
				length   int    // 歌单长度（用于获取歌曲索引）
				getAll   bool   // 是否需要获取全部歌曲
				playlist []structs.Song
				playmode = map[string]player.Mode{
					"listLoop":    player.PmListLoop,
					"order":       player.PmOrder,
					"singleLoop":  player.PmSingleLoop,
					"random":      player.PmRandom,
					"intelligent": player.PmIntelligent,
					"last":        n.player.mode,
				}
			)

			if utils.CheckUserInfo(n.user) == utils.NeedLogin {
				notice = "账号未登录"
				goto Complete
			}
			if config.AutoPlayOffset >= 1000 || config.AutoPlayOffset < 0 {
				getAll = true
			}
			if mode, ok := playmode[config.AutoPlayMode]; ok {
				n.player.mode = mode
			} else {
				notice = fmt.Sprintf("无效的播放模式：%s", config.AutoPlayMode)
				goto Complete
			}
			switch config.AutoPlayList {
			case "dailyReco":
				playlist, notice = getDailySongs()
			case "like":
				playlist, notice = getLikeSongs(n.user.UserId, getAll)
			case "no":
				playlist = n.player.playlist
			default: // name:xxx
				if !strings.HasPrefix(config.AutoPlayList, "name:") {
					notice = fmt.Sprintf("歌单格式错误：%s", config.AutoPlayList)
					goto Complete
				}
				name := config.AutoPlayList[5:]
				playlist, notice = getPlaylistByName(n.user.UserId, name, getAll)
			}
			if notice != "" {
				goto Complete
			}
			length = len(playlist)
			if config.AutoPlayList == "no" {
				// 保持原来状态
				index = n.player.curSongIndex
			} else if n.player.mode != player.PmRandom {
				if config.AutoPlayOffset >= length || -config.AutoPlayOffset > length {
					notice = fmt.Sprintf("无效的偏移量：%d", config.AutoPlayOffset)
					goto Complete
				} else {
					index = (config.AutoPlayOffset + length) % length // 无论offset正负都能工作
				}
			} else {
				// 随机播放
				index = rand.Intn(length)
			}
			n.player.playlist = playlist
			n.player.curSongIndex = index
			_ = n.player.PlaySong(n.player.playlist[index], DurationNext)
		Complete:
			if notice != "" {
				utils.Notify(utils.NotifyContent{
					Title: "自动播放失败",
					Text:  notice,
				})
			}
		}
	})
}

func (n *Netease) CloseHook(_ *model.App) {
	n.player.Close()
}

func (n *Netease) Player() *Player {
	return n.player
}

// TODO optimize
func getDailySongs() (playlist []structs.Song, notice string) {
	recommendSongs := service.RecommendSongsService{}
	code, response := recommendSongs.RecommendSongs()
	codeType := utils.CheckCode(code)
	if codeType != utils.Success {
		notice = "网络错误"
		return
	}
	playlist = utils.GetDailySongs(response)
	return
}

// TODO optimize
func getLikeSongs(userId int64, getAll bool) (playlist []structs.Song, notice string) {
	var (
		codeType  utils.ResCode
		playlists []structs.Playlist
		songs     []structs.Song
	)
	codeType, playlists, _ = getUserPlaylists(userId, 1, 0)
	if codeType != utils.Success {
		notice = "网络错误"
		return
	}
	codeType, songs = getSongsInPlaylist(playlists[0].Id, getAll)
	if codeType != utils.Success {
		notice = "网络错误"
		return
	}
	playlist = songs
	return
}

// TODO optimize
func getPlaylistByName(userId int64, playlistName string, getAll bool) (playlist []structs.Song, notice string) {
	var (
		playlistId int64
		offset     = 0
		codeType   utils.ResCode
		playlists  []structs.Playlist
		hasMore    bool
	)
	// 寻找歌单
Loop:
	for {
		codeType, playlists, hasMore = getUserPlaylists(userId, 30, offset)
		if codeType != utils.Success {
			notice = "网络错误"
			return
		}
		offset += len(playlists)
		for _, playlist := range playlists {
			if playlist.Name == playlistName {
				playlistId = playlist.Id
				break Loop
			}
		}
		if !hasMore {
			notice = fmt.Sprintf("未找到歌单：%s", playlistName)
			return
		}
	}
	codeType, songs := getSongsInPlaylist(playlistId, getAll)
	if codeType != utils.Success {
		notice = "网络错误"
		return
	}
	playlist = songs
	return
}
