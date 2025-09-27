package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)



// GetPersonalizedRecommendations 获取个性化推荐
func (p *NeteasePlugin) GetPersonalizedRecommendations(ctx context.Context, limit int) ([]*core.Playlist, error) {
	personalized := &service.PersonalizedService{
		Limit: strconv.Itoa(limit),
	}

	code, response := personalized.Personalized()
	if code != 200 {
		return nil, fmt.Errorf("failed to get personalized recommendations: code %f", code)
	}

	playlists, err := p.parsePersonalizedRecommendations(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse personalized recommendations: %w", err)
	}

	return playlists, nil
}

// GetRecommendedPlaylists 获取推荐歌单
func (p *NeteasePlugin) GetRecommendedPlaylists(ctx context.Context, category string, limit int) ([]*core.Playlist, error) {
	topPlaylistService := &service.TopPlaylistService{
		Cat:    category,
		Order:  "hot",
		Limit:  strconv.Itoa(limit),
		Offset: "0",
	}

	code, response := topPlaylistService.TopPlaylist()
	if code != 200 {
		return nil, fmt.Errorf("failed to get recommended playlists: code %f", code)
	}

	playlists, err := p.parseRecommendedPlaylists(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recommended playlists: %w", err)
	}

	return playlists, nil
}

// GetNewSongs 获取新歌速递
func (p *NeteasePlugin) GetNewSongs(ctx context.Context, areaType int) ([]*core.Track, error) {
	// TODO: 实现新歌速递功能
	// 暂时返回空结果
	return []*core.Track{}, nil
}

// GetTopArtists 获取热门歌手
func (p *NeteasePlugin) GetTopArtists(ctx context.Context, areaType, limit int) ([]*core.Artist, error) {
	// TODO: 实现热门歌手功能
	// 暂时返回空结果
	return []*core.Artist{}, nil
}

// GetTopAlbums 获取新碟上架
func (p *NeteasePlugin) GetTopAlbums(ctx context.Context, areaType, limit int) ([]*core.Album, error) {
	topAlbum := &service.TopAlbumService{
		Type:   strconv.Itoa(areaType), // ALL: 全部, ZH: 华语, EA: 欧美, KR: 韩国, JP: 日本
		Year:   "",
		Month:  "",
		Limit:  strconv.Itoa(limit),
		Offset: "0",
	}

	code, response := topAlbum.TopAlbum()
	if code != 200 {
		return nil, fmt.Errorf("failed to get top albums: code %f", code)
	}

	albums, err := p.parseTopAlbums(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse top albums: %w", err)
	}

	return albums, nil
}

// 解析函数



// parsePersonalizedRecommendations 解析个性化推荐
func (p *NeteasePlugin) parsePersonalizedRecommendations(data []byte) ([]*core.Playlist, error) {
	playlists := make([]*core.Playlist, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		playlist, parseErr := p.parsePlaylistFromRecommend(value)
		if parseErr == nil {
			playlists = append(playlists, playlist)
		}
	}, "result")

	if err != nil {
		return nil, fmt.Errorf("failed to parse personalized recommendations: %w", err)
	}

	return playlists, nil
}

// parseRecommendedPlaylists 解析推荐歌单
func (p *NeteasePlugin) parseRecommendedPlaylists(data []byte) ([]*core.Playlist, error) {
	playlists := make([]*core.Playlist, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		playlist, parseErr := p.parsePlaylistFromHighquality(value)
		if parseErr == nil {
			playlists = append(playlists, playlist)
		}
	}, "playlists")

	if err != nil {
		return nil, fmt.Errorf("failed to parse recommended playlists: %w", err)
	}

	return playlists, nil
}

// parseNewSongs 解析新歌速递
func (p *NeteasePlugin) parseNewSongs(data []byte) ([]*core.Track, error) {
	tracks := make([]*core.Track, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		track, parseErr := p.parseTrackFromNewSongs(value)
		if parseErr == nil {
			tracks = append(tracks, track)
		}
	}, "data")

	if err != nil {
		return nil, fmt.Errorf("failed to parse new songs: %w", err)
	}

	return tracks, nil
}

// parseTopArtists 解析热门歌手
func (p *NeteasePlugin) parseTopArtists(data []byte) ([]*core.Artist, error) {
	artists := make([]*core.Artist, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		artist, parseErr := p.parseArtistFromTop(value)
		if parseErr == nil {
			artists = append(artists, artist)
		}
	}, "artists")

	if err != nil {
		return nil, fmt.Errorf("failed to parse top artists: %w", err)
	}

	return artists, nil
}

// parseTopAlbums 解析新碟上架
func (p *NeteasePlugin) parseTopAlbums(data []byte) ([]*core.Album, error) {
	albums := make([]*core.Album, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		album, parseErr := p.parseAlbumFromTop(value)
		if parseErr == nil {
			albums = append(albums, album)
		}
	}, "albums")

	if err != nil {
		return nil, fmt.Errorf("failed to parse top albums: %w", err)
	}

	return albums, nil
}

// 具体解析函数

// parseTrackFromRecommend 从推荐结果解析歌曲信息
func (p *NeteasePlugin) parseTrackFromRecommend(data []byte) (*core.Track, error) {
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

	if picURL, err := jsonparser.GetString(data, "al", "picUrl"); err == nil {
		track.CoverURL = picURL
	}

	return track, nil
}

// parsePlaylistFromRecommend 从推荐结果解析播放列表信息
func (p *NeteasePlugin) parsePlaylistFromRecommend(data []byte) (*core.Playlist, error) {
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

	if picURL, err := jsonparser.GetString(data, "picUrl"); err == nil {
		playlist.CoverURL = picURL
	}

	if trackCount, err := jsonparser.GetInt(data, "trackCount"); err == nil {
		playlist.TrackCount = int(trackCount)
	}

	return playlist, nil
}

// parsePlaylistFromHighquality 从高质量歌单解析播放列表信息
func (p *NeteasePlugin) parsePlaylistFromHighquality(data []byte) (*core.Playlist, error) {
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

	// 解析创建者信息
	if creator, err := jsonparser.GetString(data, "creator", "nickname"); err == nil {
		playlist.Owner = creator
	}

	return playlist, nil
}

// parseTrackFromNewSongs 从新歌速递解析歌曲信息
func (p *NeteasePlugin) parseTrackFromNewSongs(data []byte) (*core.Track, error) {
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

	if duration, err := jsonparser.GetInt(data, "duration"); err == nil {
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
	}, "artists")
	if len(artists) > 0 {
		track.Artist = artists[0]
	}

	// 解析专辑信息
	if albumName, err := jsonparser.GetString(data, "album", "name"); err == nil {
		track.Album = albumName
	}

	if picURL, err := jsonparser.GetString(data, "album", "picUrl"); err == nil {
		track.CoverURL = picURL
	}

	return track, nil
}

// parseArtistFromTop 从热门歌手解析艺术家信息
func (p *NeteasePlugin) parseArtistFromTop(data []byte) (*core.Artist, error) {
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

	return artist, nil
}

// parseAlbumFromTop 从新碟上架解析专辑信息
func (p *NeteasePlugin) parseAlbumFromTop(data []byte) (*core.Album, error) {
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