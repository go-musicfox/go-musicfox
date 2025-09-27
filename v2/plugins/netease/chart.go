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

// GetCharts 获取排行榜列表
func (p *NeteasePlugin) GetCharts(ctx context.Context) ([]*plugin.Chart, error) {
	toplistService := &service.ToplistService{}
	code, response := toplistService.Toplist()
	if code != 200 {
		return nil, fmt.Errorf("failed to get charts: code %f", code)
	}

	charts, err := p.parseCharts(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse charts: %w", err)
	}

	return charts, nil
}

// GetChartTracks 获取排行榜歌曲
func (p *NeteasePlugin) GetChartTracks(ctx context.Context, chartID string, offset, limit int) ([]*core.Track, error) {
	playlistDetail := &service.PlaylistDetailService{
		Id: chartID,
		S:  "0",
	}

	code, response := playlistDetail.PlaylistDetail()
	if code != 200 {
		return nil, fmt.Errorf("failed to get chart tracks: code %f", code)
	}

	tracks, err := p.parsePlaylistTracks(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chart tracks: %w", err)
	}

	// 应用分页
	if offset >= len(tracks) {
		return []*core.Track{}, nil
	}

	end := offset + limit
	if limit <= 0 || end > len(tracks) {
		end = len(tracks)
	}

	return tracks[offset:end], nil
}

// GetHotSearchKeywords 获取热搜关键词
func (p *NeteasePlugin) GetHotSearchKeywords(ctx context.Context) ([]string, error) {
	searchHotService := &service.SearchHotService{}
	code, response := searchHotService.SearchHot()
	if code != 200 {
		return nil, fmt.Errorf("failed to get hot search keywords: code %f", code)
	}

	keywords, err := p.parseHotSearchKeywords(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hot search keywords: %w", err)
	}

	return keywords, nil
}

// GetTopPlaylistsByCategory 根据分类获取热门歌单
func (p *NeteasePlugin) GetTopPlaylistsByCategory(ctx context.Context, category string, offset, limit int) ([]*core.Playlist, error) {
	topPlaylistService := &service.TopPlaylistService{
		Cat:    category,
		Order:  "hot", // hot: 热门, new: 最新
		Limit:  strconv.Itoa(limit),
		Offset: strconv.Itoa(offset),
	}

	code, response := topPlaylistService.TopPlaylist()
	if code != 200 {
		return nil, fmt.Errorf("failed to get top playlists: code %f", code)
	}

	playlists, err := p.parseTopPlaylists(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse top playlists: %w", err)
	}

	return playlists, nil
}

// GetPlaylistCategories 获取歌单分类
func (p *NeteasePlugin) GetPlaylistCategories(ctx context.Context) (map[string][]string, error) {
	playlistCatlistService := &service.PlaylistCatlistService{}
	code, response := playlistCatlistService.PlaylistCatlist()
	if code != 200 {
		return nil, fmt.Errorf("failed to get playlist categories: code %f", code)
	}

	categories, err := p.parsePlaylistCategories(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse playlist categories: %w", err)
	}

	return categories, nil
}

// GetArtistTopTracks 获取歌手热门歌曲
func (p *NeteasePlugin) GetArtistTopTracks(ctx context.Context, artistID string) ([]*core.Track, error) {
	artistTopSongService := &service.ArtistTopSongService{
		Id: artistID,
	}

	code, response := artistTopSongService.ArtistTopSong()
	if code != 200 {
		return nil, fmt.Errorf("failed to get artist top tracks: code %f", code)
	}

	tracks, err := p.parseArtistTopTracks(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse artist top tracks: %w", err)
	}

	return tracks, nil
}

// GetArtistAlbums 获取歌手专辑
func (p *NeteasePlugin) GetArtistAlbums(ctx context.Context, artistID string, offset, limit int) ([]*core.Album, error) {
	artistAlbumService := &service.ArtistAlbumService{
		ID:     artistID,
		Limit:  strconv.Itoa(limit),
		Offset: strconv.Itoa(offset),
	}

	code, response := artistAlbumService.ArtistAlbum()
	if code != 200 {
		return nil, fmt.Errorf("failed to get artist albums: code %f", code)
	}

	albums, err := p.parseArtistAlbums(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse artist albums: %w", err)
	}

	return albums, nil
}

// GetAlbumTracks 获取专辑歌曲
func (p *NeteasePlugin) GetAlbumTracks(ctx context.Context, albumID string) ([]*core.Track, error) {
	albumService := &service.AlbumService{
		ID: albumID,
	}

	code, response := albumService.Album()
	if code != 200 {
		return nil, fmt.Errorf("failed to get album tracks: code %f", code)
	}

	tracks, err := p.parseAlbumTracks(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse album tracks: %w", err)
	}

	return tracks, nil
}

// 解析函数

// parseCharts 解析排行榜列表
func (p *NeteasePlugin) parseCharts(data []byte) ([]*plugin.Chart, error) {
	charts := make([]*plugin.Chart, 0)

	// 解析官方榜
	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		chart, parseErr := p.parseChart(value)
		if parseErr == nil {
			charts = append(charts, chart)
		}
	}, "list")

	if err != nil {
		return nil, fmt.Errorf("failed to parse charts: %w", err)
	}

	return charts, nil
}

