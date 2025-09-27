package plugin

import (
	"context"
	"time"
)

// Track 音轨信息
type Track struct {
	ID          string            `json:"id"`           // 音轨ID
	Title       string            `json:"title"`        // 标题
	Artist      string            `json:"artist"`       // 艺术家
	Album       string            `json:"album"`        // 专辑
	Genre       string            `json:"genre"`        // 流派
	Year        int               `json:"year"`         // 年份
	Duration    time.Duration     `json:"duration"`     // 时长
	Bitrate     int               `json:"bitrate"`      // 比特率
	SampleRate  int               `json:"sample_rate"`  // 采样率
	Channels    int               `json:"channels"`     // 声道数
	Format      AudioFormat       `json:"format"`       // 音频格式
	Quality     AudioQuality      `json:"quality"`      // 音质
	URL         string            `json:"url"`          // 播放URL
	CoverURL    string            `json:"cover_url"`    // 封面URL
	Lyrics      string            `json:"lyrics"`       // 歌词
	Tags        map[string]string `json:"tags"`         // 标签
	Metadata    map[string]interface{} `json:"metadata"` // 元数据
	Source      string            `json:"source"`       // 来源
	SourceID    string            `json:"source_id"`    // 来源ID
	CreatedAt   time.Time         `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time         `json:"updated_at"`   // 更新时间
}

// Album 专辑信息
type Album struct {
	ID          string            `json:"id"`           // 专辑ID
	Title       string            `json:"title"`        // 标题
	Artist      string            `json:"artist"`       // 艺术家
	Genre       string            `json:"genre"`        // 流派
	Year        int               `json:"year"`         // 年份
	TrackCount  int               `json:"track_count"`  // 音轨数量
	Duration    time.Duration     `json:"duration"`     // 总时长
	CoverURL    string            `json:"cover_url"`    // 封面URL
	Description string            `json:"description"`  // 描述
	Tracks      []Track           `json:"tracks"`       // 音轨列表
	Tags        map[string]string `json:"tags"`         // 标签
	Metadata    map[string]interface{} `json:"metadata"` // 元数据
	Source      string            `json:"source"`       // 来源
	SourceID    string            `json:"source_id"`    // 来源ID
	CreatedAt   time.Time         `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time         `json:"updated_at"`   // 更新时间
}

// Artist 艺术家信息
type Artist struct {
	ID          string            `json:"id"`           // 艺术家ID
	Name        string            `json:"name"`         // 名称
	Genre       string            `json:"genre"`        // 流派
	Country     string            `json:"country"`      // 国家
	Bio         string            `json:"bio"`          // 简介
	AvatarURL   string            `json:"avatar_url"`   // 头像URL
	Followers   int64             `json:"followers"`    // 粉丝数
	Albums      []Album           `json:"albums"`       // 专辑列表
	Tracks      []Track           `json:"tracks"`       // 音轨列表
	Tags        map[string]string `json:"tags"`         // 标签
	Metadata    map[string]interface{} `json:"metadata"` // 元数据
	Source      string            `json:"source"`       // 来源
	SourceID    string            `json:"source_id"`    // 来源ID
	CreatedAt   time.Time         `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time         `json:"updated_at"`   // 更新时间
}

// Playlist 播放列表信息
type Playlist struct {
	ID          string            `json:"id"`           // 播放列表ID
	Name        string            `json:"name"`         // 名称
	Description string            `json:"description"`  // 描述
	Owner       string            `json:"owner"`        // 所有者
	Public      bool              `json:"public"`       // 是否公开
	TrackCount  int               `json:"track_count"`  // 音轨数量
	Duration    time.Duration     `json:"duration"`     // 总时长
	CoverURL    string            `json:"cover_url"`    // 封面URL
	Tracks      []Track           `json:"tracks"`       // 音轨列表
	Tags        map[string]string `json:"tags"`         // 标签
	Metadata    map[string]interface{} `json:"metadata"` // 元数据
	Source      string            `json:"source"`       // 来源
	SourceID    string            `json:"source_id"`    // 来源ID
	CreatedAt   time.Time         `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time         `json:"updated_at"`   // 更新时间
}

// SearchResult 搜索结果
type SearchResult struct {
	Query     string     `json:"query"`      // 搜索查询
	Type      SearchType `json:"type"`       // 搜索类型
	Total     int        `json:"total"`      // 总数
	Offset    int        `json:"offset"`     // 偏移量
	Limit     int        `json:"limit"`      // 限制数量
	Tracks    []Track    `json:"tracks"`     // 音轨结果
	Albums    []Album    `json:"albums"`     // 专辑结果
	Artists   []Artist   `json:"artists"`    // 艺术家结果
	Playlists []Playlist `json:"playlists"`  // 播放列表结果
	Source    string     `json:"source"`     // 来源
	Timestamp time.Time  `json:"timestamp"`  // 时间戳
}

// SearchType 搜索类型枚举
type SearchType int

const (
	SearchTypeAll SearchType = iota
	SearchTypeTrack
	SearchTypeAlbum
	SearchTypeArtist
	SearchTypePlaylist
)

