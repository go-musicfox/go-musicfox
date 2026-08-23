package artist

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// ArtistDetailMenu 展示歌手详情（热门歌曲 / 热门专辑两个入口）。子菜单经
// 注册表构建 artist_song / artist_album（key 与提取前一致；ArtistSongOpts /
// ArtistAlbumOpts 仅集群内部使用，随集群移入本插件）。
type ArtistDetailMenu struct {
	ui.BaseMenu
	menus    []model.MenuItem
	artistID int64
}

func NewArtistDetailMenu(base ui.BaseMenu, artistID int64, artistName string) *ArtistDetailMenu {
	artistMenu := &ArtistDetailMenu{
		BaseMenu: base,
		menus: []model.MenuItem{
			{Title: "热门歌曲", Subtitle: artistName},
			{Title: "热门专辑", Subtitle: artistName},
		},
		artistID: artistID,
	}

	return artistMenu
}

func (m *ArtistDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("artist_detail_%d", m.artistID)
}

func (m *ArtistDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

// ArtistID returns the artist this detail menu shows (the operate.go re-entry
// dedup reads it through ui.ArtistDetailIDGetter).
func (m *ArtistDetailMenu) ArtistID() int64 {
	return m.artistID
}

func (m *ArtistDetailMenu) SubMenu(_ *model.App, index int) model.Menu {
	switch index {
	case 0:
		return ui.BuildMenuOrToast("artist_song", m.BaseMenu, ArtistSongOpts{ArtistID: m.artistID})
	case 1:
		return ui.BuildMenuOrToast("artist_album", m.BaseMenu, ArtistAlbumOpts{ArtistID: m.artistID})
	}

	return nil
}
