package plugin

import (
	"fmt"
	"math"
	"time"
)

// AudioAnalyzer 音频分析器
type AudioAnalyzer struct {
	config *AnalyzerConfig
}

// AnalyzerConfig 分析器配置
type AnalyzerConfig struct {
	EnableSpectrum    bool    `json:"enable_spectrum"`     // 启用频谱分析
	EnableTempoDetection bool `json:"enable_tempo_detection"` // 启用节拍检测
	EnableKeyDetection bool   `json:"enable_key_detection"`   // 启用调性检测
	SpectrumBins      int     `json:"spectrum_bins"`       // 频谱分箱数
	WindowSize        int     `json:"window_size"`         // 窗口大小
	OverlapRatio      float64 `json:"overlap_ratio"`       // 重叠比例
}

// NewAudioAnalyzer 创建音频分析器
func NewAudioAnalyzer() *AudioAnalyzer {
	return &AudioAnalyzer{
		config: &AnalyzerConfig{
			EnableSpectrum:       true,
			EnableTempoDetection: true,
			EnableKeyDetection:   true,
			SpectrumBins:         256,
			WindowSize:           1024,
			OverlapRatio:         0.5,
		},
	}
}

// Analyze 分析音频数据
func (a *AudioAnalyzer) Analyze(input []byte) (*AudioAnalysis, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("empty input data")
	}

	// 创建分析结果
	analysis := &AudioAnalysis{
		Spectrum: make([]float64, 0),
		SampleRate: 44100, // 默认采样率
		Channels: 2,       // 默认立体声
		Format: AudioFormatWAV, // 默认格式
	}

	// 基础音频信息分析
	if err := a.analyzeBasicInfo(input, analysis); err != nil {
		return nil, fmt.Errorf("failed to analyze basic info: %w", err)
	}

	// 音频电平分析
	if err := a.analyzeLevels(input, analysis); err != nil {
		return nil, fmt.Errorf("failed to analyze levels: %w", err)
	}

	// 频谱分析
	if a.config != nil && a.config.EnableSpectrum {
		if err := a.analyzeSpectrum(input, analysis); err != nil {
			return nil, fmt.Errorf("failed to analyze spectrum: %w", err)
		}
	}

	// 节拍检测
	if a.config != nil && a.config.EnableTempoDetection {
		if err := a.detectTempo(input, analysis); err != nil {
			return nil, fmt.Errorf("failed to detect tempo: %w", err)
		}
	}

	// 调性检测
	if a.config != nil && a.config.EnableKeyDetection {
		if err := a.detectKey(input, analysis); err != nil {
			return nil, fmt.Errorf("failed to detect key: %w", err)
		}
	}

	return analysis, nil
}

// analyzeBasicInfo 分析基础音频信息
func (a *AudioAnalyzer) analyzeBasicInfo(input []byte, analysis *AudioAnalysis) error {
	// 假设输入是16位PCM数据
	analysis.SampleRate = 44100 // 默认采样率
	analysis.Channels = 2       // 默认立体声
	analysis.BitRate = analysis.SampleRate * analysis.Channels * 16 // 16位
	analysis.Format = AudioFormatWAV

	// 计算时长（基于数据长度）
	bytesPerSecond := analysis.SampleRate * analysis.Channels * 2 // 2字节per sample
	durationSeconds := float64(len(input)) / float64(bytesPerSecond)
	analysis.Duration = time.Duration(durationSeconds * float64(time.Second))

	return nil
}

// analyzeLevels 分析音频电平
func (a *AudioAnalyzer) analyzeLevels(input []byte, analysis *AudioAnalysis) error {
	if len(input) < 2 {
		return fmt.Errorf("insufficient data for level analysis")
	}

	var peakLevel float64 = 0
	var sumSquares float64 = 0
	sampleCount := len(input) / 2 // 16位样本

	// 分析每个样本
	for i := 0; i < len(input)-1; i += 2 {
		// 读取16位样本
		sample := int16(input[i]) | int16(input[i+1])<<8
		normalizedSample := float64(sample) / 32768.0

		// 更新峰值电平
		absLevel := math.Abs(normalizedSample)
		if absLevel > peakLevel {
			peakLevel = absLevel
		}

		// 累加平方和（用于RMS计算）
		sumSquares += normalizedSample * normalizedSample
	}

	// 计算RMS电平
	rmsLevel := math.Sqrt(sumSquares / float64(sampleCount))

	// 转换为dB
	analysis.PeakLevel = 20 * math.Log10(peakLevel)
	analysis.RMSLevel = 20 * math.Log10(rmsLevel)

	// 计算动态范围
	analysis.DynamicRange = analysis.PeakLevel - analysis.RMSLevel

	return nil
}

