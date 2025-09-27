package e2e

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/kernel"
	"github.com/go-musicfox/go-musicfox/v2/test/fixtures"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MusicPlaybackE2ETestSuite 音乐播放端到端测试套件
type MusicPlaybackE2ETestSuite struct {
	suite.Suite
	kernel        kernel.Kernel
	pluginManager kernel.PluginManager
	eventBus      event.EventBus
	ctx           context.Context
	cancel        context.CancelFunc

	// 插件实例
	musicSourcePlugin *fixtures.MockMusicSourcePlugin
	audioProcessorPlugin *fixtures.MockAudioProcessorPlugin
	uiPlugin          *fixtures.MockPlugin

	// 事件跟踪
	eventsReceived map[string]int
	eventsMutex    sync.RWMutex
}

// SetupSuite 设置测试套件
func (suite *MusicPlaybackE2ETestSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 120*time.Second)
	suite.eventsReceived = make(map[string]int)
}

// TearDownSuite 清理测试套件
func (suite *MusicPlaybackE2ETestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

// SetupTest 设置每个测试
func (suite *MusicPlaybackE2ETestSuite) SetupTest() {
	// 创建并初始化微内核
	suite.kernel = kernel.NewMicroKernel()
	err := suite.kernel.Initialize(suite.ctx)
	suite.Require().NoError(err)
	err = suite.kernel.Start(suite.ctx)
	suite.Require().NoError(err)

	// 获取核心组件
	suite.pluginManager = suite.kernel.GetPluginManager()
	suite.eventBus = suite.kernel.GetEventBus()
	suite.Require().NotNil(suite.pluginManager)
	suite.Require().NotNil(suite.eventBus)

	// 创建插件实例
	suite.musicSourcePlugin = fixtures.NewMockMusicSourcePlugin()
	suite.audioProcessorPlugin = fixtures.NewMockAudioProcessorPlugin()
	suite.uiPlugin = fixtures.NewMockPlugin("ui-plugin", "1.0.0")

	// 设置事件监听
	suite.setupEventListeners()

	// 注册所有插件
	suite.registerAllPlugins()
}

// TearDownTest 清理每个测试
func (suite *MusicPlaybackE2ETestSuite) TearDownTest() {
	if suite.kernel != nil {
		_ = suite.kernel.Shutdown(suite.ctx)
	}

	// 清理事件计数
	suite.eventsMutex.Lock()
	suite.eventsReceived = make(map[string]int)
	suite.eventsMutex.Unlock()
}

// setupEventListeners 设置事件监听器
func (suite *MusicPlaybackE2ETestSuite) setupEventListeners() {
	// 监听音乐播放相关事件
	events := []event.EventType{
		event.EventType("music.search.started"),
		event.EventType("music.search.completed"),
		event.EventType("music.song.loaded"),
		event.EventType("audio.processing.started"),
		event.EventType("audio.processing.completed"),
		event.EventType("ui.updated"),
		event.EventType("playback.started"),
		event.EventType("playback.paused"),
		event.EventType("playback.stopped"),
	}

	for _, eventType := range events {
		_, err := suite.eventBus.Subscribe(eventType, suite.eventHandler(string(eventType)))
		suite.Require().NoError(err)
	}
}

// eventHandler 创建事件处理器
func (suite *MusicPlaybackE2ETestSuite) eventHandler(eventName string) event.EventHandler {
	return func(ctx context.Context, e event.Event) error {
		suite.eventsMutex.Lock()
		defer suite.eventsMutex.Unlock()
		suite.eventsReceived[eventName]++
		return nil
	}
}

// registerAllPlugins 注册所有插件
func (suite *MusicPlaybackE2ETestSuite) registerAllPlugins() {
	// 创建插件上下文
	pluginCtx := fixtures.NewMockPluginContext()
	pluginCtx.SetEventBus(suite.eventBus)
	pluginCtx.SetServiceRegistry(suite.kernel.GetServiceRegistry())
	pluginCtx.SetSecurityManager(suite.kernel.GetSecurityManager())

	// 初始化并注册音乐源插件
	err := suite.musicSourcePlugin.Initialize(pluginCtx)
	suite.Require().NoError(err)
	err = suite.pluginManager.RegisterPlugin(suite.musicSourcePlugin)
	suite.Require().NoError(err)

	// 初始化并注册音频处理插件
	err = suite.audioProcessorPlugin.Initialize(pluginCtx)
	suite.Require().NoError(err)
	err = suite.pluginManager.RegisterPlugin(suite.audioProcessorPlugin)
	suite.Require().NoError(err)

	// 初始化并注册UI插件
	err = suite.uiPlugin.Initialize(pluginCtx)
	suite.Require().NoError(err)
	err = suite.pluginManager.RegisterPlugin(suite.uiPlugin)
	suite.Require().NoError(err)

	// 启动所有插件
	plugins := []kernel.Plugin{
		suite.musicSourcePlugin,
		suite.audioProcessorPlugin,
		suite.uiPlugin,
	}

	for _, plugin := range plugins {
		err := suite.pluginManager.StartPlugin(plugin.GetInfo().Name)
		suite.Require().NoError(err)
	}
}

// getEventCount 获取事件计数
func (suite *MusicPlaybackE2ETestSuite) getEventCount(eventName string) int {
	suite.eventsMutex.RLock()
	defer suite.eventsMutex.RUnlock()
	return suite.eventsReceived[eventName]
}

// waitForEvent 等待特定事件
func (suite *MusicPlaybackE2ETestSuite) waitForEvent(eventName string, expectedCount int, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if suite.getEventCount(eventName) >= expectedCount {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// TestCompletePlaybackFlow 测试完整的播放流程
func (suite *MusicPlaybackE2ETestSuite) TestCompletePlaybackFlow() {
	// 1. 搜索音乐
	suite.T().Log("Step 1: Searching for music...")
	searchStartedEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("music.search.started"),
		Data:      map[string]interface{}{"query": "test song"},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err := suite.eventBus.Publish(suite.ctx, searchStartedEvent)
	assert.NoError(suite.T(), err)

	searchResults, err := suite.musicSourcePlugin.Search(suite.ctx, "test song")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), searchResults, 2)
	assert.Equal(suite.T(), 1, suite.musicSourcePlugin.GetSearchCount())

	searchCompletedEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("music.search.completed"),
		Data:      map[string]interface{}{"results": searchResults, "count": len(searchResults)},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = suite.eventBus.Publish(suite.ctx, searchCompletedEvent)
	assert.NoError(suite.T(), err)

	// 2. 选择并加载歌曲
	suite.T().Log("Step 2: Loading selected song...")
	selectedSongID := searchResults[0]["id"].(string)
	song, err := suite.musicSourcePlugin.GetSong(suite.ctx, selectedSongID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), selectedSongID, song["id"])
	assert.Equal(suite.T(), 1, suite.musicSourcePlugin.GetSongRequests())

	songLoadedEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("music.song.loaded"),
		Data:      map[string]interface{}{"song": song},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = suite.eventBus.Publish(suite.ctx, songLoadedEvent)
	assert.NoError(suite.T(), err)

	// 3. 音频处理
	suite.T().Log("Step 3: Processing audio...")
	audioProcessingStartedEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("audio.processing.started"),
		Data:      map[string]interface{}{"song_id": selectedSongID},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = suite.eventBus.Publish(suite.ctx, audioProcessingStartedEvent)
	assert.NoError(suite.T(), err)

	// 模拟音频数据
	audioData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	// 处理音频
	processedAudio, err := suite.audioProcessorPlugin.ProcessAudio(audioData, 44100, 2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), audioData, processedAudio)
	assert.Equal(suite.T(), 1, suite.audioProcessorPlugin.GetProcessCount())

	// 调节音量
	volumeAdjustedAudio, err := suite.audioProcessorPlugin.AdjustVolume(processedAudio, 0.8)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), processedAudio, volumeAdjustedAudio)
	assert.Equal(suite.T(), 1, suite.audioProcessorPlugin.GetVolumeAdjustments())

	// 应用音效
	finalAudio, err := suite.audioProcessorPlugin.ApplyEffect(volumeAdjustedAudio, "reverb")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), volumeAdjustedAudio, finalAudio)
	assert.Equal(suite.T(), 1, suite.audioProcessorPlugin.GetEffectsApplied())

	audioProcessingCompletedEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("audio.processing.completed"),
		Data:      map[string]interface{}{"song_id": selectedSongID, "processed_size": len(finalAudio)},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = suite.eventBus.Publish(suite.ctx, audioProcessingCompletedEvent)
	assert.NoError(suite.T(), err)

	// 4. UI更新
	suite.T().Log("Step 4: Updating UI...")
	uiUpdatedEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("ui.updated"),
		Data:      map[string]interface{}{"current_song": song, "status": "ready_to_play"},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = suite.eventBus.Publish(suite.ctx, uiUpdatedEvent)
	assert.NoError(suite.T(), err)

	// 5. 开始播放
	suite.T().Log("Step 5: Starting playback...")
	playbackStartedEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("playback.started"),
		Data:      map[string]interface{}{"song_id": selectedSongID, "song": song},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = suite.eventBus.Publish(suite.ctx, playbackStartedEvent)
	assert.NoError(suite.T(), err)

	// 验证事件接收
	suite.T().Log("Step 6: Verifying events...")
	assert.True(suite.T(), suite.waitForEvent("music.search.started", 1, 5*time.Second))
	assert.True(suite.T(), suite.waitForEvent("music.search.completed", 1, 5*time.Second))
	assert.True(suite.T(), suite.waitForEvent("music.song.loaded", 1, 5*time.Second))
	assert.True(suite.T(), suite.waitForEvent("audio.processing.started", 1, 5*time.Second))
	assert.True(suite.T(), suite.waitForEvent("audio.processing.completed", 1, 5*time.Second))
	assert.True(suite.T(), suite.waitForEvent("ui.updated", 1, 5*time.Second))
	assert.True(suite.T(), suite.waitForEvent("playback.started", 1, 5*time.Second))

	suite.T().Log("Complete playback flow test passed!")
}

