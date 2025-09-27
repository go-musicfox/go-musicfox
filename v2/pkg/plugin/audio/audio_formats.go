package plugin

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// FormatConverter 格式转换器接口
type FormatConverter interface {
	Convert(input []byte) ([]byte, error)
	GetSourceFormat() AudioFormat
	GetTargetFormat() AudioFormat
	GetConversionQuality() ConversionQuality
	SetConversionQuality(quality ConversionQuality)
}

// ConversionQuality 转换质量枚举
type ConversionQuality int

const (
	ConversionQualityFast ConversionQuality = iota
	ConversionQualityGood
	ConversionQualityBest
)

// String 返回转换质量的字符串表示
func (q ConversionQuality) String() string {
	switch q {
	case ConversionQualityFast:
		return "fast"
	case ConversionQualityGood:
		return "good"
	case ConversionQualityBest:
		return "best"
	default:
		return "unknown"
	}
}

// BaseFormatConverter 基础格式转换器
type BaseFormatConverter struct {
	FromFormat AudioFormat
	ToFormat   AudioFormat
	Quality    ConversionQuality
}

// NewBaseFormatConverter 创建基础格式转换器
func NewBaseFormatConverter(fromFormat, toFormat AudioFormat) *BaseFormatConverter {
	return &BaseFormatConverter{
		FromFormat: fromFormat,
		ToFormat:   toFormat,
		Quality:    ConversionQualityGood,
	}
}

// Convert 转换音频格式
func (c *BaseFormatConverter) Convert(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("empty input data")
	}

	if c.FromFormat == c.ToFormat {
		return input, nil // 无需转换
	}

	// 根据格式组合选择转换策略
	switch {
	case c.FromFormat == AudioFormatWAV && c.ToFormat == AudioFormatMP3:
		return c.convertWAVToMP3(input)
	case c.FromFormat == AudioFormatMP3 && c.ToFormat == AudioFormatWAV:
		return c.convertMP3ToWAV(input)
	case c.FromFormat == AudioFormatFLAC && c.ToFormat == AudioFormatWAV:
		return c.convertFLACToWAV(input)
	case c.FromFormat == AudioFormatWAV && c.ToFormat == AudioFormatFLAC:
		return c.convertWAVToFLAC(input)
	default:
		// 通用转换：先转换为WAV，再转换为目标格式
		return c.convertViaWAV(input)
	}
}

// GetSourceFormat 获取源格式
func (c *BaseFormatConverter) GetSourceFormat() AudioFormat {
	return c.FromFormat
}

// GetTargetFormat 获取目标格式
func (c *BaseFormatConverter) GetTargetFormat() AudioFormat {
	return c.ToFormat
}

// GetConversionQuality 获取转换质量
func (c *BaseFormatConverter) GetConversionQuality() ConversionQuality {
	return c.Quality
}

// SetConversionQuality 设置转换质量
func (c *BaseFormatConverter) SetConversionQuality(quality ConversionQuality) {
	c.Quality = quality
}

// convertWAVToMP3 WAV转MP3（简化实现）
func (c *BaseFormatConverter) convertWAVToMP3(input []byte) ([]byte, error) {
	// 简化实现：实际应该使用专业的编码库如LAME
	// 这里只是模拟转换过程
	if len(input) < 44 {
		return nil, fmt.Errorf("invalid WAV file: too short")
	}

	// 跳过WAV头部（44字节）
	audioData := input[44:]
	
	// 模拟MP3编码（实际需要使用专业编码器）
	compressedData := c.simulateMP3Compression(audioData)
	
	// 添加简化的MP3头部
	mp3Header := c.createSimpleMP3Header(len(compressedData))
	
	result := make([]byte, len(mp3Header)+len(compressedData))
	copy(result, mp3Header)
	copy(result[len(mp3Header):], compressedData)
	
	return result, nil
}

// convertMP3ToWAV MP3转WAV（简化实现）
func (c *BaseFormatConverter) convertMP3ToWAV(input []byte) ([]byte, error) {
	// 简化实现：实际应该使用专业的解码库
	if len(input) < 10 {
		return nil, fmt.Errorf("invalid MP3 file: too short")
	}

	// 模拟MP3解码
	decodedData := c.simulateMP3Decompression(input)
	
	// 创建WAV头部
	wavHeader := c.createWAVHeader(len(decodedData), 44100, 2, 16)
	
	result := make([]byte, len(wavHeader)+len(decodedData))
	copy(result, wavHeader)
	copy(result[len(wavHeader):], decodedData)
	
	return result, nil
}

// convertFLACToWAV FLAC转WAV（简化实现）
func (c *BaseFormatConverter) convertFLACToWAV(input []byte) ([]byte, error) {
	// 简化实现
	if len(input) < 4 {
		return nil, fmt.Errorf("invalid FLAC file: too short")
	}

	// 模拟FLAC解码
	decodedData := c.simulateFLACDecompression(input)
	
	// 创建WAV头部
	wavHeader := c.createWAVHeader(len(decodedData), 44100, 2, 16)
	
	result := make([]byte, len(wavHeader)+len(decodedData))
	copy(result, wavHeader)
	copy(result[len(wavHeader):], decodedData)
	
	return result, nil
}

