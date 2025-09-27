package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/kernel"
	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MockAudioPlugin 模拟音频插件
type MockAudioPlugin struct {
	info         *kernel.PluginInfo
	state        *model.PlayerState
	volume       int
	mutex        sync.RWMutex
	eventBus     event.EventBus
	isPlaying    bool
	currentSong  *model.Song
	initialized  bool
	started      bool
}

// NewMockAudioPlugin 创建模拟音频插件
func NewMockAudioPlugin(eventBus event.EventBus) *MockAudioPlugin {
	info := &kernel.PluginInfo{
		Name:        "Mock Audio Plugin",
		Version:     "1.0.0",
		Description: "Mock audio plugin for testing",
		Author:      "test",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return &MockAudioPlugin{
		info: info,
		state: &model.PlayerState{
			Status: model.PlayStatusStopped,
		},
		volume:   80,
		eventBus: eventBus,
	}
}

// GetInfo 获取插件信息
func (p *MockAudioPlugin) GetInfo() *kernel.PluginInfo {
	return p.info
}

// GetCapabilities 获取插件能力
func (p *MockAudioPlugin) GetCapabilities() []string {
	return []string{"audio_playback", "volume_control"}
}

// GetDependencies 获取插件依赖
func (p *MockAudioPlugin) GetDependencies() []string {
	return []string{"event_bus"}
}

// Initialize 初始化插件
func (p *MockAudioPlugin) Initialize(ctx kernel.PluginContext) error {
	p.initialized = true
	return nil
}

// Start 启动插件
func (p *MockAudioPlugin) Start() error {
	p.started = true
	return nil
}

// Stop 停止插件
func (p *MockAudioPlugin) Stop() error {
	p.started = false
	return nil
}

// Cleanup 清理插件
func (p *MockAudioPlugin) Cleanup() error {
	p.initialized = false
	p.started = false
	return nil
}

// HealthCheck 健康检查
func (p *MockAudioPlugin) HealthCheck() error {
	if !p.initialized {
		return fmt.Errorf("plugin not initialized")
	}
	return nil
}

// ValidateConfig 验证配置
func (p *MockAudioPlugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

// UpdateConfig 更新配置
func (p *MockAudioPlugin) UpdateConfig(config map[string]interface{}) error {
	return nil
}

// Play 播放音乐
func (p *MockAudioPlugin) Play(song *model.Song) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if song.URL == "invalid://url" {
		p.state.Status = model.PlayStatusError
		return fmt.Errorf("invalid URL: %s", song.URL)
	}

	p.currentSong = song
	p.state.CurrentSong = song
	p.state.Status = model.PlayStatusPlaying
	p.isPlaying = true

	// 发布播放开始事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("audio-play-start-%d", time.Now().UnixNano()),
			Type:      event.EventType("audio.play.start"),
			Source:    p.GetInfo().Name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"song": song,
			},
		}
		p.eventBus.PublishAsync(context.Background(), event)
	}

	return nil
}

// Pause 暂停播放
func (p *MockAudioPlugin) Pause() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.state.Status = model.PlayStatusPaused
	p.isPlaying = false
	return nil
}

// Resume 恢复播放
func (p *MockAudioPlugin) Resume() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.state.Status = model.PlayStatusPlaying
	p.isPlaying = true
	return nil
}

// StopPlayback 停止播放
func (p *MockAudioPlugin) StopPlayback() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.state.Status = model.PlayStatusStopped
	p.state.Position = 0
	p.isPlaying = false
	return nil
}

// SetVolume 设置音量
func (p *MockAudioPlugin) SetVolume(volume int) error {
	if volume < 0 || volume > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.volume = volume
	return nil
}

// GetVolume 获取音量
func (p *MockAudioPlugin) GetVolume() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.volume
}

// GetState 获取播放状态
func (p *MockAudioPlugin) GetState() *model.PlayerState {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return &model.PlayerState{
		Status:      p.state.Status,
		CurrentSong: p.state.CurrentSong,
		Position:    p.state.Position,
	}
}