// TestPlaylistPlayback 测试播放列表播放
func (suite *MusicPlaybackE2ETestSuite) TestPlaylistPlayback() {
	// 1. 获取播放列表
	suite.T().Log("Step 1: Loading playlist...")
	playlist, err := suite.musicSourcePlugin.GetPlaylist(suite.ctx, "test-playlist")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-playlist", playlist["id"])
	assert.Equal(suite.T(), 1, suite.musicSourcePlugin.GetPlaylistRequests())

	// 2. 获取播放列表中的歌曲
	playlistSongs := playlist["songs"].([]map[string]interface{})
	assert.Len(suite.T(), playlistSongs, 1)

	// 3. 逐个播放歌曲
	for i, songInfo := range playlistSongs {
		suite.T().Logf("Step %d: Playing song %s...", i+2, songInfo["title"])

		// 获取完整歌曲信息
		songID := songInfo["id"].(string)
		fullSong, err := suite.musicSourcePlugin.GetSong(suite.ctx, songID)
		assert.NoError(suite.T(), err)

		// 处理音频
		audioData := []byte{0x01, 0x02, 0x03, 0x04}
		processedAudio, err := suite.audioProcessorPlugin.ProcessAudio(audioData, 44100, 2)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), audioData, processedAudio)

		// 发布播放事件
		playbackEvent := &event.BaseEvent{
			ID:        uuid.New().String(),
			Type:      event.EventType("playback.started"),
			Data:      map[string]interface{}{"song_id": songID, "song": fullSong, "playlist_id": "test-playlist", "track_index": i},
			Source:    "test",
			Timestamp: time.Now(),
		}
		err = suite.eventBus.Publish(suite.ctx, playbackEvent)
		assert.NoError(suite.T(), err)
	}

	// 验证统计信息
	assert.Equal(suite.T(), 1, suite.musicSourcePlugin.GetSongRequests())
	assert.Equal(suite.T(), 1, suite.audioProcessorPlugin.GetProcessCount())
	assert.True(suite.T(), suite.waitForEvent("playback.started", 1, 5*time.Second))
}

