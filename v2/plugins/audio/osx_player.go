//go:build darwin
// +build darwin

package audio

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/internal/macdriver/avcore"
	"github.com/go-musicfox/go-musicfox/v2/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/v2/internal/macdriver/core"
)

// URLMusic 音乐URL结构体
type URLMusic struct {
	URL string
}

// State 播放状态枚举
type State int

const (
	Stopped State = iota
	Playing
	Paused
)

// OSXPlayer macOS播放器后端实现
type OSXPlayer struct {
	*BasePlayer
	mutex sync.RWMutex

	// AVPlayer相关字段
	player  *avcore.AVPlayer
	handler *playerHandler

	// 播放状态
	currentURL  string
	curMusic    URLMusic
	timer       *Timer
	volume      int
	state       State
	initialized bool

	// 通道
	timeChan  chan time.Duration
	stateChan chan State
	musicChan chan URLMusic
	close     chan struct{}

	// 配置
	config map[string]interface{}
}

// newOSXPlayer 创建新的OSX播放器实例（跨平台兼容）
func newOSXPlayer(config map[string]interface{}) (PlayerBackend, error) {
	return NewOSXPlayer(config), nil
}

// NewOSXPlayer 创建新的OSX播放器实例
func NewOSXPlayer(config map[string]interface{}) *OSXPlayer {
	formats := []string{"mp3", "wav", "m4a", "aac", "flac", "ogg"}
	return &OSXPlayer{
		BasePlayer: NewBasePlayerWithInfo("OSX Player", "2.0.0", formats),
		config:     config,
		state:      Stopped,
		timeChan:   make(chan time.Duration),
		stateChan:  make(chan State),
		musicChan:  make(chan URLMusic),
		close:      make(chan struct{}),
		volume:     100, // 默认音量100%
	}
}

// Initialize 初始化播放器
func (p *OSXPlayer) Initialize(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.initialized {
		return nil
	}

	if runtime.GOOS != "darwin" {
		return fmt.Errorf("OSX player is only available on macOS")
	}

	// 创建playerHandler
	p.handler = newPlayerHandler(p)
	if p.handler == nil {
		return fmt.Errorf("failed to create player handler")
	}

	// 初始化AVPlayer
	core.Autorelease(func() {
		player := avcore.AVPlayer_alloc().Init()
		p.player = &player
		p.player.SetActionAtItemEnd(0) // AVPlayerActionAtItemEndNone
	})

	// 注册通知监听器
	// 简化实现，暂时跳过通知监听器注册
	// cocoa.NSNotificationCenter_defaultCenter().
	//	AddObserverSelectorNameObject(p.handler.ID, sel_handleFinish, core.String("AVPlayerItemDidPlayToEndTimeNotification"), core.NSObject{})
	// cocoa.NSNotificationCenter_defaultCenter().
	//	AddObserverSelectorNameObject(p.handler.ID, sel_handleFailed, core.String("AVPlayerItemFailedToPlayToEndTimeNotification"), core.NSObject{})

	// 启动监听goroutine
	go p.listen()

	p.initialized = true
	return nil
}

// Cleanup 清理资源
func (p *OSXPlayer) Cleanup() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized {
		return nil
	}

	// 停止定时器
	if p.timer != nil {
		p.timer.Stop()
	}

	// 关闭通道
	if p.close != nil {
		close(p.close)
		p.close = nil
	}

	// 清理AVPlayer资源
	if p.handler != nil {
		p.handler.release()
	}
	if p.player != nil {
		p.player.Release()
	}

	p.initialized = false
	p.currentURL = ""
	return nil
}

// Play 播放音频文件
func (p *OSXPlayer) Play(url string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized {
		return fmt.Errorf("player not initialized")
	}

	// 如果是同一个文件且正在播放，直接返回
	if p.currentURL == url && p.IsPlaying() {
		return nil
	}

	// 通过通道发送播放请求
	music := URLMusic{URL: url}
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case p.musicChan <- music:
		p.currentURL = url
		return nil
	case <-timer.C:
		return fmt.Errorf("play request timeout")
	}
}

// Pause 暂停播放
func (p *OSXPlayer) Pause() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized || p.state != Playing {
		return fmt.Errorf("no audio is playing")
	}

	p.player.Pause()
	if p.timer != nil {
		p.timer.Pause()
	}
	p.setState(Paused)
	return nil
}

// Resume 恢复播放
func (p *OSXPlayer) Resume() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized || p.state == Playing {
		return fmt.Errorf("player is already playing or not initialized")
	}

	if p.timer != nil {
		go p.timer.Start()
	}
	p.player.Play()
	p.setState(Playing)
	return nil
}

// Stop 停止播放
func (p *OSXPlayer) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized || p.state == Stopped {
		return nil
	}

	p.player.Pause()
	if p.timer != nil {
		p.timer.Pause()
	}
	p.setState(Stopped)
	return nil
}

// Seek 跳转到指定位置
func (p *OSXPlayer) Seek(position time.Duration) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized || p.currentURL == "" {
		return fmt.Errorf("no audio loaded")
	}

	// 获取时间刻度
	scale := p.player.CurrentItem().Duration().Timescale
	if scale == 0 {
		return fmt.Errorf("unable to get timescale")
	}

	// 转换时间格式并跳转
	p.player.SeekToTime(avcore.CMTime{
		Value:     int64(float64(scale) * position.Seconds()),
		Timescale: scale,
		Flags:     1,
	})

	// 更新定时器位置
	if p.timer != nil {
		p.timer.SetPassed(position)
	}

	return nil
}

