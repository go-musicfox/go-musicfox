package ui

import "fmt"

type ArtistDetailMenu struct {
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

func (m *ArtistDetailMenu) MenuData() interface{} {
    return nil
}

func (m *ArtistDetailMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *ArtistDetailMenu) IsPlayable() bool {
    return false
}

func (m *ArtistDetailMenu) ResetPlaylistWhenPlay() bool {
    return false
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

func (m *ArtistDetailMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *ArtistDetailMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *ArtistDetailMenu) BeforeEnterMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *ArtistDetailMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *ArtistDetailMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}