// TestPlaybackControls 测试播放控制
func (suite *MusicPlaybackE2ETestSuite) TestPlaybackControls() {
	// 1. 开始播放
	suite.T().Log("Step 1: Starting playback...")
	song, err := suite.musicSourcePlugin.GetSong(suite.ctx, "test-song")
	assert.NoError(suite.T(), err)

	playbackStartEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("playback.started"),
		Data:      map[string]interface{}{"song_id": "test-song", "song": song},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = suite.eventBus.Publish(suite.ctx, playbackStartEvent)
	assert.NoError(suite.T(), err)

	// 2. 暂停播放
	suite.T().Log("Step 2: Pausing playback...")
	playbackPausedEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("playback.paused"),
		Data:      map[string]interface{}{"song_id": "test-song", "position": 30.5},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = suite.eventBus.Publish(suite.ctx, playbackPausedEvent)
	assert.NoError(suite.T(), err)

	// 3. 恢复播放
	suite.T().Log("Step 3: Resuming playback...")
	playbackResumedEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("playback.started"),
		Data:      map[string]interface{}{"song_id": "test-song", "song": song, "resumed": true, "position": 30.5},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = suite.eventBus.Publish(suite.ctx, playbackResumedEvent)
	assert.NoError(suite.T(), err)

	// 4. 停止播放
	suite.T().Log("Step 4: Stopping playback...")
	playbackStoppedEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("playback.stopped"),
		Data:      map[string]interface{}{"song_id": "test-song", "reason": "user_requested"},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = suite.eventBus.Publish(suite.ctx, playbackStoppedEvent)
	assert.NoError(suite.T(), err)

	// 验证事件接收
	assert.True(suite.T(), suite.waitForEvent("playback.started", 2, 5*time.Second)) // 开始+恢复
	assert.True(suite.T(), suite.waitForEvent("playback.paused", 1, 5*time.Second))
	assert.True(suite.T(), suite.waitForEvent("playback.stopped", 1, 5*time.Second))
}