// SetVolume 设置音量 (0.0-1.0)
func (p *OSXPlayer) SetVolume(volume float64) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.initialized {
		return fmt.Errorf("player not initialized")
	}

	if volume < 0.0 || volume > 1.0 {
		return fmt.Errorf("volume must be between 0.0 and 1.0")
	}

	// 转换为0-100范围
	p.volume = int(volume * 100)

	// 设置AVPlayer音量
	if p.player != nil {
		p.player.SetVolume(float32(volume))
	}

	// 更新BasePlayer音量
	p.BasePlayer.SetVolume(volume)
	return nil
}

// IsAvailable 检查播放器是否可用
func (p *OSXPlayer) IsAvailable() bool {
	return runtime.GOOS == "darwin"
}

// getDuration 获取音频时长
func (p *OSXPlayer) getDuration() {
	if !p.initialized || p.currentURL == "" {
		return
	}

	// 获取媒体时长
	core.Autorelease(func() {
		duration := p.player.CurrentItem().Duration()
		if duration.Timescale > 0 {
			durationTime := time.Duration(duration.Value/int64(duration.Timescale)) * time.Second
			p.BasePlayer.setDuration(durationTime)
		}
	})
}

// startPositionUpdater 启动位置更新器
func (p *OSXPlayer) startPositionUpdater() {
	// 获取音频时长
	go p.getDuration()

	// 位置更新逻辑已集成在定时器的OnTick回调中
	// 当创建定时器时会自动处理位置更新
}

// listen 开始监听音乐切换
func (p *OSXPlayer) listen() {
	for {
		select {
		case <-p.close:
			return
		case p.curMusic = <-p.musicChan:
			core.Autorelease(func() {
				// 暂停当前播放
				p.player.Pause()
				if p.timer != nil {
					p.timer.SetPassed(0)
					p.timer.Stop()
				}

				// 创建新的播放项目
				item := avcore.AVPlayerItem_playerItemWithURL(core.NSURL_URLWithString(core.String(p.curMusic.URL)))
				p.player.ReplaceCurrentItemWithPlayerItem(item)

				// 重新注册通知监听器
				cocoa.NSNotificationCenter_defaultCenter().
					AddObserverSelectorNameObject(p.handler.ID, sel_handleFinish, core.String("AVPlayerItemDidPlayToEndTimeNotification"), p.player.CurrentItem().NSObject)
				cocoa.NSNotificationCenter_defaultCenter().
					AddObserverSelectorNameObject(p.handler.ID, sel_handleFailed, core.String("AVPlayerItemFailedToPlayToEndTimeNotification"), p.player.CurrentItem().NSObject)

				// 创建定时器
				p.timer = NewTimer(TimerOptions{
					Duration:       8760 * time.Hour,
					TickerInternal: 500 * time.Millisecond,
					OnRun:          func(started bool) {},
					OnPause:        func() {},
					OnDone:         func(stopped bool) {},
					OnTick: func() {
						var curTime time.Duration
						core.Autorelease(func() {
							t := p.player.CurrentTime()
							if t.Timescale <= 0 {
								return
							}
							curTime = time.Duration(t.Value/int64(t.Timescale)) * time.Second
						})
						// 更新位置，添加800ms延迟补偿
						p.BasePlayer.setPosition(curTime + time.Millisecond*800)
						select {
						case p.timeChan <- curTime + time.Millisecond*800:
						default:
						}
					},
				})

				// 开始播放
				p.Resume()
			})
		}
	}
}

// setState 设置播放状态
func (p *OSXPlayer) setState(state State) {
	p.state = state
	p.BasePlayer.setPlaying(state == Playing)
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
	}
}

// CurMusic 获取当前音乐
func (p *OSXPlayer) CurMusic() URLMusic {
	return p.curMusic
}

// PassedTime 获取已播放时间
func (p *OSXPlayer) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	var curTime time.Duration
	core.Autorelease(func() {
		t := p.player.CurrentTime()
		if t.Timescale <= 0 {
			return
		}
		curTime = time.Duration(float64(t.Value*1000.0)/float64(t.Timescale)) * time.Millisecond
	})
	return curTime
}

// PlayedTime 获取实际播放时间
func (p *OSXPlayer) PlayedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.ActualRuntime()
}

// TimeChan 获取时间通道
func (p *OSXPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

// State 获取播放状态
func (p *OSXPlayer) State() State {
	return p.state
}

// StateChan 获取状态通道
func (p *OSXPlayer) StateChan() <-chan State {
	return p.stateChan
}

// Toggle 切换播放状态
func (p *OSXPlayer) Toggle() {
	switch p.State() {
	case Paused, Stopped:
		p.Resume()
	case Playing:
		p.Pause()
	default:
		p.Resume()
	}
}

// UpVolume 增加音量
func (p *OSXPlayer) UpVolume() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.volume+5 >= 100 {
		p.volume = 100
	} else {
		p.volume += 5
	}
	if p.player != nil {
		p.player.SetVolume(float32(p.volume) / 100.0)
	}
	p.BasePlayer.SetVolume(float64(p.volume) / 100.0)
}

// DownVolume 降低音量
func (p *OSXPlayer) DownVolume() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.volume-5 <= 0 {
		p.volume = 0
	} else {
		p.volume -= 5
	}
	if p.player != nil {
		p.player.SetVolume(float32(p.volume) / 100.0)
	}
	p.BasePlayer.SetVolume(float64(p.volume) / 100.0)
}

// Volume 获取音量
func (p *OSXPlayer) Volume() int {
	return p.volume
}
