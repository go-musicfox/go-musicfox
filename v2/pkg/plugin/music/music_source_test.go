package plugin

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMusicSourcePlugin Mock音乐源插件
type MockMusicSourcePlugin struct {
	mock.Mock
	*BaseMusicSourcePlugin
}

// NewMockMusicSourcePlugin 创建Mock音乐源插件
func NewMockMusicSourcePlugin() *MockMusicSourcePlugin {
	info := &PluginInfo{
		ID:          "mock-source",
		Name:        "Mock Music Source",
		Version:     "1.0.0",
		Description: "Mock music source for testing",
		Author:      "test",
		Type:        PluginTypeMusicSource,
	}

	mockPlugin := &MockMusicSourcePlugin{
		BaseMusicSourcePlugin: NewBaseMusicSourcePlugin(info),
	}

	// 添加所有功能
	mockPlugin.AddFeature(MusicSourceFeatureSearch)
	mockPlugin.AddFeature(MusicSourceFeaturePlaylist)
	mockPlugin.AddFeature(MusicSourceFeatureUser)
	mockPlugin.AddFeature(MusicSourceFeatureLyrics)

	return mockPlugin
}

// Search Mock搜索方法
func (m *MockMusicSourcePlugin) Search(ctx context.Context, query string, options SearchOptions) (*SearchResult, error) {
	args := m.Called(ctx, query, options)
	return args.Get(0).(*SearchResult), args.Error(1)
}

// GetPlaylist Mock获取播放列表方法
func (m *MockMusicSourcePlugin) GetPlaylist(ctx context.Context, id string) (*Playlist, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Playlist), args.Error(1)
}

// GetPlaylistTracks Mock获取播放列表音轨方法
func (m *MockMusicSourcePlugin) GetPlaylistTracks(ctx context.Context, playlistID string, offset, limit int) ([]*Track, error) {
	args := m.Called(ctx, playlistID, offset, limit)
	return args.Get(0).([]*Track), args.Error(1)
}

// GetTrackURL Mock获取音轨URL方法
func (m *MockMusicSourcePlugin) GetTrackURL(ctx context.Context, trackID string, quality AudioQuality) (string, error) {
	args := m.Called(ctx, trackID, quality)
	return args.String(0), args.Error(1)
}

// GetTrackLyrics Mock获取歌词方法
func (m *MockMusicSourcePlugin) GetTrackLyrics(ctx context.Context, trackID string) (*Lyrics, error) {
	args := m.Called(ctx, trackID)
	return args.Get(0).(*Lyrics), args.Error(1)
}

// GetTrackDetail Mock获取音轨详情方法
func (m *MockMusicSourcePlugin) GetTrackDetail(ctx context.Context, trackID string) (*Track, error) {
	args := m.Called(ctx, trackID)
	return args.Get(0).(*Track), args.Error(1)
}

// Login Mock登录方法
func (m *MockMusicSourcePlugin) Login(ctx context.Context, credentials map[string]string) error {
	args := m.Called(ctx, credentials)
	return args.Error(0)
}

// GetUserPlaylists Mock获取用户播放列表方法
func (m *MockMusicSourcePlugin) GetUserPlaylists(ctx context.Context, userID string) ([]*Playlist, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*Playlist), args.Error(1)
}

// GetUserLikedTracks Mock获取用户喜欢音轨方法
func (m *MockMusicSourcePlugin) GetUserLikedTracks(ctx context.Context, userID string) ([]*Track, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*Track), args.Error(1)
}

// 实现其他必需的接口方法
func (m *MockMusicSourcePlugin) CreatePlaylist(ctx context.Context, name, description string) (*Playlist, error) {
	args := m.Called(ctx, name, description)
	return args.Get(0).(*Playlist), args.Error(1)
}

func (m *MockMusicSourcePlugin) UpdatePlaylist(ctx context.Context, playlistID string, updates map[string]interface{}) error {
	args := m.Called(ctx, playlistID, updates)
	return args.Error(0)
}

func (m *MockMusicSourcePlugin) DeletePlaylist(ctx context.Context, playlistID string) error {
	args := m.Called(ctx, playlistID)
	return args.Error(0)
}

func (m *MockMusicSourcePlugin) GetTrackComments(ctx context.Context, trackID string, offset, limit int) ([]*Comment, error) {
	args := m.Called(ctx, trackID, offset, limit)
	return args.Get(0).([]*Comment), args.Error(1)
}