// TestAudioEffectsChain 测试音效链处理
func (suite *MusicPlaybackE2ETestSuite) TestAudioEffectsChain() {
	// 1. 获取歌曲
	_, err := suite.musicSourcePlugin.GetSong(suite.ctx, "effects-test-song")
	assert.NoError(suite.T(), err)

	// 2. 创建音频数据
	audioData := []byte{0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}

	// 3. 应用音效链
	suite.T().Log("Step 1: Applying reverb effect...")
	reverbAudio, err := suite.audioProcessorPlugin.ApplyEffect(audioData, "reverb")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), audioData, reverbAudio)

	suite.T().Log("Step 2: Applying echo effect...")
	echoAudio, err := suite.audioProcessorPlugin.ApplyEffect(reverbAudio, "echo")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), reverbAudio, echoAudio)

	suite.T().Log("Step 3: Applying chorus effect...")
	chorusAudio, err := suite.audioProcessorPlugin.ApplyEffect(echoAudio, "chorus")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), echoAudio, chorusAudio)

	// 4. 最终音量调节
	suite.T().Log("Step 4: Final volume adjustment...")
	finalAudio, err := suite.audioProcessorPlugin.AdjustVolume(chorusAudio, 0.7)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), chorusAudio, finalAudio)

	// 验证统计信息
	assert.Equal(suite.T(), 3, suite.audioProcessorPlugin.GetEffectsApplied())
	assert.Equal(suite.T(), 1, suite.audioProcessorPlugin.GetVolumeAdjustments())

	// 发布处理完成事件
	audioCompletedEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("audio.processing.completed"),
		Data:      map[string]interface{}{"song_id": "effects-test-song", "effects_chain": []string{"reverb", "echo", "chorus"}, "final_volume": 0.7, "processed_size": len(finalAudio)},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = suite.eventBus.Publish(suite.ctx, audioCompletedEvent)
	assert.NoError(suite.T(), err)

	assert.True(suite.T(), suite.waitForEvent("audio.processing.completed", 1, 5*time.Second))
}

