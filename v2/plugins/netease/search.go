package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
	plugin "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/music"
)

// Search 搜索功能实现
func (p *NeteasePlugin) Search(ctx context.Context, query string, options plugin.SearchOptions) (*plugin.SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// 设置默认参数
	if options.Limit <= 0 {
		options.Limit = 30
	}
	if options.Offset < 0 {
		options.Offset = 0
	}

	// 根据搜索类型调用不同的API
	switch options.Type {
	case plugin.SearchTypeTrack, plugin.SearchTypeAll:
		return p.searchTracks(ctx, query, options)
	case plugin.SearchTypeAlbum:
		return p.searchAlbums(ctx, query, options)
	case plugin.SearchTypeArtist:
		return p.searchArtists(ctx, query, options)
	case plugin.SearchTypePlaylist:
		return p.searchPlaylists(ctx, query, options)
	default:
		return p.searchTracks(ctx, query, options)
	}
}

// searchTracks 搜索歌曲
func (p *NeteasePlugin) searchTracks(ctx context.Context, query string, options plugin.SearchOptions) (*plugin.SearchResult, error) {
	searchService := &service.SearchService{
		S:      query,
		Type:   "1", // 1: 单曲
		Limit:  strconv.Itoa(options.Limit),
		Offset: strconv.Itoa(options.Offset),
	}

	code, response := searchService.Search()
	if code != 200 {
		return nil, fmt.Errorf("search failed with code: %f", code)
	}

	result := &plugin.SearchResult{
		Query:   query,
		Type:    options.Type,
		Tracks:  make([]core.Track, 0),
		Total:   0,
		Offset:  options.Offset,
		Limit:   options.Limit,
	}

	// 解析搜索结果
	if total, err := jsonparser.GetInt(response, "result", "songCount"); err == nil {
		result.Total = int(total)
	}

	// 解析歌曲列表
	_, err := jsonparser.ArrayEach(response, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		track, parseErr := p.parseTrackFromSearch(value)
		if parseErr == nil {
			result.Tracks = append(result.Tracks, *track)
		}
	}, "result", "songs")

	if err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	return result, nil
}

// searchAlbums 搜索专辑
func (p *NeteasePlugin) searchAlbums(ctx context.Context, query string, options plugin.SearchOptions) (*plugin.SearchResult, error) {
	searchService := &service.SearchService{
		S:      query,
		Type:   "10", // 10: 专辑
		Limit:  strconv.Itoa(options.Limit),
		Offset: strconv.Itoa(options.Offset),
	}

	code, response := searchService.Search()
	if code != 200 {
		return nil, fmt.Errorf("search albums failed with code: %f", code)
	}

	result := &plugin.SearchResult{
		Query:  query,
		Type:   options.Type,
		Albums: make([]core.Album, 0),
		Total:  0,
		Offset: options.Offset,
		Limit:  options.Limit,
	}

	// 解析专辑搜索结果
	if total, err := jsonparser.GetInt(response, "result", "albumCount"); err == nil {
		result.Total = int(total)
	}

	_, err := jsonparser.ArrayEach(response, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		album, parseErr := p.parseAlbumFromSearch(value)
		if parseErr == nil {
			result.Albums = append(result.Albums, *album)
		}
	}, "result", "albums")

	if err != nil {
		return nil, fmt.Errorf("failed to parse album search results: %w", err)
	}

	return result, nil
}

// searchArtists 搜索艺术家
func (p *NeteasePlugin) searchArtists(ctx context.Context, query string, options plugin.SearchOptions) (*plugin.SearchResult, error) {
	searchService := &service.SearchService{
		S:      query,
		Type:   "100", // 100: 歌手
		Limit:  strconv.Itoa(options.Limit),
		Offset: strconv.Itoa(options.Offset),
	}

	code, response := searchService.Search()
	if code != 200 {
		return nil, fmt.Errorf("search artists failed with code: %f", code)
	}

	result := &plugin.SearchResult{
		Query:   query,
		Type:    options.Type,
		Artists: make([]core.Artist, 0),
		Total:   0,
		Offset:  options.Offset,
		Limit:   options.Limit,
	}

	// 解析艺术家搜索结果
	if total, err := jsonparser.GetInt(response, "result", "artistCount"); err == nil {
		result.Total = int(total)
	}

	_, err := jsonparser.ArrayEach(response, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		artist, parseErr := p.parseArtistFromSearch(value)
		if parseErr == nil {
			result.Artists = append(result.Artists, *artist)
		}
	}, "result", "artists")

	if err != nil {
		return nil, fmt.Errorf("failed to parse artist search results: %w", err)
	}

	return result, nil
}