// MockPlaylistPlugin 模拟播放列表插件
type MockPlaylistPlugin struct {
	info         *kernel.PluginInfo
	playlists    map[string]*model.Playlist
	currentQueue []*model.Song
	history      []*model.Song
	playMode     model.PlayMode
	eventBus     event.EventBus
	mutex        sync.RWMutex
	initialized  bool
	started      bool
}

// NewMockPlaylistPlugin 创建模拟播放列表插件
func NewMockPlaylistPlugin(eventBus event.EventBus) *MockPlaylistPlugin {
	info := &kernel.PluginInfo{
		Name:        "Mock Playlist Plugin",
		Version:     "1.0.0",
		Description: "Mock playlist plugin for testing",
		Author:      "test",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return &MockPlaylistPlugin{
		info:         info,
		playlists:    make(map[string]*model.Playlist),
		currentQueue: make([]*model.Song, 0),
		history:      make([]*model.Song, 0),
		playMode:     model.PlayModeSequential,
		eventBus:     eventBus,
	}
}

// GetInfo 获取插件信息
func (p *MockPlaylistPlugin) GetInfo() *kernel.PluginInfo {
	return p.info
}

// GetCapabilities 获取插件能力
func (p *MockPlaylistPlugin) GetCapabilities() []string {
	return []string{"playlist_management", "queue_management"}
}

// GetDependencies 获取插件依赖
func (p *MockPlaylistPlugin) GetDependencies() []string {
	return []string{"event_bus"}
}

// Initialize 初始化插件
func (p *MockPlaylistPlugin) Initialize(ctx kernel.PluginContext) error {
	p.initialized = true
	return nil
}

// Start 启动插件
func (p *MockPlaylistPlugin) Start() error {
	p.started = true
	return nil
}

// Stop 停止插件
func (p *MockPlaylistPlugin) Stop() error {
	p.started = false
	return nil
}

// Cleanup 清理插件
func (p *MockPlaylistPlugin) Cleanup() error {
	p.initialized = false
	p.started = false
	return nil
}

// HealthCheck 健康检查
func (p *MockPlaylistPlugin) HealthCheck() error {
	if !p.initialized {
		return fmt.Errorf("plugin not initialized")
	}
	return nil
}

// ValidateConfig 验证配置
func (p *MockPlaylistPlugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

// UpdateConfig 更新配置
func (p *MockPlaylistPlugin) UpdateConfig(config map[string]interface{}) error {
	return nil
}

// CreatePlaylist 创建播放列表
func (p *MockPlaylistPlugin) CreatePlaylist(ctx context.Context, name, description string) (*model.Playlist, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	playlist := &model.Playlist{
		ID:          fmt.Sprintf("playlist-%d", time.Now().UnixNano()),
		Name:        name,
		Description: description,
		Songs:       make([]*model.Song, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	p.playlists[playlist.ID] = playlist

	// 发布播放列表创建事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("playlist-created-%d", time.Now().UnixNano()),
			Type:      event.EventType("playlist.created"),
			Source:    p.GetInfo().Name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"playlist": playlist,
			},
		}
		p.eventBus.PublishAsync(ctx, event)
	}

	return playlist, nil
}

// GetPlaylist 获取播放列表
func (p *MockPlaylistPlugin) GetPlaylist(ctx context.Context, playlistID string) (*model.Playlist, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	playlist, exists := p.playlists[playlistID]
	if !exists {
		return nil, fmt.Errorf("playlist not found: %s", playlistID)
	}

	return playlist, nil
}

// AddSong 添加歌曲到播放列表
func (p *MockPlaylistPlugin) AddSong(ctx context.Context, playlistID string, song *model.Song) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	playlist, exists := p.playlists[playlistID]
	if !exists {
		return fmt.Errorf("playlist not found: %s", playlistID)
	}

	playlist.Songs = append(playlist.Songs, song)
	playlist.UpdatedAt = time.Now()

	return nil
}

// AddToQueue 添加到播放队列
func (p *MockPlaylistPlugin) AddToQueue(ctx context.Context, song *model.Song) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.currentQueue = append(p.currentQueue, song)
	return nil
}

