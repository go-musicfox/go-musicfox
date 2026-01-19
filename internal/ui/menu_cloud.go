package ui

import (
	"log/slog"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type CloudMenu struct {
	baseMenu
	menus   []model.MenuItem
	songs   []structs.Song
	limit   int
	offset  int
	hasMore bool
}

func NewCloudMenu(base baseMenu) *CloudMenu {
	return &CloudMenu{
		baseMenu: base,
		limit:    100,
		offset:   0,
	}
}

func (m *CloudMenu) ClearCache() {
	m.menus = nil
	m.songs = nil
	m.offset = 0
	m.hasMore = true
}

func (m *CloudMenu) IsSearchable() bool {
	return true
}

func (m *CloudMenu) IsPlayable() bool {
	return true
}

func (m *CloudMenu) GetMenuKey() string {
	return "could"
}

func (m *CloudMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *CloudMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if _struct.CheckUserInfo(m.netease.user) == _struct.NeedLogin {
			slog.Info("Cloud menu requires login, redirecting to login page")
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		}

		cloudService := service.UserCloudService{
			Offset: strconv.Itoa(m.offset),
			Limit:  strconv.Itoa(m.limit),
		}
		code, response := cloudService.UserCloud()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			slog.Warn("Cloud session expired, redirecting to login page")
			page, _ := m.netease.ToLoginPage(EnterMenuCallback(main))
			return false, page
		} else if codeType != _struct.Success {
			slog.Error("Failed to fetch cloud songs", "code", code, "response", string(response))
			return false, nil
		}

		if hasMore, err := jsonparser.GetBoolean(response, "hasMore"); err == nil {
			m.hasMore = hasMore
		}

		m.songs = _struct.GetSongsOfCloud(response)
		m.menus = menux.GetViewFromSongs(m.songs)
		slog.Info("Cloud songs loaded successfully", "count", len(m.songs))

		return true, nil
	}
}

func (m *CloudMenu) BottomOutHook() model.Hook {
	if !m.hasMore {
		return nil
	}
	return func(main *model.Main) (bool, model.Page) {
		m.offset += m.limit

		cloudService := service.UserCloudService{
			Offset: strconv.Itoa(m.offset),
			Limit:  strconv.Itoa(m.limit),
		}
		code, response := cloudService.UserCloud()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			slog.Warn("Cloud session expired while loading more songs, redirecting to login page")
			page, _ := m.netease.ToLoginPage(func() model.Page {
				main.RefreshMenuList()
				return nil
			})
			return false, page
		} else if codeType != _struct.Success {
			slog.Error("Failed to fetch more cloud songs", "code", code, "offset", m.offset)
			return false, nil
		}

		if hasMore, err := jsonparser.GetBoolean(response, "hasMore"); err == nil {
			m.hasMore = hasMore
		}

		songs := _struct.GetSongsOfCloud(response)
		menus := menux.GetViewFromSongs(songs)

		m.songs = append(m.songs, songs...)
		m.menus = append(m.menus, menus...)
		slog.Info("More cloud songs loaded", "count", len(songs), "total", len(m.songs))

		return true, nil
	}
}

func (m *CloudMenu) Songs() []structs.Song {
	return m.songs
}
