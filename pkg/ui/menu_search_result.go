package ui

import (
	"fmt"
	"strconv"

	"go-musicfox/pkg/constants"
	ds2 "go-musicfox/pkg/structs"
	"go-musicfox/utils"

	"github.com/anhoder/netease-music/service"
)

type SearchResultMenu struct {
	DefaultMenu
	menus      []MenuItem
	offset     int
	searchType SearchType
	keyword    string
	result     interface{}
}

func NewSearchResultMenu(searchType SearchType) *SearchResultMenu {
	return &SearchResultMenu{
		offset:     0,
		searchType: searchType,
	}
}

func (m *SearchResultMenu) IsSearchable() bool {
	return true
}

func (m *SearchResultMenu) BeforeBackMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		if model.searchModel.wordsInput.Value() != "" {
			model.searchModel.wordsInput.SetValue("")
		}

		return true
	}
}

func (m *SearchResultMenu) IsPlayable() bool {
	playableMap := map[SearchType]bool{
		StSingleSong: true,
		StAlbum:      false,
		StSinger:     false,
		StPlaylist:   false,
		StUser:       false,
		StLyric:      true,
		StRadio:      false,
	}

	if playable, ok := playableMap[m.searchType]; ok {
		return playable
	}

	return false
}

func (m *SearchResultMenu) GetMenuKey() string {
	return fmt.Sprintf("search_result_%d_%s", m.searchType, m.keyword)
}

func (m *SearchResultMenu) MenuViews() []MenuItem {
	return m.menus
}

func (m *SearchResultMenu) SubMenu(_ *NeteaseModel, index int) Menu {
	switch resultWithType := m.result.(type) {
	case []ds2.Song:
		return nil
	case []ds2.Album:
		if index >= len(resultWithType) {
			return nil
		}
		return NewAlbumDetailMenu(resultWithType[index].Id)
	case []ds2.Playlist:
		if index >= len(resultWithType) {
			return nil
		}
		return NewPlaylistDetailMenu(resultWithType[index].Id)
	case []ds2.Artist:
		if index >= len(resultWithType) {
			return nil
		}
		return NewArtistDetailMenu(resultWithType[index].Id, resultWithType[index].Name)
	case []ds2.User:
		if index >= len(resultWithType) {
			return nil
		}
		return NewUserPlaylistMenu(resultWithType[index].UserId)
	case []ds2.DjRadio:
		if index >= len(resultWithType) {
			return nil
		}
		return NewDjRadioDetailMenu(resultWithType[index].Id)
	}

	return nil
}

func (m *SearchResultMenu) BottomOutHook() Hook {
	return func(model *NeteaseModel) bool {
		var (
			code     float64
			response []byte
		)
		m.offset += constants.SearchPageSize
		searchService := service.SearchService{
			S:      m.keyword,
			Type:   strconv.Itoa(int(m.searchType)),
			Limit:  strconv.Itoa(constants.SearchPageSize),
			Offset: strconv.Itoa(m.offset),
		}
		code, response = searchService.Search()

		codeType := utils.CheckCode(code)
		if codeType != utils.Success {
			m.offset -= constants.SearchPageSize
			return false
		}

		m.appendResult(response)
		m.convertMenus()
		return true
	}
}

func (m *SearchResultMenu) BeforeEnterMenuHook() Hook {
	return func(model *NeteaseModel) bool {
		if model.searchModel.wordsInput.Value() == "" {
			// 显示搜索页面
			model.searchModel.SearchHandle(model, m.searchType)
			return false
		}

		m.result = model.searchModel.result
		m.searchType = model.searchModel.searchType
		m.keyword = model.searchModel.wordsInput.Value()
		m.convertMenus()
		return true
	}
}

func (m *SearchResultMenu) appendResult(response []byte) {
	switch m.searchType {
	case StSingleSong, StLyric:
		appendSongs := utils.GetSongsOfSearchResult(response)
		songs, _ := m.result.([]ds2.Song)
		songs = append(songs, appendSongs...)
		m.result = songs
	case StAlbum:
		appendAlbums := utils.GetAlbumsOfSearchResult(response)
		albums, _ := m.result.([]ds2.Album)
		albums = append(albums, appendAlbums...)
		m.result = albums
	case StSinger:
		appendArtists := utils.GetArtistsOfSearchResult(response)
		artists, _ := m.result.([]ds2.Artist)
		artists = append(artists, appendArtists...)
		m.result = artists
	case StPlaylist:
		appendPlaylists := utils.GetPlaylistsOfSearchResult(response)
		playlists, _ := m.result.([]ds2.Playlist)
		playlists = append(playlists, appendPlaylists...)
		m.result = playlists
	case StUser:
		appendUsers := utils.GetUsersOfSearchResult(response)
		users, _ := m.result.([]ds2.User)
		users = append(users, appendUsers...)
		m.result = users
	case StRadio:
		appendDjRadios := utils.GetDjRadiosOfSearchResult(response)
		djRadios, _ := m.result.([]ds2.DjRadio)
		djRadios = append(djRadios, appendDjRadios...)
		m.result = djRadios
	}
}

func (m *SearchResultMenu) convertMenus() {
	switch resultWithType := m.result.(type) {
	case []ds2.Song:
		m.menus = GetViewFromSongs(resultWithType)
	case []ds2.Album:
		m.menus = GetViewFromAlbums(resultWithType)
	case []ds2.Playlist:
		m.menus = GetViewFromPlaylists(resultWithType)
	case []ds2.Artist:
		m.menus = GetViewFromArtists(resultWithType)
	case []ds2.User:
		m.menus = GetViewFromUsers(resultWithType)
	case []ds2.DjRadio:
		m.menus = GetViewFromDjRadios(resultWithType)
	}
}

func (m *SearchResultMenu) Songs() []ds2.Song {
	if songs, ok := m.result.([]ds2.Song); ok {
		return songs
	}
	return nil
}

func (m *SearchResultMenu) Playlists() []ds2.Playlist {
	if playlists, ok := m.result.([]ds2.Playlist); ok {
		return playlists
	}
	return nil
}

func (m *SearchResultMenu) Albums() []ds2.Album {
	if albums, ok := m.result.([]ds2.Album); ok {
		return albums
	}
	return nil
}

func (m *SearchResultMenu) Artists() []ds2.Artist {
	if artists, ok := m.result.([]ds2.Artist); ok {
		return artists
	}
	return nil
}
