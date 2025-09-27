package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// NeteaseAdapter 网易云音乐适配器
type NeteaseAdapter struct {
	*BaseMusicSourcePlugin
	rpcClient *RPCClient
	apiBase   string
	cookies   map[string]string
}

// NewNeteaseAdapter 创建网易云音乐适配器
func NewNeteaseAdapter(apiBase string) *NeteaseAdapter {
	info := &PluginInfo{
		ID:          "netease-music",
		Name:        "Netease Music",
		Version:     "1.0.0",
		Description: "Netease Cloud Music adapter",
		Author:      "go-musicfox",
		Type:        PluginTypeMusicSource,
	}

	adapter := &NeteaseAdapter{
		BaseMusicSourcePlugin: NewBaseMusicSourcePlugin(info),
		rpcClient:             NewRPCClient(apiBase, ""),
		apiBase:               apiBase,
		cookies:               make(map[string]string),
	}

	// 添加支持的功能
	adapter.AddFeature(MusicSourceFeatureSearch)
	adapter.AddFeature(MusicSourceFeaturePlaylist)
	adapter.AddFeature(MusicSourceFeatureUser)
	adapter.AddFeature(MusicSourceFeatureRecommendation)
	adapter.AddFeature(MusicSourceFeatureChart)
	adapter.AddFeature(MusicSourceFeatureLyrics)
	adapter.AddFeature(MusicSourceFeatureComment)

	return adapter
}

// Search 搜索音乐
func (n *NeteaseAdapter) Search(ctx context.Context, query string, options SearchOptions) (*SearchResult, error) {
	params := map[string]string{
		"keywords": query,
		"type":     n.convertSearchType(options.Type),
		"limit":    fmt.Sprintf("%d", options.Limit),
		"offset":   fmt.Sprintf("%d", options.Offset),
	}

	resp, err := n.rpcClient.Get(ctx, "/search", params)
	if err != nil {
		return nil, err
	}

	// 解析网易云API响应
	var neteaseResp struct {
		Code   int `json:"code"`
		Result struct {
			Songs     []map[string]interface{} `json:"songs"`
			Albums    []map[string]interface{} `json:"albums"`
			Artists   []map[string]interface{} `json:"artists"`
			Playlists []map[string]interface{} `json:"playlists"`
		} `json:"result"`
	}

	err = json.Unmarshal(resp.Body, &neteaseResp)
	if err != nil {
		return nil, err
	}

	if neteaseResp.Code != 200 {
		return nil, fmt.Errorf("netease api error: %d", neteaseResp.Code)
	}

	// 转换为标准格式
	result := &SearchResult{
		Query:     query,
		Type:      options.Type,
		Offset:    options.Offset,
		Limit:     options.Limit,
		Source:    "netease",
		Timestamp: time.Now(),
	}

	// 转换歌曲
	for _, song := range neteaseResp.Result.Songs {
		track := n.convertNeteaseTrack(song)
		result.Tracks = append(result.Tracks, track)
	}

	// 转换专辑
	for _, album := range neteaseResp.Result.Albums {
		albumData := n.convertNeteaseAlbum(album)
		result.Albums = append(result.Albums, albumData)
	}

	// 转换艺术家
	for _, artist := range neteaseResp.Result.Artists {
		artistData := n.convertNeteaseArtist(artist)
		result.Artists = append(result.Artists, artistData)
	}

	// 转换播放列表
	for _, playlist := range neteaseResp.Result.Playlists {
		playlistData := n.convertNeteasePlaylist(playlist)
		result.Playlists = append(result.Playlists, playlistData)
	}

	result.Total = len(result.Tracks) + len(result.Albums) + len(result.Artists) + len(result.Playlists)

	return result, nil
}

// GetTrackURL 获取音轨播放URL
func (n *NeteaseAdapter) GetTrackURL(ctx context.Context, trackID string, quality core.AudioQuality) (string, error) {
	params := map[string]string{
		"id": trackID,
		"br": fmt.Sprintf("%d", quality.GetBitrate()*1000), // 网易云使用bps
	}

	resp, err := n.rpcClient.Get(ctx, "/song/url", params)
	if err != nil {
		return "", err
	}

	var urlResp struct {
		Code int `json:"code"`
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp.Body, &urlResp)
	if err != nil {
		return "", err
	}

	if urlResp.Code != 200 || len(urlResp.Data) == 0 {
		return "", fmt.Errorf("failed to get track url")
	}

	return urlResp.Data[0].URL, nil
}