// String 返回搜索类型的字符串表示
func (s SearchType) String() string {
	switch s {
	case SearchTypeAll:
		return "all"
	case SearchTypeTrack:
		return "track"
	case SearchTypeAlbum:
		return "album"
	case SearchTypeArtist:
		return "artist"
	case SearchTypePlaylist:
		return "playlist"
	default:
		return "unknown"
	}
}

// AudioFormat 音频格式枚举
type AudioFormat int

const (
	AudioFormatUnknown AudioFormat = iota
	AudioFormatMP3
	AudioFormatFLAC
	AudioFormatWAV
	AudioFormatAAC
	AudioFormatOGG
	AudioFormatM4A
	AudioFormatWMA
)

// String 返回音频格式的字符串表示
func (f AudioFormat) String() string {
	switch f {
	case AudioFormatMP3:
		return "mp3"
	case AudioFormatFLAC:
		return "flac"
	case AudioFormatWAV:
		return "wav"
	case AudioFormatAAC:
		return "aac"
	case AudioFormatOGG:
		return "ogg"
	case AudioFormatM4A:
		return "m4a"
	case AudioFormatWMA:
		return "wma"
	default:
		return "unknown"
	}
}

// AudioBuffer 音频缓冲区
type AudioBuffer struct {
	Data       []byte `json:"data"`        // 音频数据
	SampleRate int    `json:"sample_rate"` // 采样率
	Channels   int    `json:"channels"`    // 声道数
	Length     int    `json:"length"`      // 数据长度
	Format     AudioFormat `json:"format"` // 音频格式
}

// AudioQuality 音质枚举
type AudioQuality int

const (
	AudioQualityLow AudioQuality = iota
	AudioQualityStandard
	AudioQualityHigh
	AudioQualityLossless
	AudioQualityHiRes
)

// String 返回音质的字符串表示
func (q AudioQuality) String() string {
	switch q {
	case AudioQualityLow:
		return "low"
	case AudioQualityStandard:
		return "standard"
	case AudioQualityHigh:
		return "high"
	case AudioQualityLossless:
		return "lossless"
	case AudioQualityHiRes:
		return "hires"
	default:
		return "unknown"
	}
}

// GetBitrate 获取音质对应的比特率
func (q AudioQuality) GetBitrate() int {
	switch q {
	case AudioQualityLow:
		return 128
	case AudioQualityStandard:
		return 192
	case AudioQualityHigh:
		return 320
	case AudioQualityLossless:
		return 1411
	case AudioQualityHiRes:
		return 2304
	default:
		return 192
	}
}

// PlaybackState 播放状态
type PlaybackState struct {
	Track       *Track        `json:"track"`        // 当前音轨
	State       PlayState     `json:"state"`        // 播放状态
	Position    time.Duration `json:"position"`     // 播放位置
	Duration    time.Duration `json:"duration"`     // 总时长
	Volume      float64       `json:"volume"`       // 音量 (0.0-1.0)
	Muted       bool          `json:"muted"`        // 是否静音
	Shuffle     bool          `json:"shuffle"`      // 是否随机播放
	Repeat      RepeatMode    `json:"repeat"`       // 重复模式
	Playlist    *Playlist     `json:"playlist"`     // 当前播放列表
	Index       int           `json:"index"`        // 当前索引
	Buffering   bool          `json:"buffering"`    // 是否缓冲中
	Error       string        `json:"error"`        // 错误信息
	Timestamp   time.Time     `json:"timestamp"`    // 时间戳
}

// HealthStatus 健康状态
type HealthStatus struct {
	Healthy   bool              `json:"healthy"`   // 是否健康
	Message   string            `json:"message"`   // 状态消息
	Details   map[string]interface{} `json:"details"` // 详细信息
	Timestamp time.Time         `json:"timestamp"` // 时间戳
}

// PlayState 播放状态枚举
type PlayState int

const (
	PlayStateStopped PlayState = iota
	PlayStatePlaying
	PlayStatePaused
	PlayStateBuffering
	PlayStateError
)

// String 返回播放状态的字符串表示
func (s PlayState) String() string {
	switch s {
	case PlayStateStopped:
		return "stopped"
	case PlayStatePlaying:
		return "playing"
	case PlayStatePaused:
		return "paused"
	case PlayStateBuffering:
		return "buffering"
	case PlayStateError:
		return "error"
	default:
		return "unknown"
	}
}

// RepeatMode 重复模式枚举
type RepeatMode int

const (
	RepeatModeOff RepeatMode = iota
	RepeatModeOne
	RepeatModeAll
)

// String 返回重复模式的字符串表示
func (r RepeatMode) String() string {
	switch r {
	case RepeatModeOff:
		return "off"
	case RepeatModeOne:
		return "one"
	case RepeatModeAll:
		return "all"
	default:
		return "unknown"
	}
}



