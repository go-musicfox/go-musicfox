package plugin

import (
	"context"
	"fmt"
	"time"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
)

// BaseAudioProcessor 音频处理插件基础实现
type BaseAudioProcessor struct {
	*core.BasePlugin
	supportedFormats []AudioFormat
	supportedEffects []AudioEffectType
	processorConfig  *AudioProcessorConfig
	pipeline         *AudioPipeline
	bufferPool       *AudioBufferPool
}

// AudioProcessorConfig 音频处理器配置
type AudioProcessorConfig struct {
	MaxConcurrency     int           `json:"max_concurrency"`     // 最大并发数
	BufferSize         int           `json:"buffer_size"`         // 缓冲区大小
	ProcessTimeout     time.Duration `json:"process_timeout"`     // 处理超时时间
	EnableOptimization bool          `json:"enable_optimization"` // 启用优化
	QualityPreference  AudioQuality  `json:"quality_preference"`  // 质量偏好
	MemoryLimit        int64         `json:"memory_limit"`        // 内存限制
	CPULimit           float64       `json:"cpu_limit"`           // CPU限制
}

// NewBaseAudioProcessor 创建基础音频处理器
func NewBaseAudioProcessor(info *core.PluginInfo, config *AudioProcessorConfig) *BaseAudioProcessor {
	if config == nil {
		config = &AudioProcessorConfig{
			MaxConcurrency:     4,
			BufferSize:         4096,
			ProcessTimeout:     30 * time.Second,
			EnableOptimization: true,
			QualityPreference:  AudioQualityHigh,
			MemoryLimit:        100 * 1024 * 1024, // 100MB
			CPULimit:           0.8,               // 80%
		}
	}

	return &BaseAudioProcessor{
		BasePlugin:      core.NewBasePlugin(info),
		processorConfig: config,
		pipeline:        NewAudioPipeline(),
		bufferPool:      NewAudioBufferPool(config.BufferSize),
		supportedFormats: []AudioFormat{
			AudioFormatMP3,
			AudioFormatFLAC,
			AudioFormatWAV,
			AudioFormatAAC,
			AudioFormatOGG,
		},
		supportedEffects: []AudioEffectType{
			AudioEffectTypeReverb,
			AudioEffectTypeEcho,
			AudioEffectTypeChorus,
			AudioEffectTypeDistortion,
			AudioEffectTypeCompressor,
			AudioEffectTypeEqualizer,
			AudioEffectTypeNormalize,
			AudioEffectTypeFade,
		},
	}
}

// ProcessAudio 处理音频数据
func (p *BaseAudioProcessor) ProcessAudio(input []byte, sampleRate int, channels int) ([]byte, error) {

	if len(input) == 0 {
		return nil, fmt.Errorf("empty input data")
	}

	if sampleRate <= 0 || channels <= 0 {
		return nil, fmt.Errorf("invalid audio parameters: sampleRate=%d, channels=%d", sampleRate, channels)
	}

	// 创建处理上下文
	ctx, cancel := context.WithTimeout(context.Background(), p.processorConfig.ProcessTimeout)
	defer cancel()

	// 从缓冲池获取缓冲区
	buffer := p.bufferPool.Get(len(input))
	defer p.bufferPool.Put(buffer)

	// 复制输入数据
	copy(buffer.Data, input)
	buffer.SampleRate = sampleRate
	buffer.Channels = channels
	buffer.Length = len(input)

	// 执行音频处理管道
	result, err := p.pipeline.Process(ctx, buffer)
	if err != nil {
		return nil, fmt.Errorf("audio processing failed: %w", err)
	}

	// 返回处理结果
	output := make([]byte, result.Length)
	copy(output, result.Data[:result.Length])

	return output, nil
}

// ApplyEffect 应用音效
func (p *BaseAudioProcessor) ApplyEffect(input []byte, effect AudioEffect) ([]byte, error) {

	if !effect.Enabled {
		return input, nil
	}

	// 检查音效类型是否支持
	if !p.isEffectSupported(effect.Type) {
		return nil, fmt.Errorf("unsupported effect type: %s", effect.Type.String())
	}

	// 创建音效处理器
	processor, err := p.createEffectProcessor(effect.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to create effect processor: %w", err)
	}

	// 应用音效
	output, err := processor.Apply(input, effect.Parameters, effect.Strength)
	if err != nil {
		return nil, fmt.Errorf("failed to apply effect: %w", err)
	}

	return output, nil
}

