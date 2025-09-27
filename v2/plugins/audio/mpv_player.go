package audio

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MPVPlayer MPV播放器后端实现
type MPVPlayer struct {
	*BasePlayer
	cmd         *exec.Cmd
	cancel      context.CancelFunc
	currentURL  string
	mutex       sync.RWMutex
	config      map[string]interface{}
	initialized bool
}

// NewMPVPlayer 创建MPV播放器
func NewMPVPlayer(config map[string]interface{}) *MPVPlayer {
	formats := []string{"mp3", "wav", "flac", "ogg", "m4a", "aac", "wma", "ape"}
	return &MPVPlayer{
		BasePlayer: NewBasePlayerWithInfo("MPV Player", "1.0.0", formats),
		config:     config,
	}
}

// Initialize 初始化播放器
func (p *MPVPlayer) Initialize(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.initialized {
		return nil
	}

	// 检查MPV是否可用
	if !p.IsAvailable() {
		return fmt.Errorf("mpv is not available on this system")
	}

	p.initialized = true
	return nil
}

// Cleanup 清理资源
func (p *MPVPlayer) Cleanup() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}

	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Kill()
		p.cmd.Wait()
		p.cmd = nil
	}

	p.setPlaying(false)
	p.currentURL = ""

	return nil
}

// Play 播放音频
func (p *MPVPlayer) Play(url string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 停止当前播放
	if p.cmd != nil {
		if p.cancel != nil {
			p.cancel()
		}
		if p.cmd.Process != nil {
			p.cmd.Process.Kill()
			p.cmd.Wait()
		}
	}

	// 创建新的上下文
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	// 构建MPV命令
	args := []string{
		"--no-video",     // 只播放音频
		"--no-terminal",  // 不显示终端输出
		"--really-quiet", // 静默模式
		"--no-config",    // 不加载配置文件
		"--volume=" + fmt.Sprintf("%.0f", p.volume*100), // 设置音量
		url,
	}

	// 添加自定义配置
	if p.config != nil {
		if extraArgs, ok := p.config["extra_args"].([]string); ok {
			args = append(args[:len(args)-1], append(extraArgs, url)...)
		}
	}

	// 启动MPV进程
	p.cmd = exec.CommandContext(ctx, "mpv", args...)
	err := p.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	p.currentURL = url
	p.setPlaying(true)
	p.setPosition(0)

	// 启动监控协程
	go p.monitorPlayback()

	// 获取音频时长
	go p.getDuration(url)

	return nil
}

// Pause 暂停播放
func (p *MPVPlayer) Pause() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("no active playback")
	}

	// MPV没有直接的暂停命令，这里使用SIGSTOP信号
	// 注意：这在Windows上可能不工作
	err := p.cmd.Process.Signal(nil) // 发送SIGSTOP信号暂停进程
	if err != nil {
		return fmt.Errorf("failed to pause: %w", err)
	}

	p.setPlaying(false)
	return nil
}

// Resume 恢复播放
func (p *MPVPlayer) Resume() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("no active playback")
	}

	// 发送SIGCONT信号恢复进程
	err := p.cmd.Process.Signal(nil) // 发送SIGCONT信号恢复进程
	if err != nil {
		return fmt.Errorf("failed to resume: %w", err)
	}

	p.setPlaying(true)
	return nil
}

// Stop 停止播放
func (p *MPVPlayer) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}

	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Kill()
		p.cmd.Wait()
		p.cmd = nil
	}

	p.setPlaying(false)
	p.setPosition(0)
	p.currentURL = ""

	return nil
}

// Seek 跳转到指定位置
func (p *MPVPlayer) Seek(position time.Duration) error {
	// MPV的命令行版本不支持运行时seek
	// 这里返回不支持的错误
	return fmt.Errorf("seek not supported in command-line mpv mode")
}

// SetVolume 设置音量
func (p *MPVPlayer) SetVolume(volume float64) error {
	if err := p.BasePlayer.SetVolume(volume); err != nil {
		return err
	}

	// MPV命令行版本不支持运行时音量调节
	// 新的播放会使用新的音量设置
	return nil
}

// IsAvailable 检查播放器是否可用
func (p *MPVPlayer) IsAvailable() bool {
	// 检查系统中是否安装了MPV
	_, err := exec.LookPath("mpv")
	return err == nil
}

// monitorPlayback 监控播放状态
func (p *MPVPlayer) monitorPlayback() {
	if p.cmd == nil {
		return
	}

	// 等待进程结束
	err := p.cmd.Wait()

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.setPlaying(false)
	p.setPosition(0)
	p.cmd = nil

	if err != nil && !strings.Contains(err.Error(), "killed") {
		// 进程异常退出
		fmt.Printf("MPV process exited with error: %v\n", err)
	}
}

// getDuration 获取音频时长
func (p *MPVPlayer) getDuration(url string) {
	// 使用ffprobe获取音频时长
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries", "format=duration", "-of", "csv=p=0", url)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return
	}

	duration := time.Duration(durationFloat * float64(time.Second))
	p.setDuration(duration)
}

// updatePosition 更新播放位置（简化版本）
func (p *MPVPlayer) updatePosition() {
	// MPV命令行版本难以获取实时播放位置
	// 这里使用简单的时间估算
	startTime := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		p.mutex.RLock()
		if !p.playing {
			p.mutex.RUnlock()
			break
		}
		p.mutex.RUnlock()

		elapsed := time.Since(startTime)
		p.setPosition(elapsed)
	}
}
