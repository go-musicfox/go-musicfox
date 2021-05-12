package ui

import (
    "github.com/anhoder/bubbles/textinput"
    tea "github.com/anhoder/bubbletea"
    "github.com/anhoder/netease-music/service"
    "github.com/mattn/go-runewidth"
    "github.com/muesli/termenv"
    "go-musicfox/constants"
    "go-musicfox/db"
    "go-musicfox/ds"
    "go-musicfox/utils"
    "strings"
    "time"
)

type LoginModel struct {
    index         int
    accountInput  textinput.Model
    passwordInput textinput.Model
    submitButton  string
    tips          string
    AfterLogin    func (m *NeteaseModel)
}

func NewLogin() (login *LoginModel) {
    login = new(LoginModel)
    login.accountInput = textinput.NewModel()
    login.accountInput.Placeholder = " 手机号或邮箱"
    login.accountInput.Focus()
    login.accountInput.Prompt = focusedPrompt
    login.accountInput.TextColor = primaryColorStr
    login.accountInput.CharLimit = 32

    login.passwordInput = textinput.NewModel()
    login.passwordInput.Placeholder = " 密码"
    login.passwordInput.Prompt = "> "
    login.passwordInput.EchoMode = textinput.EchoPassword
    login.passwordInput.EchoCharacter = '•'
    login.passwordInput.CharLimit = 32

    login.submitButton = blurredSubmitButton

    return
}

// update main ui
func updateLogin(msg tea.Msg, m *NeteaseModel) (tea.Model, tea.Cmd) {

    switch msg := msg.(type) {
    case tickLoginMsg:
        return m, nil
    case tea.KeyMsg:
        switch msg.String() {

        case "esc":
            m.modelType = MtMain
            m.loginModel.tips = ""
            return m, tickMainUI(time.Nanosecond)

        // Cycle between inputs
        case "tab", "shift+tab", "enter", "up", "down":

            inputs := []textinput.Model{
                m.loginModel.accountInput,
                m.loginModel.passwordInput,
            }

            s := msg.String()

            // Did the user press enter while the submit button was focused?
            // If so, exit.
            if s == "enter" && m.loginModel.index == len(inputs) {
                if len(m.loginModel.accountInput.Value()) <= 0 || len(m.loginModel.passwordInput.Value()) <= 0 {
                    m.loginModel.tips = SetFgStyle("请输入账号或密码", termenv.ANSIBrightRed)
                    return m, nil
                }
                var (
                	code     float64
                	response []byte
                )
                if strings.ContainsRune(m.loginModel.accountInput.Value(), '@') {
                    loginService := service.LoginEmailService{
                        Email: m.loginModel.accountInput.Value(),
                        Password: m.loginModel.passwordInput.Value(),
                    }
                    code, response = loginService.LoginEmail()
                } else {
                    loginService := service.LoginCellphoneService{
                        Phone: m.loginModel.accountInput.Value(),
                        Password: m.loginModel.passwordInput.Value(),
                    }
                    code, response = loginService.LoginCellphone()
                }

                codeType := utils.CheckCode(code)
                switch codeType {
                case utils.UnknownError:
                    m.loginModel.tips = SetFgStyle("未知错误，请稍后再试~", termenv.ANSIBrightRed)
                    return m, tickLogin(time.Nanosecond)
                case utils.NetworkError:
                    m.loginModel.tips = SetFgStyle("网络异常，请稍后再试~", termenv.ANSIBrightRed)
                    return m, tickLogin(time.Nanosecond)
                case utils.Success:
                    if user, err := ds.NewUserFromJson(response); err == nil {
                        m.user = &user

                        // 写入本地数据库
                        table := db.NewTable()
                        _ = table.SetByKVModel(db.User{}, user)

                        if m.loginModel.AfterLogin != nil {
                            m.loginModel.AfterLogin(m)
                        }
                    }
                default:
                    m.loginModel.tips = SetFgStyle("你是个好人，但我们不合适(╬▔皿▔)凸 ", termenv.ANSIBrightRed) +
                        SetFgStyle("(账号或密码错误)", termenv.ANSIBrightBlack)
                    return m, tickLogin(time.Nanosecond)
                }

                m.modelType = MtMain
                m.loginModel.tips = ""
                return m, tickMainUI(time.Nanosecond)
            }

            // Cycle indexes
            if s == "up" || s == "shift+tab" {
                m.loginModel.index--
            } else {
                m.loginModel.index++
            }

            if m.loginModel.index > len(inputs) {
                m.loginModel.index = 0
            } else if m.loginModel.index < 0 {
                m.loginModel.index = len(inputs)
            }

            for i := 0; i <= len(inputs)-1; i++ {
                if i == m.loginModel.index {
                    // Set focused state
                    inputs[i].Focus()
                    inputs[i].Prompt = focusedPrompt
                    inputs[i].TextColor = primaryColorStr
                    continue
                }
                // Remove focused state
                inputs[i].Blur()
                inputs[i].Prompt = blurredPrompt
                inputs[i].TextColor = ""
            }

            m.loginModel.accountInput = inputs[0]
            m.loginModel.passwordInput = inputs[1]

            if m.loginModel.index == len(inputs) {
                m.loginModel.submitButton = focusedSubmitButton
            } else {
                m.loginModel.submitButton = blurredSubmitButton
            }

            return m, nil
        }
    }

    // Handle character input and blinks
    return updateLoginInputs(msg, m)
}