func (m *MockMusicSourcePlugin) Logout(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMusicSourcePlugin) FollowUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockMusicSourcePlugin) UnfollowUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockMusicSourcePlugin) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*UserInfo), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetUserFollowing(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetUserFollowers(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetRecommendations(ctx context.Context, options RecommendationOptions) ([]*Track, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]*Track), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetDailyRecommendations(ctx context.Context) ([]*Track, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*Track), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetSimilarTracks(ctx context.Context, trackID string, limit int) ([]*Track, error) {
	args := m.Called(ctx, trackID, limit)
	return args.Get(0).([]*Track), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetCharts(ctx context.Context) ([]*Chart, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*Chart), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetChartTracks(ctx context.Context, chartID string) ([]*Track, error) {
	args := m.Called(ctx, chartID)
	return args.Get(0).([]*Track), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetAlbum(ctx context.Context, albumID string) (*Album, error) {
	args := m.Called(ctx, albumID)
	return args.Get(0).(*Album), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetAlbumTracks(ctx context.Context, albumID string) ([]*Track, error) {
	args := m.Called(ctx, albumID)
	return args.Get(0).([]*Track), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetArtist(ctx context.Context, artistID string) (*Artist, error) {
	args := m.Called(ctx, artistID)
	return args.Get(0).(*Artist), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetArtistTracks(ctx context.Context, artistID string, offset, limit int) ([]*Track, error) {
	args := m.Called(ctx, artistID, offset, limit)
	return args.Get(0).([]*Track), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetArtistAlbums(ctx context.Context, artistID string, offset, limit int) ([]*Album, error) {
	args := m.Called(ctx, artistID, offset, limit)
	return args.Get(0).([]*Album), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetRadioStations(ctx context.Context) ([]*RadioStation, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*RadioStation), args.Error(1)
}

func (m *MockMusicSourcePlugin) GetRadioTracks(ctx context.Context, stationID string) ([]*Track, error) {
	args := m.Called(ctx, stationID)
	return args.Get(0).([]*Track), args.Error(1)
}

// 测试用例

// TestBaseMusicSourcePlugin 测试基础音乐源插件
func TestBaseMusicSourcePlugin(t *testing.T) {
	info := &PluginInfo{
		ID:          "test-source",
		Name:        "Test Music Source",
		Version:     "1.0.0",
		Description: "Test music source",
		Author:      "test",
		Type:        PluginTypeMusicSource,
	}

	plugin := NewBaseMusicSourcePlugin(info)

	// 测试基本信息
	assert.Equal(t, "test-source", plugin.GetInfo().ID)
	assert.Equal(t, "Test Music Source", plugin.GetInfo().Name)
	assert.Equal(t, PluginTypeMusicSource, plugin.GetInfo().Type)

	// 测试功能管理
	assert.False(t, plugin.HasFeature(MusicSourceFeatureSearch))
	plugin.AddFeature(MusicSourceFeatureSearch)
	assert.True(t, plugin.HasFeature(MusicSourceFeatureSearch))

	features := plugin.GetSupportedFeatures()
	assert.Len(t, features, 1)
	assert.Equal(t, MusicSourceFeatureSearch, features[0])

	// 测试服务信息
	serviceInfo := plugin.GetServiceInfo()
	assert.NotNil(t, serviceInfo)
	assert.Equal(t, "test-source", serviceInfo.ID)
	assert.Equal(t, ServiceStatusStopped, serviceInfo.Status)

	// 测试状态更新
	plugin.UpdateServiceStatus(ServiceStatusRunning)
	assert.Equal(t, ServiceStatusRunning, plugin.GetServiceInfo().Status)

	healthyStatus := HealthStatusHealthy()
	plugin.UpdateHealthStatus(healthyStatus)
	assert.Equal(t, healthyStatus, plugin.GetServiceInfo().Health)
}

// TestMusicSearchEngine 测试音乐搜索引擎
func TestMusicSearchEngine(t *testing.T) {
	engine := NewMusicSearchEngine()
	mockSource := NewMockMusicSourcePlugin()

	// 注册音乐源
	engine.RegisterSource("mock", mockSource)

	// 测试搜索
	ctx := context.Background()
	query := "test song"
	options := SearchOptions{
		Query:  query,
		Type:   SearchTypeTrack,
		Limit:  10,
		Offset: 0,
	}

	expectedResult := &SearchResult{
		Query:  query,
		Type:   SearchTypeTrack,
		Tracks: []Track{{ID: "1", Title: "Test Song", Artist: "Test Artist"}},
		Total:  1,
	}

	// 设置mock的HasFeature方法
	mockSource.BaseMusicSourcePlugin.AddFeature(MusicSourceFeatureSearch)
	mockSource.On("Search", ctx, query, options).Return(expectedResult, nil)

	result, err := engine.Search(ctx, query, options)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, query, result.Query)
	assert.Len(t, result.Tracks, 1)

	mockSource.AssertExpectations(t)
}

// TestPlaylistManager 测试播放列表管理器
func TestPlaylistManager(t *testing.T) {
	manager := NewPlaylistManager()
	mockSource := NewMockMusicSourcePlugin()

	// 注册音乐源
	manager.RegisterSource("mock", mockSource)

	// 测试获取播放列表
	ctx := context.Background()
	playlistID := "playlist123"

	expectedPlaylist := &Playlist{
		ID:         playlistID,
		Name:       "Test Playlist",
		TrackCount: 5,
	}

	// 设置mock的HasFeature方法
	mockSource.BaseMusicSourcePlugin.AddFeature(MusicSourceFeaturePlaylist)
	mockSource.On("GetPlaylist", ctx, playlistID).Return(expectedPlaylist, nil)

	playlist, err := manager.GetPlaylist(ctx, playlistID, "mock")
	assert.NoError(t, err)
	assert.NotNil(t, playlist)
	assert.Equal(t, playlistID, playlist.ID)
	assert.Equal(t, "Test Playlist", playlist.Name)

	mockSource.AssertExpectations(t)
}

// TestSongService 测试歌曲服务
func TestSongService(t *testing.T) {
	service := NewSongService()
	mockSource := NewMockMusicSourcePlugin()

	// 注册音乐源
	service.RegisterSource("mock", mockSource)

	// 测试获取音轨URL
	ctx := context.Background()
	trackID := "track123"
	quality := AudioQualityHigh
	expectedURL := "https://example.com/track.mp3"

	mockSource.On("GetTrackURL", ctx, trackID, quality).Return(expectedURL, nil)

	url, err := service.GetTrackURL(ctx, trackID, "mock", quality)
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)

	// 测试获取歌词
	expectedLyrics := &Lyrics{
		TrackID: trackID,
		Content: "Test lyrics content",
		Source:  "mock",
	}

	// 设置mock的HasFeature方法
	mockSource.BaseMusicSourcePlugin.AddFeature(MusicSourceFeatureLyrics)
	mockSource.On("GetTrackLyrics", ctx, trackID).Return(expectedLyrics, nil)

	lyrics, err := service.GetTrackLyrics(ctx, trackID, "mock")
	assert.NoError(t, err)
	assert.NotNil(t, lyrics)
	assert.Equal(t, trackID, lyrics.TrackID)
	assert.Equal(t, "Test lyrics content", lyrics.Content)

	mockSource.AssertExpectations(t)
}

// TestUserService 测试用户服务
func TestUserService(t *testing.T) {
	userService := NewUserService()
	mockSource := NewMockMusicSourcePlugin()

	// 注册音乐源
	userService.RegisterSource("mock", mockSource)

	// 测试用户登录
	ctx := context.Background()
	credentials := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}

	// 设置mock的HasFeature方法
	mockSource.BaseMusicSourcePlugin.AddFeature(MusicSourceFeatureUser)
	mockSource.On("Login", ctx, credentials).Return(nil)

	session, err := userService.Login(ctx, "mock", credentials)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "testuser", session.UserID)
	assert.Equal(t, "mock", session.SourceName)

	// 测试会话管理
	retrievedSession, err := userService.GetUserSession("mock", "testuser")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedSession)
	assert.Equal(t, session.Token, retrievedSession.Token)

	mockSource.AssertExpectations(t)
}

// TestMemoryCache 测试内存缓存
func TestMemoryCache(t *testing.T) {
	cache := NewMemoryCache(100, 10*time.Minute)
	ctx := context.Background()

	// 测试设置和获取
	key := "test_key"
	value := "test_value"
	ttl := 5 * time.Minute

	err := cache.Set(ctx, key, value, ttl)
	assert.NoError(t, err)

	retrievedValue, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, value, retrievedValue)

	// 测试存在性检查
	exists, err := cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 测试TTL
	ttlRemaining, err := cache.TTL(ctx, key)
	assert.NoError(t, err)
	assert.True(t, ttlRemaining > 0)
	assert.True(t, ttlRemaining <= ttl)

	// 测试删除
	err = cache.Delete(ctx, key)
	assert.NoError(t, err)

	exists, err = cache.Exists(ctx, key)
	assert.NoError(t, err)
	assert.False(t, exists)

	// 测试统计信息
	stats := cache.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.Sets)
	assert.Equal(t, int64(1), stats.Deletes)
}