// GetTrackLyrics 获取歌词
func (n *NeteaseAdapter) GetTrackLyrics(ctx context.Context, trackID string) (*Lyrics, error) {
	params := map[string]string{
		"id": trackID,
	}

	resp, err := n.rpcClient.Get(ctx, "/lyric", params)
	if err != nil {
		return nil, err
	}

	var lyricsResp struct {
		Code int `json:"code"`
		Lrc  struct {
			Lyric string `json:"lyric"`
		} `json:"lrc"`
		Tlyric struct {
			Lyric string `json:"lyric"`
		} `json:"tlyric"`
	}

	err = json.Unmarshal(resp.Body, &lyricsResp)
	if err != nil {
		return nil, err
	}

	if lyricsResp.Code != 200 {
		return nil, fmt.Errorf("failed to get lyrics")
	}

	lyrics := &Lyrics{
		TrackID:     trackID,
		Content:     lyricsResp.Lrc.Lyric,
		Translation: lyricsResp.Tlyric.Lyric,
		Language:    "zh-CN",
		Source:      "netease",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 解析时间轴歌词
	lyrics.TimedLyrics = n.parseLRC(lyricsResp.Lrc.Lyric)

	return lyrics, nil
}

// Login 用户登录
func (n *NeteaseAdapter) Login(ctx context.Context, credentials map[string]string) error {
	phone, ok := credentials["phone"]
	if !ok {
		return fmt.Errorf("phone number required")
	}

	password, ok := credentials["password"]
	if !ok {
		return fmt.Errorf("password required")
	}

	body := map[string]string{
		"phone":    phone,
		"password": password,
	}

	resp, err := n.rpcClient.Post(ctx, "/login/cellphone", body)
	if err != nil {
		return err
	}

	var loginResp struct {
		Code    int `json:"code"`
		Message string `json:"message"`
		Cookie  string `json:"cookie"`
	}

	err = json.Unmarshal(resp.Body, &loginResp)
	if err != nil {
		return err
	}

	if loginResp.Code != 200 {
		return fmt.Errorf("login failed: %s", loginResp.Message)
	}

	// 保存cookie
	n.rpcClient.SetHeader("Cookie", loginResp.Cookie)

	return nil
}

// 辅助方法
func (n *NeteaseAdapter) convertSearchType(searchType SearchType) string {
	switch searchType {
	case SearchTypeTrack:
		return "1"
	case SearchTypeAlbum:
		return "10"
	case SearchTypeArtist:
		return "100"
	case SearchTypePlaylist:
		return "1000"
	default:
		return "1"
	}
}

func (n *NeteaseAdapter) convertNeteaseTrack(data map[string]interface{}) core.Track {
	// 简化的转换逻辑
	return core.Track{
		ID:       fmt.Sprintf("%v", data["id"]),
		Title:    fmt.Sprintf("%v", data["name"]),
		Artist:   n.extractArtistNames(data["artists"]),
		Album:    n.extractAlbumName(data["album"]),
		Duration: time.Duration(n.getInt64(data["duration"])) * time.Millisecond,
		Source:   "netease",
		SourceID: fmt.Sprintf("%v", data["id"]),
	}
}

func (n *NeteaseAdapter) convertNeteaseAlbum(data map[string]interface{}) Album {
	return Album{
		ID:       fmt.Sprintf("%v", data["id"]),
		Title:    fmt.Sprintf("%v", data["name"]),
		Artist:   n.extractArtistNames(data["artists"]),
		Source:   "netease",
		SourceID: fmt.Sprintf("%v", data["id"]),
	}
}

func (n *NeteaseAdapter) convertNeteaseArtist(data map[string]interface{}) Artist {
	return Artist{
		ID:       fmt.Sprintf("%v", data["id"]),
		Name:     fmt.Sprintf("%v", data["name"]),
		Source:   "netease",
		SourceID: fmt.Sprintf("%v", data["id"]),
	}
}

func (n *NeteaseAdapter) convertNeteasePlaylist(data map[string]interface{}) Playlist {
	return Playlist{
		ID:          fmt.Sprintf("%v", data["id"]),
		Name:        fmt.Sprintf("%v", data["name"]),
		Description: fmt.Sprintf("%v", data["description"]),
		TrackCount:  int(n.getInt64(data["trackCount"])),
		Source:      "netease",
		SourceID:    fmt.Sprintf("%v", data["id"]),
	}
}

func (n *NeteaseAdapter) extractArtistNames(artists interface{}) string {
	if artistList, ok := artists.([]interface{}); ok {
		names := make([]string, 0, len(artistList))
		for _, artist := range artistList {
			if artistMap, ok := artist.(map[string]interface{}); ok {
				if name, ok := artistMap["name"].(string); ok {
					names = append(names, name)
				}
			}
		}
		return strings.Join(names, ", ")
	}
	return ""
}

func (n *NeteaseAdapter) extractAlbumName(album interface{}) string {
	if albumMap, ok := album.(map[string]interface{}); ok {
		if name, ok := albumMap["name"].(string); ok {
			return name
		}
	}
	return ""
}

func (n *NeteaseAdapter) getInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func (n *NeteaseAdapter) parseLRC(lrcContent string) []TimedLyric {
	// 简化的LRC解析
	lines := strings.Split(lrcContent, "\n")
	timedLyrics := make([]TimedLyric, 0)

	for _, line := range lines {
		if strings.HasPrefix(line, "[") {
			// 解析时间标签 [mm:ss.xx]
			// 这里简化处理，实际需要更复杂的解析逻辑
			timedLyrics = append(timedLyrics, TimedLyric{
				Time:    0, // 需要解析时间
				Content: line,
			})
		}
	}

	return timedLyrics
}

// SpotifyAdapter Spotify适配器
type SpotifyAdapter struct {
	*BaseMusicSourcePlugin
	rpcClient   *RPCClient
	clientID    string
	clientSecret string
	accessToken string
}

// NewSpotifyAdapter 创建Spotify适配器
func NewSpotifyAdapter(clientID, clientSecret string) *SpotifyAdapter {
	info := &PluginInfo{
		ID:          "spotify",
		Name:        "Spotify",
		Version:     "1.0.0",
		Description: "Spotify music adapter",
		Author:      "go-musicfox",
		Type:        PluginTypeMusicSource,
	}

	adapter := &SpotifyAdapter{
		BaseMusicSourcePlugin: NewBaseMusicSourcePlugin(info),
		rpcClient:             NewRPCClient("https://api.spotify.com/v1", ""),
		clientID:              clientID,
		clientSecret:          clientSecret,
	}

	// 添加支持的功能
	adapter.AddFeature(MusicSourceFeatureSearch)
	adapter.AddFeature(MusicSourceFeaturePlaylist)
	adapter.AddFeature(MusicSourceFeatureUser)
	adapter.AddFeature(MusicSourceFeatureRecommendation)

	return adapter
}

// Search Spotify搜索
func (s *SpotifyAdapter) Search(ctx context.Context, query string, options SearchOptions) (*SearchResult, error) {
	params := map[string]string{
		"q":      query,
		"type":   s.convertSpotifySearchType(options.Type),
		"limit":  fmt.Sprintf("%d", options.Limit),
		"offset": fmt.Sprintf("%d", options.Offset),
	}

	// 设置认证头
	s.rpcClient.SetHeader("Authorization", "Bearer "+s.accessToken)

	resp, err := s.rpcClient.Get(ctx, "/search", params)
	if err != nil {
		return nil, err
	}

	// 解析Spotify API响应
	var spotifyResp map[string]interface{}
	err = json.Unmarshal(resp.Body, &spotifyResp)
	if err != nil {
		return nil, err
	}

	// 转换为标准格式
	result := &SearchResult{
		Query:     query,
		Type:      options.Type,
		Offset:    options.Offset,
		Limit:     options.Limit,
		Source:    "spotify",
		Timestamp: time.Now(),
	}

	// 处理搜索结果...
	// 这里需要根据Spotify API的实际响应格式进行转换

	return result, nil
}

func (s *SpotifyAdapter) convertSpotifySearchType(searchType SearchType) string {
	switch searchType {
	case SearchTypeTrack:
		return "track"
	case SearchTypeAlbum:
		return "album"
	case SearchTypeArtist:
		return "artist"
	case SearchTypePlaylist:
		return "playlist"
	default:
		return "track,album,artist,playlist"
	}
}

// LocalMusicAdapter 本地音乐适配器
type LocalMusicAdapter struct {
	*BaseMusicSourcePlugin
	musicDirs []string
	scanCache map[string][]Track
}

// NewLocalMusicAdapter 创建本地音乐适配器
func NewLocalMusicAdapter(musicDirs []string) *LocalMusicAdapter {
	info := &PluginInfo{
		ID:          "local-music",
		Name:        "Local Music",
		Version:     "1.0.0",
		Description: "Local music files adapter",
		Author:      "go-musicfox",
		Type:        PluginTypeMusicSource,
	}

	adapter := &LocalMusicAdapter{
		BaseMusicSourcePlugin: NewBaseMusicSourcePlugin(info),
		musicDirs:             musicDirs,
		scanCache:             make(map[string][]Track),
	}

	// 添加支持的功能
	adapter.AddFeature(MusicSourceFeatureSearch)
	adapter.AddFeature(MusicSourceFeaturePlaylist)

	return adapter
}

// Search 搜索本地音乐
func (l *LocalMusicAdapter) Search(ctx context.Context, query string, options SearchOptions) (*SearchResult, error) {
	// 扫描本地音乐文件
	tracks, err := l.scanLocalMusic()
	if err != nil {
		return nil, err
	}

	// 过滤匹配的音轨
	matchedTracks := make([]Track, 0)
	queryLower := strings.ToLower(query)

	for _, track := range tracks {
		if strings.Contains(strings.ToLower(track.Title), queryLower) ||
			strings.Contains(strings.ToLower(track.Artist), queryLower) ||
			strings.Contains(strings.ToLower(track.Album), queryLower) {
			matchedTracks = append(matchedTracks, track)
		}
	}

	// 应用分页
	start := options.Offset
	end := start + options.Limit
	if start > len(matchedTracks) {
		start = len(matchedTracks)
	}
	if end > len(matchedTracks) {
		end = len(matchedTracks)
	}

	result := &SearchResult{
		Query:     query,
		Type:      options.Type,
		Total:     len(matchedTracks),
		Offset:    options.Offset,
		Limit:     options.Limit,
		Tracks:    matchedTracks[start:end],
		Source:    "local",
		Timestamp: time.Now(),
	}

	return result, nil
}

// GetTrackURL 获取本地音轨URL
func (l *LocalMusicAdapter) GetTrackURL(ctx context.Context, trackID string, quality core.AudioQuality) (string, error) {
	// 本地文件直接返回文件路径
	return trackID, nil // trackID就是文件路径
}

// scanLocalMusic 扫描本地音乐文件
func (l *LocalMusicAdapter) scanLocalMusic() ([]core.Track, error) {
	tracks := make([]core.Track, 0)

	for _, dir := range l.musicDirs {
		// 检查缓存
		if cachedTracks, exists := l.scanCache[dir]; exists {
			tracks = append(tracks, cachedTracks...)
			continue
		}

		// 扫描目录
		dirTracks, err := l.scanDirectory(dir)
		if err != nil {
			continue // 忽略错误，继续扫描其他目录
		}

		// 缓存结果
		l.scanCache[dir] = dirTracks
		tracks = append(tracks, dirTracks...)
	}

	return tracks, nil
}

// scanDirectory 扫描目录
func (l *LocalMusicAdapter) scanDirectory(dir string) ([]core.Track, error) {
	tracks := make([]core.Track, 0)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续扫描
		}

		if d.IsDir() {
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(path))
		if !l.isSupportedAudioFile(ext) {
			return nil
		}

		// 创建音轨信息
		track := l.createTrackFromFile(path)
		tracks = append(tracks, track)

		return nil
	})

	return tracks, err
}

