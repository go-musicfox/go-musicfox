package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
	plugin "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/music"
)

// GetTrackURL 获取歌曲播放URL
func (p *NeteasePlugin) GetTrackURL(ctx context.Context, trackID string, quality plugin.AudioQuality) (string, error) {
	// 转换质量参数
	qualityLevel := p.convertQuality(quality)

	songURL := &service.SongUrlService{
		ID: trackID,
		Br: qualityLevel,
	}

	code, response := songURL.SongUrl()
	if code != 200 {
		return "", fmt.Errorf("failed to get song URL: code %f", code)
	}

	// 解析URL
	url, err := p.parseSongURL(response)
	if err != nil {
		return "", fmt.Errorf("failed to parse song URL: %w", err)
	}

	if url == "" {
		return "", fmt.Errorf("song URL is empty, may be due to copyright restrictions")
	}

	return url, nil
}

// GetTrackLyrics 获取歌曲歌词
func (p *NeteasePlugin) GetTrackLyrics(ctx context.Context, trackID string) (*plugin.Lyrics, error) {
	lyricService := &service.LyricService{
		ID: trackID,
	}

	code, response := lyricService.Lyric()
	if code != 200 {
		return nil, fmt.Errorf("failed to get lyrics: code %f", code)
	}

	lyrics, err := p.parseLyrics(response, trackID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lyrics: %w", err)
	}

	return lyrics, nil
}

// GetTrackDetail 获取歌曲详细信息
func (p *NeteasePlugin) GetTrackDetail(ctx context.Context, trackID string) (*core.Track, error) {
	songDetail := &service.SongDetailService{
		Ids: trackID,
	}

	code, response := songDetail.SongDetail()
	if code != 200 {
		return nil, fmt.Errorf("failed to get song detail: code %f", code)
	}

	track, err := p.parseTrackDetail(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse track detail: %w", err)
	}

	return track, nil
}

// GetTrackComments 获取歌曲评论
func (p *NeteasePlugin) GetTrackComments(ctx context.Context, trackID string, offset, limit int) ([]*plugin.Comment, error) {
	commentService := &service.CommentMusicService{
		ID:     trackID,
		Limit:  strconv.Itoa(limit),
		Offset: strconv.Itoa(offset),
	}

	code, response := commentService.CommentMusic()
	if code != 200 {
		return nil, fmt.Errorf("failed to get comments: code %f", code)
	}

	comments, err := p.parseComments(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse comments: %w", err)
	}

	return comments, nil
}

// convertQuality 转换音质参数
func (p *NeteasePlugin) convertQuality(quality plugin.AudioQuality) string {
	switch quality {
	case plugin.AudioQualityLow:
		return "128000" // 128kbps
	case plugin.AudioQualityStandard:
		return "192000" // 192kbps
	case plugin.AudioQualityHigh:
		return "320000" // 320kbps
	case plugin.AudioQualityLossless:
		return "999000" // 无损
	case plugin.AudioQualityHiRes:
		return "1999000" // Hi-Res
	default:
		return "320000" // 默认320kbps
	}
}

// parseSongURL 解析歌曲URL
func (p *NeteasePlugin) parseSongURL(data []byte) (string, error) {
	var url string
	
	// 尝试从data数组中获取第一个URL
	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}
		
		if songURL, err := jsonparser.GetString(value, "url"); err == nil && songURL != "" {
			url = songURL
			return // 找到第一个有效URL就返回
		}
	}, "data")

	if err != nil {
		return "", err
	}

	return url, nil
}

// parseLyrics 解析歌词
func (p *NeteasePlugin) parseLyrics(data []byte, trackID string) (*plugin.Lyrics, error) {
	lyrics := &plugin.Lyrics{
		TrackID:   trackID,
		Source:    "netease",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 获取原始歌词
	if lrc, err := jsonparser.GetString(data, "lrc", "lyric"); err == nil {
		lyrics.Content = lrc
		lyrics.TimedLyrics = p.parseTimedLyrics(lrc)
	}

	// 获取翻译歌词
	if tlyric, err := jsonparser.GetString(data, "tlyric", "lyric"); err == nil && tlyric != "" {
		lyrics.Translation = tlyric
	}

	// 设置语言
	lyrics.Language = "zh" // 网易云音乐主要是中文歌词

	return lyrics, nil
}

// parseTimedLyrics 解析时间轴歌词
func (p *NeteasePlugin) parseTimedLyrics(lrcContent string) []plugin.TimedLyric {
	lines := strings.Split(lrcContent, "\n")
	timedLyrics := make([]plugin.TimedLyric, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析时间标签 [mm:ss.xx]
		if strings.HasPrefix(line, "[") {
			closeIdx := strings.Index(line, "]")
			if closeIdx > 0 {
				timeStr := line[1:closeIdx]
				content := strings.TrimSpace(line[closeIdx+1:])
				
				// 解析时间
				duration := p.parseTimeString(timeStr)
				if duration >= 0 && content != "" {
					timedLyrics = append(timedLyrics, plugin.TimedLyric{
						Time:    duration,
						Content: content,
					})
				}
			}
		}
	}

	return timedLyrics
}

// parseTimeString 解析时间字符串 "mm:ss.xx" 为 time.Duration
func (p *NeteasePlugin) parseTimeString(timeStr string) time.Duration {
	// 分割分钟和秒
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return -1
	}

	// 解析分钟
	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1
	}

	// 解析秒和毫秒
	secondsParts := strings.Split(parts[1], ".")
	seconds, err := strconv.Atoi(secondsParts[0])
	if err != nil {
		return -1
	}

	milliseconds := 0
	if len(secondsParts) > 1 {
		// 补齐到3位数
		msStr := secondsParts[1]
		for len(msStr) < 3 {
			msStr += "0"
		}
		if len(msStr) > 3 {
			msStr = msStr[:3]
		}
		milliseconds, _ = strconv.Atoi(msStr)
	}

	totalMs := int64(minutes)*60*1000 + int64(seconds)*1000 + int64(milliseconds)
	return time.Duration(totalMs) * time.Millisecond
}