// TestMusicSourceFactory 测试音乐源工厂
func TestMusicSourceFactory(t *testing.T) {
	config := &FactoryConfig{
		LoadBalancing: true,
		HealthCheck:   false, // 禁用健康检查以简化测试
	}

	factory := NewMusicSourceFactory(config)

	// 测试注册音乐源
	creator := func(config map[string]interface{}) (MusicSourcePlugin, error) {
		return NewMockMusicSourcePlugin(), nil
	}

	err := factory.RegisterSource("test", creator)
	assert.NoError(t, err)

	// 测试重复注册
	err = factory.RegisterSource("test", creator)
	assert.Error(t, err)

	// 测试创建音乐源
	source, err := factory.CreateSource("test", map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, source)

	// 测试获取音乐源
	retrievedSource, err := factory.GetSource("test")
	assert.NoError(t, err)
	assert.Equal(t, source, retrievedSource)

	// 测试获取可用源列表
	availableSources := factory.GetAvailableSources()
	assert.Len(t, availableSources, 1)
	assert.Equal(t, "test", availableSources[0])

	// 测试工厂统计
	stats := factory.GetFactoryStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 4, stats["registered_sources"]) // 包括内置源
	assert.Equal(t, 1, stats["active_sources"])
}

// TestLocalMusicAdapter 测试本地音乐适配器
func TestLocalMusicAdapter(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()

	adapter := NewLocalMusicAdapter([]string{tempDir})

	// 测试基本信息
	assert.Equal(t, "local-music", adapter.GetInfo().ID)
	assert.True(t, adapter.HasFeature(MusicSourceFeatureSearch))

	// 测试获取音轨URL（本地文件）
	ctx := context.Background()
	filePath := "/path/to/song.mp3"
	url, err := adapter.GetTrackURL(ctx, filePath, AudioQualityHigh)
	assert.NoError(t, err)
	assert.Equal(t, filePath, url)
}

