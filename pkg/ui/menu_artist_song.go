package ui

import (
	"fmt"
	"strconv"

	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/anhoder/netease-music/service"
)

type ArtistSongMenu struct {
	DefaultMenu
	menus    []MenuItem
	songs    []structs.Song
	artistId int64
}

func NewArtistSongMenu(artistId int64) *ArtistSongMenu {
	return &ArtistSongMenu{
		artistId: artistId,
	}
}

func (m *ArtistSongMenu) IsSearchable() bool {
	return true
}

func (m *ArtistSongMenu) IsPlayable() bool {
	return true
}

func (m *ArtistSongMenu) GetMenuKey() string {
	return fmt.Sprintf("artist_song_%d", m.artistId)
}

func (m *ArtistSongMenu) MenuViews() []MenuItem {
	return m.menus
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

func (m *ArtistSongMenu) Songs() []structs.Song {
	return m.songs
}
