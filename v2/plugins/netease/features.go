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

// GetRecommendations 获取推荐歌曲
func (p *NeteasePlugin) GetRecommendations(ctx context.Context, options plugin.RecommendationOptions) ([]*core.Track, error) {
	if !p.IsLoggedIn() {
		return nil, fmt.Errorf("user not logged in")
	}

	recommendService := &service.RecommendResourceService{}
	code, response := recommendService.RecommendResource()
	if code != 200 {
		return nil, fmt.Errorf("failed to get recommendations: code %f", code)
	}

	tracks, err := p.parseRecommendationTracks(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recommendation tracks: %w", err)
	}

	// 应用限制
	if options.Limit > 0 && len(tracks) > options.Limit {
		tracks = tracks[:options.Limit]
	}

	return tracks, nil
}

// GetDailyRecommendations 获取每日推荐
func (p *NeteasePlugin) GetDailyRecommendations(ctx context.Context) ([]*core.Track, error) {
	if !p.IsLoggedIn() {
		return nil, fmt.Errorf("user not logged in")
	}

	recommendSongs := &service.RecommendSongsService{}
	code, response := recommendSongs.RecommendSongs()
	if code != 200 {
		return nil, fmt.Errorf("failed to get daily recommendations: code %f", code)
	}

	tracks, err := p.parseDailyRecommendations(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse daily recommendations: %w", err)
	}

	return tracks, nil
}

// GetSimilarTracks 获取相似歌曲
func (p *NeteasePlugin) GetSimilarTracks(ctx context.Context, trackID string, limit int) ([]*core.Track, error) {
	simiSongService := &service.SimiSongService{
		ID: trackID,
	}

	code, response := simiSongService.SimiSong()
	if code != 200 {
		return nil, fmt.Errorf("failed to get similar tracks: code %f", code)
	}

	tracks, err := p.parseSimilarTracks(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse similar tracks: %w", err)
	}

	// 应用限制
	if limit > 0 && len(tracks) > limit {
		tracks = tracks[:limit]
	}

	return tracks, nil
}



// GetAlbum 获取专辑信息
func (p *NeteasePlugin) GetAlbum(ctx context.Context, albumID string) (*core.Album, error) {
	albumService := &service.AlbumService{
		ID: albumID,
	}

	code, response := albumService.Album()
	if code != 200 {
		return nil, fmt.Errorf("failed to get album: code %f", code)
	}

	album, err := p.parseAlbum(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse album: %w", err)
	}

	return album, nil
}



// GetArtist 获取艺术家信息
func (p *NeteasePlugin) GetArtist(ctx context.Context, artistID string) (*core.Artist, error) {
	artistService := &service.ArtistsService{
		ID: artistID,
	}

	code, response := artistService.Artists()
	if code != 200 {
		return nil, fmt.Errorf("failed to get artist: code %f", code)
	}

	artist, err := p.parseArtist(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse artist: %w", err)
	}

	return artist, nil
}

// GetArtistTracks 获取艺术家歌曲
func (p *NeteasePlugin) GetArtistTracks(ctx context.Context, artistID string, offset, limit int) ([]*core.Track, error) {
	artistService := &service.ArtistsService{
		ID: artistID,
	}

	code, response := artistService.Artists()
	if code != 200 {
		return nil, fmt.Errorf("failed to get artist tracks: code %f", code)
	}

	tracks, err := p.parseArtistTracks(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse artist tracks: %w", err)
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



// GetRadioStations 获取电台列表
func (p *NeteasePlugin) GetRadioStations(ctx context.Context) ([]*plugin.RadioStation, error) {
	djRadioService := &service.DjRadioHotService{
		Limit:  "30",
		Offset: "0",
	}

	code, response := djRadioService.DjRadioHot()
	if code != 200 {
		return nil, fmt.Errorf("failed to get radio stations: code %f", code)
	}

	stations, err := p.parseRadioStations(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse radio stations: %w", err)
	}

	return stations, nil
}

// GetRadioTracks 获取电台节目
func (p *NeteasePlugin) GetRadioTracks(ctx context.Context, stationID string) ([]*core.Track, error) {
	djProgramService := &service.DjProgramService{
		RID:    stationID,
		Limit:  "30",
		Offset: "0",
	}

	code, response := djProgramService.DjProgram()
	if code != 200 {
		return nil, fmt.Errorf("failed to get radio tracks: code %f", code)
	}

	tracks, err := p.parseRadioTracks(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse radio tracks: %w", err)
	}

	return tracks, nil
}

// 解析函数

// parseRecommendationTracks 解析推荐歌曲
func (p *NeteasePlugin) parseRecommendationTracks(data []byte) ([]*core.Track, error) {
	tracks := make([]*core.Track, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		// 从推荐资源中解析歌曲
		_, _ = jsonparser.ArrayEach(value, func(trackValue []byte, dataType jsonparser.ValueType, offset int, err error) {
			if err != nil {
				return
			}

			track, parseErr := p.parseTrackFromSearch(trackValue)
			if parseErr == nil {
				tracks = append(tracks, track)
			}
		}, "creatives")
	}, "recommend")

	return tracks, err
}

// parseDailyRecommendations 解析每日推荐
func (p *NeteasePlugin) parseDailyRecommendations(data []byte) ([]*core.Track, error) {
	tracks := make([]*core.Track, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		track, parseErr := p.parseTrackFromSearch(value)
		if parseErr == nil {
			tracks = append(tracks, track)
		}
	}, "data", "dailySongs")

	if err != nil {
		// 尝试另一种格式
		_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if err != nil {
				return
			}

			track, parseErr := p.parseTrackFromSearch(value)
			if parseErr == nil {
				tracks = append(tracks, track)
			}
		}, "recommend")
	}

	return tracks, err
}