// TestConcurrentPlayback 测试并发播放场景
func (suite *MusicPlaybackE2ETestSuite) TestConcurrentPlayback() {
	const numConcurrentSongs = 3
	// 使用带缓冲的channel避免死锁
	done := make(chan bool, numConcurrentSongs)
	// 添加错误channel收集goroutine中的错误
	errorChan := make(chan error, numConcurrentSongs)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(suite.ctx, 30*time.Second)
	defer cancel()

	// 并发处理多首歌曲
	for i := 0; i < numConcurrentSongs; i++ {
		go func(songIndex int) {
			defer func() {
				// 使用select避免阻塞
				select {
				case done <- true:
				case <-ctx.Done():
					return
				}
			}()

			songID := fmt.Sprintf("concurrent-song-%d", songIndex)
			suite.T().Logf("Processing song %s...", songID)

			// 获取歌曲信息
			song, err := suite.musicSourcePlugin.GetSong(ctx, songID)
			if err != nil {
				select {
				case errorChan <- fmt.Errorf("failed to get song %s: %w", songID, err):
				case <-ctx.Done():
				}
				return
			}

			// 处理音频
			audioData := []byte{byte(0x10 + songIndex), byte(0x20 + songIndex), byte(0x30 + songIndex)}
			processedAudio, err := suite.audioProcessorPlugin.ProcessAudio(audioData, 44100, 2)
			if err != nil {
				select {
				case errorChan <- fmt.Errorf("failed to process audio for %s: %w", songID, err):
				case <-ctx.Done():
				}
				return
			}
			if !assert.Equal(suite.T(), audioData, processedAudio) {
				select {
				case errorChan <- fmt.Errorf("audio data mismatch for %s", songID):
				case <-ctx.Done():
				}
				return
			}

			// 发布播放事件
			concurrentPlaybackEvent := &event.BaseEvent{
				ID:        uuid.New().String(),
				Type:      event.EventType("playback.started"),
				Data:      map[string]interface{}{"song_id": songID, "song": song, "concurrent": true},
				Source:    "test",
				Timestamp: time.Now(),
			}
			// 使用异步发布避免阻塞
			err = suite.eventBus.PublishAsync(ctx, concurrentPlaybackEvent)
			if err != nil {
				select {
				case errorChan <- fmt.Errorf("failed to publish event for %s: %w", songID, err):
				case <-ctx.Done():
				}
				return
			}
		}(i)
	}

	// 等待所有协程完成或超时
	completedCount := 0
	for completedCount < numConcurrentSongs {
		select {
		case <-done:
			completedCount++
		case err := <-errorChan:
			suite.T().Errorf("Goroutine error: %v", err)
			completedCount++
		case <-ctx.Done():
			suite.T().Fatal("Concurrent playback test timeout")
			return
		}
	}

	// 验证统计信息
	assert.Equal(suite.T(), numConcurrentSongs, suite.musicSourcePlugin.GetSongRequests())
	assert.Equal(suite.T(), numConcurrentSongs, suite.audioProcessorPlugin.GetProcessCount())
	assert.True(suite.T(), suite.waitForEvent("playback.started", numConcurrentSongs, 10*time.Second))
}

// TestErrorHandlingInPlaybackFlow 测试播放流程中的错误处理
func (suite *MusicPlaybackE2ETestSuite) TestErrorHandlingInPlaybackFlow() {
	// 1. 测试无效音量调节
	suite.T().Log("Step 1: Testing invalid volume adjustment...")
	audioData := []byte{0x01, 0x02, 0x03, 0x04}
	_, err := suite.audioProcessorPlugin.AdjustVolume(audioData, 1.5) // 无效音量
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "volume must be between 0 and 1")

	// 2. 测试插件健康检查失败后的恢复
	suite.T().Log("Step 2: Testing plugin recovery after health check failure...")
	
	// 停止音频处理插件
	err = suite.pluginManager.StopPlugin(suite.audioProcessorPlugin.GetInfo().Name)
	assert.NoError(suite.T(), err)

	// 健康检查应该失败
	err = suite.audioProcessorPlugin.HealthCheck()
	assert.Error(suite.T(), err)

	// 重新启动插件
	err = suite.pluginManager.StartPlugin(suite.audioProcessorPlugin.GetInfo().Name)
	assert.NoError(suite.T(), err)

	// 健康检查应该成功
	err = suite.audioProcessorPlugin.HealthCheck()
	assert.NoError(suite.T(), err)

	// 3. 测试恢复后的正常功能
	suite.T().Log("Step 3: Testing normal functionality after recovery...")
	processedAudio, err := suite.audioProcessorPlugin.ProcessAudio(audioData, 44100, 2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), audioData, processedAudio)

	// 发布恢复事件
	pluginRecoveredEvent := &event.BaseEvent{
		ID:        uuid.New().String(),
		Type:      event.EventType("plugin.recovered"),
		Data:      map[string]interface{}{"plugin_name": suite.audioProcessorPlugin.GetInfo().Name, "recovery_time": time.Now().Unix()},
		Source:    "test",
		Timestamp: time.Now(),
	}
	err = suite.eventBus.Publish(suite.ctx, pluginRecoveredEvent)
	assert.NoError(suite.T(), err)
}

// TestMusicPlaybackE2E 运行音乐播放端到端测试
func TestMusicPlaybackE2E(t *testing.T) {
	suite.Run(t, new(MusicPlaybackE2ETestSuite))
}