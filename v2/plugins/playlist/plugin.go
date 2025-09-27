package playlist

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
	plugin "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// PlaylistPlugin 播放列表插件接口
type PlaylistPlugin interface {
	plugin.Plugin

	// 播放列表管理
	CreatePlaylist(ctx context.Context, name, description string) (*model.Playlist, error)
	DeletePlaylist(ctx context.Context, playlistID string) error
	UpdatePlaylist(ctx context.Context, playlistID string, updates map[string]interface{}) error
	GetPlaylist(ctx context.Context, playlistID string) (*model.Playlist, error)
	ListPlaylists(ctx context.Context) ([]*model.Playlist, error)

	// 播放列表歌曲管理
	AddSong(ctx context.Context, playlistID string, song *model.Song) error
	RemoveSong(ctx context.Context, playlistID string, songID string) error
	MoveSong(ctx context.Context, playlistID string, songID string, newIndex int) error
	ClearPlaylist(ctx context.Context, playlistID string) error

	// 播放队列管理
	SetCurrentQueue(ctx context.Context, songs []*model.Song) error
	GetCurrentQueue(ctx context.Context) ([]*model.Song, error)
	AddToQueue(ctx context.Context, song *model.Song) error
	RemoveFromQueue(ctx context.Context, songID string) error
	ClearQueue(ctx context.Context) error
	ShuffleQueue(ctx context.Context) error

	// 播放历史管理
	AddToHistory(ctx context.Context, song *model.Song) error
	GetHistory(ctx context.Context, limit int) ([]*model.Song, error)
	ClearHistory(ctx context.Context) error

	// 播放模式管理
	SetPlayMode(ctx context.Context, mode model.PlayMode) error
	GetPlayMode(ctx context.Context) model.PlayMode
	GetNextSong(ctx context.Context, currentSong *model.Song) (*model.Song, error)
	GetPreviousSong(ctx context.Context, currentSong *model.Song) (*model.Song, error)
}

// PlaylistPluginImpl 播放列表插件实现
type PlaylistPluginImpl struct {
	info         *plugin.PluginInfo
	playlists    map[string]*model.Playlist
	currentQueue []*model.Song
	history      []*model.Song
	playMode     model.PlayMode
	eventBus     event.EventBus
	mu           sync.RWMutex
	maxHistory   int
	shuffleIndex []int
	currentIndex int
}

// NewPlaylistPlugin 创建新的播放列表插件实例
func NewPlaylistPlugin() *PlaylistPluginImpl {
	return &PlaylistPluginImpl{
		info: &plugin.PluginInfo{
			ID:          "playlist",
			Name:        "Playlist Plugin",
			Version:     "1.0.0",
			Description: "Core playlist management plugin",
			Author:      "go-musicfox",
			Type:        "playlist",
			License:     "MIT",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		playlists:    make(map[string]*model.Playlist),
		currentQueue: make([]*model.Song, 0),
		history:      make([]*model.Song, 0),
		playMode:     model.PlayModeSequential,
		maxHistory:   100,
		currentIndex: -1,
	}
}

// GetInfo 获取插件信息
func (p *PlaylistPluginImpl) GetInfo() *plugin.PluginInfo {
	return p.info
}

// GetCapabilities 获取插件能力
func (p *PlaylistPluginImpl) GetCapabilities() []string {
	return []string{
		"playlist_management",
		"queue_management",
		"history_management",
		"play_mode_management",
		"event_integration",
	}
}

// GetDependencies 获取插件依赖
func (p *PlaylistPluginImpl) GetDependencies() []string {
	return []string{"event_bus"}
}

// Initialize 初始化插件
func (p *PlaylistPluginImpl) Initialize(ctx plugin.PluginContext) error {
	// 获取事件总线
	if eventBus, err := ctx.GetServiceRegistry().GetService("eventBus"); err == nil {
		if eb, ok := eventBus.(event.EventBus); ok {
			p.eventBus = eb
		} else {
			return fmt.Errorf("event bus service has wrong type: %T", eventBus)
		}
	} else {
		return fmt.Errorf("failed to get event bus service: %v", err)
	}

	// 订阅播放器事件
	if err := p.subscribeToPlayerEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to player events: %w", err)
	}

	return nil
}