// parseChart 解析单个排行榜
func (p *NeteasePlugin) parseChart(data []byte) (*plugin.Chart, error) {
	chart := &plugin.Chart{}

	if id, err := jsonparser.GetInt(data, "id"); err == nil {
		chart.ID = strconv.FormatInt(id, 10)
	}

	if name, err := jsonparser.GetString(data, "name"); err == nil {
		chart.Name = name
	}

	if description, err := jsonparser.GetString(data, "description"); err == nil {
		chart.Description = description
	}

	if coverImgURL, err := jsonparser.GetString(data, "coverImgUrl"); err == nil {
		chart.CoverURL = coverImgURL
	}

	if updateTime, err := jsonparser.GetInt(data, "updateTime"); err == nil {
		chart.UpdateTime = time.Unix(updateTime/1000, 0)
	}

	// 设置周期
	chart.Period = "weekly" // 默认周榜

	return chart, nil
}

// parseHotSearchKeywords 解析热搜关键词
func (p *NeteasePlugin) parseHotSearchKeywords(data []byte) ([]string, error) {
	keywords := make([]string, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		if first, err := jsonparser.GetString(value, "first"); err == nil {
			keywords = append(keywords, first)
		}
	}, "result", "hots")

	if err != nil {
		return nil, fmt.Errorf("failed to parse hot search keywords: %w", err)
	}

	return keywords, nil
}

// parseTopPlaylists 解析热门歌单
func (p *NeteasePlugin) parseTopPlaylists(data []byte) ([]*core.Playlist, error) {
	playlists := make([]*core.Playlist, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		playlist, parseErr := p.parsePlaylistFromTop(value)
		if parseErr == nil {
			playlists = append(playlists, playlist)
		}
	}, "playlists")

	if err != nil {
		return nil, fmt.Errorf("failed to parse top playlists: %w", err)
	}

	return playlists, nil
}

// parsePlaylistFromTop 从热门歌单解析播放列表信息
func (p *NeteasePlugin) parsePlaylistFromTop(data []byte) (*core.Playlist, error) {
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

	if createTime, err := jsonparser.GetInt(data, "createTime"); err == nil {
		playlist.CreatedAt = time.Unix(createTime/1000, 0)
	}

	return playlist, nil
}

// parsePlaylistCategories 解析歌单分类
func (p *NeteasePlugin) parsePlaylistCategories(data []byte) (map[string][]string, error) {
	categories := make(map[string][]string)

	// 解析分类
	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		if categoryName, err := jsonparser.GetString(value, "name"); err == nil {
			if _, exists := categories[categoryName]; !exists {
				categories[categoryName] = make([]string, 0)
			}
		}
	}, "categories")

	// 解析子分类
	err = jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		keyStr := string(key)
		if keyStr == "sub" {
			_, subErr := jsonparser.ArrayEach(value, func(subValue []byte, subDataType jsonparser.ValueType, subOffset int, subErr error) {
				if subErr != nil {
					return
				}

				if subName, err := jsonparser.GetString(subValue, "name"); err == nil {
					if category, err := jsonparser.GetString(subValue, "category"); err == nil {
						if categoryList, exists := categories[category]; exists {
							categories[category] = append(categoryList, subName)
						} else {
							categories["其他"] = append(categories["其他"], subName)
						}
					}
				}
			})
			return subErr
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse playlist categories: %w", err)
	}

	return categories, nil
}

// parseArtistTopTracks 解析歌手热门歌曲
func (p *NeteasePlugin) parseArtistTopTracks(data []byte) ([]*core.Track, error) {
	tracks := make([]*core.Track, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		track, parseErr := p.parseTrackFromArtist(value)
		if parseErr == nil {
			tracks = append(tracks, track)
		}
	}, "songs")

	if err != nil {
		return nil, fmt.Errorf("failed to parse artist top tracks: %w", err)
	}

	return tracks, nil
}

// parseTrackFromArtist 从歌手歌曲解析歌曲信息
func (p *NeteasePlugin) parseTrackFromArtist(data []byte) (*core.Track, error) {
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

// parseArtistAlbums 解析歌手专辑
func (p *NeteasePlugin) parseArtistAlbums(data []byte) ([]*core.Album, error) {
	albums := make([]*core.Album, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		album, parseErr := p.parseAlbumFromArtist(value)
		if parseErr == nil {
			albums = append(albums, album)
		}
	}, "hotAlbums")

	if err != nil {
		return nil, fmt.Errorf("failed to parse artist albums: %w", err)
	}

	return albums, nil
}

// parseAlbumFromArtist 从歌手专辑解析专辑信息
func (p *NeteasePlugin) parseAlbumFromArtist(data []byte) (*core.Album, error) {
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

// parseAlbumTracks 解析专辑歌曲
func (p *NeteasePlugin) parseAlbumTracks(data []byte) ([]*core.Track, error) {
	tracks := make([]*core.Track, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		track, parseErr := p.parseTrackFromAlbum(value)
		if parseErr == nil {
			tracks = append(tracks, track)
		}
	}, "songs")

	if err != nil {
		return nil, fmt.Errorf("failed to parse album tracks: %w", err)
	}

	return tracks, nil
}

// parseTrackFromAlbum 从专辑歌曲解析歌曲信息
func (p *NeteasePlugin) parseTrackFromAlbum(data []byte) (*core.Track, error) {
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