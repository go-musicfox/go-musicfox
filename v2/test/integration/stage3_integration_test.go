package integration

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	event "github.com/go-musicfox/go-musicfox/v2/pkg/event"
	kernel "github.com/go-musicfox/go-musicfox/v2/pkg/kernel"
	model "github.com/go-musicfox/go-musicfox/v2/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// SearchType 搜索类型
type SearchType int

const (
	SearchTypeSong SearchType = iota
	SearchTypeAlbum
	SearchTypeArtist
	SearchTypePlaylist
)

// SearchResult 搜索结果
type SearchResult struct {
	Query     string           `json:"query"`
	Type      SearchType       `json:"type"`
	Songs     []*model.Song    `json:"songs"`
	Playlists []*model.Playlist `json:"playlists"`
	Total     int              `json:"total"`
}

// MockNeteasePlugin 模拟网易云音乐插件
type MockNeteasePlugin struct {
	info         *kernel.PluginInfo
	loggedIn     bool
	userID       string
	username     string
	playlists    map[string]*model.Playlist
	songs        map[string]*model.Song
	mutex        sync.RWMutex
	eventBus     event.EventBus
	initialized  bool
	started      bool
}

// NewMockNeteasePlugin 创建模拟网易云音乐插件
func NewMockNeteasePlugin(eventBus event.EventBus) *MockNeteasePlugin {
	info := &kernel.PluginInfo{
		Name:        "Mock Netease Plugin",
		Version:     "1.0.0",
		Description: "Mock netease music plugin for testing",
		Author:      "test",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 预填充一些测试数据
	playlists := make(map[string]*model.Playlist)
	songs := make(map[string]*model.Song)

	// 创建测试歌曲
	testSongs := []*model.Song{
		{
			ID:       "netease-song-1",
			Title:    "网易云测试歌曲1",
			Artist:   "测试歌手1",
			Album:    "测试专辑1",
			Duration: 180 * time.Second,
			URL:      "http://music.163.com/song/1.mp3",
			Source:   "netease",
		},
		{
			ID:       "netease-song-2",
			Title:    "网易云测试歌曲2",
			Artist:   "测试歌手2",
			Album:    "测试专辑2",
			Duration: 200 * time.Second,
			URL:      "http://music.163.com/song/2.mp3",
			Source:   "netease",
		},
	}

	// 添加歌曲到map
	for _, song := range testSongs {
		songs[song.ID] = song
	}

	// 创建测试播放列表
	testPlaylist := &model.Playlist{
		ID:          "netease-playlist-1",
		Name:        "我的收藏",
		Description: "测试播放列表",
		Songs:       testSongs,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Source:      "netease",
	}
	playlists[testPlaylist.ID] = testPlaylist

	return &MockNeteasePlugin{
		info:      info,
		playlists: playlists,
		songs:     songs,
		eventBus:  eventBus,
	}
}

// GetInfo 获取插件信息
func (p *MockNeteasePlugin) GetInfo() *kernel.PluginInfo {
	return p.info
}

// GetCapabilities 获取插件能力
func (p *MockNeteasePlugin) GetCapabilities() []string {
	return []string{"music_search", "playlist_management", "user_auth"}
}

// GetDependencies 获取插件依赖
func (p *MockNeteasePlugin) GetDependencies() []string {
	return []string{"event_bus"}
}

// Initialize 初始化插件
func (p *MockNeteasePlugin) Initialize(ctx kernel.PluginContext) error {
	p.initialized = true
	return nil
}

// Start 启动插件
func (p *MockNeteasePlugin) Start() error {
	p.started = true
	return nil
}

// Stop 停止插件
func (p *MockNeteasePlugin) Stop() error {
	p.started = false
	return nil
}

// Cleanup 清理插件
func (p *MockNeteasePlugin) Cleanup() error {
	p.initialized = false
	p.started = false
	return nil
}

// HealthCheck 健康检查
func (p *MockNeteasePlugin) HealthCheck() error {
	if !p.initialized {
		return fmt.Errorf("plugin not initialized")
	}
	return nil
}

// ValidateConfig 验证配置
func (p *MockNeteasePlugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

// UpdateConfig 更新配置
func (p *MockNeteasePlugin) UpdateConfig(config map[string]interface{}) error {
	return nil
}

// Login 用户登录
func (p *MockNeteasePlugin) Login(ctx context.Context, username, password string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if username == "test@example.com" && password == "password123" {
		p.loggedIn = true
		p.userID = "test-user-123"
		p.username = "测试用户"

		// 发布登录成功事件
		if p.eventBus != nil {
			now := time.Now()
			event := &event.BaseEvent{
				ID:        fmt.Sprintf("netease-login-%d-%s", now.UnixNano(), p.userID),
				Type:      event.EventType("netease.user.login"),
				Source:    p.GetInfo().Name,
				Timestamp: now,
				Data: map[string]interface{}{
					"user_id":  p.userID,
					"username": p.username,
				},
			}
			p.eventBus.PublishAsync(ctx, event)
		}
		return nil
	}

	return fmt.Errorf("invalid credentials")
}

// Logout 用户登出
func (p *MockNeteasePlugin) Logout(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.loggedIn = false
	p.userID = ""
	p.username = ""

	// 发布登出事件
	if p.eventBus != nil {
		now := time.Now()
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("netease-logout-%d", now.UnixNano()),
			Type:      event.EventType("netease.user.logout"),
			Source:    p.GetInfo().Name,
			Timestamp: now,
			Data:      map[string]interface{}{},
		}
		p.eventBus.PublishAsync(ctx, event)
	}

	return nil
}

// IsLoggedIn 检查是否已登录
func (p *MockNeteasePlugin) IsLoggedIn() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.loggedIn
}