// Start 启动插件
func (p *PlaylistPluginImpl) Start() error {
	// 发送插件启动事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("playlist_started_%d", time.Now().UnixNano()),
			Type:      event.EventPluginStarted,
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"plugin_name": p.info.Name,
				"plugin_id":   p.info.ID,
			},
		}
		p.eventBus.Publish(context.Background(), event)
	}
	return nil
}

// Stop 停止插件
func (p *PlaylistPluginImpl) Stop() error {
	// 发送插件停止事件
	if p.eventBus != nil {
		event := &event.BaseEvent{
			ID:        fmt.Sprintf("playlist_stopped_%d", time.Now().UnixNano()),
			Type:      event.EventPluginStopped,
			Source:    p.info.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"plugin_name": p.info.Name,
				"plugin_id":   p.info.ID,
			},
		}
		p.eventBus.Publish(context.Background(), event)
	}
	return nil
}

// Cleanup 清理插件资源
func (p *PlaylistPluginImpl) Cleanup() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 清理所有数据
	p.playlists = make(map[string]*model.Playlist)
	p.currentQueue = make([]*model.Song, 0)
	p.history = make([]*model.Song, 0)
	p.shuffleIndex = nil
	p.currentIndex = -1

	return nil
}

// HealthCheck 健康检查
func (p *PlaylistPluginImpl) HealthCheck() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 检查基本状态
	if p.playlists == nil {
		return fmt.Errorf("playlists map is nil")
	}
	if p.currentQueue == nil {
		return fmt.Errorf("current queue is nil")
	}
	if p.history == nil {
		return fmt.Errorf("history is nil")
	}

	return nil
}

// ValidateConfig 验证配置
func (p *PlaylistPluginImpl) ValidateConfig(config map[string]interface{}) error {
	if maxHistory, ok := config["max_history"]; ok {
		if val, ok := maxHistory.(int); ok && val <= 0 {
			return fmt.Errorf("max_history must be positive")
		}
	}
	return nil
}

// UpdateConfig 更新配置
func (p *PlaylistPluginImpl) UpdateConfig(config map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if maxHistory, ok := config["max_history"]; ok {
		if val, ok := maxHistory.(int); ok {
			p.maxHistory = val
			// 如果历史记录超过新的限制，截断它
			if len(p.history) > p.maxHistory {
				p.history = p.history[len(p.history)-p.maxHistory:]
			}
		}
	}

	return nil
}

// GetMetrics 获取插件指标
func (p *PlaylistPluginImpl) GetMetrics() (*plugin.PluginMetrics, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return &plugin.PluginMetrics{
		StartTime:     p.info.CreatedAt,
		Uptime:        time.Since(p.info.CreatedAt),
		RequestCount:  0,
		ErrorCount:    0,
		MemoryUsage:   0,
		CPUUsage:      0.0,
		CustomMetrics: map[string]interface{}{
			"playlists_count": len(p.playlists),
			"queue_length":    len(p.currentQueue),
			"history_length":  len(p.history),
			"play_mode":       p.playMode.String(),
			"current_index":   p.currentIndex,
		},
	}, nil
}

// HandleEvent 处理事件
func (p *PlaylistPluginImpl) HandleEvent(evt interface{}) error {
	// 处理播放器事件
	if e, ok := evt.(event.Event); ok {
		return p.handlePlayerEvent(e)
	}
	return nil
}

// subscribeToPlayerEvents 订阅播放器事件
func (p *PlaylistPluginImpl) subscribeToPlayerEvents() error {
	if p.eventBus == nil {
		return fmt.Errorf("event bus is not available")
	}

	// 订阅歌曲变化事件
	_, err := p.eventBus.Subscribe(event.EventPlayerSongChanged, func(ctx context.Context, e event.Event) error {
		return p.HandleEvent(e)
	})

	return err
}

// handlePlayerEvent 处理播放器事件
func (p *PlaylistPluginImpl) handlePlayerEvent(e event.Event) error {
	switch e.GetType() {
	case event.EventPlayerSongChanged:
		// 当歌曲变化时，添加到历史记录
		if songData, ok := e.GetData().(map[string]interface{}); ok {
			if songID, ok := songData["song_id"].(string); ok {
				// 从队列中找到歌曲并添加到历史
				p.mu.RLock()
				for _, song := range p.currentQueue {
					if song.ID == songID {
						p.mu.RUnlock()
						p.AddToHistory(context.Background(), song)
						return nil
					}
				}
				p.mu.RUnlock()
			}
		}
	}
	return nil
}