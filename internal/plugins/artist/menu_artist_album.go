package artist

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// ArtistAlbumMenu 展示歌手的专辑列表。子菜单跳转 album_detail（key 与提取前
// 一致；ui.AlbumDetailOpts 是 ui 共享 opts，留在 ui，专辑集群在 album 插件）。
type ArtistAlbumMenu struct {
	ui.BaseMenu
	menus    []model.MenuItem
	albums   []structs.Album
	artistID int64
}

func NewArtistAlbumMenu(base ui.BaseMenu, artistID int64) *ArtistAlbumMenu {
	return &ArtistAlbumMenu{
		BaseMenu: base,
		artistID: artistID,
	}
}

func (m *ArtistAlbumMenu) IsSearchable() bool {
	return true
}

func (m *ArtistAlbumMenu) GetMenuKey() string {
	return fmt.Sprintf("artist_album_%d", m.artistID)
}

func (m *ArtistAlbumMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistAlbumMenu) SubMenu(_ *model.App, index int) model.Menu {
	if len(m.albums) < index {
		return nil
	}

	return ui.BuildMenuOrToast("album_detail", m.BaseMenu, ui.AlbumDetailOpts{AlbumID: m.albums[index].Id})
}

func (m *ArtistAlbumMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		artistAlbumService := service.ArtistAlbumService{
			ID:     strconv.FormatInt(m.artistID, 10),
			Offset: "0",
			Limit:  "50",
		}
		code, response := artistAlbumService.ArtistAlbum()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}

		m.albums = _struct.GetArtistHotAlbums(response)
		m.menus = menux.GetViewFromAlbums(m.albums)

		return true, nil
	}
}

func (m *ArtistAlbumMenu) Albums() []structs.Album {
	return m.albums
}