// Search 搜索音乐
func (p *MockNeteasePlugin) Search(ctx context.Context, query string, searchType SearchType) (*SearchResult, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	result := &SearchResult{
		Query: query,
		Type:  searchType,
		Songs: make([]*model.Song, 0),
		Playlists: make([]*model.Playlist, 0),
	}

	switch searchType {
	case SearchTypeSong:
		// 模拟搜索歌曲
		for _, song := range p.songs {
			if song.Title == query || song.Artist == query {
				result.Songs = append(result.Songs, song)
			}
		}
	case SearchTypePlaylist:
		// 模拟搜索播放列表
		for _, playlist := range p.playlists {
			if playlist.Name == query {
				result.Playlists = append(result.Playlists, playlist)
			}
		}
	}

	result.Total = len(result.Songs) + len(result.Playlists)
	return result, nil
}

// GetUserPlaylists 获取用户播放列表
func (p *MockNeteasePlugin) GetUserPlaylists(ctx context.Context) ([]*model.Playlist, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.loggedIn {
		return nil, fmt.Errorf("user not logged in")
	}

	playlists := make([]*model.Playlist, 0, len(p.playlists))
	for _, playlist := range p.playlists {
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

// GetSongURL 获取歌曲播放URL
func (p *MockNeteasePlugin) GetSongURL(ctx context.Context, songID string) (string, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	song, exists := p.songs[songID]
	if !exists {
		return "", fmt.Errorf("song not found: %s", songID)
	}

	return song.URL, nil
}

// MockTUIPlugin 模拟TUI插件
type MockTUIPlugin struct {
	info         *kernel.PluginInfo
	currentView  string
	statusMsg    string
	components   map[string]interface{}
	keyHandlers  map[string]func() error
	mutex        sync.RWMutex
	eventBus     event.EventBus
	initialized  bool
	started      bool
	isRendering  bool
}

// NewMockTUIPlugin 创建模拟TUI插件
func NewMockTUIPlugin(eventBus event.EventBus) *MockTUIPlugin {
	info := &kernel.PluginInfo{
		Name:        "Mock TUI Plugin",
		Version:     "1.0.0",
		Description: "Mock TUI plugin for testing",
		Author:      "test",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return &MockTUIPlugin{
		info:        info,
		currentView: "main",
		statusMsg:   "Welcome to MusicFox!",
		components:  make(map[string]interface{}),
		keyHandlers: make(map[string]func() error),
		eventBus:    eventBus,
	}
}

// GetInfo 获取插件信息
func (p *MockTUIPlugin) GetInfo() *kernel.PluginInfo {
	return p.info
}

// GetCapabilities 获取插件能力
func (p *MockTUIPlugin) GetCapabilities() []string {
	return []string{"ui_rendering", "key_handling", "event_display"}
}

// GetDependencies 获取插件依赖
func (p *MockTUIPlugin) GetDependencies() []string {
	return []string{"event_bus"}
}

// Initialize 初始化插件
func (p *MockTUIPlugin) Initialize(ctx kernel.PluginContext) error {
	p.initialized = true
	return nil
}

// Start 启动插件
func (p *MockTUIPlugin) Start() error {
	p.started = true
	p.isRendering = true
	return nil
}

// Stop 停止插件
func (p *MockTUIPlugin) Stop() error {
	p.started = false
	p.isRendering = false
	return nil
}

// Cleanup 清理插件
func (p *MockTUIPlugin) Cleanup() error {
	p.initialized = false
	p.started = false
	p.isRendering = false
	return nil
}

// HealthCheck 健康检查
func (p *MockTUIPlugin) HealthCheck() error {
	if !p.initialized {
		return fmt.Errorf("plugin not initialized")
	}
	return nil
}

// ValidateConfig 验证配置
func (p *MockTUIPlugin) ValidateConfig(config map[string]interface{}) error {
	return nil
}

// UpdateConfig 更新配置
func (p *MockTUIPlugin) UpdateConfig(config map[string]interface{}) error {
	return nil
}

// Render 渲染界面
func (p *MockTUIPlugin) Render(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isRendering {
		return fmt.Errorf("TUI not rendering")
	}

	// 模拟渲染过程
	time.Sleep(10 * time.Millisecond)
	return nil
}

// HandleKeyEvent 处理键盘事件
func (p *MockTUIPlugin) HandleKeyEvent(key string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if handler, exists := p.keyHandlers[key]; exists {
		return handler()
	}

	// 默认键盘处理
	switch key {
	case "q":
		p.currentView = "quit"
	case "h":
		p.currentView = "help"
	case "m":
		p.currentView = "main"
	case "s":
		p.currentView = "search"
	case "p":
		p.currentView = "player"
	}

	// 发布键盘事件
		if p.eventBus != nil {
			now := time.Now()
			event := &event.BaseEvent{
				ID:        fmt.Sprintf("tui-key-%d-%s", now.UnixNano(), key),
				Type:      event.EventType("tui.key.pressed"),
				Source:    p.GetInfo().Name,
				Timestamp: now,
				Data: map[string]interface{}{
					"key":  key,
					"view": p.currentView,
				},
			}
			p.eventBus.PublishAsync(context.Background(), event)
		}

	return nil
}

// SetCurrentView 设置当前视图
func (p *MockTUIPlugin) SetCurrentView(view string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.currentView = view
}

// GetCurrentView 获取当前视图
func (p *MockTUIPlugin) GetCurrentView() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.currentView
}

// SetStatusMessage 设置状态消息
func (p *MockTUIPlugin) SetStatusMessage(msg string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.statusMsg = msg
}

// GetStatusMessage 获取状态消息
func (p *MockTUIPlugin) GetStatusMessage() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.statusMsg
}

// AddComponent 添加UI组件
func (p *MockTUIPlugin) AddComponent(name string, component interface{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.components[name] = component
}

// GetComponent 获取UI组件
func (p *MockTUIPlugin) GetComponent(name string) (interface{}, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	component, exists := p.components[name]
	return component, exists
}

// RegisterKeyHandler 注册键盘处理器
func (p *MockTUIPlugin) RegisterKeyHandler(key string, handler func() error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.keyHandlers[key] = handler
}

// Stage3IntegrationTestSuite 阶段3集成测试套件
type Stage3IntegrationTestSuite struct {
	suite.Suite
	ctx             context.Context
	cancel          context.CancelFunc
	eventBus        event.EventBus
	audioPlugin     *MockAudioPlugin
	playlistPlugin  *MockPlaylistPlugin
	neteasePlugin   *MockNeteasePlugin
	tuiPlugin       *MockTUIPlugin
	receivedEvents  []event.Event
	eventMutex      sync.RWMutex
}

// SetupSuite 设置测试套件
func (suite *Stage3IntegrationTestSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithCancel(context.Background())

	// 创建事件总线
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // 减少测试日志输出
	}))
	suite.eventBus = event.NewEventBus(logger)

	// 启动事件总线
	err := suite.eventBus.Start(suite.ctx)
	suite.Require().NoError(err)

	// 创建插件实例
	suite.audioPlugin = NewMockAudioPlugin(suite.eventBus)
	suite.playlistPlugin = NewMockPlaylistPlugin(suite.eventBus)
	suite.neteasePlugin = NewMockNeteasePlugin(suite.eventBus)
	suite.tuiPlugin = NewMockTUIPlugin(suite.eventBus)

	// 初始化插件
	mockCtx := &MockPluginContext{}
	suite.audioPlugin.Initialize(mockCtx)
	suite.playlistPlugin.Initialize(mockCtx)
	suite.neteasePlugin.Initialize(mockCtx)
	suite.tuiPlugin.Initialize(mockCtx)

	// 启动插件
	suite.audioPlugin.Start()
	suite.playlistPlugin.Start()
	suite.neteasePlugin.Start()
	suite.tuiPlugin.Start()

	// 订阅具体事件类型用于测试验证
	_, err = suite.eventBus.Subscribe("netease.user.login", suite.handleTestEvent)
	suite.Require().NoError(err)
	_, err = suite.eventBus.Subscribe("netease.user.logout", suite.handleTestEvent)
	suite.Require().NoError(err)
	_, err = suite.eventBus.Subscribe("tui.key.pressed", suite.handleTestEvent)
	suite.Require().NoError(err)
	_, err = suite.eventBus.Subscribe("audio.play.start", suite.handleTestEvent)
	suite.Require().NoError(err)
	_, err = suite.eventBus.Subscribe("audio.play.stop", suite.handleTestEvent)
	suite.Require().NoError(err)
	_, err = suite.eventBus.Subscribe("playlist.created", suite.handleTestEvent)
	suite.Require().NoError(err)
}