// GetCurrentQueue 获取当前播放队列
func (p *MockPlaylistPlugin) GetCurrentQueue(ctx context.Context) ([]*model.Song, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	queue := make([]*model.Song, len(p.currentQueue))
	copy(queue, p.currentQueue)
	return queue, nil
}

// SetCurrentQueue 设置当前播放队列
func (p *MockPlaylistPlugin) SetCurrentQueue(ctx context.Context, songs []*model.Song) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.currentQueue = make([]*model.Song, len(songs))
	copy(p.currentQueue, songs)
	return nil
}

// AddToHistory 添加到历史记录
func (p *MockPlaylistPlugin) AddToHistory(ctx context.Context, song *model.Song) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.history = append(p.history, song)
	// 限制历史记录长度
	if len(p.history) > 100 {
		p.history = p.history[1:]
	}
	return nil
}

// GetHistory 获取历史记录
func (p *MockPlaylistPlugin) GetHistory(ctx context.Context, limit int) ([]*model.Song, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if limit > len(p.history) {
		limit = len(p.history)
	}

	history := make([]*model.Song, limit)
	copy(history, p.history[len(p.history)-limit:])
	return history, nil
}

// SetPlayMode 设置播放模式
func (p *MockPlaylistPlugin) SetPlayMode(ctx context.Context, mode model.PlayMode) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.playMode = mode
	return nil
}

// GetPlayMode 获取播放模式
func (p *MockPlaylistPlugin) GetPlayMode(ctx context.Context) model.PlayMode {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.playMode
}

// GetNextSong 获取下一首歌
func (p *MockPlaylistPlugin) GetNextSong(ctx context.Context, currentSong *model.Song) (*model.Song, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if len(p.currentQueue) == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	// 找到当前歌曲在队列中的位置
	for i, song := range p.currentQueue {
		if song.ID == currentSong.ID {
			if i+1 < len(p.currentQueue) {
				return p.currentQueue[i+1], nil
			}
			break
		}
	}

	// 如果是最后一首或找不到，返回第一首
	return p.currentQueue[0], nil
}

// GetPreviousSong 获取上一首歌
func (p *MockPlaylistPlugin) GetPreviousSong(ctx context.Context, currentSong *model.Song) (*model.Song, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if len(p.currentQueue) == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	// 找到当前歌曲在队列中的位置
	for i, song := range p.currentQueue {
		if song.ID == currentSong.ID {
			if i > 0 {
				return p.currentQueue[i-1], nil
			}
			break
		}
	}

	// 如果是第一首或找不到，返回最后一首
	return p.currentQueue[len(p.currentQueue)-1], nil
}

// Stage2IntegrationTestSuite 阶段2集成测试套件
type Stage2IntegrationTestSuite struct {
	suite.Suite
	kernel         kernel.Kernel
	pluginManager  kernel.PluginManager
	eventBus       event.EventBus
	audioPlugin    *MockAudioPlugin
	playlistPlugin *MockPlaylistPlugin
	ctx            context.Context
	cancel         context.CancelFunc
	events         []event.Event
	eventMutex     sync.RWMutex
}

// SetupSuite 设置测试套件
func (suite *Stage2IntegrationTestSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 120*time.Second)
	suite.events = make([]event.Event, 0)
}

