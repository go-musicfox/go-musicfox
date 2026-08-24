package song

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

type SimilarSongsMenu struct {
	ui.BaseMenu
	menus        []model.MenuItem
	songs        []structs.Song
	existSongIds map[int64]struct{}
	relateSongID int64
}

var _ ui.SongsMenu = (*SimilarSongsMenu)(nil)

func NewSimilarSongsMenu(base ui.BaseMenu, song structs.Song) *SimilarSongsMenu {
	songs := []structs.Song{song}
	return &SimilarSongsMenu{
		BaseMenu:     base,
		menus:        menux.GetViewFromSongs(songs),
		songs:        songs,
		existSongIds: map[int64]struct{}{song.Id: {}},
		relateSongID: song.Id,
	}
}

func (m *SimilarSongsMenu) IsSearchable() bool {
	return true
}

func (m *SimilarSongsMenu) IsPlayable() bool {
	return true
}

func (m *SimilarSongsMenu) GetMenuKey() string {
	return fmt.Sprintf("simi_songs_%d", m.relateSongID)
}

func (m *SimilarSongsMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *SimilarSongsMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		return m.fetchSimilarSongs(m.relateSongID, 2), nil
	}
}

func (m *SimilarSongsMenu) Songs() []structs.Song {
	return m.songs
}

// RelateSongID returns the source song ID the similar-songs list was built
// from. ui's operate.go uses it (via the SimilarSongsRelateSongIDGetter
// interface) to avoid re-entering the same list.
func (m *SimilarSongsMenu) RelateSongID() int64 {
	return m.relateSongID
}

func (m *SimilarSongsMenu) BottomOutHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		return m.fetchSimilarSongs(m.songs[len(m.songs)-1].Id, 2), nil
	}
}

func (m *SimilarSongsMenu) fetchSimilarSongs(songID int64, maxTry int) bool {
	simiSongService := service.SimiSongService{
		ID: strconv.FormatInt(songID, 10),
	}
	code, response := simiSongService.SimiSong()
	codeType := _struct.CheckCode(code)
	if codeType != _struct.Success {
		return false
	}

	var songs []structs.Song
	for _, song := range _struct.GetSimiSongs(response) {
		if _, ok := m.existSongIds[song.Id]; !ok {
			m.existSongIds[song.Id] = struct{}{}
			songs = append(songs, song)
		}
	}

	if len(songs) == 0 {
		if maxTry <= 0 {
			return false
		}
		return m.fetchSimilarSongs(songID, maxTry-1)
	}

	m.songs = append(m.songs, songs...)
	m.menus = menux.GetViewFromSongs(m.songs)
	m.Player().ReinitializePlaylist(m.Player().CurSongIndex(), m.songs)
	m.Player().MarkPlaylistUpdated()
	return true
}