// TeardownSuite 清理测试套件
func (suite *Stage3IntegrationTestSuite) TeardownSuite() {
	// 停止插件
	if suite.audioPlugin != nil {
		suite.audioPlugin.Stop()
		suite.audioPlugin.Cleanup()
	}
	if suite.playlistPlugin != nil {
		suite.playlistPlugin.Stop()
		suite.playlistPlugin.Cleanup()
	}
	if suite.neteasePlugin != nil {
		suite.neteasePlugin.Stop()
		suite.neteasePlugin.Cleanup()
	}
	if suite.tuiPlugin != nil {
		suite.tuiPlugin.Stop()
		suite.tuiPlugin.Cleanup()
	}

	// 停止事件总线
	if suite.eventBus != nil {
		err := suite.eventBus.Stop(suite.ctx)
		if err != nil {
			suite.T().Logf("Error stopping event bus: %v", err)
		}
	}

	// 取消上下文
	if suite.cancel != nil {
		suite.cancel()
	}

	// 等待一小段时间确保所有goroutine都停止
	time.Sleep(100 * time.Millisecond)
}

// MockPluginContext 模拟插件上下文
type MockPluginContext struct{}

// GetServiceRegistry 获取服务注册表
func (m *MockPluginContext) GetServiceRegistry() kernel.ServiceRegistry {
	return nil
}