// TearDownSuite 清理测试套件
func (suite *Stage2IntegrationTestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

// SetupTest 设置每个测试
func (suite *Stage2IntegrationTestSuite) SetupTest() {
	// 创建并初始化微内核
	suite.kernel = kernel.NewMicroKernel()
	err := suite.kernel.Initialize(suite.ctx)
	suite.Require().NoError(err)
	err = suite.kernel.Start(suite.ctx)
	suite.Require().NoError(err)

	// 获取核心组件
	suite.pluginManager = suite.kernel.GetPluginManager()
	suite.Require().NotNil(suite.pluginManager)

	suite.eventBus = suite.kernel.GetEventBus()
	suite.Require().NotNil(suite.eventBus)

	// 创建并注册模拟插件
	suite.audioPlugin = NewMockAudioPlugin(suite.eventBus)
	err = suite.pluginManager.RegisterPlugin(suite.audioPlugin)
	suite.Require().NoError(err)

	suite.playlistPlugin = NewMockPlaylistPlugin(suite.eventBus)
	err = suite.pluginManager.RegisterPlugin(suite.playlistPlugin)
	suite.Require().NoError(err)

	// 启动插件
	err = suite.pluginManager.StartPlugin(suite.audioPlugin.GetInfo().Name)
	suite.Require().NoError(err)
	err = suite.pluginManager.StartPlugin(suite.playlistPlugin.GetInfo().Name)
	suite.Require().NoError(err)

	// 订阅事件用于测试
	suite.subscribeToEvents()

	// 清空事件记录
	suite.eventMutex.Lock()
	suite.events = make([]event.Event, 0)
	suite.eventMutex.Unlock()
}

// TearDownTest 清理每个测试
func (suite *Stage2IntegrationTestSuite) TearDownTest() {
	if suite.kernel != nil {
		_ = suite.kernel.Shutdown(suite.ctx)
	}
}

// subscribeToEvents 订阅事件用于测试验证
func (suite *Stage2IntegrationTestSuite) subscribeToEvents() {
	// 订阅所有音频相关事件
	audioEvents := []event.EventType{
		event.EventType("audio.play.start"),
		event.EventType("audio.play.pause"),
		event.EventType("audio.play.resume"),
		event.EventType("audio.play.stop"),
	}

	for _, eventType := range audioEvents {
		_, err := suite.eventBus.Subscribe(eventType, suite.handleTestEvent)
		suite.Require().NoError(err)
	}

	// 订阅播放列表相关事件
	playlistEvents := []event.EventType{
		event.EventType("playlist.created"),
		event.EventType("playlist.updated"),
		event.EventType("playlist.song.added"),
	}

	for _, eventType := range playlistEvents {
		_, err := suite.eventBus.Subscribe(eventType, suite.handleTestEvent)
		suite.Require().NoError(err)
	}
}

// handleTestEvent 处理测试事件
func (suite *Stage2IntegrationTestSuite) handleTestEvent(ctx context.Context, event event.Event) error {
	suite.eventMutex.Lock()
	defer suite.eventMutex.Unlock()
	suite.events = append(suite.events, event)
	return nil
}

// getReceivedEvents 获取接收到的事件
func (suite *Stage2IntegrationTestSuite) getReceivedEvents() []event.Event {
	suite.eventMutex.RLock()
	defer suite.eventMutex.RUnlock()
	events := make([]event.Event, len(suite.events))
	copy(events, suite.events)
	return events
}

// waitForEvents 等待指定数量的事件
func (suite *Stage2IntegrationTestSuite) waitForEvents(expectedCount int, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if len(suite.getReceivedEvents()) >= expectedCount {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// TestAudioPlaylistPluginIntegration 测试音频插件和播放列表插件集成
func (suite *Stage2IntegrationTestSuite) TestAudioPlaylistPluginIntegration() {
	// 创建测试歌曲
	testSong := &model.Song{
		ID:       "test-song-1",
		Title:    "Test Song 1",
		Artist:   "Test Artist",
		Album:    "Test Album",
		Duration: 180 * time.Second,
		URL:      "http://example.com/test1.mp3",
		Source:   "test",
	}

	// 1. 测试播放列表创建和歌曲添加
	playlist, err := suite.playlistPlugin.CreatePlaylist(suite.ctx, "Test Playlist", "Integration test playlist")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), playlist)
	assert.Equal(suite.T(), "Test Playlist", playlist.Name)

	err = suite.playlistPlugin.AddSong(suite.ctx, playlist.ID, testSong)
	assert.NoError(suite.T(), err)

	// 2. 测试将歌曲添加到播放队列
	err = suite.playlistPlugin.AddToQueue(suite.ctx, testSong)
	assert.NoError(suite.T(), err)

	queue, err := suite.playlistPlugin.GetCurrentQueue(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), queue, 1)
	assert.Equal(suite.T(), testSong.ID, queue[0].ID)

	// 3. 测试音频播放
	err = suite.audioPlugin.Play(testSong)
	assert.NoError(suite.T(), err)

	// 验证播放状态
	state := suite.audioPlugin.GetState()
	assert.Equal(suite.T(), model.PlayStatusPlaying, state.Status)
	assert.Equal(suite.T(), testSong.ID, state.CurrentSong.ID)

	// 4. 测试播放历史记录
	err = suite.playlistPlugin.AddToHistory(suite.ctx, testSong)
	assert.NoError(suite.T(), err)

	history, err := suite.playlistPlugin.GetHistory(suite.ctx, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), history, 1)
	assert.Equal(suite.T(), testSong.ID, history[0].ID)

	// 5. 验证事件通信
	suite.waitForEvents(1, 2*time.Second)
	events := suite.getReceivedEvents()
	assert.GreaterOrEqual(suite.T(), len(events), 1)

	// 查找播放开始事件
	playStartEventFound := false
	for _, evt := range events {
		if string(evt.GetType()) == "audio.play.start" {
			playStartEventFound = true
			break
		}
	}
	assert.True(suite.T(), playStartEventFound, "Should receive audio play start event")
}

