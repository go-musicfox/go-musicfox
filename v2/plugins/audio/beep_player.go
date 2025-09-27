package audio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
)

// BeepPlayer Beep播放器后端实现
type BeepPlayer struct {
	*BasePlayer
	streamer    beep.StreamSeekCloser
	ctrl        *beep.Ctrl
	volume      *effects.Volume
	speakerInit bool
	mutex       sync.RWMutex
	config      map[string]interface{}

	// 缓存相关字段
	cacheReader     *os.File
	cacheWriter     *os.File
	cacheDownloaded bool
	cacheFile       string
	httpClient      *http.Client
	downloadCtx     context.Context
	downloadCancel  context.CancelFunc
	curFormat       beep.Format
}

// NewBeepPlayer 创建新的Beep播放器实例
func NewBeepPlayer(config map[string]interface{}) *BeepPlayer {
	formats := []string{"mp3", "wav", "flac", "ogg"}
	return &BeepPlayer{
		BasePlayer: NewBasePlayerWithInfo("Beep Player", "1.0.0", formats),
		config:     config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Initialize 初始化播放器
func (p *BeepPlayer) Initialize(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.speakerInit {
		// 初始化音频输出
		sampleRate := beep.SampleRate(44100)
		if p.config != nil {
			if sr, ok := p.config["sample_rate"].(int); ok {
				sampleRate = beep.SampleRate(sr)
			}
		}

		err := speaker.Init(sampleRate, sampleRate.N(time.Second/10))
		if err != nil {
			return fmt.Errorf("failed to initialize speaker: %w", err)
		}
		p.speakerInit = true
	}

	return nil
}

// Cleanup 清理资源
func (p *BeepPlayer) Cleanup() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.reset()
	p.setPlaying(false)

	return nil
}

// Play 播放音频
func (p *BeepPlayer) Play(url string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 重置之前的状态
	p.reset()

	// 判断是否为网络URL
	isURL := strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")

	if isURL {
		// 网络URL：使用缓存机制
		return p.playFromURL(url)
	} else {
		// 本地文件：直接播放
		return p.playFromFile(url)
	}
}

// Pause 暂停播放
func (p *BeepPlayer) Pause() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.ctrl != nil {
		speaker.Lock()
		p.ctrl.Paused = true
		speaker.Unlock()
		p.setPlaying(false)
	}

	return nil
}

// Resume 恢复播放
func (p *BeepPlayer) Resume() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.ctrl != nil {
		speaker.Lock()
		p.ctrl.Paused = false
		speaker.Unlock()
		p.setPlaying(true)
	}

	return nil
}

// Stop 停止播放
func (p *BeepPlayer) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	speaker.Clear()

	if p.streamer != nil {
		p.streamer.Close()
		p.streamer = nil
	}

	p.ctrl = nil
	p.volume = nil
	p.setPlaying(false)
	p.setPosition(0)

	return nil
}

// Seek 跳转到指定位置
func (p *BeepPlayer) Seek(position time.Duration) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.streamer == nil {
		return fmt.Errorf("no audio loaded")
	}

	// 计算样本位置
	sampleRate := beep.SampleRate(44100) // 默认采样率
	samplePos := sampleRate.N(position)

	// 跳转
	err := p.streamer.Seek(samplePos)
	if err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	p.setPosition(position)
	return nil
}

// SetVolume 设置音量
func (p *BeepPlayer) SetVolume(volume float64) error {
	if err := p.BasePlayer.SetVolume(volume); err != nil {
		return err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.volume != nil {
		// 转换为分贝值 (0.0-1.0 -> -∞ to 0 dB)
		var dbVolume float64
		if volume <= 0 {
			p.volume.Silent = true
		} else {
			p.volume.Silent = false
			// 简单的线性映射到分贝
			dbVolume = (volume - 1) * 10 // -10dB to 0dB
		}

		speaker.Lock()
		p.volume.Volume = dbVolume
		speaker.Unlock()
	}

	return nil
}

// IsAvailable 检查播放器是否可用
func (p *BeepPlayer) IsAvailable() bool {
	// Beep是纯Go实现，应该在所有平台都可用
	return true
}

// playFromFile 播放本地文件
func (p *BeepPlayer) playFromFile(filePath string) error {
	// 打开本地文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	// 解码音频
	streamer, format, err := p.decodeAudio(file, filePath)
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to decode audio: %w", err)
	}

	return p.startPlayback(streamer, format)
}