// isSupportedAudioFile 检查是否为支持的音频文件
func (l *LocalMusicAdapter) isSupportedAudioFile(ext string) bool {
	supportedExts := []string{".mp3", ".flac", ".wav", ".m4a", ".aac", ".ogg"}
	for _, supportedExt := range supportedExts {
		if ext == supportedExt {
			return true
		}
	}
	return false
}

// createTrackFromFile 从文件创建音轨信息
func (l *LocalMusicAdapter) createTrackFromFile(filePath string) core.Track {
	fileName := filepath.Base(filePath)
	fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// 简单的文件名解析（实际应用中可能需要读取音频文件的元数据）
	parts := strings.Split(fileNameWithoutExt, " - ")

	var title, artist string
	if len(parts) >= 2 {
		artist = parts[0]
		title = parts[1]
	} else {
		title = fileNameWithoutExt
		artist = "Unknown Artist"
	}

	// 获取文件信息
	fileInfo, _ := os.Stat(filePath)

	return core.Track{
		ID:        filePath, // 使用文件路径作为ID
		Title:     title,
		Artist:    artist,
		Album:     "Unknown Album",
		URL:       filePath,
		Format:    l.getAudioFormatFromExt(filepath.Ext(filePath)),
		Source:    "local",
		SourceID:  filePath,
		CreatedAt: fileInfo.ModTime(),
		UpdatedAt: fileInfo.ModTime(),
	}
}

// getAudioFormatFromExt 根据扩展名获取音频格式
func (l *LocalMusicAdapter) getAudioFormatFromExt(ext string) AudioFormat {
	switch strings.ToLower(ext) {
	case ".mp3":
		return AudioFormatMP3
	case ".flac":
		return AudioFormatFLAC
	case ".wav":
		return AudioFormatWAV
	case ".m4a", ".aac":
		return AudioFormatAAC
	case ".ogg":
		return AudioFormatOGG
	default:
		return AudioFormatUnknown
	}
}