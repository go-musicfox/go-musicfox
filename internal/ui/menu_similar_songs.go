package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type SimilarSongsMenu struct {
	baseMenu
	menus        []model.MenuItem
	songs        []structs.Song
	existSongIds map[int64]struct{}
	relateSongId int64
}

var _ SongsMenu = (*SimilarSongsMenu)(nil)

func NewSimilarSongsMenu(base baseMenu, song structs.Song) *SimilarSongsMenu {
	songs := []structs.Song{song}
	return &SimilarSongsMenu{
		baseMenu:     base,
		menus:        menux.GetViewFromSongs(songs),
		songs:        songs,
		existSongIds: map[int64]struct{}{song.Id: {}},
		relateSongId: song.Id,
	}
}

func (m *SimilarSongsMenu) IsSearchable() bool {
	return true
}

func (m *SimilarSongsMenu) IsPlayable() bool {
	return true
}

func (m *SimilarSongsMenu) GetMenuKey() string {
	return fmt.Sprintf("simi_songs_%d", m.relateSongId)
}

func (m *SimilarSongsMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *SimilarSongsMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		return m.fetchSimilarSongs(m.relateSongId, 2), nil
	}
}

func (m *SimilarSongsMenu) Songs() []structs.Song {
	return m.songs
}

func (m *SimilarSongsMenu) BottomOutHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		return m.fetchSimilarSongs(m.songs[len(m.songs)-1].Id, 2), nil
	}
}

func (m *SimilarSongsMenu) fetchSimilarSongs(songId int64, maxTry int) bool {
	simiSongService := service.SimiSongService{
		ID: strconv.FormatInt(songId, 10),
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
		return m.fetchSimilarSongs(songId, maxTry-1)
	}

	m.songs = append(m.songs, songs...)
	m.menus = menux.GetViewFromSongs(m.songs)
	m.netease.player.songManager.init(m.netease.player.CurSongIndex(), m.songs)
	m.netease.player.playlistUpdateAt = time.Now()
	return true
}
