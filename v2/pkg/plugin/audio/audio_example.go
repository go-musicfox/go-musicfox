package plugin

import (
	"fmt"
	"time"
)

// AudioProcessorExample 音频处理插件使用示例
func AudioProcessorExample() {
	// 创建音频处理插件工厂
	factory := NewAudioProcessorFactory()

	// 创建基础音频处理器
	processor, err := factory.CreateProcessor("base", nil)
	if err != nil {
		fmt.Printf("Failed to create processor: %v\n", err)
		return
	}

	// 创建模拟音频数据
	audioData := createSampleAudioData(1024)

	// 1. 基础音频处理
	fmt.Println("=== 基础音频处理 ===")
	processedData, err := processor.ProcessAudio(audioData, 44100, 2)
	if err != nil {
		fmt.Printf("Audio processing failed: %v\n", err)
	} else {
		fmt.Printf("Processed %d bytes of audio data\n", len(processedData))
	}

	// 2. 音量调整
	fmt.Println("\n=== 音量调整 ===")
	volumeAdjusted, err := processor.AdjustVolume(audioData, 0.5) // 50%音量
	if err != nil {
		fmt.Printf("Volume adjustment failed: %v\n", err)
	} else {
		fmt.Printf("Adjusted volume to 50%%, processed %d bytes\n", len(volumeAdjusted))
	}

	// 3. 音效处理
	fmt.Println("\n=== 音效处理 ===")
	
	// 混响效果
	reverbEffect := AudioEffect{
		Type:    AudioEffectTypeReverb,
		Enabled: true,
		Strength: 0.7,
		Parameters: map[string]interface{}{
			"room_size": 0.8,
			"damping":   0.4,
			"wet_level": 0.3,
			"dry_level": 0.7,
		},
	}

	reverbData, err := processor.ApplyEffect(audioData, reverbEffect)
	if err != nil {
		fmt.Printf("Reverb effect failed: %v\n", err)
	} else {
		fmt.Printf("Applied reverb effect, processed %d bytes\n", len(reverbData))
	}

	// 回声效果
	echoEffect := AudioEffect{
		Type:    AudioEffectTypeEcho,
		Enabled: true,
		Strength: 0.5,
		Parameters: map[string]interface{}{
			"delay":    0.3,
			"feedback": 0.4,
			"mix":      0.5,
		},
	}

	echoData, err := processor.ApplyEffect(audioData, echoEffect)
	if err != nil {
		fmt.Printf("Echo effect failed: %v\n", err)
	} else {
		fmt.Printf("Applied echo effect, processed %d bytes\n", len(echoData))
	}

	// 4. 格式转换
	fmt.Println("\n=== 格式转换 ===")
	convertedData, err := processor.ConvertFormat(audioData, AudioFormatWAV, AudioFormatMP3)
	if err != nil {
		fmt.Printf("Format conversion failed: %v\n", err)
	} else {
		fmt.Printf("Converted WAV to MP3, output size: %d bytes\n", len(convertedData))
	}

	// 5. 音频分析
	fmt.Println("\n=== 音频分析 ===")
	analysis, err := processor.AnalyzeAudio(audioData)
	if err != nil {
		fmt.Printf("Audio analysis failed: %v\n", err)
	} else {
		fmt.Printf("Audio Analysis Results:\n")
		fmt.Printf("  Duration: %v\n", analysis.Duration)
		fmt.Printf("  Sample Rate: %d Hz\n", analysis.SampleRate)
		fmt.Printf("  Channels: %d\n", analysis.Channels)
		fmt.Printf("  Bit Rate: %d bps\n", analysis.BitRate)
		fmt.Printf("  Format: %s\n", analysis.Format.String())
		fmt.Printf("  Peak Level: %.2f dB\n", analysis.PeakLevel)
		fmt.Printf("  RMS Level: %.2f dB\n", analysis.RMSLevel)
		fmt.Printf("  Dynamic Range: %.2f dB\n", analysis.DynamicRange)
		fmt.Printf("  Tempo: %.1f BPM\n", analysis.Tempo)
		fmt.Printf("  Key: %s\n", analysis.Key)
	}

	// 6. 插件信息
	fmt.Println("\n=== 插件信息 ===")
	info := processor.GetInfo()
	fmt.Printf("Plugin: %s v%s\n", info.Name, info.Version)
	fmt.Printf("Author: %s\n", info.Author)
	fmt.Printf("Description: %s\n", info.Description)

	// 支持的格式和效果
	supportedFormats := processor.GetSupportedFormats()
	fmt.Printf("Supported Formats: ")
	for i, format := range supportedFormats {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(format.String())
	}
	fmt.Println()

	supportedEffects := processor.GetSupportedEffects()
	fmt.Printf("Supported Effects: ")
	for i, effect := range supportedEffects {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(effect.String())
	}
	fmt.Println()

	// 7. 插件指标
	fmt.Println("\n=== 插件指标 ===")
	metrics, err := processor.GetMetrics()
	if err != nil {
		fmt.Printf("Failed to get metrics: %v\n", err)
	} else {
		fmt.Printf("Request Count: %d\n", metrics.RequestCount)
		fmt.Printf("Error Count: %d\n", metrics.ErrorCount)
		fmt.Printf("Memory Usage: %d bytes\n", metrics.MemoryUsage)
		fmt.Printf("CPU Usage: %.2f%%\n", metrics.CPUUsage*100)
		fmt.Printf("Uptime: %v\n", metrics.Uptime)
	}

	// 8. 清理资源
	fmt.Println("\n=== 清理资源 ===")
	err = processor.Cleanup()
	if err != nil {
		fmt.Printf("Cleanup failed: %v\n", err)
	} else {
		fmt.Println("Plugin cleanup completed successfully")
	}
}

