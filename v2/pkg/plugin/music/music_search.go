package plugin

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MusicSearchEngine 音乐搜索引擎
type MusicSearchEngine struct {
	sources    map[string]MusicSourcePlugin
	cache      Cache
	mu         sync.RWMutex
	defaultTTL time.Duration
}

// NewMusicSearchEngine 创建音乐搜索引擎
func NewMusicSearchEngine() *MusicSearchEngine {
	return &MusicSearchEngine{
		sources:    make(map[string]MusicSourcePlugin),
		defaultTTL: 30 * time.Minute,
	}
}

// RegisterSource 注册音乐源
func (e *MusicSearchEngine) RegisterSource(name string, source MusicSourcePlugin) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sources[name] = source
}

// UnregisterSource 注销音乐源
func (e *MusicSearchEngine) UnregisterSource(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.sources, name)
}

// SetCache 设置缓存
func (e *MusicSearchEngine) SetCache(cache Cache) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cache = cache
}

// Search 搜索音乐
func (e *MusicSearchEngine) Search(ctx context.Context, query string, options SearchOptions) (*SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// 检查缓存
	if e.cache != nil {
		cacheKey := e.buildCacheKey(query, options)
		if cached, err := e.cache.Get(ctx, cacheKey); err == nil {
			if result, ok := cached.(*SearchResult); ok {
				return result, nil
			}
		}
	}

	// 执行搜索
	result, err := e.performSearch(ctx, query, options)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if e.cache != nil {
		cacheKey := e.buildCacheKey(query, options)
		e.cache.Set(ctx, cacheKey, result, e.defaultTTL)
	}

	return result, nil
}

