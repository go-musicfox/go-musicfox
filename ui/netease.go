package ui

import (
    tea "github.com/anhoder/bubbletea"
    "github.com/anhoder/netease-music/util"
    "github.com/telanflow/cookiejar"
    "go-musicfox/constants"
    "go-musicfox/utils"
    "os"
    "time"
)

type NeteaseModel struct {
    WindowWidth    int
    WindowHeight   int
    isListeningKey bool

    // startup
    *startupModel

    // main ui
    *mainUIModel
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

    return
}

func (m *NeteaseModel) Init() tea.Cmd {
    if constants.AppShowStartup {
        return tickStartup(time.Nanosecond)
    }

    // Home目录
    homeDir, err := utils.Home()
    if nil != err {
        panic("未获取到用户Home目录: " + err.Error())
    }
    projectDir := homeDir + "/.go-musicfox"

    if _, err := os.Stat(projectDir); os.IsNotExist(err) {
        _ = os.Mkdir(projectDir, os.ModePerm)
    }

    // 全局文件Jar
    cookieJar, _ := cookiejar.NewFileJar(projectDir + "/cookie", nil)
    util.SetGlobalCookieJar(cookieJar)

    return tickMainUI(time.Nanosecond)
}

func (m *NeteaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Make sure these keys always quit
    switch msgWithType := msg.(type) {
    case tea.KeyMsg:
        k := msgWithType.String()
        if k == "q" || k == "ctrl+c" {
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

    return updateMainUI(msg, m)
}

func (m *NeteaseModel) View() string {
    if m.quitting || m.WindowWidth <= 0 || m.WindowHeight <= 0 {
        return ""
    }

    if constants.AppShowStartup && !m.loaded {
        return startupView(m)
    }

    return mainUIView(m)
}