// parseSimilarTracks 解析相似歌曲
func (p *NeteasePlugin) parseSimilarTracks(data []byte) ([]*core.Track, error) {
	tracks := make([]*core.Track, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		track, parseErr := p.parseTrackFromSearch(value)
		if parseErr == nil {
			tracks = append(tracks, track)
		}
	}, "songs")

	return tracks, err
}



// parseAlbum 解析专辑信息
func (p *NeteasePlugin) parseAlbum(data []byte) (*core.Album, error) {
	album := &core.Album{
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if id, err := jsonparser.GetInt(data, "album", "id"); err == nil {
		album.ID = strconv.FormatInt(id, 10)
		album.SourceID = strconv.FormatInt(id, 10)
	}

	if name, err := jsonparser.GetString(data, "album", "name"); err == nil {
		album.Title = name
	}

	if picURL, err := jsonparser.GetString(data, "album", "picUrl"); err == nil {
		album.CoverURL = picURL
	}

	if publishTime, err := jsonparser.GetInt(data, "album", "publishTime"); err == nil {
		album.Year = time.Unix(publishTime/1000, 0).Year()
	}

	// 解析艺术家
	artists := make([]string, 0)
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}
		if name, err := jsonparser.GetString(value, "name"); err == nil {
			artists = append(artists, name)
		}
	}, "album", "artists")
	if len(artists) > 0 {
		album.Artist = artists[0]
	}

	return album, nil
}



// parseArtist 解析艺术家信息
func (p *NeteasePlugin) parseArtist(data []byte) (*core.Artist, error) {
	artist := &core.Artist{
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if id, err := jsonparser.GetInt(data, "artist", "id"); err == nil {
		artist.ID = strconv.FormatInt(id, 10)
		artist.SourceID = strconv.FormatInt(id, 10)
	}

	if name, err := jsonparser.GetString(data, "artist", "name"); err == nil {
		artist.Name = name
	}

	if picURL, err := jsonparser.GetString(data, "artist", "picUrl"); err == nil {
		artist.AvatarURL = picURL
	}

	// 注意：Albums和Tracks字段是切片类型，这里暂时不设置
	// 如果需要获取具体的专辑和歌曲列表，需要调用其他API

	return artist, nil
}

// parseArtistTracks 解析艺术家歌曲
func (p *NeteasePlugin) parseArtistTracks(data []byte) ([]*core.Track, error) {
	tracks := make([]*core.Track, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		track, parseErr := p.parseTrackFromSearch(value)
		if parseErr == nil {
			tracks = append(tracks, track)
		}
	}, "hotSongs")

	return tracks, err
}



// parseRadioStations 解析电台列表
func (p *NeteasePlugin) parseRadioStations(data []byte) ([]*plugin.RadioStation, error) {
	stations := make([]*plugin.RadioStation, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		station, parseErr := p.parseRadioStation(value)
		if parseErr == nil {
			stations = append(stations, station)
		}
	}, "djRadios")

	return stations, err
}

// parseRadioStation 解析单个电台
func (p *NeteasePlugin) parseRadioStation(data []byte) (*plugin.RadioStation, error) {
	station := &plugin.RadioStation{}

	if id, err := jsonparser.GetInt(data, "id"); err == nil {
		station.ID = strconv.FormatInt(id, 10)
	}

	if name, err := jsonparser.GetString(data, "name"); err == nil {
		station.Name = name
	}

	if desc, err := jsonparser.GetString(data, "desc"); err == nil {
		station.Description = desc
	}

	if picURL, err := jsonparser.GetString(data, "picUrl"); err == nil {
		station.CoverURL = picURL
	}

	if createTime, err := jsonparser.GetInt(data, "createTime"); err == nil {
		station.CreatedAt = time.Unix(createTime/1000, 0)
	}

	return station, nil
}

// parseRadioTracks 解析电台节目
func (p *NeteasePlugin) parseRadioTracks(data []byte) ([]*core.Track, error) {
	tracks := make([]*core.Track, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		track, parseErr := p.parseRadioTrack(value)
		if parseErr == nil {
			tracks = append(tracks, track)
		}
	}, "programs")

	return tracks, err
}

// parseRadioTrack 解析电台节目歌曲
func (p *NeteasePlugin) parseRadioTrack(data []byte) (*core.Track, error) {
	track := &core.Track{
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if id, err := jsonparser.GetInt(data, "mainSong", "id"); err == nil {
		track.ID = strconv.FormatInt(id, 10)
		track.SourceID = strconv.FormatInt(id, 10)
	}

	if name, err := jsonparser.GetString(data, "mainSong", "name"); err == nil {
		track.Title = name
	}

	if duration, err := jsonparser.GetInt(data, "mainSong", "duration"); err == nil {
		track.Duration = time.Duration(duration) * time.Millisecond
	}

	// 电台节目的艺术家通常是DJ
	if djName, err := jsonparser.GetString(data, "dj", "nickname"); err == nil {
		track.Artist = djName
	}

	return track, nil
}