// GetEventBus 获取事件总线
func (ctx *MockPluginContext) GetEventBus() event.EventBus {
	return nil
}

// GetConfig 获取配置
func (ctx *MockPluginContext) GetConfig() map[string]interface{} {
	return make(map[string]interface{})
}

// GetLogger 获取日志器
func (m *MockPluginContext) GetLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// GetSecurityManager 获取安全管理器
func (m *MockPluginContext) GetSecurityManager() kernel.SecurityManager {
	return nil
}

// handleTestEvent 处理测试事件
func (suite *Stage3IntegrationTestSuite) handleTestEvent(ctx context.Context, evt event.Event) error {
	suite.eventMutex.Lock()
	defer suite.eventMutex.Unlock()
	suite.receivedEvents = append(suite.receivedEvents, evt)
	return nil
}

// getReceivedEvents 获取接收到的事件
func (suite *Stage3IntegrationTestSuite) getReceivedEvents() []event.Event {
	suite.eventMutex.RLock()
	defer suite.eventMutex.RUnlock()
	events := make([]event.Event, len(suite.receivedEvents))
	copy(events, suite.receivedEvents)
	return events
}

// clearReceivedEvents 清空接收到的事件
func (suite *Stage3IntegrationTestSuite) clearReceivedEvents() {
	suite.eventMutex.Lock()
	defer suite.eventMutex.Unlock()
	suite.receivedEvents = suite.receivedEvents[:0]
}

// waitForEvents 等待事件
func (suite *Stage3IntegrationTestSuite) waitForEvents(expectedCount int, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if len(suite.getReceivedEvents()) >= expectedCount {
			return true
		}
		time.Sleep(1 * time.Millisecond) // 减少sleep时间
	}
	return false
}

// TestNeteasePluginAndAudioPluginIntegration 测试网易云插件和音频插件集成
func (suite *Stage3IntegrationTestSuite) TestNeteasePluginAndAudioPluginIntegration() {
	// 清空之前的事件
	suite.clearReceivedEvents()

	// 重置插件状态
	suite.neteasePlugin.Logout(suite.ctx)
	suite.audioPlugin.Stop()

	// 1. 测试用户登录
	err := suite.neteasePlugin.Login(suite.ctx, "test@example.com", "password123")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), suite.neteasePlugin.IsLoggedIn())

	// 2. 测试搜索歌曲
	searchResult, err := suite.neteasePlugin.Search(suite.ctx, "网易云测试歌曲1", SearchTypeSong)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), searchResult)
	assert.Len(suite.T(), searchResult.Songs, 1)
	assert.Equal(suite.T(), "网易云测试歌曲1", searchResult.Songs[0].Title)

	// 3. 测试获取歌曲URL并播放
	song := searchResult.Songs[0]
	songURL, err := suite.neteasePlugin.GetSongURL(suite.ctx, song.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), song.URL, songURL)

	// 4. 测试音频播放
	err = suite.audioPlugin.Play(song)
	assert.NoError(suite.T(), err)

	// 验证播放状态
	state := suite.audioPlugin.GetState()
	assert.Equal(suite.T(), model.PlayStatusPlaying, state.Status)
	assert.Equal(suite.T(), song.ID, state.CurrentSong.ID)

	// 5. 测试获取用户播放列表
	playlists, err := suite.neteasePlugin.GetUserPlaylists(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), playlists, 1)
	assert.Equal(suite.T(), "我的收藏", playlists[0].Name)

	// 6. 验证事件传播
	suite.waitForEvents(1, 500*time.Millisecond) // 减少等待时间
	events := suite.getReceivedEvents()
	assert.GreaterOrEqual(suite.T(), len(events), 1)

	// 查找关键事件
	eventTypes := make(map[string]bool)
	for _, evt := range events {
		eventTypes[string(evt.GetType())] = true
	}
	// 至少应该有一些事件
	assert.True(suite.T(), len(eventTypes) > 0, "Should receive some events")
}

