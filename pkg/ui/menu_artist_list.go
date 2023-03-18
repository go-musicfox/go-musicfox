package ui

import "github.com/go-musicfox/go-musicfox/pkg/structs"

type ArtistsOfSongMenu struct {
	DefaultMenu
	menus    []MenuItem
	menuList []Menu
	song     structs.Song
}

func NewArtistsOfSongMenu(song structs.Song) *ArtistsOfSongMenu {
	artistsMenu := &ArtistsOfSongMenu{
		song: song,
	}
	var subTitle = "「" + song.Name + "」所属歌手"
	for _, artist := range song.Artists {
		artistsMenu.menus = append(artistsMenu.menus, MenuItem{Title: artist.Name, Subtitle: subTitle})
		artistsMenu.menuList = append(artistsMenu.menuList, NewArtistDetailMenu(artist.Id, artist.Name))
	}

	return artistsMenu
}

func (m *ArtistsOfSongMenu) GetMenuKey() string {
	return "artist_of_song"
}

func (m *ArtistsOfSongMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *ArtistsOfSongMenu) Artists() []structs.Artist {
	return m.song.Artists
}

func (m *ArtistsOfSongMenu) SubMenu(_ *NeteaseModel, index int) Menu {
	if index >= len(m.menuList) {
		return nil
	}

	return m.menuList[index]
}
