package ui

import (
    tea "github.com/anhoder/bubbletea"
    "github.com/anhoder/netease-music/util"
    "github.com/telanflow/cookiejar"
    "go-musicfox/constants"
    "go-musicfox/db"
    "go-musicfox/ds"
    "go-musicfox/utils"
    "time"
)

type NeteaseModel struct {
    WindowWidth    int
    WindowHeight   int
    isListeningKey bool
    program        *tea.Program
    user           *ds.User

    // startup
    *startupModel

    // main ui
    *mainUIModel

    // login
    *LoginModel
}

// NewNeteaseModel get netease model
func NewNeteaseModel(loadingDuration time.Duration) (m *NeteaseModel) {
    m = new(NeteaseModel)
    m.isListeningKey = !constants.AppShowStartup

    // startup
    m.startupModel = NewStartup()
    m.TotalDuration = loadingDuration

    // main menu
    m.mainUIModel = NewMainUIModel(m)

    // login
    m.LoginModel = NewLogin()

    return
}

func (m *NeteaseModel) Init() tea.Cmd {

    projectDir := utils.GetLocalDataDir()

    // 全局文件Jar
    cookieJar, _ := cookiejar.NewFileJar(projectDir + "/cookie", nil)
    util.SetGlobalCookieJar(cookieJar)

    // DBManager初始化
    db.DBManager = &db.LocalDBManager{}

    // 获取用户信息
    table := db.NewTable()
    if json, err := table.GetByKVModel(db.User{}); err == nil {
        if user, err := ds.NewUserFromLocalJson(json); err == nil {
            m.user = &user
        }
    }

    if constants.AppShowStartup {
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
        if !m.showLogin && (k == "q" || k == "ctrl+c") {
            m.quitting = true
            return m, tea.Quit
        }

    case tea.WindowSizeMsg:
        m.WindowHeight = msgWithType.Height
        m.WindowWidth  = msgWithType.Width

    }

    // Hand off the message and model to the approprate update function for the
    // appropriate view based on the current state.
    if constants.AppShowStartup && !m.loaded {
        if _, ok := msg.(tea.WindowSizeMsg); ok {
            updateMainUI(msg, m)
        }
        return updateStartup(msg, m)
    }

    if m.showLogin {
        return updateLogin(msg, m)
    }

    return updateMainUI(msg, m)
}

func (m *NeteaseModel) View() string {
    if m.quitting || m.WindowWidth <= 0 || m.WindowHeight <= 0 {
        return ""
    }

    if constants.AppShowStartup && !m.loaded {
        return startupView(m)
    }

    if m.showLogin {
        return loginView(m)
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