package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/kernel"
	"github.com/go-musicfox/go-musicfox/v2/test/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// PluginSystemIntegrationTestSuite 插件系统集成测试套件
type PluginSystemIntegrationTestSuite struct {
	suite.Suite
	kernel        kernel.Kernel
	pluginManager kernel.PluginManager
	ctx           context.Context
	cancel        context.CancelFunc
}

// SetupSuite 设置测试套件
func (suite *PluginSystemIntegrationTestSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 60*time.Second)
}

// TearDownSuite 清理测试套件
func (suite *PluginSystemIntegrationTestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

// SetupTest 设置每个测试
func (suite *PluginSystemIntegrationTestSuite) SetupTest() {
	// 创建并初始化微内核
	suite.kernel = kernel.NewMicroKernel()
	err := suite.kernel.Initialize(suite.ctx)
	suite.Require().NoError(err)
	err = suite.kernel.Start(suite.ctx)
	suite.Require().NoError(err)

	// 获取插件管理器
	suite.pluginManager = suite.kernel.GetPluginManager()
	suite.Require().NotNil(suite.pluginManager)
}

// TearDownTest 清理每个测试
func (suite *PluginSystemIntegrationTestSuite) TearDownTest() {
	if suite.kernel != nil {
		_ = suite.kernel.Shutdown(suite.ctx)
	}
}

// TestPluginManagerLifecycle 测试插件管理器生命周期
func (suite *PluginSystemIntegrationTestSuite) TestPluginManagerLifecycle() {
	// 验证插件管理器初始状态
	loadedPlugins := suite.pluginManager.GetLoadedPlugins()
	assert.NotNil(suite.T(), loadedPlugins)
	assert.Equal(suite.T(), 0, len(loadedPlugins))

	pluginCount := suite.pluginManager.GetLoadedPluginCount()
	assert.Equal(suite.T(), 0, pluginCount)

	// 测试插件查询
	isLoaded := suite.pluginManager.IsPluginLoaded("non-existent")
	assert.False(suite.T(), isLoaded)

	_, err := suite.pluginManager.GetPlugin("non-existent")
	assert.Error(suite.T(), err)

	_, err = suite.pluginManager.GetPluginInfo("non-existent")
	assert.Error(suite.T(), err)
}

// TestBasicPluginOperations 测试基本插件操作
func (suite *PluginSystemIntegrationTestSuite) TestBasicPluginOperations() {
	// 创建模拟插件
	mockPlugin := fixtures.NewMockPlugin("test-plugin", "1.0.0")

	// 注册插件
	err := suite.pluginManager.RegisterPlugin(mockPlugin)
	assert.NoError(suite.T(), err)

	// 验证插件已注册
	isLoaded := suite.pluginManager.IsPluginLoaded("test-plugin")
	assert.True(suite.T(), isLoaded)

	loadedPlugin, err := suite.pluginManager.GetPlugin("test-plugin")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), loadedPlugin)
	assert.Equal(suite.T(), "test-plugin", loadedPlugin.GetInfo().Name)

	// 启动插件
	err = suite.pluginManager.StartPlugin("test-plugin")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), mockPlugin.IsStartCalled())

	// 停止插件
	err = suite.pluginManager.StopPlugin("test-plugin")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), mockPlugin.IsStopCalled())

	// 注销插件
	err = suite.pluginManager.UnregisterPlugin("test-plugin")
	assert.NoError(suite.T(), err)

	// 验证插件已注销
	isLoaded = suite.pluginManager.IsPluginLoaded("test-plugin")
	assert.False(suite.T(), isLoaded)
}

// TestMultiplePluginsManagement 测试多插件管理
func (suite *PluginSystemIntegrationTestSuite) TestMultiplePluginsManagement() {
	// 创建多个模拟插件
	plugins := []*fixtures.MockPlugin{
		fixtures.NewMockPlugin("plugin-1", "1.0.0"),
		fixtures.NewMockPlugin("plugin-2", "1.1.0"),
		fixtures.NewMockPlugin("plugin-3", "2.0.0"),
	}

	// 注册所有插件
	for _, plugin := range plugins {
		err := suite.pluginManager.RegisterPlugin(plugin)
		assert.NoError(suite.T(), err)
	}

	// 验证插件数量
	loadedPlugins := suite.pluginManager.GetLoadedPlugins()
	assert.Equal(suite.T(), 3, len(loadedPlugins))
	assert.Equal(suite.T(), 3, suite.pluginManager.GetLoadedPluginCount())

	// 启动所有插件
	for _, plugin := range plugins {
		err := suite.pluginManager.StartPlugin(plugin.GetInfo().Name)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), plugin.IsStartCalled())
	}

	// 验证所有插件都在运行
	for _, plugin := range plugins {
		assert.True(suite.T(), plugin.IsRunning())
	}

	// 停止所有插件
	for _, plugin := range plugins {
		err := suite.pluginManager.StopPlugin(plugin.GetInfo().Name)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), plugin.IsStopCalled())
	}

	// 注销所有插件
	for _, plugin := range plugins {
		err := suite.pluginManager.UnregisterPlugin(plugin.GetInfo().Name)
		assert.NoError(suite.T(), err)
	}

	// 验证所有插件都已注销
	assert.Equal(suite.T(), 0, suite.pluginManager.GetLoadedPluginCount())
}

