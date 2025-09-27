package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// 类型别名
type Track = core.Track
type Album = core.Album
type Artist = core.Artist
type Playlist = core.Playlist
type SearchResult = core.SearchResult
type AudioQuality = core.AudioQuality
type AudioFormat = core.AudioFormat
type SearchType = core.SearchType
type ServiceInfo = core.ServiceInfo
type Repository = core.Repository
type PluginInfo = core.PluginInfo
type PluginType = core.PluginType
type ServiceStatus = core.ServiceStatus
type HealthStatus = core.HealthStatus

// 辅助函数
func HealthStatusHealthy() HealthStatus {
	return HealthStatus{
		Healthy:   true,
		Message:   "Healthy",
		Timestamp: time.Now(),
	}
}

func HealthStatusUnknown() HealthStatus {
	return HealthStatus{
		Healthy:   false,
		Message:   "Unknown",
		Timestamp: time.Now(),
	}
}

// 函数别名
var NewBasePlugin = core.NewBasePlugin

// 常量别名
const (
	SearchTypeAll = core.SearchTypeAll
	SearchTypeTrack = core.SearchTypeTrack
	SearchTypeAlbum = core.SearchTypeAlbum
	SearchTypeArtist = core.SearchTypeArtist
	SearchTypePlaylist = core.SearchTypePlaylist
	PluginTypeMusicSource = core.PluginTypeMusicSource
	ServiceStatusRunning = core.ServiceStatusRunning
	ServiceStatusStopped = core.ServiceStatusStopped
	AudioQualityLow = core.AudioQualityLow
	AudioQualityStandard = core.AudioQualityStandard
	AudioQualityHigh = core.AudioQualityHigh
	AudioQualityLossless = core.AudioQualityLossless
	AudioQualityHiRes = core.AudioQualityHiRes
	AudioFormatUnknown = core.AudioFormatUnknown
	AudioFormatMP3 = core.AudioFormatMP3
	AudioFormatFLAC = core.AudioFormatFLAC
	AudioFormatWAV = core.AudioFormatWAV
	AudioFormatAAC = core.AudioFormatAAC
	AudioFormatOGG = core.AudioFormatOGG
	AudioFormatM4A = core.AudioFormatM4A
	AudioFormatWMA = core.AudioFormatWMA
)

// MusicSourcePlugin 音乐源插件接口（RPC实现）
type MusicSourcePlugin interface {
	core.Plugin

	// 搜索功能
	Search(ctx context.Context, query string, options SearchOptions) (*SearchResult, error)

	// 播放列表
	GetPlaylist(ctx context.Context, id string) (*Playlist, error)
	GetPlaylistTracks(ctx context.Context, playlistID string, offset, limit int) ([]*Track, error)
	CreatePlaylist(ctx context.Context, name, description string) (*Playlist, error)
	UpdatePlaylist(ctx context.Context, playlistID string, updates map[string]interface{}) error
	DeletePlaylist(ctx context.Context, playlistID string) error

	// 音轨信息
	GetTrackURL(ctx context.Context, trackID string, quality AudioQuality) (string, error)
	GetTrackLyrics(ctx context.Context, trackID string) (*Lyrics, error)
	GetTrackDetail(ctx context.Context, trackID string) (*Track, error)
	GetTrackComments(ctx context.Context, trackID string, offset, limit int) ([]*Comment, error)

	// 用户相关
	Login(ctx context.Context, credentials map[string]string) error
	Logout(ctx context.Context) error
	FollowUser(ctx context.Context, userID string) error
	UnfollowUser(ctx context.Context, userID string) error
	GetUserInfo(ctx context.Context, userID string) (*UserInfo, error)
	GetUserPlaylists(ctx context.Context, userID string) ([]*Playlist, error)
	GetUserLikedTracks(ctx context.Context, userID string) ([]*Track, error)
	GetUserFollowing(ctx context.Context, userID string) ([]string, error)
	GetUserFollowers(ctx context.Context, userID string) ([]string, error)

	// 推荐功能
	GetRecommendations(ctx context.Context, options RecommendationOptions) ([]*Track, error)
	GetDailyRecommendations(ctx context.Context) ([]*Track, error)
	GetSimilarTracks(ctx context.Context, trackID string, limit int) ([]*Track, error)

	// 排行榜
	GetCharts(ctx context.Context) ([]*Chart, error)
	GetChartTracks(ctx context.Context, chartID string) ([]*Track, error)

	// 专辑和艺术家
	GetAlbum(ctx context.Context, albumID string) (*Album, error)
	GetAlbumTracks(ctx context.Context, albumID string) ([]*Track, error)
	GetArtist(ctx context.Context, artistID string) (*Artist, error)
	GetArtistTracks(ctx context.Context, artistID string, offset, limit int) ([]*Track, error)
	GetArtistAlbums(ctx context.Context, artistID string, offset, limit int) ([]*Album, error)

	// 电台功能
	GetRadioStations(ctx context.Context) ([]*RadioStation, error)
	GetRadioTracks(ctx context.Context, stationID string) ([]*Track, error)

	// 获取支持的功能
	GetSupportedFeatures() []MusicSourceFeature

	// 获取服务信息
	GetServiceInfo() *ServiceInfo
}