// convertWAVToFLAC WAV转FLAC（简化实现）
func (c *BaseFormatConverter) convertWAVToFLAC(input []byte) ([]byte, error) {
	// 简化实现
	if len(input) < 44 {
		return nil, fmt.Errorf("invalid WAV file: too short")
	}

	// 跳过WAV头部
	audioData := input[44:]
	
	// 模拟FLAC编码
	compressedData := c.simulateFLACCompression(audioData)
	
	// 添加FLAC头部
	flacHeader := c.createSimpleFLACHeader(len(compressedData))
	
	result := make([]byte, len(flacHeader)+len(compressedData))
	copy(result, flacHeader)
	copy(result[len(flacHeader):], compressedData)
	
	return result, nil
}

// convertViaWAV 通过WAV进行通用转换
func (c *BaseFormatConverter) convertViaWAV(input []byte) ([]byte, error) {
	// 第一步：转换为WAV
	intermediate, err := c.convertToWAV(input)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to WAV: %w", err)
	}

	// 第二步：从WAV转换为目标格式
	result, err := c.convertFromWAV(intermediate)
	if err != nil {
		return nil, fmt.Errorf("failed to convert from WAV: %w", err)
	}

	return result, nil
}

// convertToWAV 转换为WAV格式
func (c *BaseFormatConverter) convertToWAV(input []byte) ([]byte, error) {
	switch c.FromFormat {
	case AudioFormatWAV:
		return input, nil
	case AudioFormatMP3:
		return c.convertMP3ToWAV(input)
	case AudioFormatFLAC:
		return c.convertFLACToWAV(input)
	default:
		return nil, fmt.Errorf("unsupported source format: %s", c.FromFormat.String())
	}
}

// convertFromWAV 从WAV格式转换
func (c *BaseFormatConverter) convertFromWAV(input []byte) ([]byte, error) {
	switch c.ToFormat {
	case AudioFormatWAV:
		return input, nil
	case AudioFormatMP3:
		return c.convertWAVToMP3(input)
	case AudioFormatFLAC:
		return c.convertWAVToFLAC(input)
	default:
		return nil, fmt.Errorf("unsupported target format: %s", c.ToFormat.String())
	}
}

// 辅助方法：模拟压缩和解压缩

// simulateMP3Compression 模拟MP3压缩
func (c *BaseFormatConverter) simulateMP3Compression(data []byte) []byte {
	// 简化的压缩模拟：减少数据大小
	compressionRatio := 0.1 // 10:1压缩比
	compressedSize := int(float64(len(data)) * compressionRatio)
	if compressedSize < 1 {
		compressedSize = 1
	}
	
	compressed := make([]byte, compressedSize)
	for i := 0; i < compressedSize; i++ {
		if i*10 < len(data) {
			compressed[i] = data[i*10]
		}
	}
	
	return compressed
}

// simulateMP3Decompression 模拟MP3解压缩
func (c *BaseFormatConverter) simulateMP3Decompression(data []byte) []byte {
	// 简化的解压缩模拟：扩展数据大小
	expansionRatio := 10.0 // 1:10扩展比
	decompressedSize := int(float64(len(data)) * expansionRatio)
	
	decompressed := make([]byte, decompressedSize)
	for i := 0; i < len(data) && i*10 < decompressedSize; i++ {
		for j := 0; j < 10 && i*10+j < decompressedSize; j++ {
			decompressed[i*10+j] = data[i]
		}
	}
	
	return decompressed
}

// simulateFLACCompression 模拟FLAC压缩
func (c *BaseFormatConverter) simulateFLACCompression(data []byte) []byte {
	// FLAC是无损压缩，压缩比通常在50-70%
	compressionRatio := 0.6
	compressedSize := int(float64(len(data)) * compressionRatio)
	if compressedSize < 1 {
		compressedSize = 1
	}
	
	compressed := make([]byte, compressedSize)
	for i := 0; i < compressedSize; i++ {
		if i < len(data) {
			compressed[i] = data[i]
		}
	}
	
	return compressed
}

// simulateFLACDecompression 模拟FLAC解压缩
func (c *BaseFormatConverter) simulateFLACDecompression(data []byte) []byte {
	// FLAC解压缩恢复原始数据
	expansionRatio := 1.67 // 约1:1.67扩展比
	decompressedSize := int(float64(len(data)) * expansionRatio)
	
	decompressed := make([]byte, decompressedSize)
	for i := 0; i < len(data) && i < decompressedSize; i++ {
		decompressed[i] = data[i]
	}
	
	return decompressed
}

// 头部创建方法