// TestTUIPluginAndEventSystemIntegration 测试TUI插件和事件系统集成
func (suite *Stage3IntegrationTestSuite) TestTUIPluginAndEventSystemIntegration() {
	// 清空之前的事件
	suite.clearReceivedEvents()

	// 重置TUI插件状态
	suite.tuiPlugin.SetCurrentView("main")
	suite.tuiPlugin.SetStatusMessage("Welcome to MusicFox!")

	// 1. 测试TUI插件初始状态
	assert.Equal(suite.T(), "main", suite.tuiPlugin.GetCurrentView())
	assert.Equal(suite.T(), "Welcome to MusicFox!", suite.tuiPlugin.GetStatusMessage())

	// 2. 测试键盘事件处理
	err := suite.tuiPlugin.HandleKeyEvent("s")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "search", suite.tuiPlugin.GetCurrentView())

	err = suite.tuiPlugin.HandleKeyEvent("p")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "player", suite.tuiPlugin.GetCurrentView())

	err = suite.tuiPlugin.HandleKeyEvent("m")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "main", suite.tuiPlugin.GetCurrentView())

	// 3. 测试UI组件管理
	testComponent := "test-component-data"
	suite.tuiPlugin.AddComponent("test", testComponent)
	component, exists := suite.tuiPlugin.GetComponent("test")
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), testComponent, component)

	// 4. 测试状态消息更新
	suite.tuiPlugin.SetStatusMessage("正在播放音乐...")
	assert.Equal(suite.T(), "正在播放音乐...", suite.tuiPlugin.GetStatusMessage())

	// 5. 测试界面渲染
	err = suite.tuiPlugin.Render(suite.ctx)
	assert.NoError(suite.T(), err)

	// 6. 验证键盘事件传播
	suite.waitForEvents(1, 500*time.Millisecond) // 减少等待时间
	events := suite.getReceivedEvents()
	assert.GreaterOrEqual(suite.T(), len(events), 1)

	// 验证有事件产生
	eventTypes := make(map[string]bool)
	for _, evt := range events {
		eventTypes[string(evt.GetType())] = true
	}
	assert.True(suite.T(), len(eventTypes) > 0, "Should receive some events")
}

// TestCompletePlaybackFlow 测试完整的音乐播放流程
func (suite *Stage3IntegrationTestSuite) TestCompletePlaybackFlow() {
	// 清空之前的事件
	suite.clearReceivedEvents()

	// 重置插件状态
	suite.neteasePlugin.Logout(suite.ctx)
	suite.audioPlugin.Stop()
	suite.tuiPlugin.SetCurrentView("main")
	suite.tuiPlugin.SetStatusMessage("Welcome to MusicFox!")

	// 1. 用户登录网易云
	err := suite.neteasePlugin.Login(suite.ctx, "test@example.com", "password123")
	assert.NoError(suite.T(), err)

	// 2. 在TUI中切换到搜索界面
	err = suite.tuiPlugin.HandleKeyEvent("s")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "search", suite.tuiPlugin.GetCurrentView())

	// 3. 搜索音乐
	searchResult, err := suite.neteasePlugin.Search(suite.ctx, "网易云测试歌曲1", SearchTypeSong)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), searchResult.Songs, 1)

	// 4. 将搜索结果添加到播放列表
	song := searchResult.Songs[0]
	playlist, err := suite.playlistPlugin.CreatePlaylist(suite.ctx, "搜索结果", "从搜索添加的歌曲")
	assert.NoError(suite.T(), err)

	err = suite.playlistPlugin.AddSong(suite.ctx, playlist.ID, song)
	assert.NoError(suite.T(), err)

	// 5. 添加到播放队列
	err = suite.playlistPlugin.AddToQueue(suite.ctx, song)
	assert.NoError(suite.T(), err)

	// 6. 切换到播放界面
	err = suite.tuiPlugin.HandleKeyEvent("p")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "player", suite.tuiPlugin.GetCurrentView())

	// 7. 开始播放
	err = suite.audioPlugin.Play(song)
	assert.NoError(suite.T(), err)

	// 8. 更新TUI状态
	suite.tuiPlugin.SetStatusMessage(fmt.Sprintf("正在播放: %s - %s", song.Title, song.Artist))
	expectedMsg := fmt.Sprintf("正在播放: %s - %s", song.Title, song.Artist)
	assert.Equal(suite.T(), expectedMsg, suite.tuiPlugin.GetStatusMessage())

	// 9. 验证播放状态
	state := suite.audioPlugin.GetState()
	assert.Equal(suite.T(), model.PlayStatusPlaying, state.Status)
	assert.Equal(suite.T(), song.ID, state.CurrentSong.ID)

	// 10. 测试音量控制
	err = suite.audioPlugin.SetVolume(75)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 75, suite.audioPlugin.GetVolume())

	// 11. 测试播放控制
	err = suite.audioPlugin.Pause()
	assert.NoError(suite.T(), err)
	state = suite.audioPlugin.GetState()
	assert.Equal(suite.T(), model.PlayStatusPaused, state.Status)

	err = suite.audioPlugin.Resume()
	assert.NoError(suite.T(), err)
	state = suite.audioPlugin.GetState()
	assert.Equal(suite.T(), model.PlayStatusPlaying, state.Status)

	// 12. 添加到历史记录
	err = suite.playlistPlugin.AddToHistory(suite.ctx, song)
	assert.NoError(suite.T(), err)

	history, err := suite.playlistPlugin.GetHistory(suite.ctx, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), history, 1)
	assert.Equal(suite.T(), song.ID, history[0].ID)

	// 13. 验证完整的事件流
	suite.waitForEvents(1, 500*time.Millisecond)
	events := suite.getReceivedEvents()
	assert.GreaterOrEqual(suite.T(), len(events), 1)

	// 验证有事件产生
	eventTypes := make(map[string]bool)
	for _, evt := range events {
		eventTypes[string(evt.GetType())] = true
	}
	assert.True(suite.T(), len(eventTypes) > 0, "Should have some events")
}