// SearchOptions 搜索选项
type SearchOptions struct {
	Query    string     `json:"query"`     // 搜索查询
	Type     SearchType `json:"type"`      // 搜索类型
	Limit    int        `json:"limit"`     // 限制数量
	Offset   int        `json:"offset"`    // 偏移量
	Filters  map[string]interface{} `json:"filters"`  // 过滤条件
	SortBy   string     `json:"sort_by"`   // 排序字段
	SortDesc bool       `json:"sort_desc"` // 是否降序
}

// UserInfo 用户信息
type UserInfo struct {
	ID          string            `json:"id"`           // 用户ID
	Username    string            `json:"username"`     // 用户名
	Nickname    string            `json:"nickname"`     // 昵称
	AvatarURL   string            `json:"avatar_url"`   // 头像URL
	Description string            `json:"description"`  // 描述
	Level       int               `json:"level"`        // 等级
	FollowCount int64             `json:"follow_count"` // 关注数
	FanCount    int64             `json:"fan_count"`    // 粉丝数
	PlayCount   int64             `json:"play_count"`   // 播放次数
	IsFollowed  bool              `json:"is_followed"`  // 是否已关注
	Tags        map[string]string `json:"tags"`         // 标签
	Metadata    map[string]interface{} `json:"metadata"` // 元数据
	Source      string            `json:"source"`       // 来源
	SourceID    string            `json:"source_id"`    // 来源ID
	CreatedAt   time.Time         `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time         `json:"updated_at"`   // 更新时间
}









// Lyrics 歌词信息
type Lyrics struct {
	TrackID     string       `json:"track_id"`    // 音轨ID
	Content     string       `json:"content"`     // 歌词内容
	TimedLyrics []TimedLyric `json:"timed_lyrics"` // 时间轴歌词
	Translation string       `json:"translation"`  // 翻译
	Language    string       `json:"language"`     // 语言
	Source      string       `json:"source"`       // 来源
	CreatedAt   time.Time    `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time    `json:"updated_at"`   // 更新时间
}

// TimedLyric 时间轴歌词
type TimedLyric struct {
	Time    time.Duration `json:"time"`    // 时间点
	Content string        `json:"content"` // 歌词内容
}