// TestPluginDependencies 测试插件依赖关系
func (suite *PluginSystemIntegrationTestSuite) TestPluginDependencies() {
	// 创建有依赖关系的插件
	basePlugin := fixtures.NewMockPlugin("base-plugin", "1.0.0")
	dependentPlugin := fixtures.NewMockPluginWithDependencies("dependent-plugin", "1.0.0", []string{"base-plugin"})

	// 先注册依赖插件
	err := suite.pluginManager.RegisterPlugin(basePlugin)
	assert.NoError(suite.T(), err)

	// 再注册依赖于它的插件
	err = suite.pluginManager.RegisterPlugin(dependentPlugin)
	assert.NoError(suite.T(), err)

	// 验证依赖关系
	dependencies := dependentPlugin.GetDependencies()
	assert.Contains(suite.T(), dependencies, "base-plugin")

	// 启动基础插件
	err = suite.pluginManager.StartPlugin("base-plugin")
	assert.NoError(suite.T(), err)

	// 启动依赖插件
	err = suite.pluginManager.StartPlugin("dependent-plugin")
	assert.NoError(suite.T(), err)

	// 验证两个插件都在运行
	assert.True(suite.T(), basePlugin.IsRunning())
	assert.True(suite.T(), dependentPlugin.IsRunning())
}

// TestPluginHealthCheck 测试插件健康检查
func (suite *PluginSystemIntegrationTestSuite) TestPluginHealthCheck() {
	// 创建模拟插件
	mockPlugin := fixtures.NewMockPlugin("health-test-plugin", "1.0.0")

	// 注册并启动插件
	err := suite.pluginManager.RegisterPlugin(mockPlugin)
	assert.NoError(suite.T(), err)
	err = suite.pluginManager.StartPlugin("health-test-plugin")
	assert.NoError(suite.T(), err)

	// 执行健康检查
	err = mockPlugin.HealthCheck()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, mockPlugin.GetHealthCheckCount())

	// 停止插件后健康检查应该失败
	err = suite.pluginManager.StopPlugin("health-test-plugin")
	assert.NoError(suite.T(), err)

	err = mockPlugin.HealthCheck()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), 2, mockPlugin.GetHealthCheckCount())
}

// TestPluginConfigurationManagement 测试插件配置管理
func (suite *PluginSystemIntegrationTestSuite) TestPluginConfigurationManagement() {
	// 创建模拟插件
	mockPlugin := fixtures.NewMockPlugin("config-test-plugin", "1.0.0")

	// 注册插件
	err := suite.pluginManager.RegisterPlugin(mockPlugin)
	assert.NoError(suite.T(), err)

	// 测试配置验证
	validConfig := map[string]interface{}{
		"required_key": "valid_value",
		"optional_key": "optional_value",
	}
	err = mockPlugin.ValidateConfig(validConfig)
	assert.NoError(suite.T(), err)

	// 测试无效配置
	invalidConfig := map[string]interface{}{
		"required_key": "", // 空值应该无效
	}
	err = mockPlugin.ValidateConfig(invalidConfig)
	assert.Error(suite.T(), err)

	// 测试配置更新
	err = mockPlugin.UpdateConfig(validConfig)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, mockPlugin.GetConfigUpdates())

	// 测试无效配置更新
	err = mockPlugin.UpdateConfig(invalidConfig)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), 1, mockPlugin.GetConfigUpdates()) // 更新次数不应该增加
}

// TestAudioProcessorPluginIntegration 测试音频处理插件集成
func (suite *PluginSystemIntegrationTestSuite) TestAudioProcessorPluginIntegration() {
	// 创建模拟音频处理插件
	audioPlugin := fixtures.NewMockAudioProcessorPlugin()

	// 注册并启动插件
	err := suite.pluginManager.RegisterPlugin(audioPlugin)
	assert.NoError(suite.T(), err)
	err = suite.pluginManager.StartPlugin(audioPlugin.GetInfo().Name)
	assert.NoError(suite.T(), err)

	// 测试音频处理功能
	testAudio := []byte{0x01, 0x02, 0x03, 0x04}
	processedAudio, err := audioPlugin.ProcessAudio(testAudio, 44100, 2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testAudio, processedAudio)
	assert.Equal(suite.T(), 1, audioPlugin.GetProcessCount())

	// 测试音量调节
	adjustedAudio, err := audioPlugin.AdjustVolume(testAudio, 0.5)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testAudio, adjustedAudio)
	assert.Equal(suite.T(), 1, audioPlugin.GetVolumeAdjustments())

	// 测试无效音量
	_, err = audioPlugin.AdjustVolume(testAudio, 1.5)
	assert.Error(suite.T(), err)

	// 测试音效应用
	effectedAudio, err := audioPlugin.ApplyEffect(testAudio, "reverb")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testAudio, effectedAudio)
	assert.Equal(suite.T(), 1, audioPlugin.GetEffectsApplied())
}

