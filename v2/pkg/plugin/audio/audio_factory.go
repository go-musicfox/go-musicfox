package plugin

import (
	"fmt"
	"sync"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// AudioProcessorFactory 音频处理插件工厂
type AudioProcessorFactory struct {
	registeredProcessors map[string]AudioProcessorCreator
	registeredCodecs     map[string]CodecCreator
	mu                   sync.RWMutex
}

// AudioProcessorCreator 音频处理器创建函数类型
type AudioProcessorCreator func(config *AudioProcessorConfig) (AudioProcessorPlugin, error)

// CodecCreator 编解码器创建函数类型
type CodecCreator func(config map[string]interface{}) (CodecPlugin, error)

// ProcessorInfo 处理器信息
type ProcessorInfo struct {
	Name         string              `json:"name"`         // 处理器名称
	Description  string              `json:"description"`  // 描述
	Version      string              `json:"version"`      // 版本
	Author       string              `json:"author"`       // 作者
	Capabilities []string            `json:"capabilities"` // 能力列表
	SupportedFormats []AudioFormat   `json:"supported_formats"` // 支持的格式
	SupportedEffects []AudioEffectType `json:"supported_effects"` // 支持的音效
	DefaultConfig *AudioProcessorConfig `json:"default_config"` // 默认配置
}

// CodecInfo 编解码器信息
type CodecInfo struct {
	Name             string        `json:"name"`              // 编解码器名称
	Description      string        `json:"description"`       // 描述
	Version          string        `json:"version"`           // 版本
	Author           string        `json:"author"`            // 作者
	SupportedFormats []AudioFormat `json:"supported_formats"` // 支持的格式
	CanEncode        bool          `json:"can_encode"`        // 是否支持编码
	CanDecode        bool          `json:"can_decode"`        // 是否支持解码
	DefaultConfig    map[string]interface{} `json:"default_config"` // 默认配置
}

// NewAudioProcessorFactory 创建音频处理插件工厂
func NewAudioProcessorFactory() *AudioProcessorFactory {
	factory := &AudioProcessorFactory{
		registeredProcessors: make(map[string]AudioProcessorCreator),
		registeredCodecs:     make(map[string]CodecCreator),
	}

	// 注册内置处理器
	factory.registerBuiltinProcessors()
	factory.registerBuiltinCodecs()

	return factory
}

// registerBuiltinProcessors 注册内置处理器
func (f *AudioProcessorFactory) registerBuiltinProcessors() {
	// 注册基础音频处理器
	f.RegisterProcessor("base", func(config *AudioProcessorConfig) (AudioProcessorPlugin, error) {
		info := &core.PluginInfo{
			Name:        "Base Audio Processor",
			Version:     "1.0.0",
			Description: "Basic audio processing capabilities",
			Author:      "go-musicfox",
		}
		return NewBaseAudioProcessor(info, config), nil
	})

	// 注册高质量音频处理器
	f.RegisterProcessor("high_quality", func(config *AudioProcessorConfig) (AudioProcessorPlugin, error) {
		info := &core.PluginInfo{
			Name:        "High Quality Audio Processor",
			Version:     "1.0.0",
			Description: "High quality audio processing with advanced features",
			Author:      "go-musicfox",
		}
		
		// 高质量配置
		if config == nil {
			config = &AudioProcessorConfig{
				MaxConcurrency:     8,
				BufferSize:         8192,
				EnableOptimization: true,
				QualityPreference:  AudioQualityLossless,
			}
		}
		
		processor := NewBaseAudioProcessor(info, config)
		
		// 添加更多支持的格式和效果
		processor.AddSupportedFormat(AudioFormatM4A)
		processor.AddSupportedFormat(AudioFormatWMA)
		processor.AddSupportedFormat(AudioFormatAPE)
		
		return processor, nil
	})

	// 注册实时音频处理器
	f.RegisterProcessor("realtime", func(config *AudioProcessorConfig) (AudioProcessorPlugin, error) {
		info := &core.PluginInfo{
			Name:        "Realtime Audio Processor",
			Version:     "1.0.0",
			Description: "Optimized for real-time audio processing",
			Author:      "go-musicfox",
		}
		
		// 实时处理配置
		if config == nil {
			config = &AudioProcessorConfig{
				MaxConcurrency:     2,
				BufferSize:         1024,
				ProcessTimeout:     100 * time.Millisecond,
				EnableOptimization: false, // 减少延迟
				QualityPreference:  AudioQualityStandard,
			}
		}
		
		return NewBaseAudioProcessor(info, config), nil
	})
}

// registerBuiltinCodecs 注册内置编解码器
func (f *AudioProcessorFactory) registerBuiltinCodecs() {
	// 注册基础编解码器
	f.RegisterCodec("base", func(config map[string]interface{}) (CodecPlugin, error) {
		return NewBaseCodec(), nil
	})

	// 注册高质量编解码器
	f.RegisterCodec("high_quality", func(config map[string]interface{}) (CodecPlugin, error) {
		return NewHighQualityCodec(config), nil
	})
}

// RegisterProcessor 注册音频处理器
func (f *AudioProcessorFactory) RegisterProcessor(name string, creator AudioProcessorCreator) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if name == "" {
		return fmt.Errorf("processor name cannot be empty")
	}

	if creator == nil {
		return fmt.Errorf("processor creator cannot be nil")
	}

	f.registeredProcessors[name] = creator
	return nil
}

