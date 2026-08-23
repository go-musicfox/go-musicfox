package artist

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// ArtistSongMenu 展示歌手的热门歌曲。
type ArtistSongMenu struct {
	ui.BaseMenu
	menus    []model.MenuItem
	songs    []structs.Song
	artistID int64
}

func NewArtistSongMenu(base ui.BaseMenu, artistID int64) *ArtistSongMenu {
	return &ArtistSongMenu{
		BaseMenu: base,
		artistID: artistID,
	}
}

func (m *ArtistSongMenu) IsSearchable() bool {
	return true
}

func (m *ArtistSongMenu) IsPlayable() bool {
	return true
}

func (m *ArtistSongMenu) GetMenuKey() string {
	return fmt.Sprintf("artist_song_%d", m.artistID)
}

func (m *ArtistSongMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *ArtistSongMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {

		artistSongService := service.ArtistTopSongService{Id: strconv.FormatInt(m.artistID, 10)}
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