// parseTrackDetail 解析歌曲详细信息
func (p *NeteasePlugin) parseTrackDetail(data []byte) (*core.Track, error) {
	var track *core.Track

	// 从songs数组中获取第一首歌曲
	_, err := jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil || track != nil {
			return
		}

		parsedTrack, parseErr := p.parseTrackFromDetail(value)
		if parseErr == nil {
			track = parsedTrack
		}
	}, "songs")

	if err != nil {
		return nil, err
	}

	if track == nil {
		return nil, fmt.Errorf("no track found in response")
	}

	return track, nil
}

// parseTrackFromDetail 从详情响应解析歌曲信息
func (p *NeteasePlugin) parseTrackFromDetail(data []byte) (*core.Track, error) {
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
	track.Artist = strings.Join(artists, ", ")

	// 解析专辑信息
	if albumName, err := jsonparser.GetString(data, "al", "name"); err == nil {
		track.Album = albumName
	}

	// AlbumID字段在core.Track中不存在，跳过
	// if albumID, err := jsonparser.GetInt(data, "al", "id"); err == nil {
	//	track.AlbumID = strconv.FormatInt(albumID, 10)
	// }

	if picURL, err := jsonparser.GetString(data, "al", "picUrl"); err == nil {
		track.CoverURL = picURL
	}

	// 解析其他信息
	// ReleaseDate字段在core.Track中不存在，跳过
	// if publishTime, err := jsonparser.GetInt(data, "publishTime"); err == nil {
	//	track.ReleaseDate = time.Unix(publishTime/1000, 0)
	// }

	// 解析音质信息
	if br, err := jsonparser.GetInt(data, "h", "br"); err == nil {
		track.Bitrate = int(br)
	}
	// FileSize字段在core.Track中不存在，跳过
	// if size, err := jsonparser.GetInt(data, "h", "size"); err == nil {
	//	track.FileSize = size
	// }

	return track, nil
}

// parseComments 解析评论
func (p *NeteasePlugin) parseComments(data []byte) ([]*plugin.Comment, error) {
	comments := make([]*plugin.Comment, 0)

	// 解析热门评论
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		comment, parseErr := p.parseComment(value)
		if parseErr == nil {
			comments = append(comments, comment)
		}
	}, "hotComments")

	// 解析普通评论
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}

		comment, parseErr := p.parseComment(value)
		if parseErr == nil {
			comments = append(comments, comment)
		}
	}, "comments")

	return comments, nil
}

// parseComment 解析单个评论
func (p *NeteasePlugin) parseComment(data []byte) (*plugin.Comment, error) {
	comment := &plugin.Comment{}

	if commentID, err := jsonparser.GetInt(data, "commentId"); err == nil {
		comment.ID = strconv.FormatInt(commentID, 10)
	}

	if content, err := jsonparser.GetString(data, "content"); err == nil {
		comment.Content = content
	}

	if likedCount, err := jsonparser.GetInt(data, "likedCount"); err == nil {
		comment.LikeCount = likedCount
	}

	if timeVal, err := jsonparser.GetInt(data, "time"); err == nil {
		comment.CreatedAt = time.Unix(timeVal/1000, 0)
	}

	// 解析用户信息
	if userID, err := jsonparser.GetInt(data, "user", "userId"); err == nil {
		comment.UserID = strconv.FormatInt(userID, 10)
	}

	if nickname, err := jsonparser.GetString(data, "user", "nickname"); err == nil {
		comment.Username = nickname
	}

	if avatarURL, err := jsonparser.GetString(data, "user", "avatarUrl"); err == nil {
		comment.AvatarURL = avatarURL
	}

	return comment, nil
}