// RegisterCodec 注册编解码器
func (f *AudioProcessorFactory) RegisterCodec(name string, creator CodecCreator) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if name == "" {
		return fmt.Errorf("codec name cannot be empty")
	}

	if creator == nil {
		return fmt.Errorf("codec creator cannot be nil")
	}

	f.registeredCodecs[name] = creator
	return nil
}

// CreateProcessor 创建音频处理器
func (f *AudioProcessorFactory) CreateProcessor(name string, config *AudioProcessorConfig) (AudioProcessorPlugin, error) {
	f.mu.RLock()
	creator, exists := f.registeredProcessors[name]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("processor '%s' not found", name)
	}

	processor, err := creator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create processor '%s': %w", name, err)
	}

	return processor, nil
}

// CreateCodec 创建编解码器
func (f *AudioProcessorFactory) CreateCodec(name string, config map[string]interface{}) (CodecPlugin, error) {
	f.mu.RLock()
	creator, exists := f.registeredCodecs[name]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("codec '%s' not found", name)
	}

	codec, err := creator(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create codec '%s': %w", name, err)
	}

	return codec, nil
}

// GetRegisteredProcessors 获取已注册的处理器列表
func (f *AudioProcessorFactory) GetRegisteredProcessors() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	processors := make([]string, 0, len(f.registeredProcessors))
	for name := range f.registeredProcessors {
		processors = append(processors, name)
	}

	return processors
}

// GetRegisteredCodecs 获取已注册的编解码器列表
func (f *AudioProcessorFactory) GetRegisteredCodecs() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	codecs := make([]string, 0, len(f.registeredCodecs))
	for name := range f.registeredCodecs {
		codecs = append(codecs, name)
	}

	return codecs
}

// GetProcessorInfo 获取处理器信息
func (f *AudioProcessorFactory) GetProcessorInfo(name string) (*ProcessorInfo, error) {
	f.mu.RLock()
	_, exists := f.registeredProcessors[name]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("processor '%s' not found", name)
	}

	// 创建临时实例获取信息
	processor, err := f.CreateProcessor(name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create processor for info: %w", err)
	}
	defer processor.Cleanup()

	info := &ProcessorInfo{
		Name:             processor.GetInfo().Name,
		Description:      processor.GetInfo().Description,
		Version:          processor.GetInfo().Version,
		Author:           processor.GetInfo().Author,
		Capabilities:     processor.GetCapabilities(),
		SupportedFormats: processor.GetSupportedFormats(),
		SupportedEffects: processor.GetSupportedEffects(),
	}

	// 获取默认配置
	if baseProcessor, ok := processor.(*BaseAudioProcessor); ok {
		info.DefaultConfig = baseProcessor.GetConfig()
	}

	return info, nil
}

// GetCodecInfo 获取编解码器信息
func (f *AudioProcessorFactory) GetCodecInfo(name string) (*CodecInfo, error) {
	f.mu.RLock()
	_, exists := f.registeredCodecs[name]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("codec '%s' not found", name)
	}

	// 创建临时实例获取信息
	codec, err := f.CreateCodec(name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create codec for info: %w", err)
	}
	defer codec.Cleanup()

	info := &CodecInfo{
		Name:        codec.GetInfo().Name,
		Description: codec.GetInfo().Description,
		Version:     codec.GetInfo().Version,
		Author:      codec.GetInfo().Author,
	}

	// 检查支持的格式
	allFormats := []AudioFormat{
		AudioFormatMP3, AudioFormatFLAC, AudioFormatWAV,
		AudioFormatAAC, AudioFormatOGG, AudioFormatM4A,
		AudioFormatWMA, AudioFormatAPE,
	}

	for _, format := range allFormats {
		if codec.SupportsFormat(format) {
			info.SupportedFormats = append(info.SupportedFormats, format)
		}
	}

	// 检查编解码能力（简化实现）
	info.CanEncode = true
	info.CanDecode = true

	return info, nil
}

