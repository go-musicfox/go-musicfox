package ui

import (
    "encoding/json"
    tea "github.com/anhoder/bubbletea"
    "github.com/anhoder/netease-music/service"
    "github.com/anhoder/netease-music/util"
    "github.com/mattn/go-runewidth"
    "github.com/telanflow/cookiejar"
    "go-musicfox/configs"
    "go-musicfox/pkg/constants"
    db2 "go-musicfox/pkg/storage"
    "go-musicfox/pkg/structs"
    "go-musicfox/utils"
    "strconv"
    "time"
)

type NeteaseModel struct {
    WindowWidth    int
    WindowHeight   int
    isListeningKey bool
    program        *tea.Program
    user           *structs.User

    // startup
    *startupModel

    // main ui
    *mainUIModel

    // login
    loginModel *LoginModel

    // search
    searchModel *SearchModel
}

// NewNeteaseModel get netease model
func NewNeteaseModel(loadingDuration time.Duration) (m *NeteaseModel) {
    m = new(NeteaseModel)
    m.isListeningKey = !configs.ConfigRegistry.StartupShow

    // startup
    m.startupModel = NewStartup()
    m.TotalDuration = loadingDuration

    // main menu
    m.mainUIModel = NewMainUIModel(m)

    // login
    m.loginModel = NewLogin()

    // search
    m.searchModel = NewSearch()

    // 东亚
    runewidth.EastAsianWidth = false

    return
}

func (m *NeteaseModel) Init() tea.Cmd {

    projectDir := utils.GetLocalDataDir()

    // 全局文件Jar
    cookieJar, _ := cookiejar.NewFileJar(projectDir+"/cookie", nil)
    util.SetGlobalCookieJar(cookieJar)

    // DBManager初始化
    db2.DBManager = new(db2.LocalDBManager)

    // 获取用户信息
    go func() {
        table := db2.NewTable()

        // 获取用户信息
        if jsonStr, err := table.GetByKVModel(db2.User{}); err == nil {
            if user, err := structs.NewUserFromLocalJson(jsonStr); err == nil {
                m.user = &user
            }
        }

        // 获取播放模式
        if jsonStr, err := table.GetByKVModel(db2.PlayMode{}); err == nil && len(jsonStr) > 0 {
            var playMode PlayMode
            if err = json.Unmarshal(jsonStr, &playMode); err == nil {
                m.player.mode = playMode
            }
        }

        // 获取播放歌曲信息
        if jsonStr, err := table.GetByKVModel(db2.PlayerSnapshot{}); err == nil && len(jsonStr) > 0 {
            var snapshot db2.PlayerSnapshot
            if err = json.Unmarshal(jsonStr, &snapshot); err == nil {
                m.player.curSongIndex = snapshot.CurSongIndex
                m.player.playlist = snapshot.Playlist
                m.player.curSong = m.player.playlist[m.player.curSongIndex]
                //m.player.playingMenuKey = snapshot.PlayingMenuKey
                m.player.playingMenuKey = "from_local_db" // 启动后，重置菜单Key，避免很多问题
            }
        }

        // 签到
        if configs.ConfigRegistry.StartupSignIn {
            var lastSignIn int
            if jsonStr, err := table.GetByKVModel(db2.LastSignIn{}); err == nil && len(jsonStr) > 0 {
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

                _ = table.SetByKVModel(db2.LastSignIn{}, today)

                utils.Notify("签到成功", "今日手机、PC端签到成功~", constants.AppGithubUrl)
            }
        }

        // 检查更新
        if configs.ConfigRegistry.StartupCheckUpdate && utils.CheckUpdate() {
            utils.Notify("发现新版本", "点击去看看吧~", constants.AppLatestReleases)
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
        // 登录界面输入q不退出
        if m.modelType == MtMain && (k == "q" || k == "Q" || k == "ctrl+c") {
            m.quitting = true
            return m, tea.Quit
        }

    case tea.WindowSizeMsg:
        m.WindowHeight = msgWithType.Height
        m.WindowWidth = msgWithType.Width

    }

    // Hand off the message and model to the approprate update function for the
    // appropriate view based on the current state.
    if configs.ConfigRegistry.StartupShow && !m.loaded {
        if _, ok := msg.(tea.WindowSizeMsg); ok {
            updateMainUI(msg, m)
        }
        return updateStartup(msg, m)
    }

    switch m.modelType {
    case MtLogin:
        return updateLogin(msg, m)
    case MtSearch:
        return updateSearch(msg, m)
    }

    return updateMainUI(msg, m)
}

func (m *NeteaseModel) View() string {
    if m.quitting || m.WindowWidth <= 0 || m.WindowHeight <= 0 {
        return ""
    }

    if configs.ConfigRegistry.StartupShow && !m.loaded {
        return startupView(m)
    }

    switch m.modelType {
    case MtLogin:
        return loginView(m)
    case MtSearch:
        return searchView(m)
    }

    return mainUIView(m)
}

func (m *NeteaseModel) BindProgram(program *tea.Program) {
    m.program = program
}

func (m *NeteaseModel) Rerender() {
    if m.program != nil {
        m.program.Rerender(m.View())
    }
}
