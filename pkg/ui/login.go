package ui

import (
	"log"
	"strconv"
	"strings"
	"time"

	"go-musicfox/pkg/configs"
	"go-musicfox/pkg/storage"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"

	"github.com/anhoder/bubbles/textinput"
	tea "github.com/anhoder/bubbletea"
	"github.com/anhoder/netease-music/service"
	"github.com/buger/jsonparser"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
	"github.com/pkg/errors"
	"github.com/skratchdot/open-golang/open"
)

type LoginModel struct {
	index         int
	accountInput  textinput.Model
	passwordInput textinput.Model
	submitButton  string
	qrLoginButton string
	qrLoginStep   int
	qrLoginUniKey string
	tips          string
	AfterLogin    func(m *NeteaseModel, newMenu IMenu, newTitle *MenuItem)
}

func NewLogin() (login *LoginModel) {
	login = new(LoginModel)
	login.accountInput = textinput.NewModel()
	login.accountInput.Placeholder = " 手机号或邮箱"
	login.accountInput.Focus()
	login.accountInput.Prompt = GetFocusedPrompt()
	login.accountInput.TextColor = primaryColorStr
	login.accountInput.CharLimit = 32

	login.passwordInput = textinput.NewModel()
	login.passwordInput.Placeholder = " 密码"
	login.passwordInput.Prompt = "> "
	login.passwordInput.EchoMode = textinput.EchoPassword
	login.passwordInput.EchoCharacter = '•'
	login.passwordInput.CharLimit = 32

	login.submitButton = GetBlurredSubmitButton()
	login.qrLoginButton = GetBlurredButton(login.qrButtonTextByStep())

	return
}

