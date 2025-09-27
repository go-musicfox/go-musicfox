package plugin

import (
	"context"
	"time"
)

// ThirdPartyPlugin 第三方插件接口
type ThirdPartyPlugin interface {
	Plugin

	// 服务连接
	Connect(ctx context.Context, config map[string]interface{}) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	GetConnectionStatus() *ConnectionStatus

	// 数据同步
	SyncData(ctx context.Context, options SyncOptions) (*SyncResult, error)
	GetSyncStatus() *SyncStatus
	CancelSync(ctx context.Context) error

	// 通知功能
	SendNotification(ctx context.Context, notification *Notification) error
	GetNotificationHistory(ctx context.Context, limit int) ([]*Notification, error)

	// 社交功能
	ShareContent(ctx context.Context, content *ShareContent) error
	GetSocialProfile(ctx context.Context) (*SocialProfile, error)
	UpdateSocialStatus(ctx context.Context, status string) error

	// 云存储
	UploadFile(ctx context.Context, file *FileUpload) (*FileInfo, error)
	DownloadFile(ctx context.Context, fileID string) (*FileDownload, error)
	DeleteFile(ctx context.Context, fileID string) error
	ListFiles(ctx context.Context, options ListOptions) ([]*FileInfo, error)

	// 统计分析
	TrackEvent(ctx context.Context, event *AnalyticsEvent) error
	GetAnalytics(ctx context.Context, options AnalyticsOptions) (*AnalyticsData, error)

	// 获取支持的服务类型
	GetSupportedServices() []ThirdPartyService

	// 获取服务配置
	GetServiceConfig() *ThirdPartyConfig
}

// ConnectionStatus 连接状态
type ConnectionStatus struct {
	IsConnected   bool      `json:"is_connected"`   // 是否已连接
	LastConnected time.Time `json:"last_connected"` // 最后连接时间
	LastError     string    `json:"last_error"`     // 最后错误信息
	RetryCount    int       `json:"retry_count"`    // 重试次数
	Latency       time.Duration `json:"latency"`    // 延迟
	Bandwidth     int64     `json:"bandwidth"`      // 带宽
}

// SyncOptions 同步选项
type SyncOptions struct {
	Type      SyncType  `json:"type"`       // 同步类型
	Direction SyncDirection `json:"direction"` // 同步方向
	Filter    string    `json:"filter"`     // 过滤条件
	BatchSize int       `json:"batch_size"` // 批次大小
	Timeout   time.Duration `json:"timeout"` // 超时时间
	Force     bool      `json:"force"`      // 强制同步
}

// SyncType 同步类型枚举
type SyncType int

const (
	SyncTypeAll SyncType = iota
	SyncTypePlaylists
	SyncTypeSongs
	SyncTypeUserData
	SyncTypeSettings
	SyncTypeHistory
	SyncTypeFavorites
)

// SyncDirection 同步方向枚举
type SyncDirection int

const (
	SyncDirectionBoth SyncDirection = iota
	SyncDirectionUpload
	SyncDirectionDownload
)

// SyncResult 同步结果
type SyncResult struct {
	ID            string        `json:"id"`             // 同步任务ID
	Status        SyncStatus    `json:"status"`         // 同步状态
	StartTime     time.Time     `json:"start_time"`     // 开始时间
	EndTime       time.Time     `json:"end_time"`       // 结束时间
	Duration      time.Duration `json:"duration"`       // 持续时间
	TotalItems    int           `json:"total_items"`    // 总项目数
	ProcessedItems int          `json:"processed_items"` // 已处理项目数
	SuccessItems  int           `json:"success_items"`  // 成功项目数
	FailedItems   int           `json:"failed_items"`   // 失败项目数
	Errors        []string      `json:"errors"`         // 错误列表
	Progress      float64       `json:"progress"`       // 进度百分比
}

// SyncStatus 同步状态枚举
type SyncStatus int

const (
	SyncStatusPending SyncStatus = iota
	SyncStatusRunning
	SyncStatusCompleted
	SyncStatusFailed
	SyncStatusCancelled
	SyncStatusPaused
)

// String 返回同步状态的字符串表示
func (s SyncStatus) String() string {
	switch s {
	case SyncStatusPending:
		return "pending"
	case SyncStatusRunning:
		return "running"
	case SyncStatusCompleted:
		return "completed"
	case SyncStatusFailed:
		return "failed"
	case SyncStatusCancelled:
		return "cancelled"
	case SyncStatusPaused:
		return "paused"
	default:
		return "unknown"
	}
}

