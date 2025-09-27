package plugin

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// SongService 歌曲信息服务
type SongService struct {
	sources    map[string]MusicSourcePlugin
	cache      Cache
	repository Repository
	mu         sync.RWMutex
	defaultTTL time.Duration
}

// NewSongService 创建歌曲信息服务
func NewSongService() *SongService {
	return &SongService{
		sources:    make(map[string]MusicSourcePlugin),
		defaultTTL: 60 * time.Minute, // 歌曲信息缓存时间较长
	}
}

// RegisterSource 注册音乐源
func (s *SongService) RegisterSource(name string, source MusicSourcePlugin) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sources[name] = source
}

// SetCache 设置缓存
func (s *SongService) SetCache(cache Cache) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = cache
}

// SetRepository 设置数据仓库
func (s *SongService) SetRepository(repo Repository) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.repository = repo
}

// GetTrackURL 获取音轨播放URL
func (s *SongService) GetTrackURL(ctx context.Context, trackID string, sourceName string, quality AudioQuality) (string, error) {
	if trackID == "" {
		return "", fmt.Errorf("track id cannot be empty")
	}

	// 检查缓存
	if s.cache != nil {
		cacheKey := fmt.Sprintf("track_url:%s:%s:%s", sourceName, trackID, quality.String())
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
			if url, ok := cached.(string); ok {
				return url, nil
			}
		}
	}

	s.mu.RLock()
	source, exists := s.sources[sourceName]
	s.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("source %s not found", sourceName)
	}

	url, err := source.GetTrackURL(ctx, trackID, quality)
	if err != nil {
		return "", err
	}

	// 缓存URL（较短的TTL，因为URL可能会过期）
	if s.cache != nil {
		cacheKey := fmt.Sprintf("track_url:%s:%s:%s", sourceName, trackID, quality.String())
		s.cache.Set(ctx, cacheKey, url, 10*time.Minute)
	}

	return url, nil
}

// GetTrackLyrics 获取音轨歌词
func (s *SongService) GetTrackLyrics(ctx context.Context, trackID string, sourceName string) (*Lyrics, error) {
	if trackID == "" {
		return nil, fmt.Errorf("track id cannot be empty")
	}

	// 检查缓存
	if s.cache != nil {
		cacheKey := fmt.Sprintf("track_lyrics:%s:%s", sourceName, trackID)
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
			if lyrics, ok := cached.(*Lyrics); ok {
				return lyrics, nil
			}
		}
	}

	s.mu.RLock()
	source, exists := s.sources[sourceName]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	if baseSrc, ok := source.(*BaseMusicSourcePlugin); ok {
		if !baseSrc.HasFeature(MusicSourceFeatureLyrics) {
			return nil, fmt.Errorf("source %s does not support lyrics feature", sourceName)
		}
	}

	lyrics, err := source.GetTrackLyrics(ctx, trackID)
	if err != nil {
		return nil, err
	}

	// 缓存歌词
	if s.cache != nil {
		cacheKey := fmt.Sprintf("track_lyrics:%s:%s", sourceName, trackID)
		s.cache.Set(ctx, cacheKey, lyrics, s.defaultTTL)
	}

	return lyrics, nil
}

// GetTrackDetail 获取音轨详细信息
func (s *SongService) GetTrackDetail(ctx context.Context, trackID string, sourceName string) (*Track, error) {
	if trackID == "" {
		return nil, fmt.Errorf("track id cannot be empty")
	}

	// 检查缓存
	if s.cache != nil {
		cacheKey := fmt.Sprintf("track_detail:%s:%s", sourceName, trackID)
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
			if track, ok := cached.(*Track); ok {
				return track, nil
			}
		}
	}

	s.mu.RLock()
	source, exists := s.sources[sourceName]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	track, err := source.GetTrackDetail(ctx, trackID)
	if err != nil {
		return nil, err
	}

	// 缓存音轨详情
	if s.cache != nil {
		cacheKey := fmt.Sprintf("track_detail:%s:%s", sourceName, trackID)
		s.cache.Set(ctx, cacheKey, track, s.defaultTTL)
	}

	// 保存到仓库
	if s.repository != nil {
		s.repository.SaveTrack(ctx, track)
	}

	return track, nil
}

