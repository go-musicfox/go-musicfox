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

// GetPlaylist 获取播放列表信息
func (p *NeteasePlugin) GetPlaylist(ctx context.Context, playlistID string) (*core.Playlist, error) {
	playlistDetail := &service.PlaylistDetailService{
		Id: playlistID,
		S:  "0", // 最近S个收藏者，设为0
	}

	code, response := playlistDetail.PlaylistDetail()
	if code != 200 {
		return nil, fmt.Errorf("failed to get playlist detail: code %f", code)
	}

	playlist, err := p.parsePlaylistDetail(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse playlist detail: %w", err)
	}

	return playlist, nil
}

// GetPlaylistTracks 获取播放列表中的歌曲
func (p *NeteasePlugin) GetPlaylistTracks(ctx context.Context, playlistID string, offset, limit int) ([]*core.Track, error) {
	// 如果需要获取所有歌曲，使用AllTracks接口
	if limit <= 0 || limit > 1000 {
		allTrack := &service.PlaylistTrackAllService{
			Id: playlistID,
			S:  "0",
		}
		code, response := allTrack.AllTracks()
		if code != 200 {
			return nil, fmt.Errorf("failed to get all playlist tracks: code %f", code)
		}
		return p.parsePlaylistTracks(response)
	}

	// 使用普通的播放列表详情接口
	playlistDetail := &service.PlaylistDetailService{
		Id: playlistID,
		S:  "0",
	}

	code, response := playlistDetail.PlaylistDetail()
	if code != 200 {
		return nil, fmt.Errorf("failed to get playlist detail: code %f", code)
	}

	tracks, err := p.parsePlaylistTracks(response)
	if err != nil {
		return nil, err
	}

	// 应用分页
	if offset >= len(tracks) {
		return []*core.Track{}, nil
	}

	end := offset + limit
	if end > len(tracks) {
		end = len(tracks)
	}

	return tracks[offset:end], nil
}

// CreatePlaylist 创建播放列表
func (p *NeteasePlugin) CreatePlaylist(ctx context.Context, name, description string) (*core.Playlist, error) {
	p.mu.RLock()
	user := p.user
	p.mu.RUnlock()

	if user == nil {
		return nil, fmt.Errorf("user not logged in")
	}

	createService := &service.PlaylistCreateService{
		Name:    name,
		Privacy: "0", // 公开
	}

	code, response := createService.PlaylistCreate()
	if code != 200 {
		return nil, fmt.Errorf("failed to create playlist: code %f", code)
	}

	// 解析创建的播放列表信息
	playlist, err := p.parseCreatedPlaylist(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created playlist: %w", err)
	}

	return playlist, nil
}

// UpdatePlaylist 更新播放列表
func (p *NeteasePlugin) UpdatePlaylist(ctx context.Context, playlistID string, updates map[string]interface{}) error {
	p.mu.RLock()
	user := p.user
	p.mu.RUnlock()

	if user == nil {
		return fmt.Errorf("user not logged in")
	}

	// 检查更新字段
	name, hasName := updates["name"].(string)
	desc, hasDesc := updates["description"].(string)

	if hasName || hasDesc {
		updateService := &service.PlaylistUpdateService{
			Id:   playlistID,
			Name: name,
			Desc: desc,
		}

		code, _ := updateService.PlaylistUpdate()
		if code != 200 {
			return fmt.Errorf("failed to update playlist: code %f", code)
		}
	}

	return nil
}

// DeletePlaylist 删除播放列表
func (p *NeteasePlugin) DeletePlaylist(ctx context.Context, playlistID string) error {
	p.mu.RLock()
	user := p.user
	p.mu.RUnlock()

	if user == nil {
		return fmt.Errorf("user not logged in")
	}

	deleteService := &service.PlaylistDeleteService{
		ID: playlistID,
	}

	code, _ := deleteService.PlaylistDelete()
	if code != 200 {
		return fmt.Errorf("failed to delete playlist: code %f", code)
	}

	return nil
}

// GetUserPlaylists 获取用户的播放列表
func (p *NeteasePlugin) GetUserPlaylists(ctx context.Context, userID string) ([]*core.Playlist, error) {
	userPlaylists := &service.UserPlaylistService{
		Uid:    userID,
		Limit:  "50", // 默认获取50个
		Offset: "0",
	}

	code, response := userPlaylists.UserPlaylist()
	if code != 200 {
		return nil, fmt.Errorf("failed to get user playlists: code %f", code)
	}

	playlists, err := p.parseUserPlaylists(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user playlists: %w", err)
	}

	return playlists, nil
}

