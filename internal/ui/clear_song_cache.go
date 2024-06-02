package ui

import (
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/storagex"
)

type ClearSongCacheMenu struct {
	baseMenu
	netease *Netease
}

func NewClearSongCacheMenu(base baseMenu, netease *Netease) *ClearSongCacheMenu {
	return &ClearSongCacheMenu{
		baseMenu: base,
		netease:  netease,
	}
}

func (m *ClearSongCacheMenu) IsSearchable() bool {
	return true
}

func (m *ClearSongCacheMenu) GetMenuKey() string {
	return "clear_song_cache_menu"
}

func (m *ClearSongCacheMenu) MenuViews() []model.MenuItem {
	return []model.MenuItem{
		{Title: "确定"},
	}
}

func (m *ClearSongCacheMenu) SubMenu(_ *model.App, index int) model.Menu {
	loading := model.NewLoading(m.netease.MustMain())
	loading.Start()
	defer loading.Complete()
	err := storagex.ClearMusicCache()
	if err != nil {
		notify.Notify(notify.NotifyContent{
			Title:   "清除缓存失败",
			Text:    err.Error(),
			GroupId: types.GroupID,
		})
	} else {
		notify.Notify(notify.NotifyContent{
			Title:   "清除缓存成功",
			Text:    "缓存已清除",
			GroupId: types.GroupID,
		})
	}
	m.netease.MustMain().BackMenu()
	return nil
}