// TestUserInteractionAndUIResponse 测试用户交互和UI响应
func (suite *Stage3IntegrationTestSuite) TestUserInteractionAndUIResponse() {
	// 清空之前的事件
	suite.clearReceivedEvents()

	// 重置插件状态
	suite.neteasePlugin.Logout(suite.ctx)
	suite.audioPlugin.Stop()
	suite.tuiPlugin.SetCurrentView("main")
	suite.tuiPlugin.SetStatusMessage("Welcome to MusicFox!")

	// 1. 测试登录流程的UI响应
	// 切换到登录界面（模拟）
	suite.tuiPlugin.SetCurrentView("login")
	suite.tuiPlugin.SetStatusMessage("请输入用户名和密码")

	// 执行登录
	err := suite.neteasePlugin.Login(suite.ctx, "test@example.com", "password123")
	assert.NoError(suite.T(), err)

	// 登录成功后更新UI状态
	suite.tuiPlugin.SetStatusMessage("登录成功!")
	suite.tuiPlugin.SetCurrentView("main")

	assert.Equal(suite.T(), "main", suite.tuiPlugin.GetCurrentView())
	assert.Equal(suite.T(), "登录成功!", suite.tuiPlugin.GetStatusMessage())

	// 2. 测试搜索界面交互
	err = suite.tuiPlugin.HandleKeyEvent("s")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "search", suite.tuiPlugin.GetCurrentView())

	suite.tuiPlugin.SetStatusMessage("输入搜索关键词...")

	// 执行搜索
	searchResult, err := suite.neteasePlugin.Search(suite.ctx, "网易云测试歌曲1", SearchTypeSong)
	assert.NoError(suite.T(), err)

	// 更新搜索结果显示
	if len(searchResult.Songs) > 0 {
		suite.tuiPlugin.SetStatusMessage(fmt.Sprintf("找到 %d 首歌曲", len(searchResult.Songs)))
		suite.tuiPlugin.AddComponent("search_results", searchResult.Songs)
	} else {
		suite.tuiPlugin.SetStatusMessage("未找到相关歌曲")
	}

	assert.Equal(suite.T(), "找到 1 首歌曲", suite.tuiPlugin.GetStatusMessage())
	results, exists := suite.tuiPlugin.GetComponent("search_results")
	assert.True(suite.T(), exists)
	assert.Len(suite.T(), results.([]*model.Song), 1)

	// 3. 测试播放界面交互
	err = suite.tuiPlugin.HandleKeyEvent("p")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "player", suite.tuiPlugin.GetCurrentView())

	// 开始播放并更新UI
	song := searchResult.Songs[0]
	err = suite.audioPlugin.Play(song)
	assert.NoError(suite.T(), err)

	suite.tuiPlugin.SetStatusMessage(fmt.Sprintf("♪ %s - %s", song.Title, song.Artist))
	suite.tuiPlugin.AddComponent("current_song", song)
	suite.tuiPlugin.AddComponent("player_state", suite.audioPlugin.GetState())

	// 验证播放界面状态
	expectedStatus := fmt.Sprintf("♪ %s - %s", song.Title, song.Artist)
	assert.Equal(suite.T(), expectedStatus, suite.tuiPlugin.GetStatusMessage())

	currentSong, exists := suite.tuiPlugin.GetComponent("current_song")
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), song.ID, currentSong.(*model.Song).ID)

	// 4. 测试错误处理和用户反馈
	// 尝试无效登录
	err = suite.neteasePlugin.Logout(suite.ctx)
	assert.NoError(suite.T(), err)

	err = suite.neteasePlugin.Login(suite.ctx, "invalid@example.com", "wrongpassword")
	assert.Error(suite.T(), err)

	// 更新UI显示错误信息
	suite.tuiPlugin.SetStatusMessage("登录失败: 用户名或密码错误")
	assert.Equal(suite.T(), "登录失败: 用户名或密码错误", suite.tuiPlugin.GetStatusMessage())

	// 5. 测试界面导航
	navigationKeys := []string{"m", "s", "p", "h", "m"}
	expectedViews := []string{"main", "search", "player", "help", "main"}

	for i, key := range navigationKeys {
		err = suite.tuiPlugin.HandleKeyEvent(key)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), expectedViews[i], suite.tuiPlugin.GetCurrentView())
	}

	// 6. 验证事件传播和UI响应
	suite.waitForEvents(1, 500*time.Millisecond)
	events := suite.getReceivedEvents()
	assert.GreaterOrEqual(suite.T(), len(events), 1)

	// 验证有事件产生
	eventTypes := make(map[string]bool)
	for _, evt := range events {
		eventTypes[string(evt.GetType())] = true
	}
	assert.True(suite.T(), len(eventTypes) > 0, "Should have some events")
}