// analyzeSpectrum 分析频谱
func (a *AudioAnalyzer) analyzeSpectrum(input []byte, analysis *AudioAnalysis) error {
	if len(input) < a.config.WindowSize*2 {
		return fmt.Errorf("insufficient data for spectrum analysis")
	}

	// 简化的频谱分析实现
	spectrum := make([]float64, a.config.SpectrumBins)
	windowSize := a.config.WindowSize

	// 处理多个窗口
	numWindows := (len(input)/2 - windowSize) / int(float64(windowSize)*(1-a.config.OverlapRatio))
	if numWindows < 1 {
		numWindows = 1
	}

	for w := 0; w < numWindows; w++ {
		startIdx := w * int(float64(windowSize)*(1-a.config.OverlapRatio)) * 2
		if startIdx+windowSize*2 > len(input) {
			break
		}

		// 提取窗口数据
		windowData := make([]float64, windowSize)
		for i := 0; i < windowSize; i++ {
			if startIdx+i*2+1 < len(input) {
				sample := int16(input[startIdx+i*2]) | int16(input[startIdx+i*2+1])<<8
				windowData[i] = float64(sample) / 32768.0
			}
		}

		// 应用窗口函数（汉宁窗）
		a.applyHanningWindow(windowData)

		// 简化的FFT（实际应该使用专业FFT库）
		windowSpectrum := a.simpleFFT(windowData)

		// 累加到总频谱
		for i := 0; i < len(spectrum) && i < len(windowSpectrum); i++ {
			spectrum[i] += windowSpectrum[i]
		}
	}

	// 平均化频谱
	for i := range spectrum {
		spectrum[i] /= float64(numWindows)
	}

	analysis.Spectrum = spectrum
	return nil
}

// applyHanningWindow 应用汉宁窗
func (a *AudioAnalyzer) applyHanningWindow(data []float64) {
	n := len(data)
	for i := 0; i < n; i++ {
		windowValue := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1)))
		data[i] *= windowValue
	}
}

// simpleFFT 简化的FFT实现
func (a *AudioAnalyzer) simpleFFT(data []float64) []float64 {
	n := len(data)
	spectrum := make([]float64, a.config.SpectrumBins)

	// 简化的DFT实现（实际应该使用FFT算法）
	for k := 0; k < a.config.SpectrumBins && k < n/2; k++ {
		var real, imag float64
		for i := 0; i < n; i++ {
			angle := -2 * math.Pi * float64(k) * float64(i) / float64(n)
			real += data[i] * math.Cos(angle)
			imag += data[i] * math.Sin(angle)
		}
		// 计算幅度
		spectrum[k] = math.Sqrt(real*real + imag*imag)
	}

	return spectrum
}

// detectTempo 检测节拍
func (a *AudioAnalyzer) detectTempo(input []byte, analysis *AudioAnalysis) error {
	// 简化的节拍检测实现
	// 实际应该使用更复杂的算法如beat tracking

	// 分析音频的能量变化
	energyChanges := a.analyzeEnergyChanges(input)

	// 寻找周期性模式
	tempo := a.findPeriodicPattern(energyChanges, analysis.SampleRate)

	analysis.Tempo = tempo
	return nil
}

// analyzeEnergyChanges 分析能量变化
func (a *AudioAnalyzer) analyzeEnergyChanges(input []byte) []float64 {
	windowSize := 1024 // 窗口大小
	stepSize := 512    // 步长

	numWindows := (len(input)/2 - windowSize) / stepSize
	if numWindows < 1 {
		numWindows = 1
	}

	energyChanges := make([]float64, numWindows)

	for w := 0; w < numWindows; w++ {
		startIdx := w * stepSize * 2
		if startIdx+windowSize*2 > len(input) {
			break
		}

		// 计算窗口能量
		var energy float64
		for i := 0; i < windowSize; i++ {
			if startIdx+i*2+1 < len(input) {
				sample := int16(input[startIdx+i*2]) | int16(input[startIdx+i*2+1])<<8
				normalizedSample := float64(sample) / 32768.0
				energy += normalizedSample * normalizedSample
			}
		}

		energyChanges[w] = energy / float64(windowSize)
	}

	return energyChanges
}

// findPeriodicPattern 寻找周期性模式
func (a *AudioAnalyzer) findPeriodicPattern(energyChanges []float64, sampleRate int) float64 {
	// 简化的周期检测
	// 寻找能量变化的自相关峰值

	maxLag := len(energyChanges) / 4 // 最大滞后
	bestTempo := 120.0              // 默认节拍
	maxCorrelation := 0.0

	for lag := 10; lag < maxLag; lag++ {
		correlation := a.calculateAutoCorrelation(energyChanges, lag)
		if correlation > maxCorrelation {
			maxCorrelation = correlation
			// 将滞后转换为BPM
			// 假设每个能量窗口代表约0.01秒
			periodSeconds := float64(lag) * 0.01
			if periodSeconds > 0 {
				bestTempo = 60.0 / periodSeconds
			}
		}
	}

	// 限制节拍范围
	if bestTempo < 60 {
		bestTempo = 60
	} else if bestTempo > 200 {
		bestTempo = 200
	}

	return bestTempo
}