// Comment 评论信息
type Comment struct {
	ID        string    `json:"id"`         // 评论ID
	UserID    string    `json:"user_id"`    // 用户ID
	Username  string    `json:"username"`   // 用户名
	AvatarURL string    `json:"avatar_url"` // 头像地址
	Content   string    `json:"content"`    // 评论内容
	LikeCount int64     `json:"like_count"` // 点赞数
	IsLiked   bool      `json:"is_liked"`   // 是否已点赞
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// RecommendationType 推荐类型
type RecommendationType int

const (
	RecommendationTypePersonal RecommendationType = iota // 个人推荐
	RecommendationTypeDaily                              // 每日推荐
	RecommendationTypeSimilar                            // 相似推荐
	RecommendationTypeGenre                              // 流派推荐
	RecommendationTypeMood                               // 心情推荐
)

// String 返回推荐类型的字符串表示
func (r RecommendationType) String() string {
	switch r {
	case RecommendationTypePersonal:
		return "personal"
	case RecommendationTypeDaily:
		return "daily"
	case RecommendationTypeSimilar:
		return "similar"
	case RecommendationTypeGenre:
		return "genre"
	case RecommendationTypeMood:
		return "mood"
	default:
		return "unknown"
	}
}

// RecommendationOptions 推荐选项
type RecommendationOptions struct {
	UserID   string            `json:"user_id"`   // 用户ID
	Genres   []string          `json:"genres"`    // 流派过滤
	Mood     string            `json:"mood"`      // 心情
	Activity string            `json:"activity"`  // 活动
	Limit    int               `json:"limit"`     // 限制数量
	Filters  map[string]interface{} `json:"filters"` // 过滤条件
}

// Chart 排行榜信息
type Chart struct {
	ID          string    `json:"id"`          // 排行榜ID
	Name        string    `json:"name"`        // 名称
	Description string    `json:"description"` // 描述
	CoverURL    string    `json:"cover_url"`   // 封面地址
	UpdateTime  time.Time `json:"update_time"` // 更新时间
	Period      string    `json:"period"`      // 周期（日榜、周榜、月榜等）
}

// RadioStation 电台信息
type RadioStation struct {
	ID          string    `json:"id"`          // 电台ID
	Name        string    `json:"name"`        // 名称
	Description string    `json:"description"` // 描述
	CoverURL    string    `json:"cover_url"`   // 封面地址
	Genre       string    `json:"genre"`       // 流派
	ListenerCount int64   `json:"listener_count"` // 听众数
	IsLive      bool      `json:"is_live"`     // 是否直播
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
}

// MusicSourceFeature 音乐源功能枚举
type MusicSourceFeature int

const (
	MusicSourceFeatureSearch MusicSourceFeature = iota
	MusicSourceFeaturePlaylist
	MusicSourceFeatureUser
	MusicSourceFeatureRecommendation
	MusicSourceFeatureChart
	MusicSourceFeatureRadio
	MusicSourceFeatureLyrics
	MusicSourceFeatureComment
	MusicSourceFeatureDownload
	MusicSourceFeatureUpload
)

// String 返回功能的字符串表示
func (f MusicSourceFeature) String() string {
	switch f {
	case MusicSourceFeatureSearch:
		return "search"
	case MusicSourceFeaturePlaylist:
		return "playlist"
	case MusicSourceFeatureUser:
		return "user"
	case MusicSourceFeatureRecommendation:
		return "recommendation"
	case MusicSourceFeatureChart:
		return "chart"
	case MusicSourceFeatureRadio:
		return "radio"
	case MusicSourceFeatureLyrics:
		return "lyrics"
	case MusicSourceFeatureComment:
		return "comment"
	case MusicSourceFeatureDownload:
		return "download"
	case MusicSourceFeatureUpload:
		return "upload"
	default:
		return "unknown"
	}
}

// BaseMusicSourcePlugin 音乐源插件基础实现
type BaseMusicSourcePlugin struct {
	core.BasePlugin
	features    []MusicSourceFeature
	serviceInfo *core.ServiceInfo
	cache       core.Cache
	repository  core.Repository
	mu          sync.RWMutex
}

// NewBaseMusicSourcePlugin 创建基础音乐源插件
func NewBaseMusicSourcePlugin(info *PluginInfo) *BaseMusicSourcePlugin {
	return &BaseMusicSourcePlugin{
		BasePlugin: *NewBasePlugin(info),
		features:   make([]MusicSourceFeature, 0),
		serviceInfo: &ServiceInfo{
			ID:          info.ID,
			Name:        info.Name,
			Version:     info.Version,
			Description: info.Description,
			Status:      ServiceStatusStopped,
			Health:      HealthStatusUnknown(),
			Metadata:    make(map[string]interface{}),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
}

// GetSupportedFeatures 获取支持的功能
func (p *BaseMusicSourcePlugin) GetSupportedFeatures() []MusicSourceFeature {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]MusicSourceFeature(nil), p.features...)
}

// GetServiceInfo 获取服务信息
func (p *BaseMusicSourcePlugin) GetServiceInfo() *ServiceInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.serviceInfo
}

// SetCache 设置缓存
func (p *BaseMusicSourcePlugin) SetCache(cache Cache) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cache = cache
}