// playFromURL 播放网络URL（使用缓存机制）
func (p *BeepPlayer) playFromURL(url string) error {
	// 创建缓存文件
	if err := p.createCacheFile(); err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}

	// 创建下载上下文
	p.downloadCtx, p.downloadCancel = context.WithCancel(context.Background())

	// 启动下载协程
	go p.downloadToCache(url)

	// 等待足够的缓冲数据
	bufferSize := p.getBufferSize(url)
	if err := p.waitForNBytes(bufferSize, 100*time.Millisecond, 50); err != nil {
		return fmt.Errorf("failed to buffer audio data: %w", err)
	}

	// 从缓存文件解码音频
	streamer, format, err := p.decodeAudio(p.cacheReader, url)
	if err != nil {
		return fmt.Errorf("failed to decode cached audio: %w", err)
	}

	return p.startPlayback(streamer, format)
}

// startPlayback 开始播放
func (p *BeepPlayer) startPlayback(streamer beep.StreamSeekCloser, format beep.Format) error {
	p.streamer = streamer
	p.curFormat = format
	p.setDuration(format.SampleRate.D(streamer.Len()))

	// 创建音量控制
	p.volume = &effects.Volume{
		Streamer: p.streamer,
		Base:     2,
		Volume:   0, // 0 dB = 原音量
		Silent:   false,
	}

	// 创建播放控制
	p.ctrl = &beep.Ctrl{
		Streamer: p.volume,
		Paused:   false,
	}

	// 开始播放
	speaker.Play(beep.Seq(p.ctrl, beep.Callback(func() {
		p.setPlaying(false)
		p.setPosition(0)
	})))

	p.setPlaying(true)
	p.setPosition(0)

	// 启动位置更新协程
	go p.updatePosition()

	return nil
}

// updatePosition 更新播放位置（带错误恢复）
func (p *BeepPlayer) updatePosition() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		p.mutex.RLock()
		if !p.playing || p.streamer == nil {
			p.mutex.RUnlock()
			break
		}

		// 检查流错误
		if err := p.streamer.Err(); err != nil {
			p.mutex.RUnlock()
			// 尝试错误恢复
			if p.handleStreamError(err) {
				continue
			} else {
				break
			}
		}

		// 获取当前播放位置
		sampleRate := beep.SampleRate(44100)
		if p.curFormat.SampleRate > 0 {
			sampleRate = p.curFormat.SampleRate
		}
		currentPos := p.streamer.Position()
		position := sampleRate.D(currentPos)

		p.mutex.RUnlock()
		p.setPosition(position)
	}
}

// handleStreamError 处理流错误
func (p *BeepPlayer) handleStreamError(err error) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 如果是网络流且还在下载，尝试恢复
	if !p.cacheDownloaded && p.cacheReader != nil {
		// 暂停播放等待更多数据
		if p.ctrl != nil {
			speaker.Lock()
			p.ctrl.Paused = true
			speaker.Unlock()
		}

		// 等待一段时间后恢复
		go func() {
			time.Sleep(2 * time.Second)
			p.mutex.Lock()
			defer p.mutex.Unlock()
			if p.ctrl != nil && p.playing {
				speaker.Lock()
				p.ctrl.Paused = false
				speaker.Unlock()
			}
		}()
		return true
	}

	// 其他错误，停止播放
	p.setPlaying(false)
	return false
}

// createCacheFile 创建缓存文件
func (p *BeepPlayer) createCacheFile() error {
	// 创建临时缓存文件
	tempDir := os.TempDir()
	p.cacheFile = filepath.Join(tempDir, "beep_playing_cache")

	// 创建读写文件句柄
	var err error
	p.cacheReader, err = os.OpenFile(p.cacheFile, os.O_CREATE|os.O_TRUNC|os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to create cache reader: %w", err)
	}

	p.cacheWriter, err = os.OpenFile(p.cacheFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		p.cacheReader.Close()
		return fmt.Errorf("failed to create cache writer: %w", err)
	}

	p.cacheDownloaded = false
	return nil
}

// reset 重置播放器状态和资源
func (p *BeepPlayer) reset() {
	// 取消下载
	if p.downloadCancel != nil {
		p.downloadCancel()
		p.downloadCancel = nil
	}

	// 关闭缓存文件
	if p.cacheReader != nil {
		p.cacheReader.Close()
		p.cacheReader = nil
	}
	if p.cacheWriter != nil {
		p.cacheWriter.Close()
		p.cacheWriter = nil
	}

	// 关闭音频流
	if p.streamer != nil {
		p.streamer.Close()
		p.streamer = nil
	}

	// 清理播放器状态
	speaker.Clear()
	p.ctrl = nil
	p.volume = nil
	p.cacheDownloaded = false

	// 删除缓存文件
	if p.cacheFile != "" {
		os.Remove(p.cacheFile)
		p.cacheFile = ""
	}
}