// TestEventSystemPluginCommunication 测试事件系统和插件通信
func (suite *Stage2IntegrationTestSuite) TestEventSystemPluginCommunication() {
	// 创建测试歌曲
	testSong := &model.Song{
		ID:       "test-song-2",
		Title:    "Test Song 2",
		Artist:   "Test Artist 2",
		Album:    "Test Album 2",
		Duration: 200 * time.Second,
		URL:      "http://example.com/test2.mp3",
		Source:   "test",
	}

	// 1. 测试播放事件传播
	err := suite.audioPlugin.Play(testSong)
	assert.NoError(suite.T(), err)

	// 等待事件传播
	suite.waitForEvents(1, 2*time.Second)

	// 2. 测试暂停事件
	err = suite.audioPlugin.Pause()
	assert.NoError(suite.T(), err)

	// 3. 测试恢复事件
	err = suite.audioPlugin.Resume()
	assert.NoError(suite.T(), err)

	// 4. 测试音量变化
	err = suite.audioPlugin.SetVolume(50)
	assert.NoError(suite.T(), err)

	// 5. 测试停止事件
	err = suite.audioPlugin.StopPlayback()
	assert.NoError(suite.T(), err)

	// 验证事件接收
	events := suite.getReceivedEvents()
	assert.GreaterOrEqual(suite.T(), len(events), 1)

	// 验证事件类型
	eventTypes := make(map[string]bool)
	for _, evt := range events {
		eventTypes[string(evt.GetType())] = true
	}

	// 应该至少包含播放开始事件
	assert.True(suite.T(), eventTypes["audio.play.start"], "Should have audio play start event")
}