// update main ui
func updateLogin(msg tea.Msg, m *NeteaseModel) (tea.Model, tea.Cmd) {
	inputs := []textinput.Model{
		m.loginModel.accountInput,
		m.loginModel.passwordInput,
	}
	submitIndex := len(inputs)
	qrLoginIndex := submitIndex + 1

	switch msg := msg.(type) {
	case tickLoginMsg:
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "b":
			if m.loginModel.index != submitIndex && m.loginModel.index != qrLoginIndex {
				break
			}
			fallthrough
		case "esc":
			m.pageType = PtMain
			m.loginModel.tips = ""
			m.loginModel.qrLoginStep = 0
			if m.loginModel.index == qrLoginIndex {
				m.loginModel.qrLoginButton = GetFocusedButton(m.loginModel.qrButtonTextByStep())
			} else {
				m.loginModel.qrLoginButton = GetBlurredButton(m.loginModel.qrButtonTextByStep())
			}

		case "tab", "shift+tab", "enter", "up", "down", "left", "right":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.loginModel.index >= submitIndex {
				switch m.loginModel.index {
				case submitIndex:
					// 提交
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
							Email:    m.loginModel.accountInput.Value(),
							Password: m.loginModel.passwordInput.Value(),
						}
						code, response = loginService.LoginEmail()
					} else {
						var (
							phone       = m.loginModel.accountInput.Value()
							countryCode = "86"
						)
						if strings.HasPrefix(phone, "+") && strings.ContainsRune(phone, ' ') {
							if items := strings.Split(phone, " "); len(items) == 2 {
								countryCode, phone = strings.TrimLeft(items[0], "+"), items[1]
							}
						}
						loginService := service.LoginCellphoneService{
							Phone:       phone,
							Password:    m.loginModel.passwordInput.Value(),
							Countrycode: countryCode,
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
						m.loginModel.loginSuccessHandle(m, response)
					default:
						m.loginModel.tips = SetFgStyle("你是个好人，但我们不合适(╬▔皿▔)凸 ", termenv.ANSIBrightRed) +
							SetFgStyle("(账号或密码错误)", termenv.ANSIBrightBlack)
						return m, tickLogin(time.Nanosecond)
					}
				case qrLoginIndex:
					// 扫码登录
					qrService := service.LoginQRService{}
					if m.loginModel.qrLoginStep == 0 {
						code, resp, url := qrService.GetKey()
						errHandler := func(err error) (tea.Model, tea.Cmd) {
							m.loginModel.tips = SetFgStyle("生成二维码失败，请稍候再试", termenv.ANSIBrightRed)
							if err != nil {
								utils.Logger().Printf("生成二维码失败, +v%", err)
							}
							return m, nil
						}
						if code != 200 || url == "" {
							return errHandler(errors.Errorf("code: %f, resp: %s", code, string(resp)))
						}
						path, err := utils.GenQRCode("qrcode.png", url)
						if err != nil {
							return errHandler(err)
						}
						_ = open.Start(path)
						m.loginModel.tips = SetFgStyle("请扫描二维码(MUSICFOX_ROOT/qrcode.png)登录后，点击「继续」", termenv.ANSIBrightRed)
						m.loginModel.qrLoginStep++
						m.loginModel.qrLoginButton = GetFocusedButton(m.loginModel.qrButtonTextByStep())
						m.loginModel.qrLoginUniKey = qrService.UniKey
						return m, nil
					}
					qrService.UniKey = m.loginModel.qrLoginUniKey
					errHandler := func(err error) (tea.Model, tea.Cmd) {
						m.loginModel.tips = SetFgStyle("校验二维码失败，请稍候再试", termenv.ANSIBrightRed)
						if err != nil {
							utils.Logger().Printf("生成二维码失败 +v%", err)
						}
						return m, nil
					}
					if code, resp := qrService.CheckQR(); code != 803 {
						return errHandler(errors.Errorf("checkQR code: %f, resp: %s", code, string(resp)))
					}
					if code, resp := (&service.UserAccountService{}).AccountInfo(); code != 200 {
						return errHandler(errors.Errorf("accountInfo code: %f, resp: %s", code, string(resp)))
					} else {
						m.loginModel.loginSuccessHandle(m, resp)
					}
				}

				m.pageType = PtMain
				m.loginModel.tips = ""
				return m, tickMainUI(time.Nanosecond)
			}

			// 当focus在button上时，左右按键的特殊处理
			if s == "left" || s == "right" {
				if m.loginModel.index < submitIndex {
					return updateLoginInputs(msg, m)
				}
				if s == "left" && m.loginModel.index == qrLoginIndex {
					m.loginModel.index--
				} else if s == "right" && m.loginModel.index == submitIndex {
					m.loginModel.index++
				}
			} else if s == "up" || s == "shift+tab" {
				m.loginModel.index--
			} else {
				m.loginModel.index++
			}

			if m.loginModel.index > qrLoginIndex {
				m.loginModel.index = 0
			} else if m.loginModel.index < 0 {
				m.loginModel.index = qrLoginIndex
			}

			for i := 0; i <= len(inputs)-1; i++ {
				if i != m.loginModel.index {
					// Remove focused state
					inputs[i].Blur()
					inputs[i].Prompt = GetBlurredPrompt()
					inputs[i].TextColor = ""
					continue
				}
				// Set focused state
				inputs[i].Focus()
				inputs[i].Prompt = GetFocusedPrompt()
				inputs[i].TextColor = primaryColorStr
			}

			m.loginModel.accountInput = inputs[0]
			m.loginModel.passwordInput = inputs[1]

			if m.loginModel.index == submitIndex {
				m.loginModel.submitButton = GetFocusedSubmitButton()
			} else {
				m.loginModel.submitButton = GetBlurredSubmitButton()
			}

			if m.loginModel.index == qrLoginIndex {
				m.loginModel.qrLoginButton = GetFocusedButton(m.loginModel.qrButtonTextByStep())
			} else {
				m.loginModel.qrLoginButton = GetBlurredButton(m.loginModel.qrButtonTextByStep())
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
	if configs.ConfigRegistry.MainShowTitle {
		builder.WriteString(m.titleView(m, &top))
	} else {
		top++
	}

	// menu title
	builder.WriteString(m.menuTitleView(m, &top, &MenuItem{Title: "用户登录", Subtitle: "手机号或邮箱"}))
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
		if spaceLen := m.WindowWidth - m.menuStartColumn - valueLen - 3; spaceLen > 0 {
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
	builder.WriteString("    ")
	builder.WriteString(m.loginModel.qrLoginButton)
	builder.WriteString("\n")

	if m.WindowHeight > top+3 {
		builder.WriteString(strings.Repeat("\n", m.WindowHeight-top-3))
	}

	return builder.String()
}

func (m *LoginModel) qrButtonTextByStep() string {
	switch m.qrLoginStep {
	case 1:
		return "已扫码登录，继续"
	case 0:
		fallthrough
	default:
		return "扫码登录"
	}
}

func (m *LoginModel) loginSuccessHandle(nm *NeteaseModel, userInfo []byte) {
	if user, err := structs.NewUserFromJson(userInfo); err == nil {
		nm.user = &user

		// 获取我喜欢的歌单
		userPlaylists := service.UserPlaylistService{
			Uid:    strconv.FormatInt(nm.user.UserId, 10),
			Limit:  strconv.Itoa(1),
			Offset: strconv.Itoa(0),
		}
		_, response := userPlaylists.UserPlaylist()
		nm.user.MyLikePlaylistID, err = jsonparser.GetInt(response, "playlist", "[0]", "id")
		if err != nil {
			log.Printf("获取歌单ID失败: %+v\n", err)
		}

		// 写入本地数据库
		table := storage.NewTable()
		_ = table.SetByKVModel(storage.User{}, user)

		if m.AfterLogin != nil {
			m.AfterLogin(nm, nil, nil)
		}
	}
}

// NeedLoginHandle 需要登录的处理
func NeedLoginHandle(model *NeteaseModel, callback func(m *NeteaseModel, newMenu IMenu, newTitle *MenuItem)) {
	model.pageType = PtLogin
	model.loginModel.AfterLogin = callback
}
