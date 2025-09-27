package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
	plugin "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/music"
)

// FollowUser 关注用户
func (p *NeteasePlugin) FollowUser(ctx context.Context, userID string) error {
	if !p.IsLoggedIn() {
		return fmt.Errorf("user not logged in")
	}

	followService := &service.FollowService{
		Id: userID,
		T:  "1", // 1: 关注, 0: 取消关注
	}

	code, _ := followService.Follow()
	if code != 200 {
		return fmt.Errorf("failed to follow user: code %f", code)
	}

	return nil
}

// UnfollowUser 取消关注用户
func (p *NeteasePlugin) UnfollowUser(ctx context.Context, userID string) error {
	if !p.IsLoggedIn() {
		return fmt.Errorf("user not logged in")
	}

	followService := &service.FollowService{
		Id: userID,
		T:  "0", // 0: 取消关注
	}

	code, _ := followService.Follow()
	if code != 200 {
		return fmt.Errorf("failed to unfollow user: code %f", code)
	}

	return nil
}

// GetUserInfo 获取用户信息
func (p *NeteasePlugin) GetUserInfo(ctx context.Context, userID string) (*plugin.UserInfo, error) {
	userDetailService := &service.UserDetailService{
		Uid: userID,
	}

	code, response := userDetailService.UserDetail()
	if code != 200 {
		return nil, fmt.Errorf("failed to get user info: code %f", code)
	}

	userInfo, err := p.parseUserDetail(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return userInfo, nil
}

// GetUserLikedTracks 获取用户喜欢的歌曲
func (p *NeteasePlugin) GetUserLikedTracks(ctx context.Context, userID string) ([]*core.Track, error) {
	// 首先获取用户的播放列表，找到"我喜欢的音乐"播放列表
	userPlaylists, err := p.GetUserPlaylists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user playlists: %w", err)
	}

	if len(userPlaylists) == 0 {
		return []*core.Track{}, nil
	}

	// 第一个播放列表通常是"我喜欢的音乐"
	likedPlaylistID := userPlaylists[0].ID
	tracks, err := p.GetPlaylistTracks(ctx, likedPlaylistID, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get liked tracks: %w", err)
	}

	return tracks, nil
}