func updateLoginInputs(msg tea.Msg, m *NeteaseModel) (tea.Model, tea.Cmd) {
    var (
        cmd  tea.Cmd
        cmds []tea.Cmd
    )

    m.loginModel.accountInput, cmd = m.loginModel.accountInput.Update(msg)
    cmds = append(cmds, cmd)

    m.loginModel.passwordInput, cmd = m.loginModel.passwordInput.Update(msg)
    cmds = append(cmds, cmd)

    return m, tea.Batch(cmds...)
}

func loginView(m *NeteaseModel) string {

    var builder strings.Builder

    // 距离顶部的行数
    top := 0

    // title
    if constants.MainShowTitle {

        builder.WriteString(titleView(m, &top))
    } else {
        top++
    }

    // menu title
    builder.WriteString(menuTitleView(m, &top, "用户登录 (手机号或邮箱)"))
    builder.WriteString("\n\n\n")
    top += 2

    inputs := []textinput.Model{
        m.loginModel.accountInput,
        m.loginModel.passwordInput,
    }

    for i, input := range inputs {
        if m.menuStartColumn > 0 {
            builder.WriteString(strings.Repeat(" ", m.menuStartColumn))
        }

        builder.WriteString(input.View())

        var valueLen int
        if input.Value() == "" {
            valueLen = runewidth.StringWidth(input.Placeholder)
        } else {
            valueLen = runewidth.StringWidth(input.Value())
        }
        if spaceLen := m.WindowWidth-m.menuStartColumn-valueLen-3; spaceLen > 0 {
           builder.WriteString(strings.Repeat(" ", spaceLen))
        }

        top++

        if i < len(inputs)-1 {
            builder.WriteString("\n\n")
            top++
        }
    }

    builder.WriteString("\n\n")
    top++
    if m.menuStartColumn > 0 {
        builder.WriteString(strings.Repeat(" ", m.menuStartColumn))
    }
    builder.WriteString(m.loginModel.tips)
    builder.WriteString("\n\n")
    top++
    if m.menuStartColumn > 0 {
        builder.WriteString(strings.Repeat(" ", m.menuStartColumn))
    }
    builder.WriteString(m.loginModel.submitButton)
    builder.WriteString("\n")

    if m.WindowHeight > top+3 {
        builder.WriteString(strings.Repeat("\n", m.WindowHeight-top-3))
    }

    return builder.String()
}

// NeedLoginHandle 需要登录的处理
func NeedLoginHandle(model *NeteaseModel, callback func (m *NeteaseModel)) {
    model.modelType = MtLogin
    model.loginModel.AfterLogin = callback
}