// AudioStream 音频流
type AudioStream struct {
	ID         string                 `json:"id"`          // 流ID
	Format     AudioFormat            `json:"format"`      // 音频格式
	SampleRate int                    `json:"sample_rate"` // 采样率
	Channels   int                    `json:"channels"`    // 声道数
	BitDepth   int                    `json:"bit_depth"`   // 位深度
	Bitrate    int                    `json:"bitrate"`     // 比特率
	Buffers    chan *AudioBuffer      `json:"-"`           // 缓冲区通道
	Metadata   map[string]interface{} `json:"metadata"`    // 元数据
	CreatedAt  time.Time              `json:"created_at"`  // 创建时间
}



// PluginEvent 插件事件
type PluginEvent struct {
	ID        string                 `json:"id"`         // 事件ID
	PluginID  string                 `json:"plugin_id"`  // 插件ID
	Type      EventType              `json:"type"`       // 事件类型
	Name      string                 `json:"name"`       // 事件名称
	Data      map[string]interface{} `json:"data"`       // 事件数据
	Source    string                 `json:"source"`     // 事件源
	Target    string                 `json:"target"`     // 事件目标
	Priority  EventPriority          `json:"priority"`   // 事件优先级
	Async     bool                   `json:"async"`      // 是否异步
	Timestamp time.Time              `json:"timestamp"`  // 时间戳
	Processed bool                   `json:"processed"`  // 是否已处理
	Error     string                 `json:"error"`      // 错误信息
}

// EventType 事件类型枚举
type EventType int

const (
	EventTypeSystem EventType = iota
	EventTypePlugin
	EventTypeAudio
	EventTypePlayback
	EventTypeUI
	EventTypeNetwork
	EventTypeUser
	EventTypeCustom
)

// String 返回事件类型的字符串表示
func (e EventType) String() string {
	switch e {
	case EventTypeSystem:
		return "system"
	case EventTypePlugin:
		return "plugin"
	case EventTypeAudio:
		return "audio"
	case EventTypePlayback:
		return "playback"
	case EventTypeUI:
		return "ui"
	case EventTypeNetwork:
		return "network"
	case EventTypeUser:
		return "user"
	case EventTypeCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// EventPriority 事件优先级枚举
type EventPriority int

const (
	EventPriorityLow EventPriority = iota
	EventPriorityNormal
	EventPriorityHigh
	EventPriorityCritical
)

// String 返回事件优先级的字符串表示
func (p EventPriority) String() string {
	switch p {
	case EventPriorityLow:
		return "low"
	case EventPriorityNormal:
		return "normal"
	case EventPriorityHigh:
		return "high"
	case EventPriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID          string                 `json:"id"`           // 服务ID
	Name        string                 `json:"name"`         // 服务名称
	Version     string                 `json:"version"`      // 服务版本
	Description string                 `json:"description"`  // 服务描述
	Endpoint    string                 `json:"endpoint"`     // 服务端点
	Status      ServiceStatus          `json:"status"`       // 服务状态
	Health      HealthStatus           `json:"health"`       // 健康状态
	Metadata    map[string]interface{} `json:"metadata"`     // 元数据
	CreatedAt   time.Time              `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time              `json:"updated_at"`   // 更新时间
}

// ServiceStatus 服务状态枚举
type ServiceStatus int

const (
	ServiceStatusUnknown ServiceStatus = iota
	ServiceStatusStarting
	ServiceStatusRunning
	ServiceStatusStopping
	ServiceStatusStopped
	ServiceStatusError
)

// String 返回服务状态的字符串表示
func (s ServiceStatus) String() string {
	switch s {
	case ServiceStatusStarting:
		return "starting"
	case ServiceStatusRunning:
		return "running"
	case ServiceStatusStopping:
		return "stopping"
	case ServiceStatusStopped:
		return "stopped"
	case ServiceStatusError:
		return "error"
	default:
		return "unknown"
	}
}

// HealthStatus is defined in health.go to avoid duplication

// Repository 数据仓库接口
type Repository interface {
	// Track operations
	GetTrack(ctx context.Context, id string) (*Track, error)
	SaveTrack(ctx context.Context, track *Track) error
	DeleteTrack(ctx context.Context, id string) error
	SearchTracks(ctx context.Context, query string, limit, offset int) ([]Track, error)
	
	// Album operations
	GetAlbum(ctx context.Context, id string) (*Album, error)
	SaveAlbum(ctx context.Context, album *Album) error
	DeleteAlbum(ctx context.Context, id string) error
	SearchAlbums(ctx context.Context, query string, limit, offset int) ([]Album, error)
	
	// Artist operations
	GetArtist(ctx context.Context, id string) (*Artist, error)
	SaveArtist(ctx context.Context, artist *Artist) error
	DeleteArtist(ctx context.Context, id string) error
	SearchArtists(ctx context.Context, query string, limit, offset int) ([]Artist, error)
	
	// Playlist operations
	GetPlaylist(ctx context.Context, id string) (*Playlist, error)
	SavePlaylist(ctx context.Context, playlist *Playlist) error
	DeletePlaylist(ctx context.Context, id string) error
	SearchPlaylists(ctx context.Context, query string, limit, offset int) ([]Playlist, error)
}

// Cache 缓存接口
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
	Keys(ctx context.Context, pattern string) ([]string, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
}