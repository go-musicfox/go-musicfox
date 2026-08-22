package ui

import (
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

type ArtistsOfSongMenu struct {
	baseMenu
	menus    []model.MenuItem
	menuList []Menu
	song     structs.Song
}

func NewArtistsOfSongMenu(base baseMenu, song structs.Song) *ArtistsOfSongMenu {
	artistsMenu := &ArtistsOfSongMenu{
		baseMenu: base,
		song:     song,
	}
	var subTitle = "「" + song.Name + "」所属歌手"
	for _, artist := range song.Artists {
		artistsMenu.menus = append(artistsMenu.menus, model.MenuItem{Title: artist.Name, Subtitle: subTitle})
		artistMenu, err := BuildMenuB("artist_detail", base, ArtistDetailOpts{ArtistID: artist.Id, Name: artist.Name})
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

func (m *ArtistsOfSongMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