// AdjustVolume 调整音量
func (p *BaseAudioProcessor) AdjustVolume(input []byte, volume float64) ([]byte, error) {
	if volume < 0.0 || volume > 2.0 {
		return nil, fmt.Errorf("invalid volume level: %f (must be between 0.0 and 2.0)", volume)
	}

	if volume == 1.0 {
		return input, nil // 无需调整
	}

	// 创建音量调整器
	volumeProcessor := &VolumeProcessor{
		Gain: volume,
	}

	return volumeProcessor.Process(input)
}

// ConvertFormat 转换音频格式
func (p *BaseAudioProcessor) ConvertFormat(input []byte, fromFormat, toFormat AudioFormat) ([]byte, error) {
	if fromFormat == toFormat {
		return input, nil // 无需转换
	}

	// 检查格式支持
	if !p.isFormatSupported(fromFormat) {
		return nil, fmt.Errorf("unsupported source format: %s", fromFormat.String())
	}

	if !p.isFormatSupported(toFormat) {
		return nil, fmt.Errorf("unsupported target format: %s", toFormat.String())
	}

	// 创建格式转换器
	converter, err := p.createFormatConverter(fromFormat, toFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to create format converter: %w", err)
	}

	// 执行格式转换
	output, err := converter.Convert(input)
	if err != nil {
		return nil, fmt.Errorf("format conversion failed: %w", err)
	}

	return output, nil
}

// AnalyzeAudio 分析音频
func (p *BaseAudioProcessor) AnalyzeAudio(input []byte) (*AudioAnalysis, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("empty input data")
	}

	// 创建音频分析器
	analyzer := &AudioAnalyzer{}

	// 执行音频分析
	analysis, err := analyzer.Analyze(input)
	if err != nil {
		return nil, fmt.Errorf("audio analysis failed: %w", err)
	}

	return analysis, nil
}

// GetSupportedFormats 获取支持的音频格式
func (p *BaseAudioProcessor) GetSupportedFormats() []AudioFormat {
	formats := make([]AudioFormat, len(p.supportedFormats))
	copy(formats, p.supportedFormats)
	return formats
}

// GetSupportedEffects 获取支持的音效类型
func (p *BaseAudioProcessor) GetSupportedEffects() []AudioEffectType {
	effects := make([]AudioEffectType, len(p.supportedEffects))
	copy(effects, p.supportedEffects)
	return effects
}

// AddSupportedFormat 添加支持的音频格式
func (p *BaseAudioProcessor) AddSupportedFormat(format AudioFormat) {
	// 检查是否已存在
	for _, f := range p.supportedFormats {
		if f == format {
			return
		}
	}

	p.supportedFormats = append(p.supportedFormats, format)
}

// AddSupportedEffect 添加支持的音效类型
func (p *BaseAudioProcessor) AddSupportedEffect(effect AudioEffectType) {
	// 检查是否已存在
	for _, e := range p.supportedEffects {
		if e == effect {
			return
		}
	}

	p.supportedEffects = append(p.supportedEffects, effect)
}

// isFormatSupported 检查格式是否支持
func (p *BaseAudioProcessor) isFormatSupported(format AudioFormat) bool {
	for _, f := range p.supportedFormats {
		if f == format {
			return true
		}
	}
	return false
}

// isEffectSupported 检查音效是否支持
func (p *BaseAudioProcessor) isEffectSupported(effect AudioEffectType) bool {
	for _, e := range p.supportedEffects {
		if e == effect {
			return true
		}
	}
	return false
}

// createEffectProcessor 创建音效处理器
func (p *BaseAudioProcessor) createEffectProcessor(effectType AudioEffectType) (EffectProcessor, error) {
	switch effectType {
	case AudioEffectTypeReverb:
		return &ReverbProcessor{}, nil
	case AudioEffectTypeEcho:
		return &EchoProcessor{}, nil
	case AudioEffectTypeChorus:
		return &ChorusProcessor{}, nil
	case AudioEffectTypeDistortion:
		return &DistortionProcessor{}, nil
	case AudioEffectTypeCompressor:
		return &CompressorProcessor{}, nil
	case AudioEffectTypeEqualizer:
		return &EqualizerProcessor{}, nil
	case AudioEffectTypeNormalize:
		return &NormalizeProcessor{}, nil
	case AudioEffectTypeFade:
		return &FadeProcessor{}, nil
	default:
		return nil, fmt.Errorf("unsupported effect type: %s", effectType.String())
	}
}

