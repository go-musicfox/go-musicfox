package ui

import (
    tea "github.com/anhoder/bubbletea"
    "github.com/anhoder/go-musicfox/constants"
    "strings"
    "time"
)

type NeteaseModel struct {
    WindowWidth    int
    WindowHeight   int
    isListeningKey bool
    clearScreen    bool

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

    case tickClearScreenMsg:
        m.clearScreen = true;

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

    if m.clearScreen {
        m.clearScreen = false
        return strings.Repeat(strings.Repeat(" ", m.WindowWidth), m.WindowHeight)
    }

    if constants.AppShowStartup && !m.loaded {
        return startupView(m)
    }

    return mainUIView(m)
}

