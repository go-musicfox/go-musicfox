package ui

import (
    "github.com/buger/jsonparser"
    "go-musicfox/constants"
    "go-musicfox/utils"
    "io/ioutil"
    "net/http"
    "strings"
)

type CheckUpdateMenu struct {
    hasUpdate bool
}

func NewCheckUpdateMenu() *CheckUpdateMenu {
    return &CheckUpdateMenu{}
}

func (m *CheckUpdateMenu) MenuData() interface{} {
    return nil
}

func (m *CheckUpdateMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *CheckUpdateMenu) IsPlayable() bool {
    return false
}

func (m *CheckUpdateMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *CheckUpdateMenu) GetMenuKey() string {
    return "check_update"
}

func (m *CheckUpdateMenu) MenuViews() []MenuItem {
    if m.hasUpdate {
        return []MenuItem{
            {Title: "检查到新版本，回车查看~", Subtitle: "ENTER"},
        }
    }

    return []MenuItem{
        {Title: "已是最新版本"},
    }
}

func (m *CheckUpdateMenu) SubMenu(_ *NeteaseModel, _ int) IMenu {
    if m.hasUpdate {
        _ = utils.OpenUrl(constants.AppGithubUrl)
    }
    return nil
}

func (m *CheckUpdateMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *CheckUpdateMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *CheckUpdateMenu) BeforeEnterMenuHook() Hook {
    return func(model *NeteaseModel) bool {
        response, err := http.Get(constants.AppCheckUpdateUrl)
        if err != nil {
            m.hasUpdate = false
            return true
        }
        defer response.Body.Close()

        jsonBytes, err := ioutil.ReadAll(response.Body)
        if err != nil {
            m.hasUpdate = false
            return true
        }

        tag, err := jsonparser.GetString(jsonBytes, "tag_name")
        if err != nil {
            m.hasUpdate = false
            return true
        }

        curTagArr := strings.Split(constants.AppVersion, ".")
        tagArr := strings.Split(tag, ".")
        if len(tagArr) >= 1 && len(curTagArr) >= 1 {
            if tagArr[0] > curTagArr[0] {
                m.hasUpdate = true
                return true
            }

            if tagArr[0] < curTagArr[0] {
                m.hasUpdate = false
                return true
            }

        }

        if len(tagArr) >= 2 && len(curTagArr) >= 2 {
            if tagArr[1] > curTagArr[1] {
                m.hasUpdate = true
                return true
            }

            if tagArr[1] < curTagArr[1] {
                m.hasUpdate = false
                return true
            }

        }

        if len(tagArr) >= 3 && len(curTagArr) >= 3 {
            if tagArr[2] > curTagArr[2] {
                m.hasUpdate = true
                return true
            }

            if tagArr[2] < curTagArr[2] {
                m.hasUpdate = false
                return true
            }

        }

        m.hasUpdate = false
        return true
    }
}

func (m *CheckUpdateMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *CheckUpdateMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