// createWAVHeader 创建WAV头部
func (c *BaseFormatConverter) createWAVHeader(dataSize, sampleRate, channels, bitsPerSample int) []byte {
	header := make([]byte, 44)
	
	// RIFF头部
	copy(header[0:4], "RIFF")
	binary.LittleEndian.PutUint32(header[4:8], uint32(36+dataSize))
	copy(header[8:12], "WAVE")
	
	// fmt子块
	copy(header[12:16], "fmt ")
	binary.LittleEndian.PutUint32(header[16:20], 16) // fmt子块大小
	binary.LittleEndian.PutUint16(header[20:22], 1)  // PCM格式
	binary.LittleEndian.PutUint16(header[22:24], uint16(channels))
	binary.LittleEndian.PutUint32(header[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(header[28:32], uint32(sampleRate*channels*bitsPerSample/8))
	binary.LittleEndian.PutUint16(header[32:34], uint16(channels*bitsPerSample/8))
	binary.LittleEndian.PutUint16(header[34:36], uint16(bitsPerSample))
	
	// data子块
	copy(header[36:40], "data")
	binary.LittleEndian.PutUint32(header[40:44], uint32(dataSize))
	
	return header
}

// createSimpleMP3Header 创建简化的MP3头部
func (c *BaseFormatConverter) createSimpleMP3Header(dataSize int) []byte {
	// 简化的MP3头部（实际MP3头部更复杂）
	header := make([]byte, 10)
	header[0] = 0xFF // 同步字
	header[1] = 0xFB // MPEG1 Layer3
	header[2] = 0x90 // 比特率和采样率
	header[3] = 0x00 // 其他标志
	
	return header
}

// createSimpleFLACHeader 创建简化的FLAC头部
func (c *BaseFormatConverter) createSimpleFLACHeader(dataSize int) []byte {
	// 简化的FLAC头部
	header := make([]byte, 8)
	copy(header[0:4], "fLaC")
	binary.BigEndian.PutUint32(header[4:8], uint32(dataSize))
	
	return header
}

// VolumeProcessor 音量处理器
type VolumeProcessor struct {
	Gain float64 // 增益值
}

// NewVolumeProcessor 创建音量处理器
func NewVolumeProcessor(gain float64) *VolumeProcessor {
	return &VolumeProcessor{
		Gain: gain,
	}
}

// Process 处理音量
func (v *VolumeProcessor) Process(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return input, nil
	}

	if v.Gain == 1.0 {
		return input, nil // 无需调整
	}

	output := make([]byte, len(input))

	// 处理16位音频样本
	for i := 0; i < len(input)-1; i += 2 {
		// 读取16位样本
		sample := int16(input[i]) | int16(input[i+1])<<8
		
		// 应用增益
		adjustedSample := float64(sample) * v.Gain
		
		// 防止溢出
		if adjustedSample > 32767 {
			adjustedSample = 32767
		} else if adjustedSample < -32768 {
			adjustedSample = -32768
		}
		
		// 转换回整数
		finalSample := int16(adjustedSample)
		
		// 写回输出
		output[i] = byte(finalSample & 0xFF)
		output[i+1] = byte((finalSample >> 8) & 0xFF)
	}

	return output, nil
}

// SetGain 设置增益
func (v *VolumeProcessor) SetGain(gain float64) {
	v.Gain = gain
}

// GetGain 获取增益
func (v *VolumeProcessor) GetGain() float64 {
	return v.Gain
}

// SetVolume 设置音量（0.0-1.0）
func (v *VolumeProcessor) SetVolume(volume float64) {
	if volume < 0.0 {
		volume = 0.0
	} else if volume > 1.0 {
		volume = 1.0
	}
	v.Gain = volume
}

// GetVolume 获取音量
func (v *VolumeProcessor) GetVolume() float64 {
	return math.Min(math.Max(v.Gain, 0.0), 1.0)
}

// ApplyFade 应用淡入淡出
func (v *VolumeProcessor) ApplyFade(input []byte, fadeIn, fadeOut time.Duration, sampleRate int) ([]byte, error) {
	if len(input) == 0 {
		return input, nil
	}

	output := make([]byte, len(input))
	copy(output, input)

	// 计算淡入淡出样本数
	fadeInSamples := int(fadeIn.Seconds() * float64(sampleRate) * 2) // 2字节per sample
	fadeOutSamples := int(fadeOut.Seconds() * float64(sampleRate) * 2)

	// 应用淡入
	for i := 0; i < fadeInSamples && i < len(output)-1; i += 2 {
		sample := int16(output[i]) | int16(output[i+1])<<8
		fadeGain := float64(i) / float64(fadeInSamples)
		fadedSample := int16(float64(sample) * fadeGain)
		
		output[i] = byte(fadedSample & 0xFF)
		output[i+1] = byte((fadedSample >> 8) & 0xFF)
	}

	// 应用淡出
	startFadeOut := len(output) - fadeOutSamples
	if startFadeOut < 0 {
		startFadeOut = 0
	}
	
	for i := startFadeOut; i < len(output)-1; i += 2 {
		sample := int16(output[i]) | int16(output[i+1])<<8
		fadeGain := float64(len(output)-i) / float64(fadeOutSamples)
		fadedSample := int16(float64(sample) * fadeGain)
		
		output[i] = byte(fadedSample & 0xFF)
		output[i+1] = byte((fadedSample >> 8) & 0xFF)
	}

	return output, nil
}