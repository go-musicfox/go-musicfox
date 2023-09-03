package ui

import (
	"fmt"

	"github.com/anhoder/foxful-cli/model"
)

type ArtistDetailMenu struct {
	baseMenu
	menus    []model.MenuItem
	artistId int64
}

func NewArtistDetailMenu(base baseMenu, artistId int64, artistName string) *ArtistDetailMenu {
	artistMenu := &ArtistDetailMenu{
		baseMenu: base,
		menus: []model.MenuItem{
			{Title: "热门歌曲", Subtitle: artistName},
			{Title: "热门专辑", Subtitle: artistName},
		},
		artistId: artistId,
	}

	return artistMenu
}

func (m *ArtistDetailMenu) GetMenuKey() string {
	return fmt.Sprintf("artist_detail_%d", m.artistId)
}

func (m *ArtistDetailMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistDetailMenu) SubMenu(_ *model.App, index int) model.Menu {
	switch index {
	case 0:
		return NewArtistSongMenu(m.baseMenu, m.artistId)
	case 1:
		return NewArtistAlbumMenu(m.baseMenu, m.artistId)
	}

	return nil
}