// createFormatConverter 创建格式转换器
func (p *BaseAudioProcessor) createFormatConverter(fromFormat, toFormat AudioFormat) (FormatConverter, error) {
	return &BaseFormatConverter{
		FromFormat: fromFormat,
		ToFormat:   toFormat,
	}, nil
}

// GetConfig 获取处理器配置
func (p *BaseAudioProcessor) GetConfig() *AudioProcessorConfig {
	// 返回配置副本
	config := *p.processorConfig
	return &config
}

// UpdateConfig 更新处理器配置（实现Plugin接口）
func (p *BaseAudioProcessor) UpdateConfig(config map[string]interface{}) error {
	// 调用基类的UpdateConfig
	if err := p.BasePlugin.UpdateConfig(config); err != nil {
		return err
	}

	// 如果配置中包含音频处理器特定配置，进行更新
	if audioConfig, ok := config["audio_processor"]; ok {
		if audioConfigMap, ok := audioConfig.(map[string]interface{}); ok {
			return p.updateAudioProcessorConfig(audioConfigMap)
		}
	}

	return nil
}

// UpdateAudioProcessorConfig 更新音频处理器特定配置
func (p *BaseAudioProcessor) UpdateAudioProcessorConfig(config *AudioProcessorConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 验证配置
	if err := p.validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	p.processorConfig = config
	return nil
}

// updateAudioProcessorConfig 内部更新音频处理器配置
func (p *BaseAudioProcessor) updateAudioProcessorConfig(configMap map[string]interface{}) error {

	// 从map中提取配置值
	if maxConcurrency, ok := configMap["max_concurrency"]; ok {
		if val, ok := maxConcurrency.(int); ok {
			p.processorConfig.MaxConcurrency = val
		}
	}

	if bufferSize, ok := configMap["buffer_size"]; ok {
		if val, ok := bufferSize.(int); ok {
			p.processorConfig.BufferSize = val
		}
	}

	if timeout, ok := configMap["process_timeout"]; ok {
		if val, ok := timeout.(string); ok {
			if duration, err := time.ParseDuration(val); err == nil {
				p.processorConfig.ProcessTimeout = duration
			}
		}
	}

	if enableOpt, ok := configMap["enable_optimization"]; ok {
		if val, ok := enableOpt.(bool); ok {
			p.processorConfig.EnableOptimization = val
		}
	}

	// 验证更新后的配置
	return p.validateConfig(p.processorConfig)
}

// validateConfig 验证配置
func (p *BaseAudioProcessor) validateConfig(config *AudioProcessorConfig) error {
	if config.MaxConcurrency <= 0 {
		return fmt.Errorf("max_concurrency must be positive")
	}

	if config.BufferSize <= 0 {
		return fmt.Errorf("buffer_size must be positive")
	}

	if config.ProcessTimeout <= 0 {
		return fmt.Errorf("process_timeout must be positive")
	}

	if config.MemoryLimit <= 0 {
		return fmt.Errorf("memory_limit must be positive")
	}

	if config.CPULimit <= 0 || config.CPULimit > 1.0 {
		return fmt.Errorf("cpu_limit must be between 0 and 1")
	}

	return nil
}

// GetCapabilities 获取能力列表（重写以添加音频处理特定能力）
func (p *BaseAudioProcessor) GetCapabilities() []string {
	// 获取基础能力
	baseCapabilities := p.BasePlugin.GetCapabilities()

	// 添加音频处理特定能力
	audioCapabilities := []string{
		"audio_processing",
		"effect_processing",
		"volume_control",
		"format_conversion",
		"audio_analysis",
	}

	// 合并能力列表
	allCapabilities := make([]string, 0, len(baseCapabilities)+len(audioCapabilities))
	allCapabilities = append(allCapabilities, baseCapabilities...)
	allCapabilities = append(allCapabilities, audioCapabilities...)

	return allCapabilities
}

// GetMetrics 获取处理器指标
func (p *BaseAudioProcessor) GetMetrics() (*PluginMetrics, error) {
	baseMetrics, err := p.BasePlugin.GetMetrics()
	if err != nil {
		return nil, err
	}

	// 添加音频处理特定指标
	baseMetrics.CustomMetrics["supported_formats_count"] = len(p.supportedFormats)
	baseMetrics.CustomMetrics["supported_effects_count"] = len(p.supportedEffects)
	baseMetrics.CustomMetrics["buffer_pool_size"] = p.bufferPool.Size()
	baseMetrics.CustomMetrics["pipeline_stages"] = p.pipeline.StageCount()

	return baseMetrics, nil
}
