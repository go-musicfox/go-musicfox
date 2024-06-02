package ui

import (
	"fmt"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	ds2 "github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type SearchResultMenu struct {
	baseMenu
	menus      []model.MenuItem
	offset     int
	searchType SearchType
	keyword    string
	result     any
}

var playableTypes = map[SearchType]bool{
	StSingleSong: true,
	StAlbum:      false,
	StSinger:     false,
	StPlaylist:   false,
	StUser:       false,
	StLyric:      true,
	StRadio:      false,
}

func NewSearchResultMenu(base baseMenu, searchType SearchType) *SearchResultMenu {
	return &SearchResultMenu{
		baseMenu:   base,
		offset:     0,
		searchType: searchType,
	}
}

func (m *SearchResultMenu) IsSearchable() bool {
	return true
}

func (m *SearchResultMenu) BeforeBackMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.netease.search.wordsInput.Value() != "" {
			m.netease.search.wordsInput.SetValue("")
		}

		return true, nil
	}
}

func (m *SearchResultMenu) IsPlayable() bool {
	return playableTypes[m.searchType]
}

func (m *SearchResultMenu) GetMenuKey() string {
	return fmt.Sprintf("search_result_%d_%s", m.searchType, m.keyword)
}

func (m *SearchResultMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *SearchResultMenu) SubMenu(_ *model.App, index int) model.Menu {
	switch resultWithType := m.result.(type) {
	case []ds2.Song:
		return nil
	case []ds2.Album:
		if index >= len(resultWithType) {
			return nil
		}
		return NewAlbumDetailMenu(m.baseMenu, resultWithType[index].Id)
	case []ds2.Playlist:
		if index >= len(resultWithType) {
			return nil
		}
		return NewPlaylistDetailMenu(m.baseMenu, resultWithType[index].Id)
	case []ds2.Artist:
		if index >= len(resultWithType) {
			return nil
		}
		return NewArtistDetailMenu(m.baseMenu, resultWithType[index].Id, resultWithType[index].Name)
	case []ds2.User:
		if index >= len(resultWithType) {
			return nil
		}
		return NewUserPlaylistMenu(m.baseMenu, resultWithType[index].UserId)
	case []ds2.DjRadio:
		if index >= len(resultWithType) {
			return nil
		}
		return NewDjRadioDetailMenu(m.baseMenu, resultWithType[index].Id)
	}

	return nil
}

func (m *SearchResultMenu) BottomOutHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		var (
			code     float64
			response []byte
		)
		m.offset += types.SearchPageSize
		searchService := service.SearchService{
			S:      m.keyword,
			Type:   strconv.Itoa(int(m.searchType)),
			Limit:  strconv.Itoa(types.SearchPageSize),
			Offset: strconv.Itoa(m.offset),
		}
		code, response = searchService.Search()

		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			m.offset -= types.SearchPageSize
			return false, nil
		}

		m.appendResult(response)
		m.convertMenus()
		return true, nil
	}
}

func (m *SearchResultMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		if m.netease.search.wordsInput.Value() == "" {
			// 显示搜索页面
			page, _ := m.netease.ToSearchPage(m.searchType)
			return false, page
		}

		m.result = m.netease.search.result
		m.searchType = m.netease.search.searchType
		m.keyword = m.netease.search.wordsInput.Value()
		m.convertMenus()
		return true, nil
	}
}

func (m *SearchResultMenu) appendResult(response []byte) {
	switch m.searchType {
	case StSingleSong, StLyric:
		appendSongs := _struct.GetSongsOfSearchResult(response)
		songs, _ := m.result.([]ds2.Song)
		songs = append(songs, appendSongs...)
		m.result = songs
	case StAlbum:
		appendAlbums := _struct.GetAlbumsOfSearchResult(response)
		albums, _ := m.result.([]ds2.Album)
		albums = append(albums, appendAlbums...)
		m.result = albums
	case StSinger:
		appendArtists := _struct.GetArtistsOfSearchResult(response)
		artists, _ := m.result.([]ds2.Artist)
		artists = append(artists, appendArtists...)
		m.result = artists
	case StPlaylist:
		appendPlaylists := _struct.GetPlaylistsOfSearchResult(response)
		playlists, _ := m.result.([]ds2.Playlist)
		playlists = append(playlists, appendPlaylists...)
		m.result = playlists
	case StUser:
		appendUsers := _struct.GetUsersOfSearchResult(response)
		users, _ := m.result.([]ds2.User)
		users = append(users, appendUsers...)
		m.result = users
	case StRadio:
		appendDjRadios := _struct.GetDjRadiosOfSearchResult(response)
		djRadios, _ := m.result.([]ds2.DjRadio)
		djRadios = append(djRadios, appendDjRadios...)
		m.result = djRadios
	}
}

func (m *SearchResultMenu) convertMenus() {
	switch resultWithType := m.result.(type) {
	case []ds2.Song:
		m.menus = menux.GetViewFromSongs(resultWithType)
	case []ds2.Album:
		m.menus = menux.GetViewFromAlbums(resultWithType)
	case []ds2.Playlist:
		m.menus = menux.GetViewFromPlaylists(resultWithType)
	case []ds2.Artist:
		m.menus = menux.GetViewFromArtists(resultWithType)
	case []ds2.User:
		m.menus = menux.GetViewFromUsers(resultWithType)
	case []ds2.DjRadio:
		m.menus = menux.GetViewFromDjRadios(resultWithType)
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