// searchPlaylists 搜索播放列表
func (p *NeteasePlugin) searchPlaylists(ctx context.Context, query string, options plugin.SearchOptions) (*plugin.SearchResult, error) {
	searchService := &service.SearchService{
		S:      query,
		Type:   "1000", // 1000: 歌单
		Limit:  strconv.Itoa(options.Limit),
		Offset: strconv.Itoa(options.Offset),
	}

	code, response := searchService.Search()
	if code != 200 {
		return nil, fmt.Errorf("search playlists failed with code: %f", code)
	}

	result := &plugin.SearchResult{
		Query:     query,
		Type:      options.Type,
		Playlists: make([]core.Playlist, 0),
		Total:     0,
		Offset:    options.Offset,
		Limit:     options.Limit,
	}

	// 解析播放列表搜索结果
	if total, err := jsonparser.GetInt(response, "result", "playlistCount"); err == nil {
		result.Total = int(total)
	}

	_, err := jsonparser.ArrayEach(response, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		playlist, parseErr := p.parsePlaylistFromSearch(value)
		if parseErr == nil {
			result.Playlists = append(result.Playlists, *playlist)
		}
	}, "result", "playlists")

	if err != nil {
		return nil, fmt.Errorf("failed to parse playlist search results: %w", err)
	}

	return result, nil
}

// parseTrackFromSearch 从搜索结果解析歌曲信息
func (p *NeteasePlugin) parseTrackFromSearch(data []byte) (*core.Track, error) {
	track := &core.Track{
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 解析基本信息
	if id, err := jsonparser.GetInt(data, "id"); err == nil {
		track.ID = strconv.FormatInt(id, 10)
		track.SourceID = strconv.FormatInt(id, 10)
	}

	if name, err := jsonparser.GetString(data, "name"); err == nil {
		track.Title = name
	}

	if duration, err := jsonparser.GetInt(data, "dt"); err == nil {
		track.Duration = time.Duration(duration) * time.Millisecond
	}

	// 解析艺术家信息
	artists := make([]string, 0)
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}
		if name, err := jsonparser.GetString(value, "name"); err == nil {
			artists = append(artists, name)
		}
	}, "ar")
	if len(artists) > 0 {
		track.Artist = artists[0]
	}

	// 解析专辑信息
	if albumName, err := jsonparser.GetString(data, "al", "name"); err == nil {
		track.Album = albumName
	}

	// AlbumID字段在core.Track中不存在，跳过

	if picURL, err := jsonparser.GetString(data, "al", "picUrl"); err == nil {
		track.CoverURL = picURL
	}

	return track, nil
}

// parseAlbumFromSearch 从搜索结果解析专辑信息
func (p *NeteasePlugin) parseAlbumFromSearch(data []byte) (*core.Album, error) {
	album := &core.Album{
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if id, err := jsonparser.GetInt(data, "id"); err == nil {
		album.ID = strconv.FormatInt(id, 10)
		album.SourceID = strconv.FormatInt(id, 10)
	}

	if name, err := jsonparser.GetString(data, "name"); err == nil {
		album.Title = name
	}

	if picURL, err := jsonparser.GetString(data, "picUrl"); err == nil {
		album.CoverURL = picURL
	}

	// 解析艺术家信息
	if artistName, err := jsonparser.GetString(data, "artist", "name"); err == nil {
		album.Artist = artistName
	}

	if publishTime, err := jsonparser.GetInt(data, "publishTime"); err == nil {
		releaseDate := time.Unix(publishTime/1000, 0)
		album.Year = releaseDate.Year()
	}

	return album, nil
}

// parseArtistFromSearch 从搜索结果解析艺术家信息
func (p *NeteasePlugin) parseArtistFromSearch(data []byte) (*core.Artist, error) {
	artist := &core.Artist{
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if id, err := jsonparser.GetInt(data, "id"); err == nil {
		artist.ID = strconv.FormatInt(id, 10)
		artist.SourceID = strconv.FormatInt(id, 10)
	}

	if name, err := jsonparser.GetString(data, "name"); err == nil {
		artist.Name = name
	}

	if picURL, err := jsonparser.GetString(data, "picUrl"); err == nil {
		artist.AvatarURL = picURL
	}

	// AlbumCount字段在core.Artist中不存在，跳过

	return artist, nil
}

// parsePlaylistFromSearch 从搜索结果解析播放列表信息
func (p *NeteasePlugin) parsePlaylistFromSearch(data []byte) (*core.Playlist, error) {
	playlist := &core.Playlist{
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if id, err := jsonparser.GetInt(data, "id"); err == nil {
		playlist.ID = strconv.FormatInt(id, 10)
		playlist.SourceID = strconv.FormatInt(id, 10)
	}

	if name, err := jsonparser.GetString(data, "name"); err == nil {
		playlist.Name = name
	}

	if description, err := jsonparser.GetString(data, "description"); err == nil {
		playlist.Description = description
	}

	if coverImgURL, err := jsonparser.GetString(data, "coverImgUrl"); err == nil {
		playlist.CoverURL = coverImgURL
	}

	if trackCount, err := jsonparser.GetInt(data, "trackCount"); err == nil {
		playlist.TrackCount = int(trackCount)
	}

	// PlayCount字段在core.Playlist中不存在，跳过

	// 解析创建者信息
	if creator, err := jsonparser.GetString(data, "creator", "nickname"); err == nil {
		playlist.Owner = creator
	}

	return playlist, nil
}