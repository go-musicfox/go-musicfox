package ui

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type ArtistSongMenu struct {
	baseMenu
	menus    []model.MenuItem
	songs    []structs.Song
	artistId int64
}

func NewArtistSongMenu(base baseMenu, artistId int64) *ArtistSongMenu {
	return &ArtistSongMenu{
		baseMenu: base,
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

func (m *ArtistSongMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistSongMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		artistSongService := service.ArtistTopSongService{Id: strconv.FormatInt(m.artistId, 10)}
		code, response := artistSongService.ArtistTopSong()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return false, nil
		}
		m.songs = _struct.GetSongsOfArtist(response)
		m.menus = menux.GetViewFromSongs(m.songs)

		return true, nil
	}
}

func (m *ArtistSongMenu) Songs() []structs.Song {
	return m.songs
}