// GetTrackComments 获取音轨评论
func (s *SongService) GetTrackComments(ctx context.Context, trackID string, sourceName string, offset, limit int) ([]*Comment, error) {
	if trackID == "" {
		return nil, fmt.Errorf("track id cannot be empty")
	}

	// 检查缓存
	if s.cache != nil {
		cacheKey := fmt.Sprintf("track_comments:%s:%s:%d:%d", sourceName, trackID, offset, limit)
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
			if comments, ok := cached.([]*Comment); ok {
				return comments, nil
			}
		}
	}

	s.mu.RLock()
	source, exists := s.sources[sourceName]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	if baseSrc, ok := source.(*BaseMusicSourcePlugin); ok {
		if !baseSrc.HasFeature(MusicSourceFeatureComment) {
			return nil, fmt.Errorf("source %s does not support comment feature", sourceName)
		}
	}

	comments, err := source.GetTrackComments(ctx, trackID, offset, limit)
	if err != nil {
		return nil, err
	}

	// 缓存评论（较短的TTL）
	if s.cache != nil {
		cacheKey := fmt.Sprintf("track_comments:%s:%s:%d:%d", sourceName, trackID, offset, limit)
		s.cache.Set(ctx, cacheKey, comments, 5*time.Minute)
	}

	return comments, nil
}

// GetMultipleTrackDetails 批量获取音轨详情
func (s *SongService) GetMultipleTrackDetails(ctx context.Context, trackIDs []string, sourceName string) ([]*Track, error) {
	if len(trackIDs) == 0 {
		return []*Track{}, nil
	}

	tracks := make([]*Track, 0, len(trackIDs))
	errorChan := make(chan error, len(trackIDs))
	trackChan := make(chan *Track, len(trackIDs))

	// 并发获取多个音轨详情
	for _, trackID := range trackIDs {
		go func(id string) {
			track, err := s.GetTrackDetail(ctx, id, sourceName)
			if err != nil {
				errorChan <- err
				return
			}
			trackChan <- track
		}(trackID)
	}

	// 收集结果
	for i := 0; i < len(trackIDs); i++ {
		select {
		case track := <-trackChan:
			tracks = append(tracks, track)
		case err := <-errorChan:
			// 记录错误但继续处理其他结果
			fmt.Printf("Get track detail error: %v\n", err)
		case <-ctx.Done():
			return tracks, ctx.Err()
		}
	}

	return tracks, nil
}

// GetAvailableQualities 获取可用的音质选项
func (s *SongService) GetAvailableQualities(ctx context.Context, trackID string, sourceName string) ([]AudioQuality, error) {
	// 这里可以实现获取可用音质的逻辑
	// 由于接口中没有定义这个方法，我们返回默认的音质选项
	return []AudioQuality{
		AudioQualityLow,
		AudioQualityStandard,
		AudioQualityHigh,
		AudioQualityLossless,
	}, nil
}