// SearchMultipleSources 多源搜索
func (e *MusicSearchEngine) SearchMultipleSources(ctx context.Context, query string, options SearchOptions, sourceNames []string) (*SearchResult, error) {
	if len(sourceNames) == 0 {
		return e.Search(ctx, query, options)
	}

	results := make([]*SearchResult, 0, len(sourceNames))
	errorChan := make(chan error, len(sourceNames))
	resultChan := make(chan *SearchResult, len(sourceNames))

	// 并发搜索多个源
	for _, sourceName := range sourceNames {
		go func(name string) {
			e.mu.RLock()
			source, exists := e.sources[name]
			e.mu.RUnlock()

			if !exists {
				errorChan <- fmt.Errorf("source %s not found", name)
				return
			}

			result, err := source.Search(ctx, query, options)
			if err != nil {
				errorChan <- err
				return
			}

			resultChan <- result
		}(sourceName)
	}

	// 收集结果
	for i := 0; i < len(sourceNames); i++ {
		select {
		case result := <-resultChan:
			results = append(results, result)
		case err := <-errorChan:
			// 记录错误但继续处理其他结果
			fmt.Printf("Search error: %v\n", err)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// 合并结果
	return e.mergeSearchResults(results, query, options), nil
}

// performSearch 执行搜索
func (e *MusicSearchEngine) performSearch(ctx context.Context, query string, options SearchOptions) (*SearchResult, error) {
	e.mu.RLock()
	sources := make([]MusicSourcePlugin, 0, len(e.sources))
	for _, source := range e.sources {
		sources = append(sources, source)
	}
	e.mu.RUnlock()

	if len(sources) == 0 {
		return nil, fmt.Errorf("no music sources available")
	}

	// 使用第一个可用的源进行搜索
	for _, source := range sources {
		// 检查是否支持搜索功能
		if baseSrc, ok := source.(*BaseMusicSourcePlugin); ok {
			if baseSrc.HasFeature(MusicSourceFeatureSearch) {
				return source.Search(ctx, query, options)
			}
		} else {
			// 对于非BaseMusicSourcePlugin类型（如Mock），直接尝试搜索
			result, err := source.Search(ctx, query, options)
			if err == nil {
				return result, nil
			}
		}
	}

	return nil, fmt.Errorf("no sources support search feature")
}

// mergeSearchResults 合并搜索结果
func (e *MusicSearchEngine) mergeSearchResults(results []*SearchResult, query string, options SearchOptions) *SearchResult {
	merged := &SearchResult{
		Query:     query,
		Type:      options.Type,
		Offset:    options.Offset,
		Limit:     options.Limit,
		Tracks:    make([]Track, 0),
		Albums:    make([]Album, 0),
		Artists:   make([]Artist, 0),
		Playlists: make([]Playlist, 0),
		Timestamp: time.Now(),
	}

	for _, result := range results {
		if result == nil {
			continue
		}

		merged.Total += result.Total
		merged.Tracks = append(merged.Tracks, result.Tracks...)
		merged.Albums = append(merged.Albums, result.Albums...)
		merged.Artists = append(merged.Artists, result.Artists...)
		merged.Playlists = append(merged.Playlists, result.Playlists...)
	}

	// 去重和排序
	merged.Tracks = e.deduplicateTracks(merged.Tracks)
	merged.Albums = e.deduplicateAlbums(merged.Albums)
	merged.Artists = e.deduplicateArtists(merged.Artists)
	merged.Playlists = e.deduplicatePlaylists(merged.Playlists)

	return merged
}

// buildCacheKey 构建缓存键
func (e *MusicSearchEngine) buildCacheKey(query string, options SearchOptions) string {
	return fmt.Sprintf("search:%s:%s:%d:%d", query, options.Type.String(), options.Offset, options.Limit)
}

// deduplicateTracks 去重音轨
func (e *MusicSearchEngine) deduplicateTracks(tracks []Track) []Track {
	seen := make(map[string]bool)
	result := make([]Track, 0, len(tracks))

	for _, track := range tracks {
		key := fmt.Sprintf("%s-%s-%s", track.Title, track.Artist, track.Album)
		if !seen[key] {
			seen[key] = true
			result = append(result, track)
		}
	}

	return result
}

// deduplicateAlbums 去重专辑
func (e *MusicSearchEngine) deduplicateAlbums(albums []Album) []Album {
	seen := make(map[string]bool)
	result := make([]Album, 0, len(albums))

	for _, album := range albums {
		key := fmt.Sprintf("%s-%s", album.Title, album.Artist)
		if !seen[key] {
			seen[key] = true
			result = append(result, album)
		}
	}

	return result
}

// deduplicateArtists 去重艺术家
func (e *MusicSearchEngine) deduplicateArtists(artists []Artist) []Artist {
	seen := make(map[string]bool)
	result := make([]Artist, 0, len(artists))

	for _, artist := range artists {
		if !seen[artist.Name] {
			seen[artist.Name] = true
			result = append(result, artist)
		}
	}

	return result
}

// deduplicatePlaylists 去重播放列表
func (e *MusicSearchEngine) deduplicatePlaylists(playlists []Playlist) []Playlist {
	seen := make(map[string]bool)
	result := make([]Playlist, 0, len(playlists))

	for _, playlist := range playlists {
		key := fmt.Sprintf("%s-%s", playlist.Name, playlist.Owner)
		if !seen[key] {
			seen[key] = true
			result = append(result, playlist)
		}
	}

	return result
}

// GetAvailableSources 获取可用的音乐源
func (e *MusicSearchEngine) GetAvailableSources() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	sources := make([]string, 0, len(e.sources))
	for name := range e.sources {
		sources = append(sources, name)
	}

	return sources
}

// GetSourceInfo 获取音乐源信息
func (e *MusicSearchEngine) GetSourceInfo(name string) (*ServiceInfo, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	source, exists := e.sources[name]
	if !exists {
		return nil, fmt.Errorf("source %s not found", name)
	}

	return source.GetServiceInfo(), nil
}

// SearchSuggestions 搜索建议
func (e *MusicSearchEngine) SearchSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	if strings.TrimSpace(query) == "" {
		return []string{}, nil
	}

	// 这里可以实现搜索建议逻辑
	// 例如从历史搜索、热门搜索等获取建议
	suggestions := []string{
		query + " remix",
		query + " acoustic",
		query + " live",
	}

	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions, nil
}