// Notification 通知信息
type Notification struct {
	ID        string            `json:"id"`         // 通知ID
	Type      NotificationType  `json:"type"`       // 通知类型
	Title     string            `json:"title"`      // 标题
	Message   string            `json:"message"`    // 消息内容
	Icon      string            `json:"icon"`       // 图标
	URL       string            `json:"url"`        // 链接地址
	Priority  NotificationPriority `json:"priority"` // 优先级
	Data      map[string]interface{} `json:"data"`   // 附加数据
	SentAt    time.Time         `json:"sent_at"`    // 发送时间
	ReadAt    *time.Time        `json:"read_at"`    // 阅读时间
	ExpireAt  *time.Time        `json:"expire_at"`  // 过期时间
}

// NotificationType 通知类型枚举
type NotificationType int

const (
	NotificationTypeInfo NotificationType = iota
	NotificationTypeWarning
	NotificationTypeError
	NotificationTypeSuccess
	NotificationTypeMusic
	NotificationTypeSocial
	NotificationTypeSystem
)

// NotificationPriority 通知优先级枚举
type NotificationPriority int

const (
	NotificationPriorityLow NotificationPriority = iota
	NotificationPriorityNormal
	NotificationPriorityHigh
	NotificationPriorityUrgent
)

// ShareContent 分享内容
type ShareContent struct {
	Type        ShareType         `json:"type"`        // 分享类型
	Title       string            `json:"title"`       // 标题
	Description string            `json:"description"` // 描述
	URL         string            `json:"url"`         // 链接地址
	ImageURL    string            `json:"image_url"`   // 图片地址
	Tags        []string          `json:"tags"`        // 标签
	Data        map[string]interface{} `json:"data"`   // 附加数据
	Platforms   []SocialPlatform  `json:"platforms"`   // 目标平台
}

// ShareType 分享类型枚举
type ShareType int

const (
	ShareTypeSong ShareType = iota
	ShareTypePlaylist
	ShareTypeAlbum
	ShareTypeArtist
	ShareTypeText
	ShareTypeImage
	ShareTypeLink
)

// SocialPlatform 社交平台枚举
type SocialPlatform int

const (
	SocialPlatformWeChat SocialPlatform = iota
	SocialPlatformWeibo
	SocialPlatformQQ
	SocialPlatformTwitter
	SocialPlatformFacebook
	SocialPlatformInstagram
	SocialPlatformTikTok
	SocialPlatformDiscord
)

// SocialProfile 社交档案
type SocialProfile struct {
	Platform    SocialPlatform `json:"platform"`     // 平台
	UserID      string         `json:"user_id"`      // 用户ID
	Username    string         `json:"username"`     // 用户名
	DisplayName string         `json:"display_name"` // 显示名称
	AvatarURL   string         `json:"avatar_url"`   // 头像地址
	ProfileURL  string         `json:"profile_url"`  // 档案地址
	Followers   int64          `json:"followers"`    // 关注者数
	Following   int64          `json:"following"`    // 关注数
	Posts       int64          `json:"posts"`        // 帖子数
	IsVerified  bool           `json:"is_verified"`  // 是否认证
	Bio         string         `json:"bio"`          // 个人简介
	Location    string         `json:"location"`     // 位置
	Website     string         `json:"website"`      // 网站
	JoinedAt    time.Time      `json:"joined_at"`    // 加入时间
}

// FileUpload 文件上传信息
type FileUpload struct {
	Name        string            `json:"name"`         // 文件名
	Content     []byte            `json:"content"`      // 文件内容
	ContentType string            `json:"content_type"` // 内容类型
	Size        int64             `json:"size"`         // 文件大小
	Path        string            `json:"path"`         // 存储路径
	Metadata    map[string]string `json:"metadata"`     // 元数据
	Tags        []string          `json:"tags"`         // 标签
	IsPublic    bool              `json:"is_public"`    // 是否公开
	ExpireAt    *time.Time        `json:"expire_at"`    // 过期时间
}