// TestPlaybackControlFlow 测试播放控制完整流程
func (suite *Stage2IntegrationTestSuite) TestPlaybackControlFlow() {
	// 创建测试播放列表
	testSongs := []*model.Song{
		{
			ID:       "song-1",
			Title:    "Song 1",
			Artist:   "Artist 1",
			Duration: 180 * time.Second,
			URL:      "http://example.com/song1.mp3",
			Source:   "test",
		},
		{
			ID:       "song-2",
			Title:    "Song 2",
			Artist:   "Artist 2",
			Duration: 200 * time.Second,
			URL:      "http://example.com/song2.mp3",
			Source:   "test",
		},
		{
			ID:       "song-3",
			Title:    "Song 3",
			Artist:   "Artist 3",
			Duration: 220 * time.Second,
			URL:      "http://example.com/song3.mp3",
			Source:   "test",
		},
	}

	// 1. 创建播放列表并添加歌曲
	playlist, err := suite.playlistPlugin.CreatePlaylist(suite.ctx, "Flow Test Playlist", "Test complete playback flow")
	assert.NoError(suite.T(), err)

	for _, song := range testSongs {
		err = suite.playlistPlugin.AddSong(suite.ctx, playlist.ID, song)
		assert.NoError(suite.T(), err)
	}

	// 2. 设置播放队列
	err = suite.playlistPlugin.SetCurrentQueue(suite.ctx, testSongs)
	assert.NoError(suite.T(), err)

	queue, err := suite.playlistPlugin.GetCurrentQueue(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), queue, 3)

	// 3. 测试顺序播放模式
	err = suite.playlistPlugin.SetPlayMode(suite.ctx, model.PlayModeSequential)
	assert.NoError(suite.T(), err)

	playMode := suite.playlistPlugin.GetPlayMode(suite.ctx)
	assert.Equal(suite.T(), model.PlayModeSequential, playMode)

	// 4. 播放第一首歌
	err = suite.audioPlugin.Play(testSongs[0])
	assert.NoError(suite.T(), err)

	state := suite.audioPlugin.GetState()
	assert.Equal(suite.T(), model.PlayStatusPlaying, state.Status)
	assert.Equal(suite.T(), testSongs[0].ID, state.CurrentSong.ID)

	// 5. 测试下一首歌逻辑
	nextSong, err := suite.playlistPlugin.GetNextSong(suite.ctx, testSongs[0])
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testSongs[1].ID, nextSong.ID)

	// 6. 播放下一首
	err = suite.audioPlugin.Play(nextSong)
	assert.NoError(suite.T(), err)

	// 添加到历史记录
	err = suite.playlistPlugin.AddToHistory(suite.ctx, testSongs[0])
	assert.NoError(suite.T(), err)
	err = suite.playlistPlugin.AddToHistory(suite.ctx, nextSong)
	assert.NoError(suite.T(), err)

	// 7. 验证历史记录
	history, err := suite.playlistPlugin.GetHistory(suite.ctx, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), history, 2)

	// 8. 测试上一首歌逻辑
	prevSong, err := suite.playlistPlugin.GetPreviousSong(suite.ctx, nextSong)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testSongs[0].ID, prevSong.ID)

	// 验证事件传播
	suite.waitForEvents(2, 3*time.Second)
	events := suite.getReceivedEvents()
	assert.GreaterOrEqual(suite.T(), len(events), 1)
}

// TestErrorHandlingAndRecovery 测试错误处理和恢复机制
func (suite *Stage2IntegrationTestSuite) TestErrorHandlingAndRecovery() {
	// 1. 测试无效歌曲URL的错误处理
	invalidSong := &model.Song{
		ID:       "invalid-song",
		Title:    "Invalid Song",
		Artist:   "Test Artist",
		Duration: 180 * time.Second,
		URL:      "invalid://url",
		Source:   "test",
	}

	err := suite.audioPlugin.Play(invalidSong)
	assert.Error(suite.T(), err)

	// 验证播放状态应该是错误状态
	state := suite.audioPlugin.GetState()
	assert.Equal(suite.T(), model.PlayStatusError, state.Status)

	// 2. 测试音量范围错误处理
	err = suite.audioPlugin.SetVolume(-10)
	assert.Error(suite.T(), err)

	err = suite.audioPlugin.SetVolume(150)
	assert.Error(suite.T(), err)

	// 3. 测试无效播放列表操作的错误处理
	_, err = suite.playlistPlugin.GetPlaylist(suite.ctx, "non-existent-playlist")
	assert.Error(suite.T(), err)

	// 4. 测试恢复到正常状态
	validSong := &model.Song{
		ID:       "valid-song",
		Title:    "Valid Song",
		Artist:   "Test Artist",
		Duration: 180 * time.Second,
		URL:      "http://example.com/valid.mp3",
		Source:   "test",
	}

	err = suite.audioPlugin.Play(validSong)
	assert.NoError(suite.T(), err)

	state = suite.audioPlugin.GetState()
	assert.Equal(suite.T(), model.PlayStatusPlaying, state.Status)
	assert.Equal(suite.T(), validSong.ID, state.CurrentSong.ID)

	// 验证错误恢复后的正常功能
	err = suite.audioPlugin.SetVolume(75)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 75, suite.audioPlugin.GetVolume())
}

// TestStage2Integration 运行阶段2集成测试
func TestStage2Integration(t *testing.T) {
	suite.Run(t, new(Stage2IntegrationTestSuite))
}