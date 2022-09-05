package ui

import (
	"github.com/anhoder/netease-music/service"
	"github.com/buger/jsonparser"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
	"strconv"
)

type DjSubListMenu struct {
    menus  []MenuItem
    radios []structs.DjRadio
    limit  int
    offset int
    total  int
}

func NewDjSubListMenu() *DjSubListMenu {
    return &DjSubListMenu{
        limit:  50,
        offset: 0,
        total:  -1,
    }
}

func (m *DjSubListMenu) MenuData() interface{} {
    return nil
}

func (m *DjSubListMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *DjSubListMenu) IsPlayable() bool {
    return false
}

func (m *DjSubListMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *DjSubListMenu) GetMenuKey() string {
    return "dj_sub"
}

func (m *DjSubListMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *DjSubListMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
    if index >= len(m.radios) {
        return nil
    }

    return NewDjRadioDetailMenu(m.radios[index].Id)
}

func (m *DjSubListMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjSubListMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *DjSubListMenu) BeforeEnterMenuHook() Hook {
    return func(model *NeteaseModel) bool {

        if utils.CheckUserInfo(model.user) == utils.NeedLogin {
            NeedLoginHandle(model, enterMenu)
            return false
        }

        // 不重复请求
        if len(m.menus) > 0 && len(m.radios) > 0 {
            return true
        }

        djSublistService := service.DjSublistService{
            Limit:  strconv.Itoa(m.limit),
            Offset: strconv.Itoa(m.offset),
        }
        code, response := djSublistService.DjSublist()
        codeType := utils.CheckCode(code)
        if codeType == utils.NeedLogin {
            NeedLoginHandle(model, enterMenu)
            return false
        } else if codeType != utils.Success {
            return false
        }

        if total, err := jsonparser.GetInt(response, "count"); err != nil {
            m.total = int(total)
        }

        m.radios = utils.GetDjRadios(response)
        m.menus = GetViewFromDjRadios(m.radios)

        return true
    }
}

func (m *DjSubListMenu) BottomOutHook() Hook {
    if len(m.radios) >= m.total {
        return nil
    }

    return func(model *NeteaseModel) bool {
        m.offset += m.limit

        if utils.CheckUserInfo(model.user) == utils.NeedLogin {
            NeedLoginHandle(model, enterMenu)
            return false
        }

        djSublistService := service.DjSublistService{
            Limit:  strconv.Itoa(m.limit),
            Offset: strconv.Itoa(m.offset),
        }
        code, response := djSublistService.DjSublist()
        codeType := utils.CheckCode(code)
        if codeType == utils.NeedLogin {
            NeedLoginHandle(model, enterMenu)
            return false
        } else if codeType != utils.Success {
            return false
        }

        if total, err := jsonparser.GetInt(response, "count"); err != nil {
            m.total = int(total)
        }

        radios := utils.GetDjRadios(response)
        menus := GetViewFromDjRadios(radios)

        m.radios = append(m.radios, radios...)
        m.menus = append(m.menus, menus...)

        return true
    }
}

func (m *DjSubListMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