// TestErrorHandlingAndRecovery 测试错误处理和恢复
func (suite *Stage3IntegrationTestSuite) TestErrorHandlingAndRecovery() {
	// 清空之前的事件
	suite.clearReceivedEvents()

	// 重置插件状态
	suite.neteasePlugin.Logout(suite.ctx)
	suite.audioPlugin.Stop()
	suite.tuiPlugin.SetCurrentView("main")
	suite.tuiPlugin.SetStatusMessage("Welcome to MusicFox!")

	// 1. 测试网易云插件错误处理
	// 未登录状态下获取播放列表
	assert.False(suite.T(), suite.neteasePlugin.IsLoggedIn())
	_, err := suite.neteasePlugin.GetUserPlaylists(suite.ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "user not logged in")

	// 搜索不存在的歌曲
	searchResult, err := suite.neteasePlugin.Search(suite.ctx, "不存在的歌曲12345", SearchTypeSong)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(searchResult.Songs))

	// 获取不存在的歌曲URL
	_, err = suite.neteasePlugin.GetSongURL(suite.ctx, "non-existent-song-id")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "song not found")

	// 2. 测试音频插件错误处理
	// 播放无效URL的歌曲
	invalidSong := &model.Song{
		ID:     "invalid-song",
		Title:  "Invalid Song",
		Artist: "Test Artist",
		URL:    "invalid://url",
	}

	err = suite.audioPlugin.Play(invalidSong)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid URL")

	// 验证错误状态
	state := suite.audioPlugin.GetState()
	assert.Equal(suite.T(), model.PlayStatusError, state.Status)

	// 测试音量范围错误
	err = suite.audioPlugin.SetVolume(-10)
	assert.Error(suite.T(), err)

	err = suite.audioPlugin.SetVolume(150)
	assert.Error(suite.T(), err)

	// 3. 测试播放列表插件错误处理
	// 获取不存在的播放列表
	_, err = suite.playlistPlugin.GetPlaylist(suite.ctx, "non-existent-playlist")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "playlist not found")

	// 向不存在的播放列表添加歌曲
	testSong := &model.Song{ID: "test-song", Title: "Test Song"}
	err = suite.playlistPlugin.AddSong(suite.ctx, "non-existent-playlist", testSong)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "playlist not found")

	// 4. 测试TUI插件错误处理
	// 停止TUI插件后尝试渲染
	err = suite.tuiPlugin.Stop()
	assert.NoError(suite.T(), err)

	err = suite.tuiPlugin.Render(suite.ctx)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "TUI not rendering")

	// 重新启动TUI插件
	err = suite.tuiPlugin.Start()
	assert.NoError(suite.T(), err)

	err = suite.tuiPlugin.Render(suite.ctx)
	assert.NoError(suite.T(), err)

	// 5. 测试错误恢复
	// 登录网易云恢复正常状态
	err = suite.neteasePlugin.Login(suite.ctx, "test@example.com", "password123")
	assert.NoError(suite.T(), err)

	// 现在应该能获取播放列表了
	playlists, err := suite.neteasePlugin.GetUserPlaylists(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), playlists, 1)

	// 播放有效歌曲恢复音频插件
	validSong := &model.Song{
		ID:     "valid-song",
		Title:  "Valid Song",
		Artist: "Test Artist",
		URL:    "http://example.com/valid.mp3",
	}

	err = suite.audioPlugin.Play(validSong)
	assert.NoError(suite.T(), err)

	state = suite.audioPlugin.GetState()
	assert.Equal(suite.T(), model.PlayStatusPlaying, state.Status)
	assert.Equal(suite.T(), validSong.ID, state.CurrentSong.ID)

	// 设置有效音量
	err = suite.audioPlugin.SetVolume(80)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 80, suite.audioPlugin.GetVolume())

	// 6. 验证错误恢复后的正常功能
	// 创建播放列表并添加歌曲
	playlist, err := suite.playlistPlugin.CreatePlaylist(suite.ctx, "Recovery Test", "Test playlist after error recovery")
	assert.NoError(suite.T(), err)

	err = suite.playlistPlugin.AddSong(suite.ctx, playlist.ID, validSong)
	assert.NoError(suite.T(), err)

	retrievedPlaylist, err := suite.playlistPlugin.GetPlaylist(suite.ctx, playlist.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), playlist.ID, retrievedPlaylist.ID)
	assert.Len(suite.T(), retrievedPlaylist.Songs, 1)

	// TUI插件正常响应键盘事件
	err = suite.tuiPlugin.HandleKeyEvent("m")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "main", suite.tuiPlugin.GetCurrentView())
}