// parsePlaylistDetail 解析播放列表详情
func (p *NeteasePlugin) parsePlaylistDetail(data []byte) (*core.Playlist, error) {
	playlist := &core.Playlist{
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 解析播放列表基本信息
	if id, err := jsonparser.GetInt(data, "playlist", "id"); err == nil {
		playlist.ID = strconv.FormatInt(id, 10)
		playlist.SourceID = strconv.FormatInt(id, 10)
	}

	if name, err := jsonparser.GetString(data, "playlist", "name"); err == nil {
		playlist.Name = name
	}

	if description, err := jsonparser.GetString(data, "playlist", "description"); err == nil {
		playlist.Description = description
	}

	if coverImgURL, err := jsonparser.GetString(data, "playlist", "coverImgUrl"); err == nil {
		playlist.CoverURL = coverImgURL
	}

	if trackCount, err := jsonparser.GetInt(data, "playlist", "trackCount"); err == nil {
		playlist.TrackCount = int(trackCount)
	}

	// PlayCount字段在core.Playlist中不存在，跳过

	if createTime, err := jsonparser.GetInt(data, "playlist", "createTime"); err == nil {
		playlist.CreatedAt = time.Unix(createTime/1000, 0)
	}

	if updateTime, err := jsonparser.GetInt(data, "playlist", "updateTime"); err == nil {
		playlist.UpdatedAt = time.Unix(updateTime/1000, 0)
	}

	// 解析创建者信息
	if creatorName, err := jsonparser.GetString(data, "playlist", "creator", "nickname"); err == nil {
		playlist.Owner = creatorName
	}

	// CreatorID字段在core.Playlist中不存在，跳过

	return playlist, nil
}

// parsePlaylistTracks 解析播放列表中的歌曲
func (p *NeteasePlugin) parsePlaylistTracks(data []byte) ([]*core.Track, error) {
	tracks := make([]*core.Track, 0)

	// 尝试从不同的路径解析歌曲列表
	paths := [][]string{
		{"playlist", "tracks"},
		{"songs"},
		{"playlist", "trackIds"}, // 如果只有ID，需要额外获取详情
	}

	for _, path := range paths {
		_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if err != nil {
				return
			}

			track, parseErr := p.parseTrackFromPlaylist(value)
			if parseErr == nil {
				tracks = append(tracks, track)
			}
		}, path...)

		if err == nil && len(tracks) > 0 {
			break
		}
	}

	return tracks, nil
}

// parseTrackFromPlaylist 从播放列表解析歌曲信息
func (p *NeteasePlugin) parseTrackFromPlaylist(data []byte) (*core.Track, error) {
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
	} else if duration, err := jsonparser.GetInt(data, "duration"); err == nil {
		track.Duration = time.Duration(duration) * time.Millisecond
	}

	// 解析艺术家信息
	artists := make([]string, 0)
	artistPaths := [][]string{{"ar"}, {"artists"}}
	for _, path := range artistPaths {
		_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if err != nil {
				return
			}
			if name, err := jsonparser.GetString(value, "name"); err == nil {
				artists = append(artists, name)
			}
		}, path...)
		if err == nil && len(artists) > 0 {
			break
		}
	}
	if len(artists) > 0 {
		track.Artist = artists[0] // core.Track.Artist是string类型，取第一个艺术家
	}

	// 解析专辑信息
	albumPaths := [][]string{{"al"}, {"album"}}
	for _, path := range albumPaths {
		if albumName, err := jsonparser.GetString(data, append(path, "name")...); err == nil {
			track.Album = albumName
			break
		}
	}

	// AlbumID字段在core.Track中不存在，跳过
	// for _, path := range albumPaths {
	//	if albumID, err := jsonparser.GetInt(data, append(path, "id")...); err == nil {
	//		track.AlbumID = strconv.FormatInt(albumID, 10)
	//		break
	//	}
	// }

	for _, path := range albumPaths {
		if picURL, err := jsonparser.GetString(data, append(path, "picUrl")...); err == nil {
			track.CoverURL = picURL
			break
		}
	}

	return track, nil
}

// parseUserPlaylists 解析用户播放列表
func (p *NeteasePlugin) parseUserPlaylists(data []byte) ([]*core.Playlist, error) {
	playlists := make([]*core.Playlist, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		playlist, parseErr := p.parseUserPlaylist(value)
		if parseErr == nil {
			playlists = append(playlists, playlist)
		}
	}, "playlist")

	if err != nil {
		return nil, fmt.Errorf("failed to parse user playlists: %w", err)
	}

	return playlists, nil
}

// parseUserPlaylist 解析用户播放列表项
func (p *NeteasePlugin) parseUserPlaylist(data []byte) (*core.Playlist, error) {
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
	// if playCount, err := jsonparser.GetInt(data, "playCount"); err == nil {
	//	playlist.PlayCount = playCount
	// }

	if createTime, err := jsonparser.GetInt(data, "createTime"); err == nil {
		playlist.CreatedAt = time.Unix(createTime/1000, 0)
	}

	if updateTime, err := jsonparser.GetInt(data, "updateTime"); err == nil {
		playlist.UpdatedAt = time.Unix(updateTime/1000, 0)
	}

	// 解析创建者信息
	if creatorName, err := jsonparser.GetString(data, "creator", "nickname"); err == nil {
		playlist.Owner = creatorName
	}

	// CreatorID字段在core.Playlist中不存在，跳过
	// if creatorID, err := jsonparser.GetInt(data, "creator", "userId"); err == nil {
	//	playlist.CreatorID = strconv.FormatInt(creatorID, 10)
	// }

	// 检查是否为私有播放列表
	if privacy, err := jsonparser.GetInt(data, "privacy"); err == nil {
		playlist.Public = privacy == 0
	}

	return playlist, nil
}

// parseCreatedPlaylist 解析创建的播放列表
func (p *NeteasePlugin) parseCreatedPlaylist(data []byte) (*core.Playlist, error) {
	playlist := &core.Playlist{
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if id, err := jsonparser.GetInt(data, "id"); err == nil {
		playlist.ID = strconv.FormatInt(id, 10)
		playlist.SourceID = strconv.FormatInt(id, 10)
	} else if id, err := jsonparser.GetInt(data, "playlist", "id"); err == nil {
		playlist.ID = strconv.FormatInt(id, 10)
		playlist.SourceID = strconv.FormatInt(id, 10)
	}

	if name, err := jsonparser.GetString(data, "playlist", "name"); err == nil {
		playlist.Name = name
	}

	return playlist, nil
}