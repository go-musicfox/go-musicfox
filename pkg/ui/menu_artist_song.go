package ui

import (
	"fmt"
	"github.com/anhoder/netease-music/service"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
	"strconv"
)

type ArtistSongMenu struct {
	menus 	   []MenuItem
	songs      []structs.Song
	artistId int64
}

func NewArtistSongMenu(artistId int64) *ArtistSongMenu {
	return &ArtistSongMenu{
		artistId: artistId,
	}
}

func (m *ArtistSongMenu) MenuData() interface{} {
	return m.songs
}

func (m *ArtistSongMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (m *ArtistSongMenu) IsPlayable() bool {
	return true
}

func (m *ArtistSongMenu) ResetPlaylistWhenPlay() bool {
	return false
}

func (m *ArtistSongMenu) GetMenuKey() string {
	return fmt.Sprintf("artist_song_%d", m.artistId)
}

func (m *ArtistSongMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *ArtistSongMenu) SubMenu(_ *NeteaseModel, _ int) IMenu {
	return nil
}

func (m *ArtistSongMenu) BeforePrePageHook() Hook {
	// Nothing to do
	return nil
}

func (m *ArtistSongMenu) BeforeNextPageHook() Hook {
	// Nothing to do
	return nil
}

func (m *ArtistSongMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {

		artistSongService := service.ArtistTopSongService{Id: strconv.FormatInt(m.artistId, 10)}
		code, response := artistSongService.ArtistTopSong()
		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			return false
		}
		m.songs = utils.GetSongsOfArtist(response)
		m.menus = GetViewFromSongs(m.songs)

		return true
	}
}

func (m *ArtistSongMenu) BottomOutHook() Hook {
	// Nothing to do
	return nil
}

func (m *ArtistSongMenu) TopOutHook() Hook {
	// Nothing to do
	return nil
}