// TestSearchTypes 测试搜索类型
func TestSearchTypes(t *testing.T) {
	// 测试搜索类型字符串转换
	assert.Equal(t, "all", SearchTypeAll.String())
	assert.Equal(t, "track", SearchTypeTrack.String())
	assert.Equal(t, "album", SearchTypeAlbum.String())
	assert.Equal(t, "artist", SearchTypeArtist.String())
	assert.Equal(t, "playlist", SearchTypePlaylist.String())
}

// TestAudioQuality 测试音质
func TestAudioQuality(t *testing.T) {
	// 测试音质字符串转换
	assert.Equal(t, "low", AudioQualityLow.String())
	assert.Equal(t, "standard", AudioQualityStandard.String())
	assert.Equal(t, "high", AudioQualityHigh.String())
	assert.Equal(t, "lossless", AudioQualityLossless.String())
	assert.Equal(t, "hires", AudioQualityHiRes.String())

	// 测试比特率
	assert.Equal(t, 128, AudioQualityLow.GetBitrate())
	assert.Equal(t, 192, AudioQualityStandard.GetBitrate())
	assert.Equal(t, 320, AudioQualityHigh.GetBitrate())
	assert.Equal(t, 1411, AudioQualityLossless.GetBitrate())
	assert.Equal(t, 2304, AudioQualityHiRes.GetBitrate())
}

// TestMusicSourceFeatures 测试音乐源功能
func TestMusicSourceFeatures(t *testing.T) {
	// 测试功能字符串转换
	assert.Equal(t, "search", MusicSourceFeatureSearch.String())
	assert.Equal(t, "playlist", MusicSourceFeaturePlaylist.String())
	assert.Equal(t, "user", MusicSourceFeatureUser.String())
	assert.Equal(t, "recommendation", MusicSourceFeatureRecommendation.String())
	assert.Equal(t, "lyrics", MusicSourceFeatureLyrics.String())
}

// BenchmarkMemoryCache 缓存性能基准测试
func BenchmarkMemoryCache(b *testing.B) {
	cache := NewMemoryCache(10000, 10*time.Minute)
	ctx := context.Background()

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i)
			value := fmt.Sprintf("value_%d", i)
			cache.Set(ctx, key, value, 5*time.Minute)
		}
	})

	b.Run("Get", func(b *testing.B) {
		// 预先设置一些数据
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key_%d", i)
			value := fmt.Sprintf("value_%d", i)
			cache.Set(ctx, key, value, 5*time.Minute)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%1000)
			cache.Get(ctx, key)
		}
	})
}