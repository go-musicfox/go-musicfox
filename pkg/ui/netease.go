package ui

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/pkg/lastfm"
	"github.com/go-musicfox/go-musicfox/pkg/player"
	"github.com/go-musicfox/go-musicfox/pkg/storage"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"
	"github.com/go-musicfox/go-musicfox/utils/like_list"

	"github.com/anhoder/netease-music/service"
	"github.com/anhoder/netease-music/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/telanflow/cookiejar"
)

type NeteaseModel struct {
	WindowWidth    int
	WindowHeight   int
	isListeningKey bool
	program        *tea.Program
	user           *structs.User
	lastfm         *lastfm.Client
	lastfmUser     *storage.LastfmUser

	// startup
	startup *StartupModel

	// main ui
	*MainUIModel

	// login
	loginModel *LoginModel

	// search
	searchModel *SearchModel
}

// NewNeteaseModel get netease model
func NewNeteaseModel(loadingDuration time.Duration) (m *NeteaseModel) {
	m = new(NeteaseModel)
	m.isListeningKey = !configs.ConfigRegistry.StartupShow
	m.lastfm = lastfm.NewClient()

	// startup
	m.startup = NewStartup()
	m.startup.TotalDuration = loadingDuration

	// main menu
	m.MainUIModel = NewMainUIModel(m)

	// login
	m.loginModel = NewLogin()

	// search
	m.searchModel = NewSearch()

	// 东亚
	runewidth.DefaultCondition.EastAsianWidth = false

	return
}

