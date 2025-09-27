package plugin

import (
	"context"
	"time"
	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// 类型别名
type Plugin = core.Plugin
type BasePlugin = core.BasePlugin
type PluginMetrics = core.PluginMetrics
type PluginContext = core.PluginContext
type ServiceRegistry = core.ServiceRegistry
type EventBus = core.EventBus
type Logger = core.Logger
type PluginConfig = core.PluginConfig
type EventHandler = core.EventHandler
type ResourceMonitor = core.ResourceMonitor
type SecurityManager = core.SecurityManager
type IsolationGroup = core.IsolationGroup
type PluginPriority = core.PluginPriority
type ResourceLimits = core.ResourceLimits
type SecurityConfig = core.SecurityConfig
type PluginInfo = core.PluginInfo

// 常量别名
const (
	PluginPriorityLow      = core.PluginPriorityLow
	PluginPriorityNormal   = core.PluginPriorityNormal
	PluginPriorityHigh     = core.PluginPriorityHigh
	PluginPriorityCritical = core.PluginPriorityCritical
)

// NewBasePlugin 创建基础插件
func NewBasePlugin(info *PluginInfo) *BasePlugin {
	return core.NewBasePlugin(info)
}

// AudioProcessorPlugin 音频处理插件接口（动态链接库实现）
type AudioProcessorPlugin interface {
	Plugin

	// 音频处理
	ProcessAudio(input []byte, sampleRate int, channels int) ([]byte, error)

	// 音效处理
	ApplyEffect(input []byte, effect AudioEffect) ([]byte, error)

	// 音量控制
	AdjustVolume(input []byte, volume float64) ([]byte, error)

	// 格式转换
	ConvertFormat(input []byte, fromFormat, toFormat AudioFormat) ([]byte, error)

	// 音频分析
	AnalyzeAudio(input []byte) (*AudioAnalysis, error)

	// 获取支持的音频格式
	GetSupportedFormats() []AudioFormat

	// 获取支持的音效类型
	GetSupportedEffects() []AudioEffectType
}

// AudioFormat 音频格式枚举
type AudioFormat int

const (
	AudioFormatUnknown AudioFormat = iota
	AudioFormatMP3                 // MP3格式
	AudioFormatFLAC                // FLAC格式
	AudioFormatWAV                 // WAV格式
	AudioFormatAAC                 // AAC格式
	AudioFormatOGG                 // OGG格式
	AudioFormatM4A                 // M4A格式
	AudioFormatWMA                 // WMA格式
	AudioFormatAPE                 // APE格式
)

// AudioQuality 音频质量枚举
type AudioQuality int

const (
	AudioQualityLow AudioQuality = iota
	AudioQualityMedium
	AudioQualityStandard
	AudioQualityHigh
	AudioQualityLossless
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
	case AudioFormatAPE:
		return "ape"
	default:
		return "unknown"
	}
}

// AudioEffectType 音效类型枚举
type AudioEffectType int

const (
	AudioEffectTypeNone AudioEffectType = iota
	AudioEffectTypeReverb              // 混响
	AudioEffectTypeEcho                // 回声
	AudioEffectTypeChorus              // 合唱
	AudioEffectTypeDistortion          // 失真
	AudioEffectTypeCompressor          // 压缩器
	AudioEffectTypeEqualizer           // 均衡器
	AudioEffectTypeNormalize           // 标准化
	AudioEffectTypeFade                // 淡入淡出
)

// String 返回音效类型的字符串表示
func (e AudioEffectType) String() string {
	switch e {
	case AudioEffectTypeNone:
		return "none"
	case AudioEffectTypeReverb:
		return "reverb"
	case AudioEffectTypeEcho:
		return "echo"
	case AudioEffectTypeChorus:
		return "chorus"
	case AudioEffectTypeDistortion:
		return "distortion"
	case AudioEffectTypeCompressor:
		return "compressor"
	case AudioEffectTypeEqualizer:
		return "equalizer"
	case AudioEffectTypeNormalize:
		return "normalize"
	case AudioEffectTypeFade:
		return "fade"
	default:
		return "none"
	}
}

// AudioEffect 音效配置结构体
type AudioEffect struct {
	Type       AudioEffectType        `json:"type"`       // 音效类型
	Parameters map[string]interface{} `json:"parameters"` // 音效参数
	Enabled    bool                   `json:"enabled"`    // 是否启用
	Strength   float64                `json:"strength"`   // 强度 (0.0-1.0)
}

// AudioAnalysis 音频分析结果
type AudioAnalysis struct {
	Duration     time.Duration `json:"duration"`      // 音频时长
	SampleRate   int           `json:"sample_rate"`   // 采样率
	Channels     int           `json:"channels"`      // 声道数
	BitRate      int           `json:"bit_rate"`      // 比特率
	Format       AudioFormat   `json:"format"`        // 音频格式
	PeakLevel    float64       `json:"peak_level"`    // 峰值电平
	RMSLevel     float64       `json:"rms_level"`     // RMS电平
	DynamicRange float64       `json:"dynamic_range"` // 动态范围
	Spectrum     []float64     `json:"spectrum"`      // 频谱数据
	Tempo        float64       `json:"tempo"`         // 节拍
	Key          string        `json:"key"`           // 调性
}

// CodecPlugin 编解码插件接口
type CodecPlugin interface {
	Plugin

	// 编码音频
	Encode(ctx context.Context, input []byte, format AudioFormat, options map[string]interface{}) ([]byte, error)

	// 解码音频
	Decode(ctx context.Context, input []byte, format AudioFormat) ([]byte, error)

	// 获取音频信息
	GetAudioInfo(input []byte) (*AudioInfo, error)

	// 检查格式支持
	SupportsFormat(format AudioFormat) bool

	// 获取编码器配置
	GetEncoderConfig(format AudioFormat) map[string]interface{}
}

// AudioInfo 音频信息结构体
type AudioInfo struct {
	Format     AudioFormat   `json:"format"`      // 音频格式
	Duration   time.Duration `json:"duration"`    // 时长
	SampleRate int           `json:"sample_rate"` // 采样率
	Channels   int           `json:"channels"`    // 声道数
	BitRate    int           `json:"bit_rate"`    // 比特率
	Size       int64         `json:"size"`        // 文件大小
	Metadata   AudioMetadata `json:"metadata"`    // 元数据
}

// AudioMetadata 音频元数据
type AudioMetadata struct {
	Title       string `json:"title"`        // 标题
	Artist      string `json:"artist"`       // 艺术家
	Album       string `json:"album"`        // 专辑
	AlbumArtist string `json:"album_artist"` // 专辑艺术家
	Genre       string `json:"genre"`        // 流派
	Year        int    `json:"year"`         // 年份
	Track       int    `json:"track"`        // 音轨号
	Disc        int    `json:"disc"`         // 光盘号
	Comment     string `json:"comment"`      // 注释
	Lyrics      string `json:"lyrics"`       // 歌词
	CoverArt    []byte `json:"cover_art"`    // 封面图片
}