// TestPerformanceAndStability 测试性能和稳定性
func (suite *Stage3IntegrationTestSuite) TestPerformanceAndStability() {
	// 清空之前的事件
	suite.clearReceivedEvents()

	// 重置插件状态
	suite.neteasePlugin.Logout(suite.ctx)
	suite.audioPlugin.Stop()
	suite.tuiPlugin.SetCurrentView("main")
	suite.tuiPlugin.SetStatusMessage("Welcome to MusicFox!")

	// 1. 测试并发操作
	var wg sync.WaitGroup
	errorChan := make(chan error, 100)

	// 并发登录测试
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := suite.neteasePlugin.Login(suite.ctx, "test@example.com", "password123")
			if err != nil {
				errorChan <- err
			}
		}()
	}
	wg.Wait()
	close(errorChan)

	// 检查并发登录错误
	for err := range errorChan {
		assert.NoError(suite.T(), err)
	}

	// 2. 测试大量事件处理
	start := time.Now()
	keys := []string{"m", "s", "p", "h", "q"}
	for i := 0; i < 10; i++ { // 减少事件数量
		key := keys[i%len(keys)] // 循环使用有效的键
		err := suite.tuiPlugin.HandleKeyEvent(key)
		assert.NoError(suite.T(), err)
		time.Sleep(1 * time.Millisecond) // 添加小延迟确保时间戳不同
	}
	duration := time.Since(start)
	assert.Less(suite.T(), duration, 1*time.Second, "10 key events should be processed within 1 second")

	// 3. 测试内存使用
	// 创建大量播放列表
	for i := 0; i < 50; i++ {
		playlist, err := suite.playlistPlugin.CreatePlaylist(suite.ctx, fmt.Sprintf("Test Playlist %d", i), "Performance test playlist")
		assert.NoError(suite.T(), err)

		// 向每个播放列表添加歌曲
		for j := 0; j < 10; j++ {
			song := &model.Song{
				ID:     fmt.Sprintf("song-%d-%d", i, j),
				Title:  fmt.Sprintf("Song %d-%d", i, j),
				Artist: "Test Artist",
				URL:    "http://example.com/test.mp3",
			}
			err = suite.playlistPlugin.AddSong(suite.ctx, playlist.ID, song)
			assert.NoError(suite.T(), err)
		}
	}

	// 4. 测试长时间运行稳定性
	// 模拟长时间播放
	testSong := &model.Song{
		ID:     "stability-test-song",
		Title:  "Stability Test Song",
		Artist: "Test Artist",
		URL:    "http://example.com/stability.mp3",
	}

	err := suite.audioPlugin.Play(testSong)
	assert.NoError(suite.T(), err)

	// 模拟播放过程中的各种操作
	for i := 0; i < 20; i++ {
		// 暂停和恢复
		err = suite.audioPlugin.Pause()
		assert.NoError(suite.T(), err)

		time.Sleep(10 * time.Millisecond)

		err = suite.audioPlugin.Resume()
		assert.NoError(suite.T(), err)

		// 调整音量
		volume := 50 + (i % 50)
		err = suite.audioPlugin.SetVolume(volume)
		assert.NoError(suite.T(), err)

		// UI更新
		suite.tuiPlugin.SetStatusMessage(fmt.Sprintf("Stability test iteration %d", i))
		err = suite.tuiPlugin.Render(suite.ctx)
		assert.NoError(suite.T(), err)
	}

	// 5. 测试事件系统性能
	// 清空之前的事件，避免累积
	suite.clearReceivedEvents()
	
	start2 := time.Now()
	keys2 := []string{"m", "s", "p", "h"}
	for i := 0; i < 10; i++ { // 减少事件数量
		key := keys2[i%len(keys2)] // 循环使用有效的键
		err := suite.tuiPlugin.HandleKeyEvent(key)
		assert.NoError(suite.T(), err)
		time.Sleep(1 * time.Millisecond) // 添加小延迟确保时间戳不同
	}
	eventProcessingDuration := time.Since(start2)
	assert.Less(suite.T(), eventProcessingDuration, 1*time.Second, "10 events should be processed within 1 second")

	// 6. 验证系统仍然正常工作
	// 验证音频插件状态
	state := suite.audioPlugin.GetState()
	assert.Equal(suite.T(), model.PlayStatusPlaying, state.Status)
	assert.Equal(suite.T(), testSong.ID, state.CurrentSong.ID)

	// 验证TUI插件状态（最后一个键是什么就是什么状态）
	currentView := suite.tuiPlugin.GetCurrentView()
	assert.Contains(suite.T(), []string{"main", "search", "player", "help"}, currentView)

	// 验证网易云插件状态
	assert.True(suite.T(), suite.neteasePlugin.IsLoggedIn())

	// 验证播放列表插件状态
	queue, err := suite.playlistPlugin.GetCurrentQueue(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), queue)

	// 7. 验证事件处理
	// 等待事件处理完成
	time.Sleep(100 * time.Millisecond)
	events := suite.getReceivedEvents()
	assert.GreaterOrEqual(suite.T(), len(events), 1, "Should have processed some events")

	// 验证事件ID唯一性
	eventIDs := make(map[string]bool)
	duplicateCount := 0
	for _, evt := range events {
		if eventIDs[evt.GetID()] {
			duplicateCount++
			suite.T().Logf("Duplicate event ID: %s, Type: %s", evt.GetID(), evt.GetType())
		} else {
			eventIDs[evt.GetID()] = true
		}
	}
	assert.Equal(suite.T(), 0, duplicateCount, "Should not have duplicate events")
}

// TestStage3Integration 运行阶段3集成测试
func TestStage3Integration(t *testing.T) {
	suite.Run(t, new(Stage3IntegrationTestSuite))
}