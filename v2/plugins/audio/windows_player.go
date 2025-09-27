//go:build windows
// +build windows

package audio

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// WindowsPlayer Windows播放器后端实现
type WindowsPlayer struct {
	*BasePlayer
	mediaPlayer    *ole.IDispatch
	currentURL     string
	mutex          sync.RWMutex
	config         map[string]interface{}
	initialized    bool
	positionTicker *time.Ticker
	stopTicker     chan bool

	// COM资源管理
	oleInitialized bool
}

// newWindowsPlayer 创建新的Windows播放器实例（跨平台兼容）
func newWindowsPlayer(config map[string]interface{}) (PlayerBackend, error) {
	return NewWindowsPlayer(config), nil
}

// NewWindowsPlayer 创建新的Windows播放器实例
func NewWindowsPlayer(config map[string]interface{}) *WindowsPlayer {
	formats := []string{"mp3", "wav", "wma", "m4a", "aac", "flac", "ogg", "ape"}
	return &WindowsPlayer{
		BasePlayer: NewBasePlayerWithInfo("Windows Player", "2.0.0", formats),
		config:     config,
		stopTicker: make(chan bool, 1),
	}
}

// Initialize 初始化播放器
func (p *WindowsPlayer) Initialize(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.initialized {
		return nil
	}

	if runtime.GOOS != "windows" {
		return fmt.Errorf("Windows player is only available on Windows")
	}

	// 初始化COM
	if err := ole.CoInitialize(0); err != nil {
		return fmt.Errorf("failed to initialize COM: %w", err)
	}
	p.oleInitialized = true

	// 创建Windows Media Player COM对象
	unknown, err := oleutil.CreateObject("WMPlayer.OCX")
	if err != nil {
		return fmt.Errorf("failed to create Windows Media Player: %w", err)
	}

	// 获取IDispatch接口
	mediaPlayer, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		unknown.Release()
		return fmt.Errorf("failed to get IDispatch interface: %w", err)
	}
	unknown.Release()
	p.mediaPlayer = mediaPlayer

	p.initialized = true
	return nil
}

// Cleanup 清理资源
func (p *WindowsPlayer) Cleanup() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized {
		return nil
	}

	// 停止位置更新器
	if p.positionTicker != nil {
		p.positionTicker.Stop()
		p.positionTicker = nil
	}

	// 发送停止信号
	select {
	case p.stopTicker <- true:
	default:
	}

	// 清理COM资源
	if p.mediaPlayer != nil {
		p.mediaPlayer.Release()
		p.mediaPlayer = nil
	}

	// 清理COM
	if p.oleInitialized {
		ole.CoUninitialize()
		p.oleInitialized = false
	}

	p.initialized = false
	p.currentURL = ""
	return nil
}

// Play 播放音频文件
func (p *WindowsPlayer) Play(url string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized {
		return fmt.Errorf("player not initialized")
	}

	// 如果是同一个文件且正在播放，直接返回
	if p.currentURL == url && p.IsPlaying() {
		return nil
	}

	// 停止当前播放
	if p.currentURL != "" {
		oleutil.CallMethod(p.mediaPlayer, "controls.stop")
	}

	// 设置URL
	if _, err := oleutil.PutProperty(p.mediaPlayer, "URL", url); err != nil {
		return fmt.Errorf("failed to set URL: %w", err)
	}

	// 开始播放
	if _, err := oleutil.CallMethod(p.mediaPlayer, "controls.play"); err != nil {
		return fmt.Errorf("failed to play: %w", err)
	}

	p.currentURL = url
	p.BasePlayer.setPlaying(true)

	// 启动位置更新器
	p.startPositionUpdater()

	return nil
}

// Pause 暂停播放
func (p *WindowsPlayer) Pause() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized || p.currentURL == "" {
		return fmt.Errorf("no audio is playing")
	}

	if _, err := oleutil.CallMethod(p.mediaPlayer, "controls.pause"); err != nil {
		return fmt.Errorf("failed to pause: %w", err)
	}

	p.BasePlayer.setPlaying(false)
	return nil
}

