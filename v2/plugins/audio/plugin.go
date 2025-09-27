package audio

import (
	"context"
	"fmt"
	"sync"
	"time"

	event "github.com/go-musicfox/go-musicfox/v2/pkg/event"
	model "github.com/go-musicfox/go-musicfox/v2/pkg/model"
	audio "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/audio"
	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// AudioPlugin 音频处理插件实现
type AudioPlugin struct {
	*core.BasePlugin
	playerFactory *PlayerFactory
	currentPlayer PlayerBackend
	state         *PlayState
	volume        int
	mutex         sync.RWMutex
	eventBus      event.EventBus
	config        *AudioConfig
}

// AudioConfig 音频插件配置
type AudioConfig struct {
	DefaultBackend string                 `json:"default_backend"`
	BackendConfigs map[string]interface{} `json:"backend_configs"`
	BufferSize     int                    `json:"buffer_size"`
	SampleRate     int                    `json:"sample_rate"`
	Channels       int                    `json:"channels"`
	Volume         int                    `json:"volume"`
	EnableEffects  bool                   `json:"enable_effects"`
	Effects        []audio.AudioEffect    `json:"effects"`
}

// PlayState 播放状态
type PlayState struct {
	Status      model.PlayStatus `json:"status"`
	CurrentSong *model.Song      `json:"current_song"`
	Position    time.Duration    `json:"position"`
	Duration    time.Duration    `json:"duration"`
	Playlist    *model.Playlist  `json:"playlist"`
	PlayIndex   int              `json:"play_index"`
	Shuffle     bool             `json:"shuffle"`
	PlayMode    model.PlayMode   `json:"play_mode"`
}

// NewAudioPlugin 创建音频插件实例
func NewAudioPlugin(eventBus event.EventBus) *AudioPlugin {
	info := &core.PluginInfo{
		Name:        "Audio Processor Plugin",
		Version:     "1.0.0",
		Description: "Core audio processing plugin with multiple backend support",
		Author:      "go-musicfox",
	}

	plugin := &AudioPlugin{
		BasePlugin:    core.NewBasePlugin(info),
		playerFactory: NewPlayerFactory(),
		state: &PlayState{
			Status: model.PlayStatusStopped,
		},
		volume:   80,
		eventBus: eventBus,
		config: &AudioConfig{
			DefaultBackend: "beep",
			BufferSize:     4096,
			SampleRate:     44100,
			Channels:       2,
			Volume:         80,
			EnableEffects:  false,
		},
	}

	return plugin
}

// Initialize 初始化插件
func (p *AudioPlugin) Initialize(pluginCtx core.PluginContext) error {
	if err := p.BasePlugin.Initialize(pluginCtx); err != nil {
		return fmt.Errorf("failed to initialize base plugin: %w", err)
	}

	// 解析配置
	pluginConfig := pluginCtx.GetPluginConfig()
	if pluginConfig != nil {
		customConfig := pluginConfig.GetCustomConfig()
		if backend, ok := customConfig["backend"].(string); ok {
			p.config.DefaultBackend = backend
		}
		if volume, ok := customConfig["volume"].(int); ok {
			p.volume = volume
		}
	}

	// 初始化默认播放器后端
	backend, err := p.playerFactory.CreatePlayer(p.config.DefaultBackend, p.config.BackendConfigs)
	if err != nil {
		return fmt.Errorf("failed to create default player backend: %w", err)
	}

	p.currentPlayer = backend
	return nil
}

// Start 启动插件
func (p *AudioPlugin) Start() error {
	// 启动基础插件
	if err := p.BasePlugin.Start(); err != nil {
		return err
	}

	// 初始化默认播放器
	if p.config.DefaultBackend != "" {
		if err := p.SwitchBackend(p.config.DefaultBackend); err != nil {
			return fmt.Errorf("failed to initialize default backend: %w", err)
		}
	}

	// 发布插件启动事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("audio-plugin-start-%d", time.Now().UnixNano()),
			Type:      "audio.plugin.start",
			Source:    p.GetInfo().Name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"plugin":  p.GetInfo().Name,
				"backend": p.config.DefaultBackend,
			},
		}
		p.eventBus.PublishAsync(context.Background(), event)
	}

	return nil
}

// Stop 停止插件
func (p *AudioPlugin) Stop() error {
	// 停止播放
	if p.currentPlayer != nil {
		p.currentPlayer.Stop()
		p.currentPlayer.Cleanup()
	}

	// 更新状态
	p.mutex.Lock()
	p.state.Status = model.PlayStatusStopped
	p.mutex.Unlock()

	// 停止基础插件
	if err := p.BasePlugin.Stop(); err != nil {
		return err
	}

	// 发布播放停止事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("audio-play-stop-%d", time.Now().UnixNano()),
			Type:      "audio.play.stop",
			Source:    p.GetInfo().Name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"plugin": p.GetInfo().Name,
			},
		}
		p.eventBus.PublishAsync(context.Background(), event)
	}

	return nil
}