// CodecExample 编解码器使用示例
func CodecExample() {
	fmt.Println("\n\n=== 编解码器示例 ===")

	// 创建音频处理插件工厂
	factory := NewAudioProcessorFactory()

	// 创建基础编解码器
	codec, err := factory.CreateCodec("base", nil)
	if err != nil {
		fmt.Printf("Failed to create codec: %v\n", err)
		return
	}

	// 创建模拟WAV数据
	wavData := createSampleWAVData(1024)

	// 1. 获取音频信息
	fmt.Println("\n=== 音频信息获取 ===")
	audioInfo, err := codec.GetAudioInfo(wavData)
	if err != nil {
		fmt.Printf("Failed to get audio info: %v\n", err)
	} else {
		fmt.Printf("Audio Info:\n")
		fmt.Printf("  Format: %s\n", audioInfo.Format.String())
		fmt.Printf("  Duration: %v\n", audioInfo.Duration)
		fmt.Printf("  Sample Rate: %d Hz\n", audioInfo.SampleRate)
		fmt.Printf("  Channels: %d\n", audioInfo.Channels)
		fmt.Printf("  Bit Rate: %d bps\n", audioInfo.BitRate)
		fmt.Printf("  Size: %d bytes\n", audioInfo.Size)
	}

	// 2. 格式支持检查
	fmt.Println("\n=== 格式支持检查 ===")
	formats := []AudioFormat{AudioFormatMP3, AudioFormatFLAC, AudioFormatWAV, AudioFormatAAC}
	for _, format := range formats {
		supported := codec.SupportsFormat(format)
		fmt.Printf("  %s: %v\n", format.String(), supported)
	}

	// 3. 编码示例
	fmt.Println("\n=== 音频编码 ===")
	// 解码WAV数据
	pcmData, err := codec.Decode(nil, wavData, AudioFormatWAV)
	if err != nil {
		fmt.Printf("WAV decode failed: %v\n", err)
	} else {
		fmt.Printf("Decoded WAV to PCM: %d bytes\n", len(pcmData))

		// 编码为MP3
		mp3Data, err := codec.Encode(nil, pcmData, AudioFormatMP3, map[string]interface{}{
			"bitrate": 320,
			"quality": "high",
		})
		if err != nil {
			fmt.Printf("MP3 encode failed: %v\n", err)
		} else {
			fmt.Printf("Encoded PCM to MP3: %d bytes\n", len(mp3Data))
		}

		// 编码为FLAC
		flacData, err := codec.Encode(nil, pcmData, AudioFormatFLAC, map[string]interface{}{
			"compression_level": 5,
		})
		if err != nil {
			fmt.Printf("FLAC encode failed: %v\n", err)
		} else {
			fmt.Printf("Encoded PCM to FLAC: %d bytes\n", len(flacData))
		}
	}

	// 4. 编码器配置
	fmt.Println("\n=== 编码器配置 ===")
	mp3Config := codec.GetEncoderConfig(AudioFormatMP3)
	fmt.Printf("MP3 Encoder Config:\n")
	for key, value := range mp3Config {
		fmt.Printf("  %s: %v\n", key, value)
	}

	flacConfig := codec.GetEncoderConfig(AudioFormatFLAC)
	fmt.Printf("FLAC Encoder Config:\n")
	for key, value := range flacConfig {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// 清理资源
	err = codec.Cleanup()
	if err != nil {
		fmt.Printf("Codec cleanup failed: %v\n", err)
	} else {
		fmt.Println("Codec cleanup completed successfully")
	}
}

// FactoryExample 工厂使用示例
func FactoryExample() {
	fmt.Println("\n\n=== 工厂使用示例 ===")

	factory := NewAudioProcessorFactory()

	// 1. 获取已注册的处理器和编解码器
	fmt.Println("\n=== 已注册组件 ===")
	processors := factory.GetRegisteredProcessors()
	fmt.Printf("Registered Processors: %v\n", processors)

	codecs := factory.GetRegisteredCodecs()
	fmt.Printf("Registered Codecs: %v\n", codecs)

	// 2. 获取处理器信息
	fmt.Println("\n=== 处理器信息 ===")
	for _, name := range processors {
		info, err := factory.GetProcessorInfo(name)
		if err != nil {
			fmt.Printf("Failed to get info for %s: %v\n", name, err)
			continue
		}
		fmt.Printf("Processor: %s\n", name)
		fmt.Printf("  Name: %s\n", info.Name)
		fmt.Printf("  Version: %s\n", info.Version)
		fmt.Printf("  Author: %s\n", info.Author)
		fmt.Printf("  Description: %s\n", info.Description)
		fmt.Printf("  Supported Formats: %d\n", len(info.SupportedFormats))
		fmt.Printf("  Supported Effects: %d\n", len(info.SupportedEffects))
	}

	// 3. 获取编解码器信息
	fmt.Println("\n=== 编解码器信息 ===")
	for _, name := range codecs {
		info, err := factory.GetCodecInfo(name)
		if err != nil {
			fmt.Printf("Failed to get info for %s: %v\n", name, err)
			continue
		}
		fmt.Printf("Codec: %s\n", name)
		fmt.Printf("  Name: %s\n", info.Name)
		fmt.Printf("  Version: %s\n", info.Version)
		fmt.Printf("  Author: %s\n", info.Author)
		fmt.Printf("  Description: %s\n", info.Description)
		fmt.Printf("  Can Encode: %v\n", info.CanEncode)
		fmt.Printf("  Can Decode: %v\n", info.CanDecode)
		fmt.Printf("  Supported Formats: %d\n", len(info.SupportedFormats))
	}

	// 4. 智能处理器选择
	fmt.Println("\n=== 智能处理器选择 ===")
	requirements := &ProcessorRequirements{
		RequiredFormats: []AudioFormat{AudioFormatMP3, AudioFormatFLAC},
		RequiredEffects: []AudioEffectType{AudioEffectTypeReverb, AudioEffectTypeEcho},
		QualityLevel:    AudioQualityHigh,
		RealtimeMode:    false,
		MaxLatency:      100 * time.Millisecond,
	}

	bestProcessor, err := factory.FindBestProcessor(requirements)
	if err != nil {
		fmt.Printf("Failed to find best processor: %v\n", err)
	} else {
		fmt.Printf("Best processor for requirements: %s\n", bestProcessor)
	}

	// 5. 智能编解码器选择
	fmt.Println("\n=== 智能编解码器选择 ===")
	bestCodec, err := factory.FindBestCodec(AudioFormatMP3, true) // 编码
	if err != nil {
		fmt.Printf("Failed to find best codec for MP3 encoding: %v\n", err)
	} else {
		fmt.Printf("Best codec for MP3 encoding: %s\n", bestCodec)
	}

	bestCodec, err = factory.FindBestCodec(AudioFormatFLAC, false) // 解码
	if err != nil {
		fmt.Printf("Failed to find best codec for FLAC decoding: %v\n", err)
	} else {
		fmt.Printf("Best codec for FLAC decoding: %s\n", bestCodec)
	}

	// 6. 工厂信息
	fmt.Println("\n=== 工厂信息 ===")
	factoryInfo := factory.GetFactoryInfo()
	for key, value := range factoryInfo {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

// 辅助函数

// createSampleAudioData 创建示例音频数据
func createSampleAudioData(size int) []byte {
	data := make([]byte, size)
	for i := 0; i < size; i += 2 {
		// 创建简单的正弦波
		sample := int16(16383) // 50%音量
		data[i] = byte(sample & 0xFF)
		if i+1 < size {
			data[i+1] = byte((sample >> 8) & 0xFF)
		}
	}
	return data
}

// createSampleWAVData 创建示例WAV数据
func createSampleWAVData(audioSize int) []byte {
	// 创建简化的WAV头部
	header := make([]byte, 44)
	copy(header[0:4], "RIFF")
	copy(header[8:12], "WAVE")
	copy(header[12:16], "fmt ")
	copy(header[36:40], "data")

	// 创建音频数据
	audioData := createSampleAudioData(audioSize)

	// 组合头部和数据
	wavData := make([]byte, len(header)+len(audioData))
	copy(wavData, header)
	copy(wavData[len(header):], audioData)

	return wavData
}