// SetRepository 设置数据仓库
func (p *BaseMusicSourcePlugin) SetRepository(repo Repository) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.repository = repo
}

// AddFeature 添加支持的功能
func (p *BaseMusicSourcePlugin) AddFeature(feature MusicSourceFeature) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, f := range p.features {
		if f == feature {
			return
		}
	}
	p.features = append(p.features, feature)
}

// HasFeature 检查是否支持某个功能
func (p *BaseMusicSourcePlugin) HasFeature(feature MusicSourceFeature) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, f := range p.features {
		if f == feature {
			return true
		}
	}
	return false
}

// UpdateServiceStatus 更新服务状态
func (p *BaseMusicSourcePlugin) UpdateServiceStatus(status ServiceStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.serviceInfo.Status = status
	p.serviceInfo.UpdatedAt = time.Now()
}

// UpdateHealthStatus 更新健康状态
func (p *BaseMusicSourcePlugin) UpdateHealthStatus(health HealthStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.serviceInfo.Health = health
	p.serviceInfo.UpdatedAt = time.Now()
}

// 实现MusicSourcePlugin接口的默认方法

// Search 默认搜索实现
func (p *BaseMusicSourcePlugin) Search(ctx context.Context, query string, options SearchOptions) (*SearchResult, error) {
	return nil, fmt.Errorf("search not implemented")
}

// GetPlaylist 默认获取播放列表实现
func (p *BaseMusicSourcePlugin) GetPlaylist(ctx context.Context, id string) (*Playlist, error) {
	return nil, fmt.Errorf("get playlist not implemented")
}

// GetPlaylistTracks 默认获取播放列表音轨实现
func (p *BaseMusicSourcePlugin) GetPlaylistTracks(ctx context.Context, playlistID string, offset, limit int) ([]*Track, error) {
	return nil, fmt.Errorf("get playlist tracks not implemented")
}

// CreatePlaylist 默认创建播放列表实现
func (p *BaseMusicSourcePlugin) CreatePlaylist(ctx context.Context, name, description string) (*Playlist, error) {
	return nil, fmt.Errorf("create playlist not implemented")
}

// UpdatePlaylist 默认更新播放列表实现
func (p *BaseMusicSourcePlugin) UpdatePlaylist(ctx context.Context, playlistID string, updates map[string]interface{}) error {
	return fmt.Errorf("update playlist not implemented")
}

// DeletePlaylist 默认删除播放列表实现
func (p *BaseMusicSourcePlugin) DeletePlaylist(ctx context.Context, playlistID string) error {
	return fmt.Errorf("delete playlist not implemented")
}

// GetTrackURL 默认获取音轨URL实现
func (p *BaseMusicSourcePlugin) GetTrackURL(ctx context.Context, trackID string, quality AudioQuality) (string, error) {
	return "", fmt.Errorf("get track url not implemented")
}

// GetTrackLyrics 默认获取歌词实现
func (p *BaseMusicSourcePlugin) GetTrackLyrics(ctx context.Context, trackID string) (*Lyrics, error) {
	return nil, fmt.Errorf("get track lyrics not implemented")
}

// GetTrackDetail 默认获取音轨详情实现
func (p *BaseMusicSourcePlugin) GetTrackDetail(ctx context.Context, trackID string) (*Track, error) {
	return nil, fmt.Errorf("get track detail not implemented")
}

// GetTrackComments 默认获取音轨评论实现
func (p *BaseMusicSourcePlugin) GetTrackComments(ctx context.Context, trackID string, offset, limit int) ([]*Comment, error) {
	return nil, fmt.Errorf("get track comments not implemented")
}

// Login 默认登录实现
func (p *BaseMusicSourcePlugin) Login(ctx context.Context, credentials map[string]string) error {
	return fmt.Errorf("login not implemented")
}

// Logout 默认登出实现
func (p *BaseMusicSourcePlugin) Logout(ctx context.Context) error {
	return fmt.Errorf("logout not implemented")
}

// FollowUser 默认关注用户实现
func (p *BaseMusicSourcePlugin) FollowUser(ctx context.Context, userID string) error {
	return fmt.Errorf("follow user not implemented")
}

// UnfollowUser 默认取消关注用户实现
func (p *BaseMusicSourcePlugin) UnfollowUser(ctx context.Context, userID string) error {
	return fmt.Errorf("unfollow user not implemented")
}