// parseTrackFromPlayHistory 从播放历史解析歌曲信息
func (p *NeteasePlugin) parseTrackFromPlayHistory(data []byte) (*core.Track, error) {
	// 从song字段解析歌曲信息
	_, err := jsonparser.GetString(data, "song")
	if err != nil {
		return nil, fmt.Errorf("failed to get song data: %w", err)
	}
	
	// 这里简化处理，实际应该解析完整的歌曲信息
	track := &core.Track{
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	return track, nil
}

// parseTrackFromCloud 从云盘解析歌曲信息
func (p *NeteasePlugin) parseTrackFromCloud(data []byte) (*core.Track, error) {
	// 从simpleSong字段解析歌曲信息
	_, err := jsonparser.GetString(data, "simpleSong")
	if err != nil {
		return nil, fmt.Errorf("failed to get simple song data: %w", err)
	}
	
	// 这里简化处理，实际应该解析完整的歌曲信息
	track := &core.Track{
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	return track, nil
}

// GetUserFollowing 获取用户关注列表
func (p *NeteasePlugin) GetUserFollowing(ctx context.Context, userID string) ([]string, error) {
	userFollowsService := &service.UserFollowsService{
		Uid:    userID,
		Limit:  "30",
		Offset: "0",
	}

	code, response := userFollowsService.UserFollows()
	if code != 200 {
		return nil, fmt.Errorf("failed to get user following: code %f", code)
	}

	following, err := p.parseUserFollowing(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user following: %w", err)
	}

	return following, nil
}

// GetUserFollowers 获取用户粉丝列表
func (p *NeteasePlugin) GetUserFollowers(ctx context.Context, userID string) ([]string, error) {
	userFollowedsService := &service.UserFollowedsService{
		Uid:   userID,
		Limit: "30",
	}

	code, response := userFollowedsService.UserFolloweds()
	if code != 200 {
		return nil, fmt.Errorf("failed to get user followers: code %f", code)
	}

	followers, err := p.parseUserFollowers(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user followers: %w", err)
	}

	return followers, nil
}

// ClearCache 清除缓存
func (p *NeteasePlugin) ClearCache() error {
	// 清除cookie缓存
	if p.cookieJar != nil {
		u := &url.URL{Scheme: "https", Host: "music.163.com"}
		p.cookieJar.SetCookies(u, []*http.Cookie{})
	}

	// 如果有其他缓存，也在这里清除
	// TODO: 实现更完整的缓存清除逻辑

	return nil
}

// GetCacheSize 获取缓存大小
func (p *NeteasePlugin) GetCacheSize() int64 {
	// TODO: 实现缓存大小计算
	// 这里返回一个估算值
	return 0
}

// 解析函数

// parseUserDetail 解析用户详细信息
func (p *NeteasePlugin) parseUserDetail(data []byte) (*plugin.UserInfo, error) {
	userInfo := &plugin.UserInfo{
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 尝试从不同路径解析用户信息
	paths := [][]string{
		{"profile"},
		{"data", "profile"},
	}

	for _, path := range paths {
		if userID, err := jsonparser.GetInt(data, append(path, "userId")...); err == nil {
			userInfo.ID = strconv.FormatInt(userID, 10)
			userInfo.SourceID = strconv.FormatInt(userID, 10)
			break
		}
	}

	for _, path := range paths {
		if nickname, err := jsonparser.GetString(data, append(path, "nickname")...); err == nil {
			userInfo.Username = nickname
			userInfo.Nickname = nickname
			break
		}
	}

	for _, path := range paths {
		if avatarURL, err := jsonparser.GetString(data, append(path, "avatarUrl")...); err == nil {
			userInfo.AvatarURL = avatarURL
			break
		}
	}

	for _, path := range paths {
		if description, err := jsonparser.GetString(data, append(path, "description")...); err == nil {
			userInfo.Description = description
			break
		}
	}

	for _, path := range paths {
		if level, err := jsonparser.GetInt(data, append(path, "level")...); err == nil {
			userInfo.Level = int(level)
			break
		}
	}

	for _, path := range paths {
		if followeds, err := jsonparser.GetInt(data, append(path, "followeds")...); err == nil {
			userInfo.FanCount = followeds
			break
		}
	}

	for _, path := range paths {
		if follows, err := jsonparser.GetInt(data, append(path, "follows")...); err == nil {
			userInfo.FollowCount = follows
			break
		}
	}

	for _, path := range paths {
		if playlistCount, err := jsonparser.GetInt(data, append(path, "playlistCount")...); err == nil {
			userInfo.PlayCount = playlistCount
			break
		}
	}

	return userInfo, nil
}

// parseUserFollowing 解析用户关注列表
func (p *NeteasePlugin) parseUserFollowing(data []byte) ([]string, error) {
	following := make([]string, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		if userID, err := jsonparser.GetInt(value, "userId"); err == nil {
			following = append(following, strconv.FormatInt(userID, 10))
		}
	}, "follow")

	return following, err
}

// parseUserFollowers 解析用户粉丝列表
func (p *NeteasePlugin) parseUserFollowers(data []byte) ([]string, error) {
	followers := make([]string, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		if userID, err := jsonparser.GetInt(value, "userId"); err == nil {
			followers = append(followers, strconv.FormatInt(userID, 10))
		}
	}, "followeds")

	return followers, err
}

// GetUserPlayHistory 获取用户播放历史
func (p *NeteasePlugin) GetUserPlayHistory(ctx context.Context, userID string, historyType int) ([]*core.Track, error) {
	if !p.IsLoggedIn() {
		return nil, fmt.Errorf("user not logged in")
	}

	// 获取当前登录用户信息
	p.mu.RLock()
	currentUser := p.user
	p.mu.RUnlock()

	if currentUser == nil || strconv.FormatInt(currentUser.UserID, 10) != userID {
		return nil, fmt.Errorf("can only get play history for current user")
	}

	userRecordService := &service.UserRecordService{
		UId:  userID,
		Type: strconv.Itoa(historyType), // 0: 所有时间, 1: 最近一周
	}

	code, response := userRecordService.UserRecord()
	if code != 200 {
		return nil, fmt.Errorf("failed to get user play history: code %f", code)
	}

	tracks, err := p.parseUserPlayHistory(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user play history: %w", err)
	}

	return tracks, nil
}

// parseUserPlayHistory 解析用户播放历史
func (p *NeteasePlugin) parseUserPlayHistory(data []byte) ([]*core.Track, error) {
	tracks := make([]*core.Track, 0)

	// 尝试从不同路径解析播放历史
	paths := [][]string{
		{"weekData"},
		{"allData"},
	}

	for _, path := range paths {
		_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if err != nil {
				return
			}

			// 解析歌曲信息
			track, parseErr := p.parseTrackFromPlayHistory(value)
			if parseErr == nil {
				// PlayCount字段在core.Track中不存在，跳过
				tracks = append(tracks, track)
			}
		}, path...)

		if err == nil && len(tracks) > 0 {
			break
		}
	}

	return tracks, nil
}

// GetUserCloudMusic 获取用户云盘音乐
func (p *NeteasePlugin) GetUserCloudMusic(ctx context.Context) ([]*core.Track, error) {
	if !p.IsLoggedIn() {
		return nil, fmt.Errorf("user not logged in")
	}

	userCloudService := &service.UserCloudService{
		Limit:  "200",
		Offset: "0",
	}

	code, response := userCloudService.UserCloud()
	if code != 200 {
		return nil, fmt.Errorf("failed to get user cloud music: code %f", code)
	}

	tracks, err := p.parseUserCloudMusic(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user cloud music: %w", err)
	}

	return tracks, nil
}

// parseUserCloudMusic 解析用户云盘音乐
func (p *NeteasePlugin) parseUserCloudMusic(data []byte) ([]*core.Track, error) {
	tracks := make([]*core.Track, 0)

	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		// 解析云盘歌曲信息
		track, parseErr := p.parseTrackFromCloud(value)
		if parseErr == nil {
			// 设置云盘特有的元数据
			track.Metadata = make(map[string]interface{})
			track.Metadata["cloud"] = "true"
			tracks = append(tracks, track)
		}
	}, "data")

	return tracks, err
}