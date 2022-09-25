package ui

import "fmt"

type ArtistDetailMenu struct {
	DefaultMenu
	menus    []MenuItem
	artistId int64
}

func NewArtistDetailMenu(artistId int64) *ArtistDetailMenu {
	artistMenu := new(ArtistDetailMenu)
	artistMenu.menus = []MenuItem{
		{Title: "热门歌曲"},
		{Title: "热门专辑"},
	}
	artistMenu.artistId = artistId

	return artistMenu
}

func (m *ArtistDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("artist_detail_%d", m.artistId)
}

func (m *ArtistDetailMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *ArtistDetailMenu) SubMenu(_ *NeteaseModel, index int) IMenu {
	switch index {
	case 0:
		return NewArtistSongMenu(m.artistId)
	case 1:
		return NewArtistAlbumMenu(m.artistId)
	}

	return nil
}