// downloadToCache 下载音频到缓存文件（带重试机制）
func (p *BeepPlayer) downloadToCache(url string) {
	defer func() {
		if r := recover(); r != nil {
			// 发生panic时停止播放
			p.mutex.Lock()
			p.setPlaying(false)
			p.mutex.Unlock()
		}
	}()

	// 重试机制
	maxRetries := 3
	for retry := 0; retry < maxRetries; retry++ {
		if p.downloadCtx.Err() != nil {
			return
		}

		if p.downloadWithRetry(url, retry) {
			return // 成功下载
		}

		// 重试前等待
		if retry < maxRetries-1 {
			select {
			case <-time.After(time.Duration(retry+1) * time.Second):
			case <-p.downloadCtx.Done():
				return
			}
		}
	}
}

// downloadWithRetry 单次下载尝试
func (p *BeepPlayer) downloadWithRetry(url string, retryCount int) bool {
	// 发起HTTP请求
	resp, err := p.httpClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return false
	}

	// 边下载边写入缓存文件
	buf := make([]byte, 32*1024) // 32KB缓冲区
	for {
		select {
		case <-p.downloadCtx.Done():
			return false
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := p.cacheWriter.Write(buf[:n]); writeErr != nil {
				return false
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return false
		}
	}

	// 下载完成
	p.mutex.Lock()
	p.cacheDownloaded = true
	p.mutex.Unlock()
	return true
}

// decodeAudio 解码音频文件（优化版）
func (p *BeepPlayer) decodeAudio(reader io.ReadCloser, url string) (beep.StreamSeekCloser, beep.Format, error) {
	// 根据文件扩展名选择解码器
	ext := strings.ToLower(filepath.Ext(url))

	// 优先使用扩展名匹配的解码器
	switch ext {
	case ".mp3":
		if streamer, format, err := mp3.Decode(reader); err == nil {
			return streamer, format, nil
		}
	case ".wav":
		if streamer, format, err := wav.Decode(reader); err == nil {
			return streamer, format, nil
		}
	case ".flac":
		if streamer, format, err := flac.Decode(reader); err == nil {
			return streamer, format, nil
		}
	case ".ogg":
		if streamer, format, err := vorbis.Decode(reader); err == nil {
			return streamer, format, nil
		}
	}

	// 如果扩展名匹配失败或未知格式，尝试所有解码器
	decoders := []struct {
		name   string
		decode func(io.ReadCloser) (beep.StreamSeekCloser, beep.Format, error)
	}{
		{"mp3", mp3.Decode},
		{"wav", func(rc io.ReadCloser) (beep.StreamSeekCloser, beep.Format, error) { return wav.Decode(rc) }},
		{"flac", func(rc io.ReadCloser) (beep.StreamSeekCloser, beep.Format, error) { return flac.Decode(rc) }},
		{"ogg", vorbis.Decode},
	}

	// 尝试每个解码器
	for _, decoder := range decoders {
		// 如果reader支持Seek，重置到开头
		if seeker, ok := reader.(io.ReadSeeker); ok {
			seeker.Seek(0, io.SeekStart)
		}

		if streamer, format, err := decoder.decode(reader); err == nil {
			return streamer, format, nil
		}
	}

	return nil, beep.Format{}, fmt.Errorf("unsupported audio format: %s (tried all decoders)", ext)
}

// getBufferSize 根据音频格式获取缓冲区大小（优化版）
func (p *BeepPlayer) getBufferSize(url string) int {
	ext := strings.ToLower(filepath.Ext(url))
	baseSize := 512

	// 根据格式调整基础大小
	switch ext {
	case ".flac":
		baseSize = 4096 // FLAC无损格式需要更大缓冲区
	case ".wav":
		baseSize = 2048 // WAV无压缩格式
	case ".ogg":
		baseSize = 1024 // OGG压缩格式
	case ".mp3":
		baseSize = 512 // MP3压缩格式，缓冲区较小
	default:
		baseSize = 1024 // 未知格式使用中等大小
	}

	// 根据配置调整缓冲区大小
	if p.config != nil {
		if bufferMultiplier, ok := p.config["buffer_multiplier"].(float64); ok {
			baseSize = int(float64(baseSize) * bufferMultiplier)
		}
		if minBuffer, ok := p.config["min_buffer_size"].(int); ok && baseSize < minBuffer {
			baseSize = minBuffer
		}
		if maxBuffer, ok := p.config["max_buffer_size"].(int); ok && baseSize > maxBuffer {
			baseSize = maxBuffer
		}
	}

	return baseSize
}

// waitForNBytes 等待缓存文件有足够的字节数
func (p *BeepPlayer) waitForNBytes(n int, checkInterval time.Duration, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		if p.downloadCtx != nil && p.downloadCtx.Err() != nil {
			return p.downloadCtx.Err()
		}

		// 检查文件大小
		if stat, err := os.Stat(p.cacheFile); err == nil {
			if stat.Size() >= int64(n) {
				return nil
			}
		}

		time.Sleep(checkInterval)
	}

	return fmt.Errorf("timeout waiting for %d bytes in cache", n)
}