// FileInfo 文件信息
type FileInfo struct {
	ID          string            `json:"id"`           // 文件ID
	Name        string            `json:"name"`         // 文件名
	Path        string            `json:"path"`         // 存储路径
	URL         string            `json:"url"`          // 访问地址
	ContentType string            `json:"content_type"` // 内容类型
	Size        int64             `json:"size"`         // 文件大小
	Checksum    string            `json:"checksum"`     // 校验和
	Metadata    map[string]string `json:"metadata"`     // 元数据
	Tags        []string          `json:"tags"`         // 标签
	IsPublic    bool              `json:"is_public"`    // 是否公开
	Downloads   int64             `json:"downloads"`    // 下载次数
	CreatedAt   time.Time         `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time         `json:"updated_at"`   // 更新时间
	ExpireAt    *time.Time        `json:"expire_at"`    // 过期时间
}

// FileDownload 文件下载信息
type FileDownload struct {
	FileInfo *FileInfo `json:"file_info"` // 文件信息
	Content  []byte    `json:"content"`   // 文件内容
	URL      string    `json:"url"`       // 下载地址
}

// ListOptions 列表选项
type ListOptions struct {
	Path      string   `json:"path"`       // 路径
	Filter    string   `json:"filter"`     // 过滤条件
	Tags      []string `json:"tags"`       // 标签过滤
	SortBy    string   `json:"sort_by"`    // 排序字段
	SortOrder string   `json:"sort_order"` // 排序方向
	Offset    int      `json:"offset"`     // 偏移量
	Limit     int      `json:"limit"`      // 限制数量
}

// AnalyticsEvent 分析事件
type AnalyticsEvent struct {
	Name       string                 `json:"name"`       // 事件名称
	Category   string                 `json:"category"`   // 事件分类
	Action     string                 `json:"action"`     // 动作
	Label      string                 `json:"label"`      // 标签
	Value      float64                `json:"value"`      // 数值
	Properties map[string]interface{} `json:"properties"` // 属性
	UserID     string                 `json:"user_id"`    // 用户ID
	SessionID  string                 `json:"session_id"` // 会话ID
	Timestamp  time.Time              `json:"timestamp"`  // 时间戳
}

// AnalyticsOptions 分析选项
type AnalyticsOptions struct {
	Metrics   []string  `json:"metrics"`    // 指标
	Dimensions []string `json:"dimensions"` // 维度
	Filters   []string  `json:"filters"`    // 过滤条件
	StartDate time.Time `json:"start_date"` // 开始日期
	EndDate   time.Time `json:"end_date"`   // 结束日期
	GroupBy   string    `json:"group_by"`   // 分组字段
	Limit     int       `json:"limit"`      // 限制数量
}

// AnalyticsData 分析数据
type AnalyticsData struct {
	Metrics    map[string]float64       `json:"metrics"`    // 指标数据
	Dimensions map[string]interface{}   `json:"dimensions"` // 维度数据
	TimeSeries []TimeSeriesPoint       `json:"time_series"` // 时间序列
	TotalRows  int                      `json:"total_rows"` // 总行数
	SampleRate float64                  `json:"sample_rate"` // 采样率
	GeneratedAt time.Time               `json:"generated_at"` // 生成时间
}

// TimeSeriesPoint 时间序列点
type TimeSeriesPoint struct {
	Timestamp time.Time            `json:"timestamp"` // 时间戳
	Values    map[string]float64   `json:"values"`    // 数值
}

// ThirdPartyService 第三方服务枚举
type ThirdPartyService int

const (
	ThirdPartyServiceCloudStorage ThirdPartyService = iota
	ThirdPartyServiceSocialMedia
	ThirdPartyServiceNotification
	ThirdPartyServiceAnalytics
	ThirdPartyServiceSync
	ThirdPartyServiceBackup
	ThirdPartyServiceCDN
	ThirdPartyServiceAuth
	ThirdPartyServicePayment
	ThirdPartyServiceAI
)

// String 返回第三方服务的字符串表示
func (s ThirdPartyService) String() string {
	switch s {
	case ThirdPartyServiceCloudStorage:
		return "cloud_storage"
	case ThirdPartyServiceSocialMedia:
		return "social_media"
	case ThirdPartyServiceNotification:
		return "notification"
	case ThirdPartyServiceAnalytics:
		return "analytics"
	case ThirdPartyServiceSync:
		return "sync"
	case ThirdPartyServiceBackup:
		return "backup"
	case ThirdPartyServiceCDN:
		return "cdn"
	case ThirdPartyServiceAuth:
		return "auth"
	case ThirdPartyServicePayment:
		return "payment"
	case ThirdPartyServiceAI:
		return "ai"
	default:
		return "unknown"
	}
}

// ThirdPartyConfig 第三方配置
type ThirdPartyConfig struct {
	Services    []ThirdPartyService   `json:"services"`     // 支持的服务
	Credentials map[string]string     `json:"credentials"`  // 凭证信息
	Endpoints   map[string]string     `json:"endpoints"`    // 端点配置
	Settings    map[string]interface{} `json:"settings"`    // 设置
	Limits      *ServiceLimits        `json:"limits"`       // 服务限制
	RetryPolicy *RetryPolicy          `json:"retry_policy"` // 重试策略
}

// ServiceLimits 服务限制
type ServiceLimits struct {
	RequestsPerSecond int           `json:"requests_per_second"` // 每秒请求数
	RequestsPerDay    int           `json:"requests_per_day"`    // 每日请求数
	MaxFileSize       int64         `json:"max_file_size"`       // 最大文件大小
	MaxStorageSize    int64         `json:"max_storage_size"`    // 最大存储大小
	Timeout           time.Duration `json:"timeout"`             // 超时时间
	Concurrency       int           `json:"concurrency"`         // 并发数
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`    // 最大重试次数
	InitialDelay  time.Duration `json:"initial_delay"`  // 初始延迟
	MaxDelay      time.Duration `json:"max_delay"`      // 最大延迟
	BackoffFactor float64       `json:"backoff_factor"` // 退避因子
	Jitter        bool          `json:"jitter"`         // 是否添加抖动
}