// FindBestProcessor 根据需求查找最佳处理器
func (f *AudioProcessorFactory) FindBestProcessor(requirements *ProcessorRequirements) (string, error) {
	if requirements == nil {
		return "base", nil // 默认处理器
	}

	f.mu.RLock()
	processorNames := make([]string, 0, len(f.registeredProcessors))
	for name := range f.registeredProcessors {
		processorNames = append(processorNames, name)
	}
	f.mu.RUnlock()

	bestProcessor := ""
	bestScore := -1

	for _, name := range processorNames {
		info, err := f.GetProcessorInfo(name)
		if err != nil {
			continue
		}

		score := f.calculateProcessorScore(info, requirements)
		if score > bestScore {
			bestScore = score
			bestProcessor = name
		}
	}

	if bestProcessor == "" {
		return "", fmt.Errorf("no suitable processor found")
	}

	return bestProcessor, nil
}

// ProcessorRequirements 处理器需求
type ProcessorRequirements struct {
	RequiredFormats []AudioFormat     `json:"required_formats"` // 必需的格式
	RequiredEffects []AudioEffectType `json:"required_effects"` // 必需的音效
	QualityLevel    AudioQuality      `json:"quality_level"`    // 质量要求
	RealtimeMode    bool              `json:"realtime_mode"`    // 实时模式
	MaxLatency      time.Duration     `json:"max_latency"`      // 最大延迟
	MaxMemoryUsage  int64             `json:"max_memory_usage"` // 最大内存使用
}

// calculateProcessorScore 计算处理器评分
func (f *AudioProcessorFactory) calculateProcessorScore(info *ProcessorInfo, requirements *ProcessorRequirements) int {
	score := 0

	// 检查格式支持
	for _, requiredFormat := range requirements.RequiredFormats {
		for _, supportedFormat := range info.SupportedFormats {
			if requiredFormat == supportedFormat {
				score += 10
				break
			}
		}
	}

	// 检查音效支持
	for _, requiredEffect := range requirements.RequiredEffects {
		for _, supportedEffect := range info.SupportedEffects {
			if requiredEffect == supportedEffect {
				score += 5
				break
			}
		}
	}

	// 质量匹配
	if info.DefaultConfig != nil {
		if info.DefaultConfig.QualityPreference == requirements.QualityLevel {
			score += 15
		} else if int(info.DefaultConfig.QualityPreference) >= int(requirements.QualityLevel) {
			score += 5
		}
	}

	// 实时模式匹配
	if requirements.RealtimeMode {
		if info.Name == "Realtime Audio Processor" {
			score += 20
		}
	}

	// 延迟要求
	if requirements.MaxLatency > 0 && info.DefaultConfig != nil {
		if info.DefaultConfig.ProcessTimeout <= requirements.MaxLatency {
			score += 10
		}
	}

	return score
}

// FindBestCodec 根据需求查找最佳编解码器
func (f *AudioProcessorFactory) FindBestCodec(format AudioFormat, encode bool) (string, error) {
	f.mu.RLock()
	codecNames := make([]string, 0, len(f.registeredCodecs))
	for name := range f.registeredCodecs {
		codecNames = append(codecNames, name)
	}
	f.mu.RUnlock()

	for _, name := range codecNames {
		info, err := f.GetCodecInfo(name)
		if err != nil {
			continue
		}

		// 检查格式支持
		formatSupported := false
		for _, supportedFormat := range info.SupportedFormats {
			if supportedFormat == format {
				formatSupported = true
				break
			}
		}

		if !formatSupported {
			continue
		}

		// 检查编解码能力
		if encode && !info.CanEncode {
			continue
		}
		if !encode && !info.CanDecode {
			continue
		}

		return name, nil
	}

	return "", fmt.Errorf("no suitable codec found for format %s", format.String())
}

// UnregisterProcessor 注销音频处理器
func (f *AudioProcessorFactory) UnregisterProcessor(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.registeredProcessors[name]; !exists {
		return fmt.Errorf("processor '%s' not found", name)
	}

	delete(f.registeredProcessors, name)
	return nil
}

// UnregisterCodec 注销编解码器
func (f *AudioProcessorFactory) UnregisterCodec(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.registeredCodecs[name]; !exists {
		return fmt.Errorf("codec '%s' not found", name)
	}

	delete(f.registeredCodecs, name)
	return nil
}

// Clear 清空工厂
func (f *AudioProcessorFactory) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.registeredProcessors = make(map[string]AudioProcessorCreator)
	f.registeredCodecs = make(map[string]CodecCreator)
}

// GetFactoryInfo 获取工厂信息
func (f *AudioProcessorFactory) GetFactoryInfo() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return map[string]interface{}{
		"registered_processors": len(f.registeredProcessors),
		"registered_codecs":     len(f.registeredCodecs),
		"processor_names":       f.GetRegisteredProcessors(),
		"codec_names":           f.GetRegisteredCodecs(),
	}
}