// GetUserInfo 默认获取用户信息实现
func (p *BaseMusicSourcePlugin) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	return nil, fmt.Errorf("get user info not implemented")
}

// GetUserPlaylists 默认获取用户播放列表实现
func (p *BaseMusicSourcePlugin) GetUserPlaylists(ctx context.Context, userID string) ([]*Playlist, error) {
	return nil, fmt.Errorf("get user playlists not implemented")
}

// GetUserLikedTracks 默认获取用户喜欢音轨实现
func (p *BaseMusicSourcePlugin) GetUserLikedTracks(ctx context.Context, userID string) ([]*Track, error) {
	return nil, fmt.Errorf("get user liked tracks not implemented")
}

// GetUserFollowing 默认获取用户关注列表实现
func (p *BaseMusicSourcePlugin) GetUserFollowing(ctx context.Context, userID string) ([]string, error) {
	return nil, fmt.Errorf("get user following not implemented")
}

// GetUserFollowers 默认获取用户粉丝列表实现
func (p *BaseMusicSourcePlugin) GetUserFollowers(ctx context.Context, userID string) ([]string, error) {
	return nil, fmt.Errorf("get user followers not implemented")
}

// GetRecommendations 默认获取推荐实现
func (p *BaseMusicSourcePlugin) GetRecommendations(ctx context.Context, options RecommendationOptions) ([]*Track, error) {
	return nil, fmt.Errorf("get recommendations not implemented")
}

// GetDailyRecommendations 默认获取每日推荐实现
func (p *BaseMusicSourcePlugin) GetDailyRecommendations(ctx context.Context) ([]*Track, error) {
	return nil, fmt.Errorf("get daily recommendations not implemented")
}

// GetSimilarTracks 默认获取相似音轨实现
func (p *BaseMusicSourcePlugin) GetSimilarTracks(ctx context.Context, trackID string, limit int) ([]*Track, error) {
	return nil, fmt.Errorf("get similar tracks not implemented")
}

// GetCharts 默认获取排行榜实现
func (p *BaseMusicSourcePlugin) GetCharts(ctx context.Context) ([]*Chart, error) {
	return nil, fmt.Errorf("get charts not implemented")
}

// GetChartTracks 默认获取排行榜音轨实现
func (p *BaseMusicSourcePlugin) GetChartTracks(ctx context.Context, chartID string) ([]*Track, error) {
	return nil, fmt.Errorf("get chart tracks not implemented")
}

// GetAlbum 默认获取专辑实现
func (p *BaseMusicSourcePlugin) GetAlbum(ctx context.Context, albumID string) (*Album, error) {
	return nil, fmt.Errorf("get album not implemented")
}

// GetAlbumTracks 默认获取专辑音轨实现
func (p *BaseMusicSourcePlugin) GetAlbumTracks(ctx context.Context, albumID string) ([]*Track, error) {
	return nil, fmt.Errorf("get album tracks not implemented")
}

// GetArtist 默认获取艺术家实现
func (p *BaseMusicSourcePlugin) GetArtist(ctx context.Context, artistID string) (*Artist, error) {
	return nil, fmt.Errorf("get artist not implemented")
}

// GetArtistTracks 默认获取艺术家音轨实现
func (p *BaseMusicSourcePlugin) GetArtistTracks(ctx context.Context, artistID string, offset, limit int) ([]*Track, error) {
	return nil, fmt.Errorf("get artist tracks not implemented")
}

// GetArtistAlbums 默认获取艺术家专辑实现
func (p *BaseMusicSourcePlugin) GetArtistAlbums(ctx context.Context, artistID string, offset, limit int) ([]*Album, error) {
	return nil, fmt.Errorf("get artist albums not implemented")
}

// GetRadioStations 默认获取电台列表实现
func (p *BaseMusicSourcePlugin) GetRadioStations(ctx context.Context) ([]*RadioStation, error) {
	return nil, fmt.Errorf("get radio stations not implemented")
}

// GetRadioTracks 默认获取电台音轨实现
func (p *BaseMusicSourcePlugin) GetRadioTracks(ctx context.Context, stationID string) ([]*Track, error) {
	return nil, fmt.Errorf("get radio tracks not implemented")
}