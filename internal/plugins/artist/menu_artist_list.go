package artist

import (
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// ArtistsOfSongMenu 展示歌曲所属歌手列表（多歌手时由 operate 的
// goToArtistOfSong 进入）。子菜单经注册表构建 artist_detail（key 与提取前
// 一致；ArtistDetailOpts 是 ui 共享 opts，留在 ui）。
type ArtistsOfSongMenu struct {
	ui.BaseMenu
	menus    []model.MenuItem
	menuList []ui.Menu
	song     structs.Song
}

func NewArtistsOfSongMenu(base ui.BaseMenu, song structs.Song) *ArtistsOfSongMenu {
	artistsMenu := &ArtistsOfSongMenu{
		BaseMenu: base,
		song:     song,
	}
	var subTitle = "「" + song.Name + "」所属歌手"
	for _, artist := range song.Artists {
		artistsMenu.menus = append(artistsMenu.menus, model.MenuItem{Title: artist.Name, Subtitle: subTitle})
		artistMenu, err := ui.BuildMenu("artist_detail", base, ui.ArtistDetailOpts{ArtistID: artist.Id, Name: artist.Name})
		if err != nil {
			continue // cannot happen with the static registry; skip on error
		}
		artistsMenu.menuList = append(artistsMenu.menuList, artistMenu)
	}

	return artistsMenu
}

func (m *ArtistsOfSongMenu) GetMenuKey() string {
	return "artist_of_song"
}

func (m *ArtistsOfSongMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistsOfSongMenu) Artists() []structs.Artist {
	return m.song.Artists
}

// SongID returns the song this artists-of-song menu belongs to (the operate.go
// re-entry dedup reads it through ui.ArtistsOfSongSongIDGetter).
func (m *ArtistsOfSongMenu) SongID() int64 {
	return m.song.Id
}

func (m *ArtistsOfSongMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