// Resume 恢复播放
func (p *WindowsPlayer) Resume() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized || p.currentURL == "" {
		return fmt.Errorf("no audio loaded")
	}

	if _, err := oleutil.CallMethod(p.mediaPlayer, "controls.play"); err != nil {
		return fmt.Errorf("failed to resume: %w", err)
	}

	p.BasePlayer.setPlaying(true)
	return nil
}

// Stop 停止播放
func (p *WindowsPlayer) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized {
		return fmt.Errorf("player not initialized")
	}

	if p.currentURL != "" {
		if _, err := oleutil.CallMethod(p.mediaPlayer, "controls.stop"); err != nil {
			return fmt.Errorf("failed to stop: %w", err)
		}
	}

	p.BasePlayer.setPlaying(false)
	p.BasePlayer.setPosition(0)

	// 停止位置更新器
	if p.positionTicker != nil {
		p.positionTicker.Stop()
		p.positionTicker = nil
		select {
		case p.stopTicker <- true:
		default:
		}
	}

	return nil
}

// Seek 跳转到指定位置
func (p *WindowsPlayer) Seek(position time.Duration) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized || p.currentURL == "" {
		return fmt.Errorf("no audio loaded")
	}

	// 转换为秒数
	positionSeconds := position.Seconds()

	// 设置播放位置
	if _, err := oleutil.PutProperty(p.mediaPlayer, "controls.currentPosition", positionSeconds); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	p.BasePlayer.setPosition(position)
	return nil
}

// SetVolume 设置音量 (0-100)
func (p *WindowsPlayer) SetVolume(volume float64) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized {
		return fmt.Errorf("player not initialized")
	}

	if volume < 0 || volume > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}

	// Windows Media Player音量范围是0-100
	if _, err := oleutil.PutProperty(p.mediaPlayer, "settings.volume", volume); err != nil {
		return fmt.Errorf("failed to set volume: %w", err)
	}

	// 更新BasePlayer中的音量值
	volumeFloat := float64(volume) / 100.0
	p.BasePlayer.SetVolume(volumeFloat)
	return nil
}

// IsAvailable 检查播放器是否可用
func (p *WindowsPlayer) IsAvailable() bool {
	return runtime.GOOS == "windows"
}

// getDuration 获取音频时长
func (p *WindowsPlayer) getDuration() {
	if !p.initialized || p.currentURL == "" {
		return
	}

	// 等待媒体加载
	time.Sleep(100 * time.Millisecond)

	// 获取媒体时长（秒）
	durationVariant, err := oleutil.GetProperty(p.mediaPlayer, "currentMedia.duration")
	if err != nil {
		return
	}

	// 转换为float64
	durationSeconds, ok := durationVariant.Value().(float64)
	if !ok {
		return
	}

	// 转换为Go时间格式
	duration := time.Duration(durationSeconds * float64(time.Second))
	p.BasePlayer.setDuration(duration)
}

// startPositionUpdater 启动位置更新器
func (p *WindowsPlayer) startPositionUpdater() {
	// 停止之前的更新器
	if p.positionTicker != nil {
		p.positionTicker.Stop()
		select {
		case p.stopTicker <- true:
		default:
		}
	}

	// 创建新的定时器
	p.positionTicker = time.NewTicker(500 * time.Millisecond)

	// 获取音频时长
	go p.getDuration()

	go func() {
		for {
			select {
			case <-p.positionTicker.C:
				if !p.IsPlaying() {
					continue
				}

				// 获取当前播放位置（秒）
				positionVariant, err := oleutil.GetProperty(p.mediaPlayer, "controls.currentPosition")
				if err != nil {
					continue
				}

				// 转换为float64
				positionSeconds, ok := positionVariant.Value().(float64)
				if !ok {
					continue
				}

				// 转换为Go时间格式
				position := time.Duration(positionSeconds * float64(time.Second))
				p.BasePlayer.setPosition(position)

			case <-p.stopTicker:
				return
			}
		}
	}()
}