// SearchTracks 搜索音轨
func (s *SongService) SearchTracks(ctx context.Context, query string, limit, offset int) ([]*Track, error) {
	if s.repository == nil {
		return nil, fmt.Errorf("repository not configured")
	}

	tracks, err := s.repository.SearchTracks(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	// 转换为指针切片
	result := make([]*Track, len(tracks))
	for i := range tracks {
		result[i] = &tracks[i]
	}

	return result, nil
}

// GetTracksByGenre 按流派获取音轨
func (s *SongService) GetTracksByGenre(ctx context.Context, genre string, limit int) ([]*Track, error) {
	if s.repository == nil {
		return nil, fmt.Errorf("repository not configured")
	}

	// 使用搜索功能查找特定流派的音轨
	tracks, err := s.repository.SearchTracks(ctx, genre, limit, 0)
	if err != nil {
		return nil, err
	}

	// 过滤出匹配流派的音轨
	result := make([]*Track, 0)
	for _, track := range tracks {
		if strings.EqualFold(track.Genre, genre) {
			result = append(result, &track)
			if len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

// GetTracksByArtist 按艺术家获取音轨
func (s *SongService) GetTracksByArtist(ctx context.Context, artistName string, limit int) ([]*Track, error) {
	if s.repository == nil {
		return nil, fmt.Errorf("repository not configured")
	}

	tracks, err := s.repository.SearchTracks(ctx, artistName, limit*2, 0) // 获取更多结果用于过滤
	if err != nil {
		return nil, err
	}

	// 过滤出匹配艺术家的音轨
	result := make([]*Track, 0)
	for _, track := range tracks {
		if strings.Contains(strings.ToLower(track.Artist), strings.ToLower(artistName)) {
			result = append(result, &track)
			if len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

// GetSimilarTracks 获取相似音轨
func (s *SongService) GetSimilarTracks(ctx context.Context, trackID string, sourceName string, limit int) ([]*Track, error) {
	if trackID == "" {
		return nil, fmt.Errorf("track id cannot be empty")
	}

	// 检查缓存
	if s.cache != nil {
		cacheKey := fmt.Sprintf("similar_tracks:%s:%s:%d", sourceName, trackID, limit)
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
			if tracks, ok := cached.([]*Track); ok {
				return tracks, nil
			}
		}
	}

	s.mu.RLock()
	source, exists := s.sources[sourceName]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source %s not found", sourceName)
	}

	if baseSrc, ok := source.(*BaseMusicSourcePlugin); ok {
		if !baseSrc.HasFeature(MusicSourceFeatureRecommendation) {
			return nil, fmt.Errorf("source %s does not support recommendation feature", sourceName)
		}
	}

	tracks, err := source.GetSimilarTracks(ctx, trackID, limit)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if s.cache != nil {
		cacheKey := fmt.Sprintf("similar_tracks:%s:%s:%d", sourceName, trackID, limit)
		s.cache.Set(ctx, cacheKey, tracks, s.defaultTTL)
	}

	return tracks, nil
}

// GetTrackAnalytics 获取音轨分析数据
func (s *SongService) GetTrackAnalytics(ctx context.Context, trackID string, sourceName string) (map[string]interface{}, error) {
	track, err := s.GetTrackDetail(ctx, trackID, sourceName)
	if err != nil {
		return nil, err
	}

	analytics := map[string]interface{}{
		"duration":     track.Duration.String(),
		"bitrate":      track.Bitrate,
		"sample_rate":  track.SampleRate,
		"channels":     track.Channels,
		"format":       track.Format.String(),
		"quality":      track.Quality.String(),
		"genre":        track.Genre,
		"year":         track.Year,
		"created_at":   track.CreatedAt,
		"updated_at":   track.UpdatedAt,
	}

	// 添加元数据
	if track.Metadata != nil {
		analytics["metadata"] = track.Metadata
	}

	return analytics, nil
}

// ValidateTrackURL 验证音轨URL是否有效
func (s *SongService) ValidateTrackURL(ctx context.Context, url string) (bool, error) {
	if url == "" {
		return false, fmt.Errorf("url cannot be empty")
	}

	// 这里可以实现URL验证逻辑
	// 例如发送HEAD请求检查URL是否可访问
	// 为了简化，这里只检查URL格式
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return true, nil
	}

	return false, fmt.Errorf("invalid url format")
}

// ClearTrackCache 清除音轨相关缓存
func (s *SongService) ClearTrackCache(ctx context.Context, trackID string, sourceName string) error {
	if s.cache == nil {
		return nil
	}

	// 清除所有相关的缓存键
	cacheKeys := []string{
		fmt.Sprintf("track_detail:%s:%s", sourceName, trackID),
		fmt.Sprintf("track_lyrics:%s:%s", sourceName, trackID),
	}

	// 清除不同音质的URL缓存
	qualities := []AudioQuality{AudioQualityLow, AudioQualityStandard, AudioQualityHigh, AudioQualityLossless, AudioQualityHiRes}
	for _, quality := range qualities {
		cacheKeys = append(cacheKeys, fmt.Sprintf("track_url:%s:%s:%s", sourceName, trackID, quality.String()))
	}

	for _, key := range cacheKeys {
		s.cache.Delete(ctx, key)
	}

	return nil
}