// TestMusicSourcePluginIntegration 测试音乐源插件集成
func (suite *PluginSystemIntegrationTestSuite) TestMusicSourcePluginIntegration() {
	// 创建模拟音乐源插件
	musicPlugin := fixtures.NewMockMusicSourcePlugin()

	// 注册并启动插件
	err := suite.pluginManager.RegisterPlugin(musicPlugin)
	assert.NoError(suite.T(), err)
	err = suite.pluginManager.StartPlugin(musicPlugin.GetInfo().Name)
	assert.NoError(suite.T(), err)

	// 测试搜索功能
	searchResults, err := musicPlugin.Search(suite.ctx, "test query")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), searchResults, 2)
	assert.Equal(suite.T(), 1, musicPlugin.GetSearchCount())

	// 验证搜索结果
	firstResult := searchResults[0]
	assert.Equal(suite.T(), "1", firstResult["id"])
	assert.Equal(suite.T(), "Mock Song 1", firstResult["title"])

	// 测试获取播放列表
	playlist, err := musicPlugin.GetPlaylist(suite.ctx, "test-playlist-id")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-playlist-id", playlist["id"])
	assert.Equal(suite.T(), "Mock Playlist", playlist["name"])
	assert.Equal(suite.T(), 1, musicPlugin.GetPlaylistRequests())

	// 测试获取歌曲信息
	song, err := musicPlugin.GetSong(suite.ctx, "test-song-id")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-song-id", song["id"])
	assert.Equal(suite.T(), "Mock Song", song["title"])
	assert.Equal(suite.T(), 1, musicPlugin.GetSongRequests())
}

// TestConcurrentPluginOperations 测试并发插件操作
func (suite *PluginSystemIntegrationTestSuite) TestConcurrentPluginOperations() {
	const numPlugins = 5
	const numGoroutines = 10

	// 创建多个插件
	plugins := make([]*fixtures.MockPlugin, numPlugins)
	for i := 0; i < numPlugins; i++ {
		plugins[i] = fixtures.NewMockPlugin(fmt.Sprintf("concurrent-plugin-%d", i), "1.0.0")
		err := suite.pluginManager.RegisterPlugin(plugins[i])
		assert.NoError(suite.T(), err)
	}

	// 并发启动插件
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(pluginIndex int) {
			defer func() { done <- true }()

			pluginName := fmt.Sprintf("concurrent-plugin-%d", pluginIndex%numPlugins)
			err := suite.pluginManager.StartPlugin(pluginName)
			assert.NoError(suite.T(), err)

			// 执行一些操作
			plugin, err := suite.pluginManager.GetPlugin(pluginName)
			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), plugin)

			// 健康检查
			err = plugin.HealthCheck()
			assert.NoError(suite.T(), err)
		}(i)
	}

	// 等待所有协程完成
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// 协程完成
		case <-time.After(10 * time.Second):
			suite.T().Fatal("Concurrent operations timeout")
		}
	}

	// 验证所有插件状态
	for _, plugin := range plugins {
		assert.True(suite.T(), plugin.IsStartCalled())
		assert.True(suite.T(), plugin.IsRunning())
	}
}

// TestPluginReload 测试插件重载
func (suite *PluginSystemIntegrationTestSuite) TestPluginReload() {
	// 创建模拟插件
	mockPlugin := fixtures.NewMockPlugin("reload-test-plugin", "1.0.0")

	// 注册并启动插件
	err := suite.pluginManager.RegisterPlugin(mockPlugin)
	assert.NoError(suite.T(), err)
	err = suite.pluginManager.StartPlugin("reload-test-plugin")
	assert.NoError(suite.T(), err)

	// 验证插件运行状态
	assert.True(suite.T(), mockPlugin.IsRunning())
	assert.True(suite.T(), mockPlugin.IsStartCalled())

	// 重载插件
	err = suite.pluginManager.ReloadPlugin("reload-test-plugin")
	assert.NoError(suite.T(), err)

	// 验证插件重载后的状态
	// 注意：重载可能会创建新的插件实例，这里主要测试重载操作不出错
	isLoaded := suite.pluginManager.IsPluginLoaded("reload-test-plugin")
	assert.True(suite.T(), isLoaded)
}

// TestPluginSystemIntegration 运行插件系统集成测试
func TestPluginSystemIntegration(t *testing.T) {
	suite.Run(t, new(PluginSystemIntegrationTestSuite))
}