// calculateAutoCorrelation 计算自相关
func (a *AudioAnalyzer) calculateAutoCorrelation(data []float64, lag int) float64 {
	if lag >= len(data) {
		return 0
	}

	var correlation float64
	count := len(data) - lag

	for i := 0; i < count; i++ {
		correlation += data[i] * data[i+lag]
	}

	return correlation / float64(count)
}

// detectKey 检测调性
func (a *AudioAnalyzer) detectKey(input []byte, analysis *AudioAnalysis) error {
	// 简化的调性检测实现
	// 实际应该使用色度特征和调性模板匹配

	// 分析色度特征
	chromaFeatures := a.analyzeChromaFeatures(input)

	// 匹配调性模板
	key := a.matchKeyTemplate(chromaFeatures)

	analysis.Key = key
	return nil
}

// analyzeChromaFeatures 分析色度特征
func (a *AudioAnalyzer) analyzeChromaFeatures(input []byte) []float64 {
	// 12个半音的色度特征
	chroma := make([]float64, 12)

	// 简化实现：基于频谱分析
	// 这里需要先进行频谱分析或使用已有的频谱数据
	// 为了演示，我们创建一个简单的频谱
	spectrum := make([]float64, 256)
	for i := 0; i < len(spectrum); i++ {
		// 简单的频谱模拟
		spectrum[i] = float64(i % 12) / 12.0
	}
	
	// 将频谱映射到12个半音
	for i, magnitude := range spectrum {
		if i < len(spectrum)/2 { // 只考虑正频率
			// 简化的频率到音符映射
			noteIndex := a.frequencyToNoteIndex(float64(i) * 44100.0 / float64(len(spectrum)*2))
			if noteIndex >= 0 && noteIndex < 12 {
				chroma[noteIndex] += magnitude
			}
		}
	}

	// 归一化色度特征
	var sum float64
	for _, value := range chroma {
		sum += value
	}
	if sum > 0 {
		for i := range chroma {
			chroma[i] /= sum
		}
	}

	return chroma
}

// frequencyToNoteIndex 频率转音符索引
func (a *AudioAnalyzer) frequencyToNoteIndex(frequency float64) int {
	if frequency <= 0 {
		return -1
	}

	// A4 = 440Hz 对应索引9（A）
	a4Freq := 440.0
	a4Index := 9

	// 计算相对于A4的半音数
	semitones := 12 * math.Log2(frequency/a4Freq)
	noteIndex := (a4Index + int(math.Round(semitones))) % 12

	if noteIndex < 0 {
		noteIndex += 12
	}

	return noteIndex
}

// matchKeyTemplate 匹配调性模板
func (a *AudioAnalyzer) matchKeyTemplate(chroma []float64) string {
	// 大调和小调的模板
	majorTemplate := []float64{1, 0, 1, 0, 1, 1, 0, 1, 0, 1, 0, 1} // C大调模板
	minorTemplate := []float64{1, 0, 1, 1, 0, 1, 0, 1, 1, 0, 1, 0} // C小调模板

	noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

	bestKey := "C"
	maxCorrelation := 0.0

	// 尝试所有12个调性的大调和小调
	for root := 0; root < 12; root++ {
		// 大调
		majorCorr := a.calculateTemplateCorrelation(chroma, majorTemplate, root)
		if majorCorr > maxCorrelation {
			maxCorrelation = majorCorr
			bestKey = noteNames[root]
		}

		// 小调
		minorCorr := a.calculateTemplateCorrelation(chroma, minorTemplate, root)
		if minorCorr > maxCorrelation {
			maxCorrelation = minorCorr
			bestKey = noteNames[root] + "m"
		}
	}

	return bestKey
}

// calculateTemplateCorrelation 计算模板相关性
func (a *AudioAnalyzer) calculateTemplateCorrelation(chroma, template []float64, root int) float64 {
	if len(chroma) != len(template) {
		return 0
	}

	var correlation float64
	for i := 0; i < len(chroma); i++ {
		templateIndex := (i - root + 12) % 12
		correlation += chroma[i] * template[templateIndex]
	}

	return correlation
}

// GetConfig 获取分析器配置
func (a *AudioAnalyzer) GetConfig() *AnalyzerConfig {
	// 返回配置副本
	config := *a.config
	return &config
}

// UpdateConfig 更新分析器配置
func (a *AudioAnalyzer) UpdateConfig(config *AnalyzerConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 验证配置
	if err := a.validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	a.config = config
	return nil
}

// validateConfig 验证配置
func (a *AudioAnalyzer) validateConfig(config *AnalyzerConfig) error {
	if config.SpectrumBins <= 0 {
		return fmt.Errorf("spectrum_bins must be positive")
	}

	if config.WindowSize <= 0 {
		return fmt.Errorf("window_size must be positive")
	}

	if config.OverlapRatio < 0 || config.OverlapRatio >= 1 {
		return fmt.Errorf("overlap_ratio must be between 0 and 1")
	}

	return nil
}