func (m *NeteaseModel) Init() tea.Cmd {
	projectDir := utils.GetLocalDataDir()

	// 全局文件Jar
	cookieJar, _ := cookiejar.NewFileJar(path.Join(projectDir, "cookie"), nil)
	util.SetGlobalCookieJar(cookieJar)

	// DBManager初始化
	storage.DBManager = new(storage.LocalDBManager)

	// 获取用户信息
	go func() {
		defer utils.Recover(false)

		table := storage.NewTable()

		// 获取用户信息
		if jsonStr, err := table.GetByKVModel(storage.User{}); err == nil {
			if user, err := structs.NewUserFromLocalJson(jsonStr); err == nil {
				m.user = &user
			}
		}
		m.refreshMenuTitle() // 刷新界面用户名

		// 获取lastfm用户信息
		var lastfmUser storage.LastfmUser
		if jsonStr, err := table.GetByKVModel(&lastfmUser); err == nil {
			if err = json.Unmarshal(jsonStr, &lastfmUser); err == nil {
				m.lastfmUser = &lastfmUser
				m.lastfm.SetSession(lastfmUser.SessionKey)
			}
		}
		m.refreshMenuList()

		// 获取播放模式
		if jsonStr, err := table.GetByKVModel(storage.PlayMode{}); err == nil && len(jsonStr) > 0 {
			var playMode player.Mode
			if err = json.Unmarshal(jsonStr, &playMode); err == nil {
				m.player.mode = playMode
			}
		}

		// 获取音量
		if jsonStr, err := table.GetByKVModel(storage.Volume{}); err == nil && len(jsonStr) > 0 {
			var volume int
			if err = json.Unmarshal(jsonStr, &volume); err == nil {
				v, ok := m.player.Player.(storage.VolumeStorable)
				if ok {
					v.SetVolume(volume)
				}
			}
		}

		// 获取播放歌曲信息
		if jsonStr, err := table.GetByKVModel(storage.PlayerSnapshot{}); err == nil && len(jsonStr) > 0 {
			var snapshot storage.PlayerSnapshot
			if err = json.Unmarshal(jsonStr, &snapshot); err == nil {
				p := m.player
				p.curSongIndex = snapshot.CurSongIndex
				p.playlist = snapshot.Playlist
				p.playlistUpdateAt = snapshot.PlaylistUpdateAt
				p.curSong = p.playlist[p.curSongIndex]
				p.playingMenuKey = "from_local_db" // 启动后，重置菜单Key，避免很多问题
			}
		}
		m.Rerender(false)

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
		if m.user != nil {
			like_list.RefreshLikeList(m.user.UserId)
			m.Rerender(false)
		}

		// 签到
		if configs.ConfigRegistry.StartupSignIn {
			var lastSignIn int
			if jsonStr, err := table.GetByKVModel(storage.LastSignIn{}); err == nil && len(jsonStr) > 0 {
				_ = json.Unmarshal(jsonStr, &lastSignIn)
			}
			today, err := strconv.Atoi(time.Now().Format("20060102"))
			if m.user != nil && err == nil && lastSignIn != today {
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
		if m.user != nil {
			refreshLoginService := service.LoginRefreshService{}
			refreshLoginService.LoginRefresh()
		}

		// 检查更新
		if configs.ConfigRegistry.StartupCheckUpdate && utils.CheckUpdate() {
			utils.Notify(utils.NotifyContent{
				Title: "发现新版本",
				Text:  "去看看呗",
				Url:   constants.AppLatestReleases,
			})
		}

		// 自动播放
		if configs.ConfigRegistry.AutoPlay {
			var notice string // 错误通知文本
			var index int     // 歌曲索引
			var length int    // 歌单长度（用于获取歌曲索引）
			var allFlag bool  // 用于标记是否需要获取全部歌曲的flag
			if utils.CheckUserInfo(m.user) == utils.NeedLogin {
				notice = "账号未登录"
				goto Complete
			}
			if configs.ConfigRegistry.AutoPlayOffset >= 0 && configs.ConfigRegistry.AutoPlayOffset >= 1000 || configs.ConfigRegistry.AutoPlayOffset < 0 {
				allFlag = true
			}
			switch configs.ConfigRegistry.AutoPlayList {
			case "dailyReco":
				recommendSongs := service.RecommendSongsService{}
				code, response := recommendSongs.RecommendSongs()
				codeType := utils.CheckCode(code)
				if codeType != utils.Success {
					notice = "网络错误"
					goto Complete
				}
				m.player.playlist = utils.GetDailySongs(response)
			case "like":
				userPlaylists := service.UserPlaylistService{
					Uid:    strconv.FormatInt(m.user.UserId, 10),
					Limit:  "1",
					Offset: "0",
				}
				code, response := userPlaylists.UserPlaylist()
				codeType := utils.CheckCode(code)
				if codeType != utils.Success {
					notice = "网络错误"
					goto Complete
				}
				playlistId := utils.GetPlaylists(response)[0].Id
				if !allFlag {
					playlistDetail := service.PlaylistDetailService{Id: strconv.FormatInt(playlistId, 10), S: "0"} // 最近S个收藏者，设为0
					code, response = playlistDetail.PlaylistDetail()
				} else {
					allTrack := service.PlaylistTrackAllService{Id: strconv.FormatInt(playlistId, 10), S: "0"} // 最近S个收藏者，设为0
					code, response = allTrack.AllTracks()
				}
				codeType = utils.CheckCode(code)
				if codeType != utils.Success {
					notice = "网络错误"
					goto Complete
				}
				m.player.playlist = utils.GetSongsOfPlaylist(response)
			default:
				if !strings.HasPrefix(configs.ConfigRegistry.AutoPlayList, "name:") {
					notice = fmt.Sprintf("歌单格式错误：%s", configs.ConfigRegistry.AutoPlayList)
					goto Complete
				}
				name := configs.ConfigRegistry.AutoPlayList[5:]
				var playlistId int64
				hasMore := true
				offset := 0
				for hasMore {
					userPlaylists := service.UserPlaylistService{
						Uid:    strconv.FormatInt(m.user.UserId, 10),
						Limit:  "30",
						Offset: strconv.Itoa(offset),
					}
					code, response := userPlaylists.UserPlaylist()
					codeType := utils.CheckCode(code)
					if codeType != utils.Success {
						notice = "网络错误"
						goto Complete
					}
					list := utils.GetPlaylists(response)
					offset += len(list)
					for _, playlist := range list {
						if playlist.Name == name {
							playlistId = playlist.Id
							goto Ok
						}
					}
				}
				notice = fmt.Sprintf("未找到歌单：%s", name)
				goto Complete
			Ok:
				var code float64
				var response []byte
				if !allFlag {
					playlistDetail := service.PlaylistDetailService{Id: strconv.FormatInt(playlistId, 10), S: "0"} // 最近S个收藏者，设为0
					code, response = playlistDetail.PlaylistDetail()
				} else {
					allTrack := service.PlaylistTrackAllService{Id: strconv.FormatInt(playlistId, 10), S: "0"} // 最近S个收藏者，设为0
					code, response = allTrack.AllTracks()
				}
				codeType := utils.CheckCode(code)
				if codeType != utils.Success {
					notice = "网络错误"
					goto Complete
				}
				m.player.playlist = utils.GetSongsOfPlaylist(response)
			}
			length = len(m.player.playlist)
			if !configs.ConfigRegistry.AutoPlayRandom {
				if configs.ConfigRegistry.AutoPlayOffset >= 0 {
					if configs.ConfigRegistry.AutoPlayOffset >= length {
						notice = fmt.Sprintf("无效的偏移量：%d", configs.ConfigRegistry.AutoPlayOffset)
						goto Complete
					} else {
						index = configs.ConfigRegistry.AutoPlayOffset
					}
				} else {
					if -(configs.ConfigRegistry.AutoPlayOffset) > length {
						notice = fmt.Sprintf("无效的偏移量：%d", configs.ConfigRegistry.AutoPlayOffset)
						goto Complete
					} else {
						index = (configs.ConfigRegistry.AutoPlayOffset + length) % length
					}
				}
			} else {
				rand.Seed(time.Now().UnixMicro())
				index = rand.Intn(length)
			}
			m.player.PlaySong(m.player.playlist[index], DurationNext)
		Complete:
			if notice != "" {
				utils.Notify(utils.NotifyContent{
					Title: "自动播放失败",
					Text:  notice,
				})
			}
		}

	}()

	if configs.ConfigRegistry.StartupShow {
		return tickStartup(time.Nanosecond)
	}

	return tickMainUI(time.Nanosecond)
}

func (m *NeteaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	switch msgWithType := msg.(type) {
	case tea.KeyMsg:
		k := msgWithType.String()
		// 账号登录或搜索页输入q不退出
		if m.pageType == PtMain && !m.inSearching && (k == "q" || k == "Q" || k == "ctrl+c") {
			m.startup.quitting = true
			m.Close()
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.WindowHeight = msgWithType.Height
		m.WindowWidth = msgWithType.Width
	}

	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.MainUIModel.update(msg, m)
	}

	// Hand off the message and model to the approprate update function for the
	// appropriate view based on the current state.
	if configs.ConfigRegistry.StartupShow && !m.startup.loaded {
		return m.startup.update(msg, m)
	}

	switch m.pageType {
	case PtLogin:
		return updateLogin(msg, m)
	case PtSearch:
		return m.searchModel.update(msg, m)
	}

	return m.MainUIModel.update(msg, m)
}

func (m *NeteaseModel) View() string {
	if m.startup.quitting || m.WindowWidth <= 0 || m.WindowHeight <= 0 {
		return ""
	}

	if configs.ConfigRegistry.StartupShow && !m.startup.loaded {
		return m.startup.view(m)
	}

	switch m.pageType {
	case PtLogin:
		return loginView(m)
	case PtSearch:
		return m.searchModel.view(m)
	}

	return m.MainUIModel.view(m)
}

func (m *NeteaseModel) BindProgram(program *tea.Program) {
	m.program = program
}

func (m *NeteaseModel) Rerender(cleanScreen bool) {
	if m.program == nil {
		return
	}
	ticker := m.rerenderTicker(cleanScreen)
	m.program.Send(ticker())
}

func (m *NeteaseModel) rerenderTicker(cleanScreen bool) tea.Cmd {
	return func() tea.Msg {
		if cleanScreen {
			m.program.Send(tea.ClearScreen())
		}
		return MsgOfPageType(m.pageType)
	}
}

func (m *NeteaseModel) Close() {
	m.MainUIModel.Close()
}