// Play 播放音乐
func (p *AudioPlugin) Play(song *model.Song) error {
	if p.currentPlayer == nil {
		return fmt.Errorf("no player backend available")
	}

	p.mutex.Lock()
	p.state.CurrentSong = song
	p.state.Status = model.PlayStatusPlaying
	p.mutex.Unlock()

	// 播放音乐
	if err := p.currentPlayer.Play(song.URL); err != nil {
		p.mutex.Lock()
		p.state.Status = model.PlayStatusError
		p.mutex.Unlock()
		return fmt.Errorf("failed to play song: %w", err)
	}

	// 发布播放开始事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("audio-play-start-%d", time.Now().UnixNano()),
			Type:      "audio.play.start",
			Source:    p.GetInfo().Name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"plugin": p.GetInfo().Name,
				"song":   song,
			},
		}
		p.eventBus.PublishAsync(context.Background(), event)
	}

	return nil
}

// Pause 暂停播放
func (p *AudioPlugin) Pause() error {
	if p.currentPlayer == nil {
		return fmt.Errorf("no player backend available")
	}

	if err := p.currentPlayer.Pause(); err != nil {
		return err
	}

	p.mutex.Lock()
	p.state.Status = model.PlayStatusPaused
	p.mutex.Unlock()

	return nil
}

// Resume 恢复播放
func (p *AudioPlugin) Resume() error {
	if p.currentPlayer == nil {
		return fmt.Errorf("no player backend available")
	}

	if err := p.currentPlayer.Resume(); err != nil {
		return err
	}

	p.mutex.Lock()
	p.state.Status = model.PlayStatusPlaying
	p.mutex.Unlock()

	return nil
}

// Stop 停止播放
func (p *AudioPlugin) StopPlayback() error {
	if p.currentPlayer == nil {
		return fmt.Errorf("no player backend available")
	}

	if err := p.currentPlayer.Stop(); err != nil {
		return err
	}

	p.mutex.Lock()
	p.state.Status = model.PlayStatusStopped
	p.state.Position = 0
	p.mutex.Unlock()

	return nil
}

// Seek 跳转到指定位置
func (p *AudioPlugin) Seek(position time.Duration) error {
	if p.currentPlayer == nil {
		return fmt.Errorf("no player backend available")
	}

	if err := p.currentPlayer.Seek(position); err != nil {
		return err
	}

	p.mutex.Lock()
	p.state.Position = position
	p.mutex.Unlock()

	return nil
}

// SetVolume 设置音量
func (p *AudioPlugin) SetVolume(volume int) error {
	if volume < 0 || volume > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}

	if p.currentPlayer == nil {
		return fmt.Errorf("no player backend available")
	}

	if err := p.currentPlayer.SetVolume(float64(volume) / 100.0); err != nil {
		return err
	}

	p.mutex.Lock()
	p.volume = volume
	p.mutex.Unlock()

	return nil
}

// GetVolume 获取音量
func (p *AudioPlugin) GetVolume() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.volume
}

// GetState 获取播放状态
func (p *AudioPlugin) GetState() *PlayState {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// 更新当前位置
	if p.currentPlayer != nil {
		if pos, err := p.currentPlayer.GetPosition(); err == nil {
			p.state.Position = pos
		}
		if dur, err := p.currentPlayer.GetDuration(); err == nil {
			p.state.Duration = dur
		}
	}

	return &PlayState{
		Status:      p.state.Status,
		CurrentSong: p.state.CurrentSong,
		Position:    p.state.Position,
		Duration:    p.state.Duration,
		PlayMode:    p.state.PlayMode,
	}
}

// GetPlayState 获取播放状态（别名方法）
func (p *AudioPlugin) GetPlayState() *PlayState {
	return p.GetState()
}

// GetPosition 获取播放位置
func (p *AudioPlugin) GetPosition() time.Duration {
	if p.currentPlayer == nil {
		return 0
	}

	pos, _ := p.currentPlayer.GetPosition()
	return pos
}

// GetDuration 获取音乐时长
func (p *AudioPlugin) GetDuration() time.Duration {
	if p.currentPlayer == nil {
		return 0
	}

	dur, _ := p.currentPlayer.GetDuration()
	return dur
}

// SwitchBackend 切换播放器后端
func (p *AudioPlugin) SwitchBackend(backendName string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 停止当前播放器
	if p.currentPlayer != nil {
		p.currentPlayer.Stop()
		p.currentPlayer.Cleanup()
	}

	// 创建新的播放器后端
	newPlayer, err := p.playerFactory.CreatePlayer(backendName, p.config.BackendConfigs)
	if err != nil {
		return fmt.Errorf("failed to create player backend '%s': %w", backendName, err)
	}

	// 初始化新播放器
	if err := newPlayer.Initialize(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize player backend '%s': %w", backendName, err)
	}

	p.currentPlayer = newPlayer
	p.config.DefaultBackend = backendName

	return nil
}

// GetAvailableBackends 获取可用的播放器后端
func (p *AudioPlugin) GetAvailableBackends() []string {
	return p.playerFactory.GetAvailableBackends()
}

// GetCurrentBackend 获取当前播放器后端名称
func (p *AudioPlugin) GetCurrentBackend() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.config.DefaultBackend
}
